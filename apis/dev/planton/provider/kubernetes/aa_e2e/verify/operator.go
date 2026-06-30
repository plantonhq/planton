package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// OperatorComponentVerifier checks operator/controller components by verifying
// that the namespace exists and at least one Pod is Running. Operators install
// CRD controllers that do not expose Services, so service checks are omitted.
type OperatorComponentVerifier struct {
	Namespace     string
	ComponentName string
}

func (v *OperatorComponentVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] Operator component %q in namespace %q\n", v.ComponentName, v.Namespace)

	if err := KubectlResourceExists(ctx, kubeconfig, "namespace", v.Namespace, ""); err != nil {
		return errors.Wrapf(err, "namespace %q not found for operator component %q", v.Namespace, v.ComponentName)
	}

	if err := KubectlPodsRunningInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no running pods in namespace %q for operator component %q", v.Namespace, v.ComponentName)
	}

	return nil
}

func (v *OperatorComponentVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "namespace", v.Namespace, "")
}
