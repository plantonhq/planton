package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// CRDWorkloadVerifier checks Tier 3 operator-dependent components. These
// create Custom Resources (e.g., Zalando Postgresql, Strimzi Kafka) that an
// operator reconciles into pods and services. Verification checks namespace
// exists, at least one pod Running, and at least one service present. Uses
// the same retry windows as HelmComponentVerifier because CRD reconciliation
// takes comparable time to Helm chart startup.
type CRDWorkloadVerifier struct {
	Namespace     string
	ComponentName string
}

func (v *CRDWorkloadVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] CRD workload %q in namespace %q\n", v.ComponentName, v.Namespace)

	if err := KubectlResourceExists(ctx, kubeconfig, "namespace", v.Namespace, ""); err != nil {
		return errors.Wrapf(err, "namespace %q not found for CRD workload %q", v.Namespace, v.ComponentName)
	}

	if err := KubectlPodsRunningInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no running pods in namespace %q for CRD workload %q", v.Namespace, v.ComponentName)
	}

	if err := KubectlServicesExistInNamespace(ctx, kubeconfig, v.Namespace); err != nil {
		return errors.Wrapf(err, "no services in namespace %q for CRD workload %q", v.Namespace, v.ComponentName)
	}

	return nil
}

func (v *CRDWorkloadVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "namespace", v.Namespace, "")
}
