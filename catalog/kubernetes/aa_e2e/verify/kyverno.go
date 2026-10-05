package verify

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// KyvernoVerifier checks a Kyverno policy engine to the point a customer
// can rely on it: every enabled controller rolled out, the
// runtime-registered webhook configurations present, and THE ENFORCEMENT
// PROOF on every lane — a verifier-owned ClusterPolicy in Enforce mode
// REJECTS a violating Pod at admission and ADMITS a compliant one. An
// admission engine that does not block a violation is not a policy
// engine, whatever its Deployments say.
//
// The behavioral-enforcement scenario (recognized by name) additionally
// proves the MUTATION path: a mutate ClusterPolicy stamps a label onto a
// created Pod and the verifier reads it back.
//
// The proof runs in a verifier-owned TARGET namespace (Kyverno's webhooks
// exclude the engine's own namespace by design, so policing must be
// proven somewhere else) and removes everything it created — policies,
// pods, the target namespace — leaving zero residue.
//
// DESTROY: VerifyAbsent asserts the admission controller Deployment gone
// AND the runtime-registered webhook configurations gone — the chart's
// pre-delete cleanup hook is the designed uninstall path; a stranded
// fail-closed webhook configuration would block matched admissions
// cluster-wide (the class this kind's spec teaches).
type KyvernoVerifier struct {
	Namespace string
	Name      string
	// Enabled controllers (from the scenario manifest; the chart
	// defaults all four on). The admission controller is unconditional.
	BackgroundEnabled bool
	CleanupEnabled    bool
	ReportsEnabled    bool
	// Mutation switches on the behavioral mutation proof
	// (behavioral-enforcement scenario, recognized by name).
	Mutation bool
}

// Chart-truth (verified at the pin): the controller Deployment names
// derive from the CHART name, not the fullname — constant regardless of
// the release name. The webhook configuration names are compiled into
// Kyverno itself (pkg/config).
const (
	kyvernoAdmissionDeployment  = "kyverno-admission-controller"
	kyvernoBackgroundDeployment = "kyverno-background-controller"
	kyvernoCleanupDeployment    = "kyverno-cleanup-controller"
	kyvernoReportsDeployment    = "kyverno-reports-controller"

	kyvernoPolicyWebhookConfig   = "kyverno-policy-validating-webhook-cfg"
	kyvernoResourceWebhookConfig = "kyverno-resource-validating-webhook-cfg"
)

func (v *KyvernoVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] kyverno engine %q in namespace %q\n", v.Name, v.Namespace)

	deployments := []string{kyvernoAdmissionDeployment}
	if v.BackgroundEnabled {
		deployments = append(deployments, kyvernoBackgroundDeployment)
	}
	if v.CleanupEnabled {
		deployments = append(deployments, kyvernoCleanupDeployment)
	}
	if v.ReportsEnabled {
		deployments = append(deployments, kyvernoReportsDeployment)
	}
	for _, deployment := range deployments {
		if err := waitForDeploymentReady(ctx, kubeconfig, v.Namespace, deployment, 10*time.Minute); err != nil {
			return err
		}
	}

	// The engine registers its webhook configurations at RUNTIME — their
	// presence is the proof the admission controller actually came up as
	// a webhook server, not just as a Deployment.
	if err := waitForClusterResource(ctx, kubeconfig, "validatingwebhookconfiguration", kyvernoPolicyWebhookConfig, 3*time.Minute); err != nil {
		return errors.Wrap(err, "kyverno never registered its policy webhook configuration")
	}

	return v.proveEnforcement(ctx, kubeconfig)
}

// VerifyAbsent asserts the engine AND its runtime-registered webhook
// configurations are gone — the destroy-path proof that the pre-delete
// cleanup hook did its job.
func (v *KyvernoVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", kyvernoAdmissionDeployment, v.Namespace); err != nil {
		return err
	}
	for _, webhookConfig := range []string{kyvernoPolicyWebhookConfig, kyvernoResourceWebhookConfig} {
		if err := KubectlResourceAbsent(ctx, kubeconfig, "validatingwebhookconfiguration", webhookConfig, ""); err != nil {
			return errors.Wrapf(err, "webhook configuration %q survived destroy — the pre-delete cleanup hook did not run; matched admissions are blocked cluster-wide until it is deleted by hand", webhookConfig)
		}
	}
	return nil
}

