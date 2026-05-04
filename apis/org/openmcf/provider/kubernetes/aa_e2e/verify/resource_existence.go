package verify

import "context"

// ResourceExistenceVerifier checks basic existence/absence of a named resource
// without readiness checks (e.g., secrets, services).
type ResourceExistenceVerifier struct {
	Namespace string
	Kind      string
	Name      string
}

func (v *ResourceExistenceVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	return KubectlResourceExists(ctx, kubeconfig, v.Kind, v.Name, v.Namespace)
}

func (v *ResourceExistenceVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, v.Kind, v.Name, v.Namespace)
}
