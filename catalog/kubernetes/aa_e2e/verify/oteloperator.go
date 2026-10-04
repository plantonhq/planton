package verify

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// otelCrds are the CRDs the module derives from the pinned chart at its
// default values and applies OUTSIDE the release (kept on destroy unless the
// spec says otherwise -- an operator uninstall never deletes the fleet's
// declarations). The feature-gated fifth CRD (clusterobservabilities) is
// absent at default values because its gate is off.
var otelCrds = []string{
	"opentelemetrycollectors.opentelemetry.io",
	"instrumentations.opentelemetry.io",
	"opampbridges.opentelemetry.io",
	"targetallocators.opentelemetry.io",
}

// otelDefaultChartVersion mirrors the spec's default pin. A scenario that
// leaves chartVersion unset installs this version, and the CRDs must carry
// it in their stamp.
const otelDefaultChartVersion = "0.120.0"

// OtelOperatorVerifier checks an OpenTelemetry Operator install to the
// point a KubernetesOtelCollector declaration could be applied against
// it: the manager Deployment rolled out (its pod mounts the
// cert-manager-issued webhook Secret, so a rollout IS the
// cert-issuance proof), all four opentelemetry.io CRDs established and
// stamped with the manifest's chart version -- and the TWO webhook proofs
// that only a live cluster can give:
//
//   - THE ADMISSION GATE: an invalid collector CR (a pipeline
//     referencing an undeclared receiver) is REJECTED by the
//     fail-closed validating webhook.
//   - THE CONVERSION PROOF: a v1beta1 collector CR written with
//     managementState Unmanaged (admitted, never reconciled into a
//     workload) reads back through the v1alpha1 served version -- the
//     conversion webhook call only succeeds when cert-manager's CA
//     injector has patched the MODULE-owned (kept) CRD's conversion
//     caBundle, which is the exact trust seam this kind's design
//     hangs on.
//
// Destroy asserts the CRD posture the manifest declares: kept (the
// default) or deleted (crds.keepOnUninstall: false). The upgrade lane
// re-runs VerifyExists against the upgraded manifest, so the stamp check
// is what proves "a chart bump re-applied the CRDs".
type OtelOperatorVerifier struct {
	// Namespace is the install namespace from the spec.
	Namespace string
	// ReleaseName is metadata.name -- the module pins the chart's
	// fullname to it, so the manager Deployment shares the name.
	ReleaseName string
	// The CRD-lifecycle assertions every kind on the primitive shares.
	helmCRDLifecycle
}

// newOtelOperatorVerifier reads the lifecycle-relevant spec fields so the
// same verifier serves the plain, upgrade, cleanup, and reinstall lanes.
func newOtelOperatorVerifier(manifestPath, releaseName, namespace string) *OtelOperatorVerifier {
	return &OtelOperatorVerifier{
		Namespace:        namespace,
		ReleaseName:      releaseName,
		helmCRDLifecycle: readHelmCRDLifecycle(manifestPath, otelDefaultChartVersion, "opentelemetry-operator", otelCrds),
	}
}

func (v *OtelOperatorVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] opentelemetry-operator %q in namespace %q\n", v.ReleaseName, v.Namespace)

	if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/"+v.ReleaseName, v.Namespace, 5*time.Minute); err != nil {
		return errors.Wrap(err, "the operator deployment never rolled out")
	}
	if err := v.verifyEstablishedAndStamped(ctx, kubeconfig); err != nil {
		return err
	}
	if err := v.admissionGate(ctx, kubeconfig); err != nil {
		return err
	}
	return v.conversionProof(ctx, kubeconfig)
}

