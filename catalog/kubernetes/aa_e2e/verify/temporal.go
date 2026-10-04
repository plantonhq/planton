package verify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	temporalworkflowservice "go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	temporalworker "go.temporal.io/sdk/worker"
	temporalworkflow "go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

// describeNamespaceRequest / registerNamespaceRequest build the raw
// WorkflowService requests the namespace bootstrap uses (registration
// requires an explicit retention period).
func describeNamespaceRequest(temporalNamespace string) *temporalworkflowservice.DescribeNamespaceRequest {
	return &temporalworkflowservice.DescribeNamespaceRequest{Namespace: temporalNamespace}
}

func registerNamespaceRequest(temporalNamespace string) *temporalworkflowservice.RegisterNamespaceRequest {
	return &temporalworkflowservice.RegisterNamespaceRequest{
		Namespace:                        temporalNamespace,
		WorkflowExecutionRetentionPeriod: durationpb.New(24 * time.Hour),
	}
}

// TemporalVerifier checks a Temporal cluster to the point a customer
// could run their workflows on it: all four server Deployments rolled
// out (frontend, history, matching, worker), the Web UI rolled out and
// answering (when enabled), and THE ENGINE PROOF on every lane — a real
// worker built on the official Temporal Go SDK connects through the
// frontend gRPC endpoint, and a workflow EXECUTES TO COMPLETION with its
// result read back (a workflow engine that cannot complete a workflow is
// not a workflow engine). The proof exercises the whole story: gRPC
// frontend → matching (task dispatch) → history (state in the composed
// PostgreSQL) → worker execution → visibility.
//
// The behavioral-state scenario (recognized by name) additionally
// DELETES the history and frontend pods, waits for REPLACEMENTS (new
// UIDs), then DESCRIBES the already-completed workflow (its history must
// survive in the database — the storage-separation proof) and runs a
// SECOND workflow end to end.
//
// Destroy is clean by design: Temporal installs no CRDs — everything
// leaves with the release.
type TemporalVerifier struct {
	Namespace string
	Name      string
	// WebUiEnabled gates the UI rollout + HTTP probe.
	WebUiEnabled bool
	// TemporalNamespace is the Temporal namespace the proof runs in —
	// the first declared spec.temporal_namespaces entry, or the
	// verifier registers its own.
	TemporalNamespace string
	// StateProof switches on the pod-replacement arm.
	StateProof bool
}

// temporalWebUiEnabled reads spec.web_ui.enabled (default true — the
// proto default; both manifest key forms tolerated).
func temporalWebUiEnabled(spec map[string]interface{}) bool {
	for _, key := range []string{"web_ui", "webUi"} {
		if web, ok := spec[key].(map[string]interface{}); ok {
			if raw, ok := web["enabled"]; ok {
				if enabled, ok := raw.(bool); ok {
					return enabled
				}
			}
		}
	}
	return true
}

// temporalFirstNamespace reads the first declared Temporal namespace
// ("" when the manifest declares none — the verifier registers its own).
func temporalFirstNamespace(spec map[string]interface{}) string {
	for _, key := range []string{"temporal_namespaces", "temporalNamespaces"} {
		if list, ok := spec[key].([]interface{}); ok && len(list) > 0 {
			if entry, ok := list[0].(map[string]interface{}); ok {
				if name, ok := entry["name"].(string); ok {
					return name
				}
			}
		}
	}
	return ""
}

// temporalProofWorkflow is the verifier-owned workflow definition — one
// activity-free step returning a marker the starter asserts on.
func temporalProofWorkflow(_ temporalworkflow.Context) (string, error) {
	return "e2e-proof-ok", nil
}

