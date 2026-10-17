package verify

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The resource lanes are only as honest as the verifiers' reading of their
// scenarios: which behavior each lane proves, which claims and pools it
// expects. Each committed scenario is read here offline, so a scenario edit
// that changes what a lane asserts fails a test instead of quietly proving
// something else.

const agentSandboxResourceScenarios = "../../kubernetesagentsandbox/e2e/scenarios"

func TestAgentSandboxResourceVerifierReadsTheScenarios(t *testing.T) {
	cases := map[string]struct {
		suspended, suspendLane bool
		claims                 []string
	}{
		"minimal.yaml":               {},
		"full-surface.yaml":          {claims: []string{"workspace"}},
		"behavioral-suspend.yaml":    {suspendLane: true, claims: []string{"workspace"}},
		"behavioral-suspend.to.yaml": {suspended: true, suspendLane: true, claims: []string{"workspace"}},
	}
	if _, err := os.Stat(agentSandboxResourceScenarios); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	files, err := filepath.Glob(filepath.Join(agentSandboxResourceScenarios, "*.yaml"))
	if err != nil || len(files) != len(cases) {
		t.Fatalf("every scenario needs a case here: found %d files, %d cases (%v)", len(files), len(cases), err)
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			v := newAgentSandboxVerifier(filepath.Join(agentSandboxResourceScenarios, name), "s", "ns")
			if v.Suspended != want.suspended || v.SuspendLane != want.suspendLane || !reflect.DeepEqual(v.ClaimTemplates, want.claims) {
				t.Errorf("suspended=%t suspendLane=%t claims=%v, want %t %t %v", v.Suspended, v.SuspendLane, v.ClaimTemplates, want.suspended, want.suspendLane, want.claims)
			}
		})
	}
}

const agentSandboxTemplateScenarios = "../../kubernetesagentsandboxtemplate/e2e/scenarios"

func TestAgentSandboxTemplateVerifierReadsTheScenarios(t *testing.T) {
	if _, err := os.Stat(agentSandboxTemplateScenarios); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	cases := map[string]struct{ managed, custom bool }{
		"minimal.yaml":              {managed: true},
		"full-surface.yaml":         {managed: true, custom: true},
		"behavioral-unmanaged.yaml": {custom: true},
	}
	files, err := filepath.Glob(filepath.Join(agentSandboxTemplateScenarios, "*.yaml"))
	if err != nil || len(files) != len(cases) {
		t.Fatalf("every scenario needs a case here: found %d files, %d cases (%v)", len(files), len(cases), err)
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			v := newAgentSandboxTemplateVerifier(filepath.Join(agentSandboxTemplateScenarios, name), "t", "ns")
			if v.Managed != want.managed || v.CustomRules != want.custom {
				t.Errorf("managed=%t custom=%t, want %t %t", v.Managed, v.CustomRules, want.managed, want.custom)
			}
		})
	}
}

const (
	agentSandboxWarmPoolScenarios = "../../kubernetesagentsandboxwarmpool/e2e/scenarios"
	agentSandboxClaimScenarios    = "../../kubernetesagentsandboxclaim/e2e/scenarios"
)

func TestAgentSandboxWarmPoolVerifierReadsTheScenarios(t *testing.T) {
	if _, err := os.Stat(agentSandboxWarmPoolScenarios); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	cases := map[string]int64{"minimal.yaml": 1, "full-surface.yaml": 2}
	files, err := filepath.Glob(filepath.Join(agentSandboxWarmPoolScenarios, "*.yaml"))
	if err != nil || len(files) != len(cases) {
		t.Fatalf("every scenario needs a case here: found %d files, %d cases (%v)", len(files), len(cases), err)
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			if got := newAgentSandboxWarmPoolVerifier(filepath.Join(agentSandboxWarmPoolScenarios, name), "p", "ns").Replicas; got != want {
				t.Errorf("replicas = %d, want %d", got, want)
			}
		})
	}
}

func TestAgentSandboxClaimVerifierReadsTheScenarios(t *testing.T) {
	if _, err := os.Stat(agentSandboxClaimScenarios); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	cases := map[string]struct {
		pool, refused string
		cold          bool
		labels        map[string]string
		env           []string
	}{
		"minimal.yaml":      {pool: "e2e-asbx-cl-pool", labels: map[string]string{}},
		"full-surface.yaml": {pool: "e2e-asbx-cl-pool-open", cold: true, labels: map[string]string{"sandbox.users.io/task": "e2e-full"}, env: []string{"SANDBOX_TASK"}},
		"behavioral-policy-refusal.yaml": {pool: "e2e-asbx-cl-pool", cold: true, refused: "EnvVarsInjectionRejected",
			labels: map[string]string{}, env: []string{"SANDBOX_TASK"}},
	}
	files, err := filepath.Glob(filepath.Join(agentSandboxClaimScenarios, "*.yaml"))
	if err != nil || len(files) != len(cases) {
		t.Fatalf("every scenario needs a case here: found %d files, %d cases (%v)", len(files), len(cases), err)
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			v := newAgentSandboxClaimVerifier(filepath.Join(agentSandboxClaimScenarios, name), "c", "ns")
			if v.Pool != want.pool || v.Cold != want.cold || v.RefusedReason != want.refused ||
				!reflect.DeepEqual(v.PodLabels, want.labels) || !reflect.DeepEqual(v.EnvNames, want.env) {
				t.Errorf("got pool=%q cold=%t refused=%q labels=%v env=%v, want %+v", v.Pool, v.Cold, v.RefusedReason, v.PodLabels, v.EnvNames, want)
			}
		})
	}
}
