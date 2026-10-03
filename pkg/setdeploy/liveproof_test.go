package setdeploy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/iac/provisioner"
)

// The zero-cloud live proof: a two-node set through the REAL deploy path —
// preflight wall with live probes, identity-keyed workspaces, actual `tofu
// init` + `tofu apply`, real output capture (`tofu output -json`), and
// output-fed resolution into the downstream node — with no cloud, no
// credentials, and no network. The _test kind's module manages a
// terraform_data builtin, and its annotated_ref composes on
// status.outputs.id, so the producer's REAL captured output must land in the
// consumer's manifest for the consumer's apply to succeed.
//
// Requires tofu on PATH; skips loudly otherwise (the skip is the recorded
// notice, never silence).
func TestLiveProof_TwoNodeSetDeploysThroughTofu(t *testing.T) {
	if testing.Short() {
		t.Skip("live proof runs the real tofu binary — skipped in -short")
	}
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Skip("live proof needs tofu on PATH — not found; the deploy path is otherwise covered by the fake-deployer suite")
	}

	// Everything the run writes — node workspaces, plugin cache — lands in a
	// throwaway HOME so the proof never touches the developer's ~/.planton.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("TF_PLUGIN_CACHE_DIR", t.TempDir())

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	moduleDir := filepath.Join(repoRoot, "catalog", "_test", "testcatalogkindgeneric", "iac", "tf")
	if _, err := os.Stat(filepath.Join(moduleDir, "main.tf")); err != nil {
		t.Fatalf("the _test kind's tofu module is missing at %s: %v", moduleDir, err)
	}

	docs := docsOf(t, map[string]string{
		"01-producer.yaml": producerYaml,
		"02-consumer.yaml": consumerYaml,
	})

	plan := Preflight(docs, Flags{}, LiveProbes{})
	if plan.Report.Refused() {
		t.Fatalf("the wall must pass on this machine; report:\n%s", reportDump(plan.Report))
	}

	deployer := &EngineDeployer{
		ResolveModuleDir: func(kindName string, prov provisioner.ProvisionerType) (string, bool) {
			return moduleDir, true
		},
	}
	defer deployer.Close()

	events := &recordingEvents{}
	result, err := Execute(plan, deployer, events)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !result.Succeeded() {
		t.Fatalf("deploy failed: %v (statuses %v)", result.FailedErr, result.Statuses)
	}

	// The producer's REAL captured outputs fed the consumer: its id output is
	// deterministic ("tckg-" + name), so the capture path is provable.
	producerID := plan.Set.Nodes[plan.Order[0]].Identity
	captured := result.Outputs[producerID]
	if captured == nil || captured.Flat["id"] != "tcrg-producer" {
		t.Fatalf("expected the producer's captured id output; got %+v", captured)
	}
	if captured.Flat["url"] != "test://producer" {
		t.Fatalf("expected the producer's url output; got %+v", captured.Flat)
	}

	if len(events.succeeded) != 2 {
		t.Fatalf("expected both nodes to succeed, got %v", events.succeeded)
	}

	// The re-run story, proven: local state persisted in the identity-keyed
	// workspaces, so the same set applies again as no-ops and succeeds.
	rerun, err := Execute(Preflight(docs, Flags{}, LiveProbes{}), deployer, &recordingEvents{})
	if err != nil || !rerun.Succeeded() {
		t.Fatalf("re-running the same set must succeed as no-ops: %v %+v", err, rerun)
	}

	// The workspaces are stable and identity-keyed — local state lives where
	// the report said it does.
	home, _ := os.UserHomeDir()
	stateFile := filepath.Join(home, ".planton", "setdeploy", "dev", "testcatalogkindgeneric", "producer", "terraform.tfstate")
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("the producer's local state must persist in its identity-keyed workspace: %v", err)
	}
	b, err := os.ReadFile(stateFile)
	if err != nil || !strings.Contains(string(b), "tcrg-producer") {
		t.Fatalf("the persisted state must hold the producer's resource")
	}
}