func (v *TemporalVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] temporal %q in namespace %q\n", v.Name, v.Namespace)

	// All four server services roll out (the first wait absorbs the
	// schema-job + cold-start budget; the rest are already settling).
	for _, service := range []string{"frontend", "history", "matching", "worker"} {
		if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/"+v.Name+"-"+service, v.Namespace, 15*time.Minute); err != nil {
			return errors.Wrapf(err, "the %s deployment never rolled out", service)
		}
	}

	if v.WebUiEnabled {
		if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/"+v.Name+"-web", v.Namespace, 10*time.Minute); err != nil {
			return errors.Wrap(err, "the web UI deployment never rolled out")
		}
		if err := v.proveWebUi(ctx, kubeconfig); err != nil {
			return err
		}
	}

	// THE ENGINE PROOF — on every lane.
	if err := v.proveWorkflowExecution(ctx, kubeconfig, "e2e-proof"); err != nil {
		return err
	}

	if !v.StateProof {
		return nil
	}

	// ---- the state proof: history + frontend pod replacement ---------------
	// Workflow state lives in the database, never in the pods: after
	// UID-verified replacements, the COMPLETED workflow's history must
	// still describe, and a fresh workflow must run end to end.
	for _, component := range []string{"history", "frontend"} {
		if err := deletePodAwaitReplacement(ctx, kubeconfig, v.Namespace,
			"app.kubernetes.io/instance="+v.Name+",app.kubernetes.io/component="+component, 10*time.Minute); err != nil {
			return errors.Wrapf(err, "the %s pod did not recover after deletion", component)
		}
	}
	if err := v.proveWorkflowSurvived(ctx, kubeconfig, "e2e-proof"); err != nil {
		return err
	}
	if err := v.proveWorkflowExecution(ctx, kubeconfig, "e2e-proof-recovered"); err != nil {
		return errors.Wrap(err, "a workflow started AFTER the pod replacements should still run to completion")
	}
	fmt.Printf("  [verify] STATE: the completed workflow survived history+frontend pod replacement AND a fresh workflow succeeded — state lives in the database\n")
	return nil
}

func (v *TemporalVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	for _, service := range []string{"frontend", "history", "matching", "worker"} {
		if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.Name+"-"+service, v.Namespace); err != nil {
			return err
		}
	}
	fmt.Printf("  [verify] DESTROY: all four temporal server deployments gone (Temporal installs no CRDs — destroy is clean by design)\n")
	return nil
}

// proveWebUi drives an HTTP GET through the UI Service against the UI
// server's OWN api (`GET /api/v1/settings` — the ui-server registers it
// unconditionally and answers its SettingsResponse JSON, whose
// DefaultNamespace key is part of that contract). Asserting the app's
// own API is the honest identity proof: the SPA document at / is an
// opaque asset bundle whose text content (even the product name) is an
// upstream implementation detail — verified live: the 2.52.0 index.html
// carries no "temporal" substring at all.
func (v *TemporalVerifier) proveWebUi(ctx context.Context, kubeconfig string) error {
	const uiPort = "18233"
	cancel, err := startPortForward(ctx, kubeconfig, "svc/"+v.Name+"-web", v.Namespace, uiPort+":8080")
	if err != nil {
		return errors.Wrap(err, "starting port-forward to the web UI")
	}
	defer cancel()

	body, err := httpBearerRoundTrip(ctx, "GET", "http://127.0.0.1:"+uiPort+"/api/v1/settings", "", 3*time.Minute)
	if err != nil {
		return errors.Wrap(err, "the web UI settings-API round-trip failed")
	}
	if !strings.Contains(body, "DefaultNamespace") {
		return errors.Errorf("the web UI settings API answered without the SettingsResponse contract: %s", firstLines(body, 2))
	}
	fmt.Printf("  [verify] WEB UI: the UI server's settings API answered its SettingsResponse contract\n")
	return nil
}

// temporalDial port-forwards the frontend and dials a Temporal SDK
// client against it. The caller owns both returned closers.
func (v *TemporalVerifier) temporalDial(ctx context.Context, kubeconfig, temporalNamespace string) (temporalclient.Client, func(), error) {
	const grpcPort = "17233"
	cancelForward, err := startPortForward(ctx, kubeconfig, "svc/"+v.Name+"-frontend", v.Namespace, grpcPort+":7233")
	if err != nil {
		return nil, nil, errors.Wrap(err, "starting port-forward to the temporal frontend")
	}

	var dialed temporalclient.Client
	deadline := time.Now().Add(3 * time.Minute)
	for {
		dialed, err = temporalclient.Dial(temporalclient.Options{
			HostPort:  "127.0.0.1:" + grpcPort,
			Namespace: temporalNamespace,
		})
		if err == nil || time.Now().After(deadline) {
			break
		}
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		cancelForward()
		return nil, nil, errors.Wrap(err, "dialing the temporal frontend through the port-forward")
	}

	cleanup := func() {
		dialed.Close()
		cancelForward()
	}
	return dialed, cleanup, nil
}

