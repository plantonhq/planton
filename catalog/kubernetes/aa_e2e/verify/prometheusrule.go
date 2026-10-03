package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
)

// PrometheusRuleVerifier proves a KubernetesPrometheusRule end to end, on
// either engine, to the point a customer could rely on its rules:
//
//   - The object exists with EXACTLY the declared spec under the upstream keys
//     and the declared labels and annotations, with Planton's identity labels
//     on top. Both engines build the object from one projection, so this
//     check is the same for both and fails on any key an engine spelled wrong.
//   - The behavioral-evaluation scenario (recognized by name) proves the
//     prerequisite stack's Prometheus LOADED and EVALUATED it: the always-true
//     alert is firing with the group's and the rule's labels, and the
//     recording rule's series answers a query.
//   - The behavioral-fence scenario proves the object's own labels decide who
//     loads it: a fenced Prometheus (release_managed_only) loads the labelled
//     object and never its unlabelled twin, while the prerequisite stack
//     (all_monitors) loads both.
//
// On destroy, the object must be gone.
type PrometheusRuleVerifier struct {
	Namespace string
	Name      string
	// ManifestPath is the scenario manifest, read for the declared spec,
	// labels and annotations, and to find the stacks the proofs query.
	ManifestPath string
	Evaluation   bool
	Fence        bool
}

const prometheusRuleResource = "prometheusrules.monitoring.coreos.com"

func (v *PrometheusRuleVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] PrometheusRule %q in namespace %q\n", v.Name, v.Namespace)
	if err := KubectlResourceExists(ctx, kubeconfig, prometheusRuleResource, v.Name, v.Namespace); err != nil {
		return err
	}
	if err := v.proveDeclaredObject(ctx, kubeconfig); err != nil {
		return err
	}
	if v.Evaluation {
		if err := v.proveEvaluation(ctx, kubeconfig); err != nil {
			return err
		}
	}
	if v.Fence {
		return v.proveFence(ctx, kubeconfig)
	}
	return nil
}

func (v *PrometheusRuleVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, prometheusRuleResource, v.Name, v.Namespace)
}

// proveDeclaredObject compares the live object with the scenario manifest:
// its spec minus the envelope must equal the live spec exactly, and the
// declared labels, annotations and identity labels must be on the object.
func (v *PrometheusRuleVerifier) proveDeclaredObject(ctx context.Context, kubeconfig string) error {
	declared := manifestSpecMap(v.ManifestPath)
	wantSpec := map[string]interface{}{}
	for k, val := range declared {
		if k == "namespace" || k == "labels" || k == "annotations" {
			continue
		}
		wantSpec[k] = val
	}

	liveSpec, err := kubectlGetJSONPath(ctx, kubeconfig, prometheusRuleResource, v.Name, v.Namespace, "{.spec}")
	if err != nil {
		return errors.Wrap(err, "reading the live spec")
	}
	if !jsonEqual(wantSpec, liveSpec) {
		want, _ := json.Marshal(wantSpec)
		return errors.Errorf("PROJECTION: the live spec differs from the declared one\nlive:     %s\ndeclared: %s", liveSpec, want)
	}

	liveLabels, err := liveStringMap(ctx, kubeconfig, prometheusRuleResource, v.Name, v.Namespace, "{.metadata.labels}")
	if err != nil {
		return err
	}
	wantLabels := map[string]string{
		manifestprojection.LabelResource:     "true",
		manifestprojection.LabelResourceName: v.Name,
		manifestprojection.LabelResourceKind: "KubernetesPrometheusRule",
	}
	for k, val := range stringMapOf(declared["labels"]) {
		if _, identity := wantLabels[k]; !identity {
			wantLabels[k] = val
		}
	}
	for k, val := range wantLabels {
		if liveLabels[k] != val {
			return errors.Errorf("PROJECTION: label %s = %q on the live object, want %q", k, liveLabels[k], val)
		}
	}

	if wantAnnotations := stringMapOf(declared["annotations"]); len(wantAnnotations) > 0 {
		liveAnnotations, err := liveStringMap(ctx, kubeconfig, prometheusRuleResource, v.Name, v.Namespace, "{.metadata.annotations}")
		if err != nil {
			return err
		}
		for k, val := range wantAnnotations {
			if liveAnnotations[k] != val {
				return errors.Errorf("PROJECTION: annotation %s = %q on the live object, want %q", k, liveAnnotations[k], val)
			}
		}
	}
	fmt.Printf("  [verify] PROJECTION: the live object carries exactly the declared spec, labels and annotations\n")
	return nil
}

