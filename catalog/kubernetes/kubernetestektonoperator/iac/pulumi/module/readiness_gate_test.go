package module

import (
	"os/exec"
	"strings"
	"testing"
)

// The readiness gate's offline contract. What the script asks of a live
// cluster is proven by the cold-node-pool E2E scenario on both engines;
// that the two engines run the same bytes is the cross-engine script
// parity guard's (hack/guards/ensure_cross_engine_script_parity.sh).

func TestReadinessGateScriptParses(t *testing.T) {
	out, err := exec.Command("/bin/sh", "-n", "-c", readinessGateScript).CombinedOutput()
	if err != nil {
		t.Fatalf("the readiness gate is not valid POSIX shell: %v\n%s", err, out)
	}
}

func TestReadinessGateEnvironmentNamesTheReleaseHandles(t *testing.T) {
	want := []string{
		"TEKTON_OPERATOR_NAMESPACE=tekton-operator",
		"TEKTON_OPERATOR_DEPLOYMENT=tekton-operator",
		"TEKTON_WEBHOOK_DEPLOYMENT=tekton-operator-webhook",
	}
	got := readinessGateEnvironment()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("readiness gate environment = %q, want %q (the release manifest's fixed handles)", got, want)
	}
}

// A deploy host with no kubectl is refused with the fix named, never with a
// shell's "command not found" in the middle of an apply.
func TestReadinessGateWithoutKubectlSaysSo(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := runReadinessGate()
	if err == nil {
		t.Fatal("the readiness gate passed with no kubectl on PATH")
	}
	if !strings.Contains(err.Error(), "needs kubectl on PATH") {
		t.Fatalf("the refusal does not say kubectl is missing: %v", err)
	}
}

func TestReadinessGateDataNamesTheRelease(t *testing.T) {
	if got := readinessGateData()["operatorRelease"]; got != vars.OperatorRelease {
		t.Fatalf("readiness gate data names release %q, want %q: a release move must replace the stand-in so it gates again", got, vars.OperatorRelease)
	}
}