// proveEnforcement is the engine proof: a verifier-owned Enforce policy
// rejects a violating Pod and admits a compliant one (plus the mutation
// proof on the behavioral lane).
func (v *KyvernoVerifier) proveEnforcement(ctx context.Context, kubeconfig string) error {
	targetNamespace := fmt.Sprintf("%s-e2e-target", v.Name)
	policyName := fmt.Sprintf("%s-e2e-require-label", v.Name)
	mutatePolicyName := fmt.Sprintf("%s-e2e-add-label", v.Name)

	// Always sweep the proof artifacts — enforcement proofs that leak
	// policies would police the whole cluster's future lanes.
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		_ = kubectlDeleteResource(cleanupCtx, kubeconfig, "clusterpolicy", policyName, "")
		_ = kubectlDeleteResource(cleanupCtx, kubeconfig, "clusterpolicy", mutatePolicyName, "")
		_ = kubectlDeleteResource(cleanupCtx, kubeconfig, "namespace", targetNamespace, "")
	}()

	if err := kubectlApplyStdin(ctx, kubeconfig, fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, targetNamespace)); err != nil {
		return errors.Wrap(err, "creating the enforcement target namespace")
	}

	// A validation policy scoped to the TARGET namespace only: Pods must
	// carry the e2e.planton.ai/policy label. Enforce + Fail — the posture
	// whose rejection we are here to witness.
	policyManifest := fmt.Sprintf(`apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: %s
spec:
  validationFailureAction: Enforce
  background: false
  rules:
    - name: require-e2e-label
      match:
        any:
          - resources:
              kinds:
                - Pod
              namespaces:
                - %s
      validate:
        message: "e2e enforcement proof: the e2e.planton.ai/policy label is required"
        pattern:
          metadata:
            labels:
              e2e.planton.ai/policy: "?*"
`, policyName, targetNamespace)
	if err := kubectlApplyStdin(ctx, kubeconfig, policyManifest); err != nil {
		return errors.Wrap(err, "applying the enforcement ClusterPolicy")
	}
	if err := waitForClusterPolicyReady(ctx, kubeconfig, policyName, 3*time.Minute); err != nil {
		return err
	}
	// The resource webhook configuration materializes with the first
	// policy — wait for it before attempting the rejection.
	if err := waitForClusterResource(ctx, kubeconfig, "validatingwebhookconfiguration", kyvernoResourceWebhookConfig, 3*time.Minute); err != nil {
		return errors.Wrap(err, "kyverno never registered its resource webhook for the installed policy")
	}

	violatingPod := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: e2e-violating
  namespace: %s
spec:
  containers:
    - name: pause
      image: registry.k8s.io/pause:3.9
`, targetNamespace)
	compliantPod := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: e2e-compliant
  namespace: %s
  labels:
    e2e.planton.ai/policy: "true"
spec:
  containers:
    - name: pause
      image: registry.k8s.io/pause:3.9
`, targetNamespace)

	// THE REJECTION: retry inside a window — webhook rule propagation
	// after policy admission is asynchronous, and an immediately-admitted
	// violating pod within the window is not yet a failure. The proof
	// holds when the API server rejects with the policy's own message.
	if err := v.proveRejection(ctx, kubeconfig, violatingPod, policyName, targetNamespace, 3*time.Minute); err != nil {
		return err
	}
	if err := kubectlApplyStdin(ctx, kubeconfig, compliantPod); err != nil {
		return errors.Wrap(err, "the compliant pod was rejected — enforcement is over-blocking")
	}
	fmt.Printf("  [verify] ENFORCEMENT PROVEN: violating pod REJECTED by %s, compliant pod admitted\n", policyName)

	if v.Mutation {
		if err := v.proveMutation(ctx, kubeconfig, mutatePolicyName, targetNamespace); err != nil {
			return err
		}
	}
	return nil
}