// proveEvaluation queries the prerequisite stack's Prometheus: the
// always-true alert firing with its group and rule labels, and the recording
// rule's series answering.
func (v *PrometheusRuleVerifier) proveEvaluation(ctx context.Context, kubeconfig string) error {
	stack, err := prometheusStackPrerequisite()
	if err != nil {
		return err
	}
	base, stop, err := portForwardPrometheus(ctx, kubeconfig, stack, "19190")
	if err != nil {
		return err
	}
	defer stop()

	if err := awaitHTTPCondition(ctx, base+"/api/v1/alerts", 5*time.Minute, func(body string) error {
		for _, a := range prometheusAlerts(body) {
			if a.Labels["alertname"] == "E2EPrometheusRuleAlwaysFiring" && a.State == "firing" {
				if a.Labels["e2e_group"] != "evaluation" || a.Labels["e2e_rule"] != "alerting" {
					return errors.Errorf("the alert fires without its group and rule labels: %v", a.Labels)
				}
				return nil
			}
		}
		return errors.New("the e2e alert is not firing yet")
	}); err != nil {
		return errors.Wrap(err, "EVALUATION: the alerting rule never fired in prometheus")
	}
	fmt.Printf("  [verify] EVALUATION: the alerting rule fires with its group and rule labels\n")

	if err := awaitHTTPCondition(ctx, base+"/api/v1/query?query="+url.QueryEscape(`e2e:prometheus_rule:one{e2e_rule="recording"}`), 3*time.Minute, func(body string) error {
		if !queryHasSamples(body) {
			return errors.New("the recorded series has no samples yet")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "EVALUATION: the recording rule's series never answered")
	}
	fmt.Printf("  [verify] EVALUATION: the recording rule's series answers a query\n")
	return nil
}

// proveFence checks which rule groups each Prometheus loaded: the fenced one
// only the labelled object's, the prerequisite stack both.
func (v *PrometheusRuleVerifier) proveFence(ctx context.Context, kubeconfig string) error {
	fenced, err := ParseManifestInfo(filepath.Join(filepath.Dir(filepath.Dir(v.ManifestPath)), "fixture-fenced-stack.yaml"))
	if err != nil {
		return errors.Wrap(err, "reading the fenced stack fixture")
	}
	fencedBase, stopFenced, err := portForwardPrometheus(ctx, kubeconfig, fenced, "19191")
	if err != nil {
		return err
	}
	defer stopFenced()

	if err := awaitHTTPCondition(ctx, fencedBase+"/api/v1/rules", 5*time.Minute, func(body string) error {
		groups := ruleGroupNames(body)
		if !groups["e2e-fenced-in"] {
			return errors.Errorf("the fenced prometheus has not loaded the labelled rule yet (groups %v)", groups)
		}
		if groups["e2e-fenced-out"] {
			return errors.New("the fenced prometheus loaded the UNLABELLED rule: the fence does not hold")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "FENCE: the fenced prometheus did not load exactly the labelled rule")
	}
	fmt.Printf("  [verify] FENCE: the fenced prometheus loads the labelled rule and not its unlabelled twin\n")

	stack, err := prometheusStackPrerequisite()
	if err != nil {
		return err
	}
	base, stop, err := portForwardPrometheus(ctx, kubeconfig, stack, "19192")
	if err != nil {
		return err
	}
	defer stop()
	if err := awaitHTTPCondition(ctx, base+"/api/v1/rules", 5*time.Minute, func(body string) error {
		groups := ruleGroupNames(body)
		if !groups["e2e-fenced-in"] || !groups["e2e-fenced-out"] {
			return errors.Errorf("the all_monitors prometheus has not loaded both rules yet (groups %v)", groups)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "FENCE: the all_monitors prometheus did not load both rules")
	}
	fmt.Printf("  [verify] FENCE: the all_monitors prometheus loads both rules\n")
	return nil
}

// prometheusStackPrerequisite reads the KubernetesKubePrometheusStack the
// harness installs as a kind's registry prerequisite, from the same install
// manifest the harness uses (its prerequisite.yaml, else its minimal
// scenario). The stack's e2e folder is found beside this verifier in the
// source tree: a scenario whose references the harness resolved reaches the
// verifier as a copy outside the tree, so the scenario's path is no guide.
func prometheusStackPrerequisite() (*ManifestInfo, error) {
	_, self, _, _ := runtime.Caller(0)
	stackE2E := filepath.Join(filepath.Dir(self), "..", "..", "kuberneteskubeprometheusstack", "e2e")
	for _, candidate := range []string{"prerequisite.yaml", filepath.Join("scenarios", "minimal.yaml")} {
		if info, err := ParseManifestInfo(filepath.Join(stackE2E, candidate)); err == nil {
			return info, nil
		}
	}
	return nil, errors.Errorf("no install manifest for the prerequisite stack under %s", stackE2E)
}

// portForwardPrometheus forwards a local port to a stack's Prometheus
// service (`<release>-prometheus`) and returns its base URL and a stop
// function that ends the forward.
func portForwardPrometheus(ctx context.Context, kubeconfig string, stack *ManifestInfo, localPort string) (string, func(), error) {
	pfCtx, cancel := context.WithCancel(ctx)
	pf := exec.CommandContext(pfCtx, "kubectl", "--kubeconfig", kubeconfig, "port-forward", "svc/"+stack.Name+"-prometheus", localPort+":9090", "-n", stack.Namespace)
	var out strings.Builder
	pf.Stdout = &out
	pf.Stderr = &out
	if err := pf.Start(); err != nil {
		cancel()
		return "", nil, errors.Wrapf(err, "starting port-forward to %s-prometheus", stack.Name)
	}
	stop := func() {
		cancel()
		_ = pf.Wait()
	}
	return "http://127.0.0.1:" + localPort, stop, nil
}

type prometheusAlert struct {
	Labels map[string]string `json:"labels"`
	State  string            `json:"state"`
}

func prometheusAlerts(body string) []prometheusAlert {
	var payload struct {
		Data struct {
			Alerts []prometheusAlert `json:"alerts"`
		} `json:"data"`
	}
	_ = json.Unmarshal([]byte(body), &payload)
	return payload.Data.Alerts
}

func queryHasSamples(body string) bool {
	var payload struct {
		Data struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return false
	}
	return len(payload.Data.Result) > 0 && len(payload.Data.Result[0].Value) == 2
}

func ruleGroupNames(body string) map[string]bool {
	var payload struct {
		Data struct {
			Groups []struct {
				Name string `json:"name"`
			} `json:"groups"`
		} `json:"data"`
	}
	names := map[string]bool{}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return names
	}
	for _, g := range payload.Data.Groups {
		names[g.Name] = true
	}
	return names
}

func liveStringMap(ctx context.Context, kubeconfig, resource, name, namespace, jsonPath string) (map[string]string, error) {
	raw, err := kubectlGetJSONPath(ctx, kubeconfig, resource, name, namespace, jsonPath)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s", jsonPath)
	}
	out := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, errors.Wrapf(err, "parsing %s", jsonPath)
	}
	return out, nil
}

func stringMapOf(v interface{}) map[string]string {
	raw, _ := v.(map[string]interface{})
	out := make(map[string]string, len(raw))
	for k, val := range raw {
		out[k] = fmt.Sprint(val)
	}
	return out
}

// jsonEqual compares a decoded YAML value with a JSON document by
// round-tripping both through encoding/json, so numbers and maps compare by
// value.
func jsonEqual(want interface{}, gotJSON string) bool {
	wantBytes, err := json.Marshal(want)
	if err != nil {
		return false
	}
	var a, b interface{}
	if json.Unmarshal(wantBytes, &a) != nil || json.Unmarshal([]byte(gotJSON), &b) != nil {
		return false
	}
	return reflect.DeepEqual(a, b)
}
