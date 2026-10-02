package verify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// KubePrometheusStackVerifier checks a kube-prometheus-stack installation
// to the point a customer could rely on it as their cluster's monitoring:
// the operator Deployment available (unless the stack is a receiver whose
// helm_values turn its operator off), the operator-reconciled Prometheus
// StatefulSet at its declared replica count, the Alertmanager StatefulSet
// and bundled-Grafana Deployment when enabled — and a LIVE metric-flow
// proof through Prometheus' own API (a scrape target up and a PromQL
// query answering with data; a metrics stack that cannot answer a query
// is not a metrics stack).
//
// The behavioral-notifications scenario (recognized by name) proves typed
// alert delivery reaches a sink standing in for Discord and a heartbeat
// monitor (kubeprometheusstack_notifications.go).
//
// The behavioral-alerting scenario (recognized by name) additionally
// proves the ALERTING PIPELINE end to end using the stack's own
// dead-man's-switch: the always-firing Watchdog alert must be visible as
// firing in Prometheus' alerts API AND must have propagated into
// Alertmanager's API — rule evaluation and alert delivery proven without
// waiting for anything to actually break.
//
// On destroy, the release's workloads must be GONE while the
// monitoring.coreos.com CRDs REMAIN — the crds-subchart keep posture is
// the designed behavior (ServiceMonitors and rules across the cluster
// survive removal of the stack), so the verifier asserts it rather than
// tolerating it.
type KubePrometheusStackVerifier struct {
	Namespace string
	Name      string
	// PrometheusReplicas is the declared replica count (the operator
	// names the StatefulSet `prometheus-<name>-prometheus`).
	PrometheusReplicas int
	// AlertmanagerEnabled / GrafanaEnabled mirror the manifest's typed
	// toggles (proto optional-bools defaulting true).
	AlertmanagerEnabled bool
	// AlertmanagerReplicas is the declared replica count (the operator
	// names the StatefulSet `alertmanager-<name>-alertmanager`).
	AlertmanagerReplicas int
	GrafanaEnabled       bool
	// Alerting switches on the Watchdog pipeline proof.
	Alerting bool
	// Notifications switches on the alert-delivery proof
	// (kubeprometheusstack_notifications.go).
	Notifications bool
	// OperatorEnabled is false for a receiver beside another stack: its
	// helm_values turn the operator off and the other stack's operator
	// reconciles its Prometheus, so there is no operator Deployment to wait
	// for.
	OperatorEnabled bool
	// KubeStateMetricsEnabled is false when the exporters are off; the
	// metric-flow proof then accepts any healthy target (a receiver scrapes
	// only its own release's self-monitor).
	KubeStateMetricsEnabled bool
}

// stackCrds are the monitoring.coreos.com CRDs whose keep-on-uninstall
// posture the destroy phase asserts.
var stackCrds = []string{
	"prometheuses.monitoring.coreos.com",
	"alertmanagers.monitoring.coreos.com",
	"servicemonitors.monitoring.coreos.com",
	"podmonitors.monitoring.coreos.com",
	"prometheusrules.monitoring.coreos.com",
	"scrapeconfigs.monitoring.coreos.com",
}

func (v *KubePrometheusStackVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] kube-prometheus-stack %q in namespace %q\n", v.Name, v.Namespace)

	// The operator itself: `<name>-operator` per the chart's naming
	// contract off the pinned fullname.
	if v.OperatorEnabled {
		if err := v.waitDeploymentAvailable(ctx, kubeconfig, v.Name+"-operator", 5*time.Minute); err != nil {
			return errors.Wrap(err, "the prometheus-operator deployment never became available")
		}
	}

	// The operator-reconciled Prometheus StatefulSet: the operator names
	// it `prometheus-<crname>` and the chart's CR name is
	// `<fullname>-prometheus`.
	if err := v.waitStatefulSetReady(ctx, kubeconfig,
		"prometheus-"+v.Name+"-prometheus", v.PrometheusReplicas, 10*time.Minute); err != nil {
		return errors.Wrap(err, "the prometheus statefulset never became ready")
	}

	if v.AlertmanagerEnabled {
		if err := v.waitStatefulSetReady(ctx, kubeconfig,
			"alertmanager-"+v.Name+"-alertmanager", v.AlertmanagerReplicas, 5*time.Minute); err != nil {
			return errors.Wrap(err, "the alertmanager statefulset never became ready")
		}
	}
	if v.GrafanaEnabled {
		if err := v.waitDeploymentAvailable(ctx, kubeconfig, v.Name+"-grafana", 5*time.Minute); err != nil {
			return errors.Wrap(err, "the bundled grafana deployment never became available")
		}
	}

	if err := v.proveMetricFlow(ctx, kubeconfig); err != nil {
		return err
	}
	if v.Notifications {
		return v.proveNotifications(ctx, kubeconfig)
	}
	return nil
}

