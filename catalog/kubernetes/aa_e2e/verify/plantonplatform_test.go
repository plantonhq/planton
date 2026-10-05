package verify

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

// The verifier's default chart is the kind's default option, and the two
// modules carry the same pin: a module that installs another chart than the
// one its definition names would be verified against the wrong version.
func TestPlantonOperatorDefaultChartVersionIsTheModulesPin(t *testing.T) {
	want := plantonOperatorDefaultChartVersion()
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(want) {
		t.Fatalf("the kind's chart_version default must be an exact version, got %q", want)
	}
	for path, pin := range map[string]*regexp.Regexp{
		"../../kubernetesplantonoperator/iac/pulumi/module/vars.go": regexp.MustCompile(`DefaultChartVersion:\s+"([^"]+)"`),
		"../../kubernetesplantonoperator/iac/tf/locals.tf":          regexp.MustCompile(`default_chart_version\s*=\s*"([^"]+)"`),
	} {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		m := pin.FindSubmatch(src)
		if m == nil {
			t.Errorf("%s: no default chart pin found; the pattern this test reads it by has moved", path)
		} else if string(m[1]) != want {
			t.Errorf("%s pins %s; the kind's default is %s -- move the three together", path, m[1], want)
		}
	}
}

// The operator's phase Error is transient (any component error for one
// reconcile cycle, requeued and retried); only VersionSupported=False is
// terminal. The wait must keep going through the former and stop at once
// on the latter.
func TestPlatformBootVerdict(t *testing.T) {
	cases := []struct {
		name             string
		phase            string
		versionSupported string
		want             bootVerdict
	}{
		{"ready", "Ready", "True", bootReady},
		{"deploying", "Deploying", "True", bootWaiting},
		{"pending before the first reconcile", "Pending", "", bootWaiting},
		{"a component error is a retry, not a stop", "Error", "True", bootWaiting},
		{"a component error before the condition exists", "Error", "", bootWaiting},
		{"the version floor refuses the declaration", "Error", "False", bootRefused},
		{"a refusal wins over any phase", "Deploying", "False", bootRefused},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := platformBootVerdict(tc.phase, tc.versionSupported); got != tc.want {
				t.Fatalf("platformBootVerdict(%q, %q) = %v, want %v", tc.phase, tc.versionSupported, got, tc.want)
			}
		})
	}
}

// A Deployment has rolled out at a tag only when its image carries the tag
// AND the rollout is complete: during an image change the previous pod stays
// available while the new one starts, and both counters can read 1 while
// they name different pods.
func TestDeploymentRolledOutAt(t *testing.T) {
	deploy := func(image string, generation, observed, total, updated, available int64) []byte {
		return []byte(fmt.Sprintf(`{"metadata":{"generation":%d},"spec":{"replicas":1,"template":{"spec":{"containers":[{"image":%q}]}}},`+
			`"status":{"observedGeneration":%d,"replicas":%d,"updatedReplicas":%d,"availableReplicas":%d}}`,
			generation, image, observed, total, updated, available))
	}
	cases := []struct {
		name string
		json []byte
		want bool
	}{
		{"complete at the tag", deploy("ghcr.io/x/control-plane:v0.0.46", 2, 2, 1, 1, 1), true},
		{"still on the old image", deploy("ghcr.io/x/control-plane:v0.0.45", 2, 2, 1, 1, 1), false},
		{"new image, old pod still serving beside the new one", deploy("ghcr.io/x/control-plane:v0.0.46", 2, 2, 2, 1, 1), false},
		{"new image, new pod not yet created", deploy("ghcr.io/x/control-plane:v0.0.46", 2, 2, 1, 0, 1), false},
		{"new image, spec not yet observed", deploy("ghcr.io/x/control-plane:v0.0.46", 2, 1, 1, 1, 1), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := deploymentRolledOutAt(tc.json, "v0.0.46")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("deploymentRolledOutAt = %v, want %v", got, tc.want)
			}
		})
	}
	if _, err := deploymentRolledOutAt([]byte("not json"), "v0.0.46"); err == nil {
		t.Fatal("unparseable input must be an error, not a verdict")
	}
}