// proveRejection creates the violating pod until the webhook rejects it
// (deleting any admission that slipped through the propagation window).
func (v *KyvernoVerifier) proveRejection(ctx context.Context, kubeconfig, podManifest, policyName, namespace string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastOutput string
	for time.Now().Before(deadline) {
		out, err := kubectlApplyStdinOutput(ctx, kubeconfig, podManifest)
		lastOutput = out
		if err != nil && strings.Contains(out, policyName) {
			return nil
		}
		if err == nil {
			// Admitted during propagation — remove and retry.
			_ = kubectlDeleteResource(ctx, kubeconfig, "pod", "e2e-violating", namespace)
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("the violating pod was never rejected by policy %s (last output: %s)", policyName, firstLines(lastOutput, 3))
}

// proveMutation applies a mutate ClusterPolicy and asserts a fresh pod
// comes back wearing the label the policy stamps.
func (v *KyvernoVerifier) proveMutation(ctx context.Context, kubeconfig, mutatePolicyName, targetNamespace string) error {
	mutatePolicy := fmt.Sprintf(`apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: %s
spec:
  background: false
  rules:
    - name: add-e2e-mutated-label
      match:
        any:
          - resources:
              kinds:
                - Pod
              namespaces:
                - %s
      mutate:
        patchStrategicMerge:
          metadata:
            labels:
              e2e.planton.ai/mutated: "true"
`, mutatePolicyName, targetNamespace)
	if err := kubectlApplyStdin(ctx, kubeconfig, mutatePolicy); err != nil {
		return errors.Wrap(err, "applying the mutation ClusterPolicy")
	}
	if err := waitForClusterPolicyReady(ctx, kubeconfig, mutatePolicyName, 3*time.Minute); err != nil {
		return err
	}

	mutationTarget := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: e2e-mutation-target
  namespace: %s
  labels:
    e2e.planton.ai/policy: "true"
spec:
  containers:
    - name: pause
      image: registry.k8s.io/pause:3.9
`, targetNamespace)

	// Retry the create-and-read loop across the propagation window (the
	// same asynchrony as the rejection proof).
	deadline := time.Now().Add(3 * time.Minute)
	var lastLabel string
	for time.Now().Before(deadline) {
		_ = kubectlDeleteResource(ctx, kubeconfig, "pod", "e2e-mutation-target", targetNamespace)
		if err := kubectlApplyStdin(ctx, kubeconfig, mutationTarget); err == nil {
			label, _ := kubectlGetJSONPath(ctx, kubeconfig, "pod", "e2e-mutation-target", targetNamespace,
				`{.metadata.labels.e2e\.planton\.ai/mutated}`)
			lastLabel = label
			if label == "true" {
				fmt.Printf("  [verify] MUTATION PROVEN: pod created wearing the policy-stamped label\n")
				return nil
			}
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("the mutation policy never stamped its label (last value %q)", lastLabel)
}

// waitForClusterPolicyReady waits for a ClusterPolicy's Ready condition.
func waitForClusterPolicyReady(ctx context.Context, kubeconfig, name string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastStatus string
	for time.Now().Before(deadline) {
		status, _ := kubectlGetJSONPath(ctx, kubeconfig, "clusterpolicy", name, "",
			`{.status.conditions[?(@.type=="Ready")].status}`)
		lastStatus = status
		if status == "True" {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("clusterpolicy %s never reached Ready (last status %q)", name, lastStatus)
}

// waitForClusterResource polls a cluster-scoped resource into existence.
func waitForClusterResource(ctx context.Context, kubeconfig, kind, name string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	for time.Now().Before(deadline) {
		if err := KubectlResourceExists(ctx, kubeconfig, kind, name, ""); err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("%s %q never appeared", kind, name)
}

// waitForDeploymentReady waits until a Deployment's ready replicas match
// its spec.
func waitForDeploymentReady(ctx context.Context, kubeconfig, namespace, name string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastReady, want string
	for time.Now().Before(deadline) {
		want, _ = kubectlGetJSONPath(ctx, kubeconfig, "deployment", name, namespace, "{.spec.replicas}")
		ready, _ := kubectlGetJSONPath(ctx, kubeconfig, "deployment", name, namespace, "{.status.readyReplicas}")
		lastReady = ready
		if want != "" && ready == want {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("deployment %s/%s never reached %s ready replicas (last %q)", namespace, name, want, lastReady)
}

// kubectlApplyStdin applies a manifest passed on stdin.
func kubectlApplyStdin(ctx context.Context, kubeconfig, manifest string) error {
	out, err := kubectlApplyStdinOutput(ctx, kubeconfig, manifest)
	if err != nil {
		return errors.Errorf("kubectl apply: %v: %s", err, firstLines(out, 3))
	}
	return nil
}

// kubectlApplyStdinOutput applies a manifest passed on stdin and returns
// the combined output (callers asserting REJECTION need the message).
func kubectlApplyStdinOutput(ctx context.Context, kubeconfig, manifest string) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// nestedBoolWithDefault reads spec[block][key] as a bool, returning the
// fallback when the block or key is absent (the chart-default posture).
func nestedBoolWithDefault(spec map[string]interface{}, block, key string, fallback bool) bool {
	blockMap, ok := spec[block].(map[string]interface{})
	if !ok {
		return fallback
	}
	value, ok := blockMap[key].(bool)
	if !ok {
		return fallback
	}
	return value
}

// kubectlDeleteResource deletes a resource, tolerating absence.
func kubectlDeleteResource(ctx context.Context, kubeconfig, kind, name, namespace string) error {
	args := []string{"--kubeconfig", kubeconfig, "delete", kind, name, "--ignore-not-found", "--wait=false"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Errorf("kubectl delete %s %s: %v: %s", kind, name, err, firstLines(string(out), 2))
	}
	return nil
}
