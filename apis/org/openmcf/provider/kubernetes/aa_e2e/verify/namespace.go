package verify

import "context"

// NamespaceVerifier checks that a namespace exists or is absent.
type NamespaceVerifier struct {
	Name string
}

func (v *NamespaceVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	return KubectlResourceExists(ctx, kubeconfig, "namespace", v.Name, "")
}

func (v *NamespaceVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "namespace", v.Name, "")
}
