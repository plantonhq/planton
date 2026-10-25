package verify

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/plantonhq/planton/catalog/kubernetes/aa_e2e/verify/internal/ateapipb"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"k8s.io/apimachinery/pkg/api/resource"
)

// The Substrate resource lanes are only as honest as the verifier's reading
// of their scenarios. Each committed worker pool and sandbox config scenario
// is read here offline, so a scenario edit that changes what a lane asserts
// shows up as a failing test: the pool's size, class, labels and named
// config, and which lane runs the actor lifecycle. The counter template the
// actor lane creates is held to upstream's counter demo and to the pool it
// must fit on, and the control plane's outputs are read only from the
// prerequisite the runner deployed.

const (
	agentSubstrateWorkerPoolScenarios    = "../../kubernetesagentsubstrateworkerpool/e2e/scenarios"
	agentSubstrateSandboxConfigScenarios = "../../kubernetesagentsubstratesandboxconfig/e2e/scenarios"
)

func TestAgentSubstrateWorkerPoolVerifierReadsTheScenarios(t *testing.T) {
	type lane struct {
		replicas int64
		labels   map[string]string
		actor    bool
	}
	cases := map[string]lane{
		"minimal.yaml":                    {replicas: 1, labels: map[string]string{}},
		"full-surface.yaml":               {replicas: 2, labels: map[string]string{"tenant": "acme"}},
		"behavioral-actor-lifecycle.yaml": {replicas: 2, labels: map[string]string{"workload": "e2e-counter"}, actor: true},
	}
	files, err := filepath.Glob(filepath.Join(agentSubstrateWorkerPoolScenarios, "*.yaml"))
	if err != nil || len(files) != len(cases) {
		t.Fatalf("every scenario needs a case here: found %d files, %d cases (%v)", len(files), len(cases), err)
	}
	for name, want := range cases {
		path := filepath.Join(agentSubstrateWorkerPoolScenarios, name)
		v := newAgentSubstrateWorkerPoolVerifier(path, "pool", "e2e-ate-workers")
		if v.Replicas != want.replicas || v.Class != "gvisor" || v.ActorLifecycle != want.actor || !reflect.DeepEqual(v.Labels, want.labels) {
			t.Errorf("%s: read replicas=%d class=%s actor=%t labels=%v, want replicas=%d class=gvisor actor=%t labels=%v",
				name, v.Replicas, v.Class, v.ActorLifecycle, v.Labels, want.replicas, want.actor, want.labels)
		}
		if manifestAnnotation(path, "planton.dev/e2e-cluster-profile") != "pod-certificates" {
			t.Errorf("%s: a worker pool lane runs only on the pod-certificates profile", name)
		}
	}
}

