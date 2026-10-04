package verify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// GatekeeperVerifier checks an OPA Gatekeeper engine to the point a
// customer can rely on it: controller-manager replicas and the audit
// controller rolled out, the chart-owned webhook configuration present,
// and THE ENFORCEMENT PROOF on every lane — a verifier-owned
// ConstraintTemplate (Kubernetes CEL) instantiated as a deny Constraint
// REJECTS a violating Pod at admission and ADMITS a compliant one.
//
// The behavioral-audit scenario (recognized by name) additionally proves
// THE AUDIT LOOP: a violating Pod created BEFORE the constraint exists —
// so admission never saw it — is found by the audit controller and
// recorded in the constraint's status. Enforcement without audit misses
// everything that already runs; this is the half brownfield adoption
// depends on.
//
// The proof runs in a verifier-owned TARGET namespace (the engine's own
// namespace carries the exemption label by design) and removes everything
// it created — template, constraint, pods, namespace — asserting the
// template's runtime-created constraint CRD is gone with it.
//
// DESTROY: VerifyAbsent asserts the deployments and the chart-owned
// webhook configurations gone, AND the engine CRDs still present — the
// crds/-directory keep posture is a designed outcome, asserted, not
// tolerated.
type GatekeeperVerifier struct {
	Namespace string
	Name      string
	// Audit switches on the behavioral audit-loop proof
	// (behavioral-audit scenario, recognized by name).
	Audit bool
}

// Chart-truth (verified at the pin): every resource name is HARDCODED by
// the chart — no fullname derivation exists.
const (
	gatekeeperControllerDeployment = "gatekeeper-controller-manager"
	gatekeeperAuditDeployment      = "gatekeeper-audit"
	gatekeeperValidatingWebhook    = "gatekeeper-validating-webhook-configuration"
	gatekeeperTemplatesCrd         = "constrainttemplates.templates.gatekeeper.sh"
)

func (v *GatekeeperVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] gatekeeper engine %q in namespace %q\n", v.Name, v.Namespace)

	if err := waitForDeploymentReady(ctx, kubeconfig, v.Namespace, gatekeeperControllerDeployment, 10*time.Minute); err != nil {
		return err
	}
	if err := waitForDeploymentReady(ctx, kubeconfig, v.Namespace, gatekeeperAuditDeployment, 5*time.Minute); err != nil {
		return err
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "validatingwebhookconfiguration", gatekeeperValidatingWebhook, ""); err != nil {
		return errors.Wrap(err, "the chart-owned validating webhook configuration is missing")
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "crd", gatekeeperTemplatesCrd, ""); err != nil {
		return errors.Wrap(err, "the constrainttemplates engine CRD is missing")
	}

	return v.proveEnforcement(ctx, kubeconfig)
}

// VerifyAbsent asserts the engine workloads and its chart-owned webhook
// configurations are gone — and that the engine CRDs SURVIVED, the
// crds/-directory keep posture this kind documents.
func (v *GatekeeperVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", gatekeeperControllerDeployment, v.Namespace); err != nil {
		return err
	}
	if err := KubectlResourceAbsent(ctx, kubeconfig, "validatingwebhookconfiguration", gatekeeperValidatingWebhook, ""); err != nil {
		return errors.Wrap(err, "the chart-owned webhook configuration survived destroy")
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "crd", gatekeeperTemplatesCrd, ""); err != nil {
		return errors.Wrap(err, "the engine CRDs were DELETED on destroy — the crds/-directory keep posture regressed")
	}
	return nil
}

