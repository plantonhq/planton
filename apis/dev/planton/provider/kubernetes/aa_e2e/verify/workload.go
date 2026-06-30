package verify

import "context"

// WorkloadVerifier checks that a workload (deployment, statefulset, etc.) exists or is absent.
type WorkloadVerifier struct {
	Namespace string
	Kind      string
	Name      string
}

func (v *WorkloadVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	return KubectlResourceExists(ctx, kubeconfig, v.Kind, v.Name, v.Namespace)
}

func (v *WorkloadVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, v.Kind, v.Name, v.Namespace)
}