func (v *KubePrometheusStackVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.Name+"-operator", v.Namespace); err != nil {
		return err
	}
	if err := KubectlResourceAbsent(ctx, kubeconfig, "statefulset", "prometheus-"+v.Name+"-prometheus", v.Namespace); err != nil {
		return err
	}
	// The keep posture is DESIGNED behavior, not tolerated residue: the
	// crds-subchart CRDs must still be present after destroy — a missing
	// CRD here means the lifecycle regressed and every ServiceMonitor
	// and rule in the cluster just lost its API.
	for _, crd := range stackCrds {
		if err := KubectlResourceExists(ctx, kubeconfig, "crd", crd, ""); err != nil {
			return errors.Wrapf(err, "CRD %q should SURVIVE uninstall (the crds-subchart keep posture) but is gone", crd)
		}
	}
	fmt.Printf("  [verify] DESTROY: release workloads gone; monitoring.coreos.com CRDs kept (designed posture)\n")
	return nil
}

func (v *KubePrometheusStackVerifier) waitDeploymentAvailable(ctx context.Context, kubeconfig, name string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var last string
	for time.Now().Before(deadline) {
		ready, _ := kubectlGetJSONPath(ctx, kubeconfig, "deployment", name, v.Namespace, "{.status.availableReplicas}")
		last = ready
		if ready != "" && ready != "0" {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("deployment %q never reported available replicas (last %q)", name, last)
}

func (v *KubePrometheusStackVerifier) waitStatefulSetReady(ctx context.Context, kubeconfig, name string, replicas int, budget time.Duration) error {
	want := fmt.Sprintf("%d", replicas)
	deadline := time.Now().Add(budget)
	var last string
	for time.Now().Before(deadline) {
		ready, _ := kubectlGetJSONPath(ctx, kubeconfig, "statefulset", name, v.Namespace, "{.status.readyReplicas}")
		last = ready
		if ready == want {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("statefulset %q never reached %s ready replicas (last %q)", name, want, last)
}

// proveMetricFlow drives Prometheus' own API over a port-forward: a
// healthy kube-state-metrics scrape target proves service discovery and
// scraping; a PromQL `up` query answering with samples proves the query
// path; the alerting variant additionally proves the Watchdog alert is
// firing in Prometheus AND visible in Alertmanager (the full pipeline).
func (v *KubePrometheusStackVerifier) proveMetricFlow(ctx context.Context, kubeconfig string) error {
	const localPort = "19090"

	pfCtx, cancel := context.WithCancel(ctx)
	pf := exec.CommandContext(pfCtx, "kubectl", "--kubeconfig", kubeconfig,
		"port-forward", "svc/"+v.Name+"-prometheus", localPort+":9090", "-n", v.Namespace)
	var pfOut strings.Builder
	pf.Stdout = &pfOut
	pf.Stderr = &pfOut
	if err := pf.Start(); err != nil {
		cancel()
		return errors.Wrap(err, "starting port-forward to the prometheus service")
	}
	// ONE deferred func, cancel FIRST — Wait blocks forever on a
	// port-forward that is never told to exit.
	defer func() {
		cancel()
		_ = pf.Wait()
	}()

	base := "http://127.0.0.1:" + localPort

	// A kube-state-metrics target reporting health=up proves discovery +
	// scrape end to end (the exporter ships with the stack and its job
	// name is stable). First scrapes land within the default interval;
	// the budget covers Prometheus' own startup and WAL replay.
	if err := awaitHTTPCondition(ctx, base+"/api/v1/targets?state=active", 6*time.Minute, func(body string) error {
		var targets struct {
			Data struct {
				ActiveTargets []struct {
					Health string            `json:"health"`
					Labels map[string]string `json:"labels"`
				} `json:"activeTargets"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(body), &targets); err != nil {
			return errors.Wrap(err, "parsing the targets response")
		}
		for _, t := range targets.Data.ActiveTargets {
			if t.Health != "up" {
				continue
			}
			if !v.KubeStateMetricsEnabled || strings.Contains(t.Labels["job"], "kube-state-metrics") {
				return nil
			}
		}
		if !v.KubeStateMetricsEnabled {
			return errors.New("no healthy scrape target yet")
		}
		return errors.New("no healthy kube-state-metrics target yet")
	}); err != nil {
		return errors.Wrap(err, "METRIC-FLOW: no expected scrape target became healthy")
	}
	if v.KubeStateMetricsEnabled {
		fmt.Printf("  [verify] METRIC-FLOW: kube-state-metrics target scraped and healthy\n")
	} else {
		fmt.Printf("  [verify] METRIC-FLOW: a scrape target is healthy (exporters off)\n")
	}

	// A PromQL query returning samples proves the query path over the
	// stored data.
	if err := awaitHTTPCondition(ctx, base+"/api/v1/query?query=count(up==1)", 2*time.Minute, func(body string) error {
		var query struct {
			Data struct {
				Result []struct {
					Value []interface{} `json:"value"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(body), &query); err != nil {
			return errors.Wrap(err, "parsing the query response")
		}
		if len(query.Data.Result) == 0 || len(query.Data.Result[0].Value) < 2 {
			return errors.New("the query returned no samples yet")
		}
		fmt.Printf("  [verify] METRIC-FLOW: PromQL count(up==1) = %v\n", query.Data.Result[0].Value[1])
		return nil
	}); err != nil {
		return errors.Wrap(err, "METRIC-FLOW: the PromQL query never answered with data")
	}

	if !v.Alerting {
		return nil
	}

	// The Watchdog alert is the stack's own dead-man's-switch: always
	// firing by design, so rule evaluation is proven the moment it shows
	// up as firing.
	if err := awaitHTTPCondition(ctx, base+"/api/v1/alerts", 5*time.Minute, func(body string) error {
		if watchdogFiring(body) {
			return nil
		}
		return errors.New("the Watchdog alert is not firing yet")
	}); err != nil {
		return errors.Wrap(err, "ALERTING: the Watchdog alert never fired in prometheus")
	}
	fmt.Printf("  [verify] ALERTING: the Watchdog alert is firing in prometheus\n")

	// And its presence in Alertmanager proves delivery — the
	// prometheus→alertmanager leg of the pipeline.
	const amLocalPort = "19093"
	amCtx, amCancel := context.WithCancel(ctx)
	amPf := exec.CommandContext(amCtx, "kubectl", "--kubeconfig", kubeconfig,
		"port-forward", "svc/"+v.Name+"-alertmanager", amLocalPort+":9093", "-n", v.Namespace)
	amPf.Stdout = &pfOut
	amPf.Stderr = &pfOut
	if err := amPf.Start(); err != nil {
		amCancel()
		return errors.Wrap(err, "starting port-forward to the alertmanager service")
	}
	defer func() {
		amCancel()
		_ = amPf.Wait()
	}()

	if err := awaitHTTPCondition(ctx, "http://127.0.0.1:"+amLocalPort+"/api/v2/alerts?filter=alertname%3D%22Watchdog%22", 5*time.Minute, func(body string) error {
		var alerts []struct {
			Labels map[string]string `json:"labels"`
		}
		if err := json.Unmarshal([]byte(body), &alerts); err != nil {
			return errors.Wrap(err, "parsing the alertmanager alerts response")
		}
		for _, a := range alerts {
			if a.Labels["alertname"] == "Watchdog" {
				return nil
			}
		}
		return errors.New("the Watchdog alert has not reached alertmanager yet")
	}); err != nil {
		return errors.Wrap(err, "ALERTING: the Watchdog alert never reached alertmanager")
	}
	fmt.Printf("  [verify] ALERTING: the Watchdog alert reached alertmanager — the pipeline is live end to end\n")
	return nil
}

// watchdogFiring reports whether the Prometheus alerts payload carries the
// Watchdog alert in the firing state.
func watchdogFiring(body string) bool {
	var alerts struct {
		Data struct {
			Alerts []struct {
				Labels map[string]string `json:"labels"`
				State  string            `json:"state"`
			} `json:"alerts"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &alerts); err != nil {
		return false
	}
	for _, a := range alerts.Data.Alerts {
		if a.Labels["alertname"] == "Watchdog" && a.State == "firing" {
			return true
		}
	}
	return false
}

// awaitHTTPCondition polls one GET endpoint until check passes or the budget
// runs out (body-read inside the loop — a response dying mid-stream
// retries rather than escaping).
func awaitHTTPCondition(ctx context.Context, url string, budget time.Duration, check func(body string) error) error {
	deadline := time.Now().Add(budget)
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			buf := new(bytes.Buffer)
			_, readErr := buf.ReadFrom(resp.Body)
			resp.Body.Close()
			if readErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				if lastErr = check(buf.String()); lastErr == nil {
					return nil
				}
			} else if readErr != nil {
				lastErr = readErr
			} else {
				lastErr = errors.Errorf("HTTP %d", resp.StatusCode)
			}
		} else {
			lastErr = err
		}
		time.Sleep(10 * time.Second)
	}
	return lastErr
}
