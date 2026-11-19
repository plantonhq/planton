package verify

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

// WorkloadVerifier checks that a workload (deployment, statefulset, etc.) exists or is absent.
type WorkloadVerifier struct {
	Namespace string
	Kind      string
	Name      string
	// SizeLane is the replica-ownership lane the scenario runs, if any
	// (workloadsize.go).
	SizeLane workloadSizeLane
}

// newWorkloadVerifier builds the verifier for a workload scenario.
func newWorkloadVerifier(manifestPath string, info *ManifestInfo, kind string) *WorkloadVerifier {
	return &WorkloadVerifier{
		Namespace: info.Namespace,
		Kind:      kind,
		Name:      info.Name,
		SizeLane:  workloadSizeLaneFor(manifestPath),
	}
}

func (v *WorkloadVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceExists(ctx, kubeconfig, v.Kind, v.Name, v.Namespace); err != nil {
		return err
	}
	if v.SizeLane != "" {
		return v.proveSize(ctx, kubeconfig)
	}
	return nil
}

func (v *WorkloadVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, v.Kind, v.Name, v.Namespace)
}

// failureEngineUnsupportedVolumeSource is a workload refused at plan because
// its volume mounts carry a source the OpenTofu engine cannot write (the
// clusterTrustBundle and podCertificate projections): the refusal names the
// engine to use, and nothing reaches the cluster.
const failureEngineUnsupportedVolumeSource = "engine-unsupported-volume-source"

// VerifyExpectedDeployFailure pins the one designed workload refusal: the
// plan explains itself (the unsupported sources and the engine that writes
// them) and the workload was never created.
func (v *WorkloadVerifier) VerifyExpectedDeployFailure(ctx context.Context, kubeconfig, expectation string, deployErr error) error {
	if expectation != failureEngineUnsupportedVolumeSource {
		return errors.Errorf("workload verifier pins no deploy failure %q", expectation)
	}
	if deployErr == nil {
		return errors.New("expected the plan to refuse the workload's volume sources, but the deploy succeeded")
	}
	// OpenTofu wraps its error text across lines; compare it as one line.
	text := strings.Join(strings.Fields(deployErr.Error()), " ")
	for _, want := range []string{"not supported on the OpenTofu engine", "planton.dev/provisioner to pulumi"} {
		if !strings.Contains(text, want) {
			return errors.Errorf("the refusal does not say %q -- got: %s", want, text)
		}
	}
	return v.VerifyAbsent(ctx, kubeconfig)
}
