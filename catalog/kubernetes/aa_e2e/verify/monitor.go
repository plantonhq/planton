package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
	"github.com/plantonhq/planton/pkg/manifestgraph"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// MonitorVerifier proves a KubernetesServiceMonitor or KubernetesPodMonitor
// end to end, on either engine, to the point a customer could rely on its
// scrape:
//
//   - The live object's spec is EXACTLY the manifest's projection -- the same
//     rendering both engines apply, computed here from the scenario the
//     harness deployed (its references already resolved from the fixtures'
//     outputs). A key an engine spelled
//     wrong, an IntOrString left a string, a list-valued map left wrapped, or
//     an explicitly empty selector dropped all fail this check. The declared
//     labels and annotations are on the object, under Planton's identity
//     labels.
//   - The behavioral-scrape scenario (recognized by name) proves the
//     prerequisite stack's Prometheus LOADED the monitor, DISCOVERED its
//     targets and SCRAPED them: a target of the monitor's job is up, its
//     scrape URL carries the scenario's params, and the series it yields
//     carry the label the scenario's relabeling stamps.
//
// On destroy, the object must be gone.
type MonitorVerifier struct {
	// Resource is the custom resource's kubectl name
	// (servicemonitors.monitoring.coreos.com).
	Resource string
	// JobPrefix is the first segment of the scrape job the operator names
	// after the monitor ("serviceMonitor", "podMonitor").
	JobPrefix string
	Namespace string
	Name      string
	// ManifestPath is the scenario manifest as the harness deployed it (a
	// resolved copy when it carries references), read for the expected
	// object.
	ManifestPath string
	Scrape       bool
}

func (v *MonitorVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] %s %q in namespace %q\n", v.Resource, v.Name, v.Namespace)
	if err := KubectlResourceExists(ctx, kubeconfig, v.Resource, v.Name, v.Namespace); err != nil {
		return err
	}
	if err := v.proveProjection(ctx, kubeconfig); err != nil {
		return err
	}
	if v.Scrape {
		return v.proveScrape(ctx, kubeconfig)
	}
	return nil
}

func (v *MonitorVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, v.Resource, v.Name, v.Namespace)
}

// proveProjection renders the scenario through the projection both engines
// apply and compares the live object with it.
func (v *MonitorVerifier) proveProjection(ctx context.Context, kubeconfig string) error {
	want, err := expectedProjection(v.ManifestPath)
	if err != nil {
		return err
	}

	liveSpec, err := kubectlGetJSONPath(ctx, kubeconfig, v.Resource, v.Name, v.Namespace, "{.spec}")
	if err != nil {
		return errors.Wrap(err, "reading the live spec")
	}
	withUpstreamDefaults(want.Spec)
	if !jsonEqual(want.Spec, liveSpec) {
		wantJSON, _ := json.Marshal(want.Spec)
		return errors.Errorf("PROJECTION: the live spec differs from the manifest's projection\nlive:      %s\nprojected: %s", liveSpec, wantJSON)
	}

	liveLabels, err := liveStringMap(ctx, kubeconfig, v.Resource, v.Name, v.Namespace, "{.metadata.labels}")
	if err != nil {
		return err
	}
	for k, val := range want.Labels {
		// The id, organization and environment labels carry values the
		// harness assigns at deploy time; the rest are fixed by the manifest.
		if k == manifestprojection.LabelResourceID || k == manifestprojection.LabelOrganization || k == manifestprojection.LabelEnvironment {
			continue
		}
		if liveLabels[k] != val {
			return errors.Errorf("PROJECTION: label %s = %q on the live object, want %q", k, liveLabels[k], val)
		}
	}
	if len(want.Annotations) > 0 {
		liveAnnotations, err := liveStringMap(ctx, kubeconfig, v.Resource, v.Name, v.Namespace, "{.metadata.annotations}")
		if err != nil {
			return err
		}
		for k, val := range want.Annotations {
			if liveAnnotations[k] != val {
				return errors.Errorf("PROJECTION: annotation %s = %q on the live object, want %q", k, liveAnnotations[k], val)
			}
		}
	}
	fmt.Printf("  [verify] PROJECTION: the live object is exactly the manifest's projection, labels and annotations included\n")
	return nil
}

// expectedProjection loads the scenario and renders the object both engines
// apply. The harness hands the verifier the manifest it deployed, with every
// reference already resolved from the fixtures' outputs; a reference still
// present (a scenario run outside the harness) is resolved to the name of the
// fixture it targets, which every monitor scenario's fixtures (a
// KubernetesSecret, a KubernetesConfigMap, a KubernetesNamespace) also name
// their objects after.
func expectedProjection(manifestPath string) (*manifestprojection.Object, error) {
	msg, err := manifest.LoadManifest(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "loading %s", manifestPath)
	}
	for _, use := range manifestgraph.CollectRefUses(msg) {
		use.Replace(&foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: use.Ref.GetName()},
		})
	}
	m, ok := msg.(manifestprojection.Manifest)
	if !ok {
		return nil, errors.Errorf("%s is not a catalog manifest", manifestPath)
	}
	obj, err := manifestprojection.Render(m)
	if err != nil {
		return nil, errors.Wrapf(err, "rendering %s", manifestPath)
	}
	return obj, nil
}