// ensureTemporalNamespace makes sure the proof namespace exists: the
// declared one is created by the chart's namespace Job (idempotent
// describe-or-create — the verifier waits for it); an undeclared proof
// namespace is registered by the verifier itself. Registration is
// asynchronous server-side — both paths poll DESCRIBE until the
// namespace answers.
func (v *TemporalVerifier) ensureTemporalNamespace(ctx context.Context, c temporalclient.Client, temporalNamespace string, register bool) error {
	deadline := time.Now().Add(5 * time.Minute)
	registered := false
	for time.Now().Before(deadline) {
		_, err := c.WorkflowService().DescribeNamespace(ctx, describeNamespaceRequest(temporalNamespace))
		if err == nil {
			return nil
		}
		if register && !registered {
			if _, regErr := c.WorkflowService().RegisterNamespace(ctx, registerNamespaceRequest(temporalNamespace)); regErr == nil ||
				strings.Contains(regErr.Error(), "already") {
				registered = true
			}
		}
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("the temporal namespace %q never became describable", temporalNamespace)
}

// proveWorkflowExecution runs THE ENGINE PROOF: a worker registers the
// verifier-owned workflow on a dedicated task queue and a starter
// executes it to completion, asserting the returned marker.
func (v *TemporalVerifier) proveWorkflowExecution(ctx context.Context, kubeconfig, workflowId string) error {
	temporalNamespace := v.TemporalNamespace
	register := false
	if temporalNamespace == "" {
		temporalNamespace = "e2e-proof"
		register = true
	}

	dialed, cleanup, err := v.temporalDial(ctx, kubeconfig, temporalNamespace)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := v.ensureTemporalNamespace(ctx, dialed, temporalNamespace, register); err != nil {
		return err
	}

	const taskQueue = "e2e-proof-tq"
	w := temporalworker.New(dialed, taskQueue, temporalworker.Options{})
	w.RegisterWorkflow(temporalProofWorkflow)
	if err := w.Start(); err != nil {
		return errors.Wrap(err, "starting the proof worker")
	}
	defer w.Stop()

	runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	run, err := dialed.ExecuteWorkflow(runCtx, temporalclient.StartWorkflowOptions{
		ID:        workflowId,
		TaskQueue: taskQueue,
	}, temporalProofWorkflow)
	if err != nil {
		return errors.Wrap(err, "starting the proof workflow")
	}
	fmt.Printf("  [verify] SUBMIT: workflow %q started on task queue %q (namespace %q)\n", workflowId, taskQueue, temporalNamespace)

	var result string
	if err := run.Get(runCtx, &result); err != nil {
		return errors.Wrap(err, "the proof workflow never completed")
	}
	if result != "e2e-proof-ok" {
		return errors.Errorf("the proof workflow returned %q, expected \"e2e-proof-ok\"", result)
	}
	fmt.Printf("  [verify] RUN: workflow %q executed to completion with its result read back — the engine runs workflows\n", workflowId)
	return nil
}

// proveWorkflowSurvived describes an already-completed workflow AFTER pod
// replacements — its history lives in the database, not the pods.
func (v *TemporalVerifier) proveWorkflowSurvived(ctx context.Context, kubeconfig, workflowId string) error {
	temporalNamespace := v.TemporalNamespace
	if temporalNamespace == "" {
		temporalNamespace = "e2e-proof"
	}

	dialed, cleanup, err := v.temporalDial(ctx, kubeconfig, temporalNamespace)
	if err != nil {
		return err
	}
	defer cleanup()

	deadline := time.Now().Add(5 * time.Minute)
	var lastErr error
	for time.Now().Before(deadline) {
		describeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		desc, err := dialed.DescribeWorkflowExecution(describeCtx, workflowId, "")
		cancel()
		if err == nil {
			status := desc.GetWorkflowExecutionInfo().GetStatus().String()
			if strings.EqualFold(status, "Completed") {
				fmt.Printf("  [verify] SURVIVED: workflow %q still describes as Completed after pod replacement\n", workflowId)
				return nil
			}
			lastErr = errors.Errorf("workflow %q describes with status %s, expected Completed", workflowId, status)
		} else {
			lastErr = err
		}
		time.Sleep(5 * time.Second)
	}
	return errors.Wrapf(lastErr, "describing the completed workflow %q after pod replacement", workflowId)
}
