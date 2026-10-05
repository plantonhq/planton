package setdeploy

import (
	"os"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/outputs"
)

// fakeDeployer records handoffs and answers with configured outputs/errors.
type fakeDeployer struct {
	deployed  []string          // identities in handoff order
	manifests map[string]string // identity -> resolved manifest content at handoff
	outputs   map[string]*outputs.CaptureResult
	failOn    string
}

func (d *fakeDeployer) Deploy(node NodePlan, manifestPath string) (*outputs.CaptureResult, error) {
	id := node.Identity.String()
	d.deployed = append(d.deployed, id)
	if d.manifests == nil {
		d.manifests = map[string]string{}
	}
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	d.manifests[id] = string(b)
	if d.failOn == id {
		return nil, errors.New("exit status 1: Error: everything is on fire")
	}
	if d.outputs != nil {
		if captured, ok := d.outputs[id]; ok {
			return captured, nil
		}
	}
	return &outputs.CaptureResult{Flat: map[string]string{}}, nil
}

// recordingEvents captures the event stream for assertions.
type recordingEvents struct {
	started   []string
	succeeded []string
	warnings  []string
	failed    []string
}

func (e *recordingEvents) NodeStarting(_, _ int, node NodePlan) {
	e.started = append(e.started, node.Identity.String())
}
func (e *recordingEvents) NodeSucceeded(node NodePlan, _ *outputs.CaptureResult) {
	e.succeeded = append(e.succeeded, node.Identity.String())
}
func (e *recordingEvents) NodeWarning(_ NodePlan, message string) {
	e.warnings = append(e.warnings, message)
}
func (e *recordingEvents) NodeFailed(node NodePlan, _ error) {
	e.failed = append(e.failed, node.Identity.String())
}

func passingTwoNodePlan(t *testing.T) *Plan {
	t.Helper()
	docs := docsOf(t, map[string]string{"01-producer.yaml": producerYaml, "02-consumer.yaml": consumerYaml})
	plan := Preflight(docs, Flags{}, newFakeProbes())
	if plan.Report.Refused() {
		t.Fatalf("fixture wall must pass: %+v", plan.Report)
	}
	return plan
}

func TestExecute_OutputsFeedDownstreamReferences(t *testing.T) {
	plan := passingTwoNodePlan(t)
	deployer := &fakeDeployer{
		outputs: map[string]*outputs.CaptureResult{
			"TestCatalogKindGeneric/producer@dev": {
				Flat: map[string]string{"id": "tckg-producer"},
			},
		},
	}
	events := &recordingEvents{}
	result, err := Execute(plan, deployer, events)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !result.Succeeded() {
		t.Fatalf("expected success; %+v", result)
	}
	if len(deployer.deployed) != 2 || deployer.deployed[0] != "TestCatalogKindGeneric/producer@dev" {
		t.Fatalf("expected producer first, got %v", deployer.deployed)
	}
	// The consumer's annotated reference must have become the producer's
	// literal id output by handoff time.
	consumerManifest := deployer.manifests["TestCatalogKindGeneric/consumer@dev"]
	if !strings.Contains(consumerManifest, "tckg-producer") {
		t.Fatalf("the consumer's handoff manifest must carry the resolved literal; got:\n%s", consumerManifest)
	}
	if strings.Contains(consumerManifest, "valueFrom") {
		t.Fatalf("no valueFrom may survive resolution; got:\n%s", consumerManifest)
	}
}

func TestExecute_MissingOutputFailsNamingTheField(t *testing.T) {
	plan := passingTwoNodePlan(t)
	deployer := &fakeDeployer{
		outputs: map[string]*outputs.CaptureResult{
			// The producer deploys but exports nothing — the consumer's
			// reference names an output that does not exist.
			"TestCatalogKindGeneric/producer@dev": {Flat: map[string]string{"name": "producer"}},
		},
	}
	events := &recordingEvents{}
	result, err := Execute(plan, deployer, events)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result.Succeeded() {
		t.Fatalf("expected failure")
	}
	if !strings.Contains(result.FailedErr.Error(), "does not export the referenced field") {
		t.Fatalf("the failure must name the missing output, got: %v", result.FailedErr)
	}
}

func TestExecute_FailureStopsAndStatusesTellTheTruth(t *testing.T) {
	plan := passingTwoNodePlan(t)
	deployer := &fakeDeployer{failOn: "TestCatalogKindGeneric/producer@dev"}
	events := &recordingEvents{}
	result, err := Execute(plan, deployer, events)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result.Succeeded() {
		t.Fatalf("expected failure")
	}
	// The engine's error must survive verbatim.
	if !strings.Contains(result.FailedErr.Error(), "everything is on fire") {
		t.Fatalf("engine error must be verbatim, got: %v", result.FailedErr)
	}
	var failed, neverStarted int
	for _, s := range result.Statuses {
		switch s {
		case NodeStatusFailed:
			failed++
		case NodeStatusNeverStarted:
			neverStarted++
		}
	}
	if failed != 1 || neverStarted != 1 {
		t.Fatalf("expected 1 failed + 1 never-started, got %v", result.Statuses)
	}
	if len(events.failed) != 1 {
		t.Fatalf("exactly one NodeFailed event, got %v", events.failed)
	}
}

func TestExecute_SensitiveResolutionWarns(t *testing.T) {
	plan := passingTwoNodePlan(t)
	deployer := &fakeDeployer{
		outputs: map[string]*outputs.CaptureResult{
			"TestCatalogKindGeneric/producer@dev": {
				Flat:    map[string]string{"id": "super-secret-value"},
				Secrets: map[string]bool{"id": true},
			},
		},
	}
	events := &recordingEvents{}
	result, err := Execute(plan, deployer, events)
	if err != nil || !result.Succeeded() {
		t.Fatalf("execute: %v %+v", err, result)
	}
	if len(events.warnings) != 1 || !strings.Contains(events.warnings[0], "sensitive output") {
		t.Fatalf("resolving from a sensitive output must warn, got %v", events.warnings)
	}
	if !strings.Contains(events.warnings[0], "provider-native secret reference") {
		t.Fatalf("the warning must point at the right pattern, got %v", events.warnings)
	}
}

func TestExecute_RefusedPlanNeverRuns(t *testing.T) {
	docs := docsOf(t, map[string]string{"01-consumer.yaml": consumerYaml}) // external ref -> refusal
	plan := Preflight(docs, Flags{}, newFakeProbes())
	if !plan.Report.Refused() {
		t.Fatalf("fixture must refuse")
	}
	deployer := &fakeDeployer{}
	if _, err := Execute(plan, deployer, &recordingEvents{}); err == nil {
		t.Fatalf("executing a refused plan must error")
	}
	if len(deployer.deployed) != 0 {
		t.Fatalf("nothing may deploy behind a refused wall")
	}
}
