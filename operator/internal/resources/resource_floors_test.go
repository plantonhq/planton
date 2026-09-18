package resources

import (
	"fmt"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Every workload the operator renders carries a sizing the operator chose:
// requests for CPU and memory so it schedules honestly, a memory limit so a
// leak or a runaway dataset is an eviction or a restart instead of a node
// taken down, and NO CPU limit so a burst is never throttled into failing its
// own probes (the house pattern). A chart's smallest preset by omission is
// how the bundled cache crash-looped on a live instance; this file is the
// floor under every component, at the level Kubernetes sees.

// The workload kinds whose pod templates carry containers to judge.
const (
	floorKindDeployment  = "Deployment"
	floorKindStatefulSet = "StatefulSet"
	floorKindDaemonSet   = "DaemonSet"
	floorKindJob         = "Job"
)

// assertHousePattern is the one judgment for a rendered container's resources.
func assertHousePattern(t *testing.T, who string, res corev1.ResourceRequirements) {
	t.Helper()
	if res.Requests.Cpu().IsZero() || res.Requests.Memory().IsZero() {
		t.Errorf("%s: expected CPU + memory requests (starvation guard)", who)
	}
	if res.Limits.Memory().IsZero() {
		t.Errorf("%s: expected a memory limit", who)
	}
	if !res.Limits.Cpu().IsZero() {
		t.Errorf("%s: CPU must not be limited (requests-only, the house pattern)", who)
	}
	if res.Limits.Memory().Cmp(*res.Requests.Memory()) < 0 {
		t.Errorf("%s: the memory limit %s is below the request %s", who, res.Limits.Memory(), res.Requests.Memory())
	}
}

// assertHousePatternUnstructured reads a rendered container's resources block
// the way the API server will and applies the same judgment.
func assertHousePatternUnstructured(t *testing.T, who string, container map[string]any) {
	t.Helper()
	cpuReq, _, _ := unstructured.NestedString(container, "resources", "requests", "cpu")
	memReq, _, _ := unstructured.NestedString(container, "resources", "requests", "memory")
	memLim, _, _ := unstructured.NestedString(container, "resources", "limits", "memory")
	_, cpuLimited, _ := unstructured.NestedString(container, "resources", "limits", "cpu")
	if cpuReq == "" || memReq == "" {
		t.Errorf("%s: expected CPU + memory requests on the rendered container, got cpu=%q memory=%q", who, cpuReq, memReq)
	}
	if memLim == "" {
		t.Errorf("%s: expected a memory limit on the rendered container", who)
	}
	if cpuLimited {
		t.Errorf("%s: CPU must not be limited (requests-only, the house pattern)", who)
	}
}

func TestResourceFloor_TypedDeployments(t *testing.T) {
	assertHousePattern(t, "control-plane",
		ControlPlaneDeployment(testControlPlaneConfig()).Spec.Template.Spec.Containers[0].Resources)
	assertHousePattern(t, "runner",
		RunnerDeployment(testRunnerConfig()).Spec.Template.Spec.Containers[0].Resources)
	assertHousePattern(t, "gateway",
		GatewayDeployment(GatewayConfig{CRName: "planton", Namespace: "default"}).Spec.Template.Spec.Containers[0].Resources)
	assertHousePattern(t, "console",
		ConsoleDeployment(ConsoleConfig{CRName: "planton", Namespace: "default", Version: "v1.0.0", Replicas: 1}).Spec.Template.Spec.Containers[0].Resources)

	identity := IdentityDeployment(testIdentityConfig()).Spec.Template.Spec
	assertHousePattern(t, "identity", identity.Containers[0].Resources)
	assertHousePattern(t, "identity ensure-database init container", identity.InitContainers[0].Resources)
}

// The control plane's pair is the hosted product's pair for the same service:
// one number in both homes, and the heap follows from the image's own
// MaxRAMPercentage, so the limit is the heap rule.
func TestResourceFloor_ControlPlaneMatchesTheHostedPair(t *testing.T) {
	res := ControlPlaneDeployment(testControlPlaneConfig()).Spec.Template.Spec.Containers[0].Resources
	if got := res.Requests.Memory().String(); got != controlPlaneMemoryRequest {
		t.Errorf("expected the memory request %s, got %s", controlPlaneMemoryRequest, got)
	}
	if got := res.Limits.Memory().String(); got != controlPlaneMemoryLimit {
		t.Errorf("expected the memory limit %s, got %s", controlPlaneMemoryLimit, got)
	}
}

// renderedContainers returns every container of every workload the chart
// renders, keyed "Kind/name/container", so a chart that grows a workload the
// operator forgot to size fails here.
func renderedContainers(t *testing.T, chartData []byte, releaseName string, values map[string]any) map[string]map[string]any {
	t.Helper()
	objs, err := RenderHelmChart(chartData, releaseName, "default", values)
	if err != nil {
		t.Fatalf("failed to render chart: %v", err)
	}
	out := map[string]map[string]any{}
	for _, obj := range objs {
		var containers []any
		switch obj.GetKind() {
		case floorKindDeployment, floorKindStatefulSet, floorKindDaemonSet, floorKindJob:
			containers, _, _ = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
		default:
			continue
		}
		for _, raw := range containers {
			c, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			name, _, _ := unstructured.NestedString(c, "name")
			out[fmt.Sprintf("%s/%s/%s", obj.GetKind(), obj.GetName(), name)] = c
		}
	}
	if len(out) == 0 {
		t.Fatalf("%s rendered no workload containers", releaseName)
	}
	return out
}

// The OpenFGA chart exposes one resources key, for the server Deployment; its
// migrate Job is a one-shot the chart does not let the operator size.
func TestResourceFloor_OpenFGAChart(t *testing.T) {
	containers := renderedContainers(t, LoadOpenFGAChart(), "test-openfga", OpenFGAHelmValues("test", "default"))
	sized := 0
	for who, c := range containers {
		if strings.HasPrefix(who, floorKindDeployment+"/") {
			assertHousePatternUnstructured(t, who, c)
			sized++
		}
	}
	if sized == 0 {
		t.Fatalf("expected the server Deployment among the rendered workloads, got %v", keys(containers))
	}
}

// Temporal renders six workloads; the four server services take the server
// default with history above it, and the web UI, the admin tools, and the
// schema Jobs take the small auxiliary floor. Every one of them carries a
// floor, so a chart bump that adds a workload fails here.
func TestResourceFloor_TemporalChart(t *testing.T) {
	containers := renderedContainers(t, LoadTemporalChart(), "test-temporal", TemporalHelmValues("test", "default"))
	seenHistory := false
	for who, c := range containers {
		assertHousePatternUnstructured(t, who, c)
		if who == "Deployment/test-temporal-history/temporal-history" {
			seenHistory = true
			memLim, _, _ := unstructured.NestedString(c, "resources", "limits", "memory")
			if memLim != temporalHistoryMemoryLimit {
				t.Errorf("history must carry its own, larger limit %s; got %q", temporalHistoryMemoryLimit, memLim)
			}
		}
	}
	if !seenHistory {
		t.Errorf("expected the history service among the rendered workloads, got %v", keys(containers))
	}
}

func TestResourceFloor_BarmanPluginChart(t *testing.T) {
	objs, err := LoadBarmanCloudPluginManifests()
	if err != nil {
		t.Fatalf("rendering plugin chart: %v", err)
	}
	sized := 0
	for _, obj := range objs {
		if obj.GetKind() != floorKindDeployment {
			continue
		}
		containers, _, _ := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
		for _, raw := range containers {
			c := raw.(map[string]any)
			name, _, _ := unstructured.NestedString(c, "name")
			assertHousePatternUnstructured(t, "barman plugin "+name, c)
			sized++
		}
	}
	if sized == 0 {
		t.Fatal("expected the plugin's Deployment container among the rendered objects")
	}
}

func keys(m map[string]map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