// The component trace must read the same whatever order the API server's
// map arrives in, and must not swallow output it cannot parse.
func TestComponentPhaseSummary(t *testing.T) {
	raw := `{"identity":{"message":"Waiting for the identity server","phase":"Deploying"},` +
		`"controlPlane":{"message":"Waiting for dependency: identity","phase":"Pending"},` +
		`"gateway":{"message":"Front door ready","phase":"Ready"}}`
	if got, want := componentPhaseSummary(raw), "controlPlane=Pending gateway=Ready identity=Deploying"; got != want {
		t.Fatalf("componentPhaseSummary = %q, want %q", got, want)
	}
	if got := componentPhaseSummary("  not json  "); got != "not json" {
		t.Fatalf("unparseable input must pass through trimmed, got %q", got)
	}
	if got := componentPhaseSummary(""); got != "" {
		t.Fatalf("empty input must stay empty, got %q", got)
	}
}

// A declared size is compared quantity by quantity, and only what was
// declared: the operator's defaults beside it are not the manifest's to
// check, and a quantity the pod does not carry reads as the gap it is.
func TestSizingMismatches(t *testing.T) {
	declared := declaredSizing(map[string]interface{}{"limits": map[string]interface{}{"memory": "7Gi"}})
	if len(declared) != 1 || declared["limits.memory"] != "7Gi" {
		t.Fatalf("declaredSizing flattened to %v", declared)
	}
	if got := sizingMismatches(declared, `{"limits":{"memory":"7Gi"},"requests":{"cpu":"250m","memory":"1Gi"}}`, "the container"); len(got) != 0 {
		t.Errorf("a pod running the declared limit beside the default requests matches, got %v", got)
	}
	got := sizingMismatches(declared, `{"limits":{"memory":"6Gi"},"requests":{"cpu":"250m","memory":"1Gi"}}`, "the container")
	if len(got) != 1 || got[0] != `the container runs limits.memory "6Gi" where the manifest declares "7Gi"` {
		t.Errorf("the operator's default where a size was declared is the mismatch, got %v", got)
	}
	if got := sizingMismatches(declared, ``, "the readback"); len(got) != 1 {
		t.Errorf("an absent readback is a mismatch, never a pass, got %v", got)
	}
	if declaredSizing(nil) != nil || declaredSizing(map[string]interface{}{}) != nil {
		t.Error("a manifest that declares no size has nothing to verify")
	}
}

// The verifier reads the declared trace store the way manifests write it, in
// either key spelling, and reads nothing when a scenario traces nothing -- so
// the tracing check runs exactly on the scenarios that declare one.
func TestPlantonPlatformVerifierReadsTheDeclaredTraceStore(t *testing.T) {
	const endpoint = "http://cluster-traces-collector.observability.svc.cluster.local:4318"
	for name, spec := range map[string]string{
		"snake_case": "observability:\n    otlp_http_endpoint:\n      value: " + endpoint,
		"camelCase":  "observability:\n    otlpHttpEndpoint:\n      value: " + endpoint,
		"absent":     "create_namespace: true",
	} {
		path := t.TempDir() + "/scenario.yaml"
		manifest := fmt.Sprintf("kind: KubernetesPlantonPlatform\nspec:\n  version: v0.0.140\n  %s\n", spec)
		if err := os.WriteFile(path, []byte(manifest), 0o600); err != nil {
			t.Fatal(err)
		}
		want := endpoint
		if name == "absent" {
			want = ""
		}
		if got := newPlantonPlatformVerifier("planton", "planton", path).TracesEndpoint; got != want {
			t.Errorf("%s: TracesEndpoint = %q, want %q", name, got, want)
		}
	}
}

// The vault keys check reads only the Secret's key names, from the data map
// kubectl prints as JSON (its JSONPath cannot list a map's keys).
func TestMapKeysListsASecretsKeyNames(t *testing.T) {
	keys, err := mapKeys(`{"unseal-keys":"c2hhcmVz","root-token":"dG9rZW4="}`)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(keys) != "[root-token unseal-keys]" {
		t.Errorf("keys = %v, want [root-token unseal-keys]", keys)
	}
	if keys, err := mapKeys(""); err != nil || len(keys) != 0 {
		t.Errorf("an empty Secret = %v, %v; want no keys", keys, err)
	}
}