// proveScrape asks the prerequisite stack's Prometheus for the monitor's
// targets and series.
func (v *MonitorVerifier) proveScrape(ctx context.Context, kubeconfig string) error {
	want, err := expectedProjection(v.ManifestPath)
	if err != nil {
		return err
	}
	marker, params := scrapeExpectations(want.Spec)
	if marker == "" {
		return errors.New("SCRAPE: the scenario's first endpoint stamps no e2e_monitor label, so there is nothing to look for")
	}

	stack, err := prometheusStackPrerequisite()
	if err != nil {
		return err
	}
	base, stop, err := portForwardPrometheus(ctx, kubeconfig, stack, "19193")
	if err != nil {
		return err
	}
	defer stop()

	jobPrefix := fmt.Sprintf("%s/%s/%s/", v.JobPrefix, v.Namespace, v.Name)
	if err := awaitHTTPCondition(ctx, base+"/api/v1/targets?state=active", 6*time.Minute, func(body string) error {
		targets := monitorTargets(body, jobPrefix)
		if len(targets) == 0 {
			return errors.Errorf("no active target from %s* yet", jobPrefix)
		}
		for _, t := range targets {
			if t.Health != "up" {
				continue
			}
			for key, values := range params {
				for _, value := range values {
					if !strings.Contains(t.ScrapeURL, url.QueryEscape(key)+"="+url.QueryEscape(value)) {
						return errors.Errorf("the target %s is up but its scrape URL lacks the param %s=%s", t.ScrapeURL, key, value)
					}
				}
			}
			return nil
		}
		return errors.Errorf("the monitor's targets are not up yet: %+v", targets)
	}); err != nil {
		return errors.Wrap(err, "SCRAPE: the monitor's target never came up")
	}
	fmt.Printf("  [verify] SCRAPE: a target of %s* is up and its scrape URL carries the declared params\n", jobPrefix)

	query := fmt.Sprintf(`up{e2e_monitor=%q} == 1`, marker)
	if err := awaitHTTPCondition(ctx, base+"/api/v1/query?query="+url.QueryEscape(query), 3*time.Minute, func(body string) error {
		if !queryHasSamples(body) {
			return errors.New("no scraped series carries the relabeled marker yet")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "SCRAPE: the relabeling never reached the scraped series")
	}
	fmt.Printf("  [verify] SCRAPE: the scraped series carry e2e_monitor=%q from the relabeling\n", marker)
	return nil
}

// withUpstreamDefaults applies to a projected spec the defaults the pinned
// ServiceMonitor and PodMonitor CRDs declare, which the API server writes into
// the stored object: a relabeling step's `action` is "replace" when unset. (The
// CRDs' only other default, an empty selector `name`, never applies: the spec
// requires a name.) Comparing against the defaulted projection keeps the
// projection proof exact instead of tolerating any extra key.
func withUpstreamDefaults(spec map[string]interface{}) {
	for _, list := range []string{"endpoints", "podMetricsEndpoints"} {
		endpoints, _ := spec[list].([]interface{})
		for _, e := range endpoints {
			endpoint, _ := e.(map[string]interface{})
			for _, key := range []string{"relabelings", "metricRelabelings"} {
				steps, _ := endpoint[key].([]interface{})
				for _, st := range steps {
					if step, ok := st.(map[string]interface{}); ok {
						if _, set := step["action"]; !set {
							step["action"] = "replace"
						}
					}
				}
			}
		}
	}
}

// scrapeExpectations reads, from the projected spec's first endpoint, the
// value its relabeling stamps on e2e_monitor and the params its scrape URL
// must carry.
func scrapeExpectations(spec map[string]interface{}) (string, map[string][]string) {
	var endpoints []interface{}
	for _, key := range []string{"endpoints", "podMetricsEndpoints"} {
		if list, ok := spec[key].([]interface{}); ok {
			endpoints = list
		}
	}
	if len(endpoints) == 0 {
		return "", nil
	}
	endpoint, _ := endpoints[0].(map[string]interface{})
	marker := ""
	relabelings, _ := endpoint["relabelings"].([]interface{})
	for _, r := range relabelings {
		step, _ := r.(map[string]interface{})
		if step["targetLabel"] == "e2e_monitor" {
			marker, _ = step["replacement"].(string)
		}
	}
	params := map[string][]string{}
	raw, _ := endpoint["params"].(map[string]interface{})
	for key, values := range raw {
		list, _ := values.([]interface{})
		for _, value := range list {
			params[key] = append(params[key], fmt.Sprint(value))
		}
	}
	return marker, params
}

type monitorTarget struct {
	ScrapePool string `json:"scrapePool"`
	ScrapeURL  string `json:"scrapeUrl"`
	Health     string `json:"health"`
	LastError  string `json:"lastError"`
}

func monitorTargets(body, jobPrefix string) []monitorTarget {
	var payload struct {
		Data struct {
			ActiveTargets []monitorTarget `json:"activeTargets"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil
	}
	var out []monitorTarget
	for _, t := range payload.Data.ActiveTargets {
		if strings.HasPrefix(t.ScrapePool, jobPrefix) {
			out = append(out, t)
		}
	}
	return out
}
