package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// HelmComponentVerifier checks Helm-based (Tier 2) components by verifying
// that the namespace exists, at least one Pod is Running, and at least one
// Service is present. This avoids coupling to chart-internal resource names.
type HelmComponentVerifier struct {
	Namespace     string
	ComponentName string
}

func (v *HelmComponentVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] Helm component %q in namespace %q\n", v.ComponentName, v.Namespace)

	if err := KubectlResourceExists(ctx, kubeconfig, "namespace", v.Namespace, ""); err != nil {
		return errors.Wrapf(err, "namespace %q not found for helm component %q", v.Namespace, v.ComponentName)
	}

	if err := KubectlPodsRunningInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no running pods in namespace %q for helm component %q", v.Namespace, v.ComponentName)
	}

	if err := KubectlServicesExistInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no services in namespace %q for helm component %q", v.Namespace, v.ComponentName)
	}

	return nil
}

func (v *HelmComponentVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "namespace", v.Namespace, "")
}