// The actor lane's pool names a non-default SandboxConfig; once the runner
// resolves the reference, the verifier must expect the pool to run exactly
// that config.
func TestAgentSubstrateWorkerPoolVerifierReadsAResolvedConfigReference(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(agentSubstrateWorkerPoolScenarios, "behavioral-actor-lifecycle.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	resolved := strings.Replace(string(raw), "          valueFrom:\n            name: e2e-ate-sandbox-config", "          value: e2e-gvisor-workers", 1)
	if resolved == string(raw) {
		t.Fatal("the actor lane no longer references its SandboxConfig fixture by valueFrom; update this test with the scenario")
	}
	path := filepath.Join(t.TempDir(), "behavioral-actor-lifecycle.yaml")
	if err := os.WriteFile(path, []byte(resolved), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := newAgentSubstrateWorkerPoolVerifier(path, "pool", "ns").ConfigRef; got != "e2e-gvisor-workers" {
		t.Errorf("ConfigRef = %q, want e2e-gvisor-workers (the fixture's spec.name)", got)
	}
}

func TestReferenceValue(t *testing.T) {
	for _, tc := range []struct {
		in   interface{}
		want interface{}
	}{
		{"gvisor-default", "gvisor-default"},
		{map[string]interface{}{"value": "gvisor-default"}, "gvisor-default"},
		{nil, nil},
	} {
		if got := referenceValue(tc.in); got != tc.want {
			t.Errorf("referenceValue(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestAgentSubstrateSandboxConfigVerifierReadsTheScenarios(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(agentSubstrateSandboxConfigScenarios, "*.yaml"))
	if err != nil || len(files) != 2 {
		t.Fatalf("expected the minimal and full-surface scenarios, found %v (%v)", files, err)
	}
	for _, path := range files {
		v := newAgentSubstrateSandboxConfigVerifier(path, "config")
		if v.Class != "gvisor" {
			t.Errorf("%s: class = %q, want gvisor", path, v.Class)
		}
		if manifestAnnotation(path, "planton.dev/e2e-cluster-profile") != "pod-certificates" {
			t.Errorf("%s: a sandbox config lane runs only on the pod-certificates profile", path)
		}
		if strings.Contains(readFile(t, path), "is-class-default") {
			t.Errorf("%s: a lane must not make its config the class default; the next lanes on the shared cluster rely on gvisor-default", path)
		}
	}
}

// The counter template is upstream's counter demo (demos/counter) with
// VOLUMES fidelity, so the lane's two counters can tell a restored actor from
// one that never slept. It names no storage location, so its snapshots prove
// the control plane's own.
func TestCounterTemplateIsUpstreamsCounter(t *testing.T) {
	tpl := counterTemplate("lane", "t", "img@sha256:x", map[string]string{"workload": "e2e-counter"})
	c := tpl.GetContainers()[0]
	if c.GetCommand()[0] != "/ko-app/counter" || c.GetWakeupProbe().GetHttpGet().GetPort() != 80 || c.GetWakeupProbe().GetHttpGet().GetPath() != "/readyz" {
		t.Errorf("container = %v, want upstream's counter: /ko-app/counter, woken on :80/readyz", c)
	}
	if len(c.GetVolumeMounts()) != 1 || c.GetVolumeMounts()[0].GetMountPath() != "/home/counter" || tpl.GetVolumes()[0].GetDurableDir() == nil {
		t.Errorf("the counter's file counter lives in a durable directory at /home/counter; got mounts %v volumes %v", c.GetVolumeMounts(), tpl.GetVolumes())
	}
	if tpl.GetSnapshotConfig().GetPreferredFidelity() != ateapipb.SnapshotFidelity_SNAPSHOT_FIDELITY_VOLUMES {
		t.Errorf("fidelity = %v, want VOLUMES", tpl.GetSnapshotConfig().GetPreferredFidelity())
	}
	if tpl.GetWorkerSelector().GetMatchLabels()["workload"] != "e2e-counter" || tpl.GetSnapshotConfig().GetStorageLocation() != "" {
		t.Errorf("template must select the pool by its labels and store under the control plane's location; got %v %v", tpl.GetWorkerSelector(), tpl.GetSnapshotConfig())
	}
	if limits := tpl.GetResources().GetLimits(); limits.GetCpuMillicores() == 0 || limits.GetMemoryBytes() == 0 {
		t.Errorf("every actor declares cpu and memory limits; got %v", limits)
	}
}

// An actor is placed only on a worker whose unspent limits cover the limits
// its template declares, so the actor lane's counter must fit the lane
// pool's worker limits.
func TestCounterActorFitsTheActorLanePool(t *testing.T) {
	spec := manifestSpecMap(filepath.Join(agentSubstrateWorkerPoolScenarios, "behavioral-actor-lifecycle.yaml"))
	template, _ := specField(spec, "template").(map[string]interface{})
	resources, _ := specField(template, "resources").(map[string]interface{})
	limits, _ := specField(resources, "limits").(map[string]interface{})
	for name, want := range map[string]*resource.Quantity{
		"cpu":    resource.NewMilliQuantity(counterActorMillicores, resource.DecimalSI),
		"memory": resource.NewQuantity(counterActorMemory, resource.BinarySI),
	} {
		worker := resource.MustParse(stringOr(limits[name], "0"))
		if worker.Cmp(*want) < 0 {
			t.Errorf("the actor lane's workers have %s limit %s, below the counter actor's %s", name, worker.String(), want.String())
		}
	}
}

func TestAgentSubstrateOutputsComeFromThePrerequisite(t *testing.T) {
	if _, err := agentSubstrateOutputsFromPrerequisite(context.Background()); err == nil {
		t.Error("with no prerequisite deployed the verifier must refuse to guess the control plane's outputs")
	}
	ctx := context.WithValue(context.Background(), provider.DependencyOutputsKey{}, provider.DependencyOutputs{
		agentSubstrateKind: {"e2e-ate-prereq": {"api_address": "api.ate-system.svc:443"}},
	})
	outputs, err := agentSubstrateOutputsFromPrerequisite(ctx)
	if err != nil || outputs["api_address"] != "api.ate-system.svc:443" {
		t.Errorf("outputs = %v, %v; want the deployed control plane's", outputs, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Every run names its own atespace, a DNS label, so a run never reuses an
// interrupted run's objects.
func TestLaneAtespaceIsFreshPerRun(t *testing.T) {
	label := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$`)
	first, second := laneAtespace("actors"), laneAtespace("actors")
	if !label.MatchString(first) || !label.MatchString(second) {
		t.Errorf("lane atespaces %q and %q must be DNS labels", first, second)
	}
	if first == second {
		t.Errorf("two runs named the same atespace %q", first)
	}
}