// admissionGate proves the fail-closed validating webhook is LIVE: a
// collector CR whose pipeline references an undeclared receiver must be
// REJECTED at apply time (the webhook validates pipeline wiring).
func (v *OtelOperatorVerifier) admissionGate(ctx context.Context, kubeconfig string) error {
	invalid := fmt.Sprintf(`apiVersion: opentelemetry.io/v1beta1
kind: OpenTelemetryCollector
metadata:
  name: e2e-invalid-probe
  namespace: %s
spec:
  managementState: unmanaged
  config:
    exporters:
      debug: {}
    service:
      pipelines:
        traces:
          receivers: [otlp]
          exporters: [debug]
`, v.Namespace)

	out, err := applyManifestString(ctx, kubeconfig, invalid)
	if err == nil {
		// It went through — remove it and fail loudly.
		_, _ = exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"delete", "opentelemetrycollectors.opentelemetry.io", "e2e-invalid-probe",
			"-n", v.Namespace, "--ignore-not-found").CombinedOutput()
		return errors.New("ADMISSION GATE FAILED: a collector CR with a pipeline referencing an undeclared receiver was ADMITTED — the validating webhook is not enforcing")
	}
	fmt.Printf("  [verify] ADMISSION GATE: invalid collector CR REJECTED by the webhook (%s)\n", firstLines(out, 1))
	return nil
}

// conversionProof exercises the CRD's version-conversion webhook through
// the CA-injected trust chain: write v1beta1 (structured config), read
// back v1alpha1 (string config). Unmanaged, so the operator never
// creates a workload for the probe.
func (v *OtelOperatorVerifier) conversionProof(ctx context.Context, kubeconfig string) error {
	probe := fmt.Sprintf(`apiVersion: opentelemetry.io/v1beta1
kind: OpenTelemetryCollector
metadata:
  name: e2e-conversion-probe
  namespace: %s
spec:
  managementState: unmanaged
  config:
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
    exporters:
      debug: {}
    service:
      pipelines:
        traces:
          receivers: [otlp]
          exporters: [debug]
`, v.Namespace)

	if out, err := applyManifestString(ctx, kubeconfig, probe); err != nil {
		return errors.Wrapf(err, "the valid conversion-probe CR was not admitted: %s", firstLines(out, 3))
	}
	defer func() {
		_, _ = exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"delete", "opentelemetrycollectors.opentelemetry.io", "e2e-conversion-probe",
			"-n", v.Namespace, "--ignore-not-found").CombinedOutput()
	}()

	// GET through the v1alpha1 served version — this call round-trips
	// through the conversion webhook (storage is v1beta1), which only
	// works when the kept CRD's conversion caBundle is current.
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"get", "opentelemetrycollectors.v1alpha1.opentelemetry.io", "e2e-conversion-probe",
		"-n", v.Namespace, "-o", "jsonpath={.spec.config}").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "CONVERSION PROOF FAILED — the v1alpha1 read did not convert (is the CA injector patching the kept CRD?): %s", firstLines(string(out), 3))
	}
	// v1alpha1 carries config as a STRING — the converted document must
	// contain the receiver we declared.
	if !strings.Contains(string(out), "otlp") {
		return errors.Errorf("CONVERSION PROOF FAILED — the v1alpha1 config did not carry the declared receiver: %s", firstLines(string(out), 3))
	}
	fmt.Printf("  [verify] CONVERSION PROOF: v1beta1-written CR read back through v1alpha1 (the CA-injected conversion webhook is live)\n")
	return nil
}

func (v *OtelOperatorVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.ReleaseName, v.Namespace); err != nil {
		return err
	}
	return v.verifyDestroyPosture(ctx, kubeconfig)
}

// VerifyExpectedDeployFailure pins a refused deploy or upgrade to the
// primitive's three-part refusal (see helmCRDLifecycle).
func (v *OtelOperatorVerifier) VerifyExpectedDeployFailure(ctx context.Context, kubeconfig, expectation string, deployErr error) error {
	return v.verifyRefusal(ctx, kubeconfig, expectation, deployErr)
}

// applyManifestString kubectl-applies a YAML document from a temp file
// and returns the combined output.
func applyManifestString(ctx context.Context, kubeconfig, manifest string) (string, error) {
	f, err := os.CreateTemp("", "e2e-otel-probe-*.yaml")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.WriteString(manifest); err != nil {
		_ = f.Close()
		return "", err
	}
	_ = f.Close()
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"apply", "-f", filepath.Clean(f.Name())).CombinedOutput()
	return string(out), err
}