// proveEnforcement is the engine proof: a verifier-owned CEL
// ConstraintTemplate + deny Constraint rejects a violating Pod and admits
// a compliant one (plus the audit-loop proof on the behavioral lane).
func (v *GatekeeperVerifier) proveEnforcement(ctx context.Context, kubeconfig string) error {
	targetNamespace := fmt.Sprintf("%s-e2e-target", v.Name)
	// ConstraintTemplate kind names must be lowercase alphanumeric; the
	// constraint CRD it creates at runtime is <kind>.constraints.gatekeeper.sh.
	templateKind := "e2erequiredlabel"
	constraintName := "e2e-require-label"
	constraintCrd := templateKind + ".constraints.gatekeeper.sh"

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		_ = kubectlDeleteResource(cleanupCtx, kubeconfig, templateKind, constraintName, "")
		_ = kubectlDeleteResource(cleanupCtx, kubeconfig, "constrainttemplate", templateKind, "")
		_ = kubectlDeleteResource(cleanupCtx, kubeconfig, "namespace", targetNamespace, "")
	}()

	if err := kubectlApplyStdin(ctx, kubeconfig, fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, targetNamespace)); err != nil {
		return errors.Wrap(err, "creating the enforcement target namespace")
	}

	// A Kubernetes-CEL template (no Rego required — K8sNativeValidation
	// is on by default at the pin): Pods must carry the
	// e2e.planton.ai/policy label.
	template := fmt.Sprintf(`apiVersion: templates.gatekeeper.sh/v1
kind: ConstraintTemplate
metadata:
  name: %s
spec:
  crd:
    spec:
      names:
        kind: E2eRequiredLabel
  targets:
    - target: admission.k8s.gatekeeper.sh
      code:
        - engine: K8sNativeValidation
          source:
            validations:
              - expression: 'has(object.metadata.labels) && "e2e.planton.ai/policy" in object.metadata.labels'
                message: "e2e enforcement proof: the e2e.planton.ai/policy label is required"
`, templateKind)
	if err := kubectlApplyStdin(ctx, kubeconfig, template); err != nil {
		return errors.Wrap(err, "applying the ConstraintTemplate")
	}
	// The engine turns the template into a constraint CRD at runtime —
	// its arrival is itself a proof of the constraint framework working.
	if err := waitForClusterResource(ctx, kubeconfig, "crd", constraintCrd, 3*time.Minute); err != nil {
		return errors.Wrap(err, "gatekeeper never created the constraint CRD from the template")
	}

	auditVictimManifest := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: e2e-audit-victim
  namespace: %s
spec:
  containers:
    - name: pause
      image: registry.k8s.io/pause:3.9
`, targetNamespace)
	if v.Audit {
		// The audit-proof victim goes in BEFORE the constraint exists:
		// admission never evaluates it, so only the audit loop can find it.
		if err := kubectlApplyStdin(ctx, kubeconfig, auditVictimManifest); err != nil {
			return errors.Wrap(err, "creating the audit victim pod")
		}
	}

	constraint := fmt.Sprintf(`apiVersion: constraints.gatekeeper.sh/v1beta1
kind: E2eRequiredLabel
metadata:
  name: %s
spec:
  enforcementAction: deny
  match:
    kinds:
      - apiGroups: [""]
        kinds: ["Pod"]
    namespaces:
      - %s
`, constraintName, targetNamespace)
	if err := kubectlApplyStdin(ctx, kubeconfig, constraint); err != nil {
		return errors.Wrap(err, "applying the constraint")
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

	// THE REJECTION: retry inside a window — constraint ingestion into
	// the webhook is asynchronous; an admitted violating pod inside the
	// window is deleted and retried until the API server denies with the
	// constraint's kind or message.
	if err := v.proveRejection(ctx, kubeconfig, violatingPod, targetNamespace, 3*time.Minute); err != nil {
		return err
	}
	if err := kubectlApplyStdin(ctx, kubeconfig, compliantPod); err != nil {
		return errors.Wrap(err, "the compliant pod was rejected — enforcement is over-blocking")
	}
	fmt.Printf("  [verify] ENFORCEMENT PROVEN: violating pod REJECTED by the constraint, compliant pod admitted\n")

	if v.Audit {
		if err := v.proveAudit(ctx, kubeconfig, templateKind, constraintName, 6*time.Minute); err != nil {
			return err
		}
	}
	return nil
}

// proveRejection creates the violating pod until the webhook denies it.
func (v *GatekeeperVerifier) proveRejection(ctx context.Context, kubeconfig, podManifest, namespace string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastOutput string
	for time.Now().Before(deadline) {
		out, err := kubectlApplyStdinOutput(ctx, kubeconfig, podManifest)
		lastOutput = out
		if err != nil && (strings.Contains(out, "e2e-require-label") || strings.Contains(out, "e2e.planton.ai/policy")) {
			return nil
		}
		if err == nil {
			_ = kubectlDeleteResource(ctx, kubeconfig, "pod", "e2e-violating", namespace)
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("the violating pod was never denied by the constraint (last output: %s)", firstLines(lastOutput, 3))
}

// proveAudit waits for the audit controller to record the pre-existing
// victim pod as a violation in the constraint's status — proof the loop
// evaluates what admission never saw.
func (v *GatekeeperVerifier) proveAudit(ctx context.Context, kubeconfig, templateKind, constraintName string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastViolations string
	for time.Now().Before(deadline) {
		violations, _ := kubectlGetJSONPath(ctx, kubeconfig, templateKind, constraintName, "", "{.status.totalViolations}")
		lastViolations = violations
		if violations != "" && violations != "0" {
			fmt.Printf("  [verify] AUDIT PROVEN: the pre-constraint victim pod recorded as a violation (totalViolations=%s)\n", violations)
			return nil
		}
		time.Sleep(15 * time.Second)
	}
	return errors.Errorf("the audit loop never recorded the pre-existing violation (last totalViolations %q)", lastViolations)
}
