package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// CRDInstallVerifier checks components that install Custom Resource Definitions
// without deploying any pods or services. Verification confirms each expected
// CRD is registered in the API server.
//
// This is the first non-pod verifier type in the framework. It exists because
// components like GatewayAPICRDs apply upstream YAML manifests that only create
// cluster-scoped CRD objects.
type CRDInstallVerifier struct {
	ComponentName string
	CRDNames      []string
}

func (v *CRDInstallVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] CRD install component %q -- checking %d CRDs\n", v.ComponentName, len(v.CRDNames))

	for _, crdName := range v.CRDNames {
		if err := KubectlResourceExists(ctx, kubeconfig, "crd", crdName, ""); err != nil {
			return errors.Wrapf(err, "CRD %q not registered for component %q", crdName, v.ComponentName)
		}
	}

	return nil
}

func (v *CRDInstallVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	for _, crdName := range v.CRDNames {
		if err := KubectlResourceAbsent(ctx, kubeconfig, "crd", crdName, ""); err != nil {
			return errors.Wrapf(err, "CRD %q still present after destroy for component %q", crdName, v.ComponentName)
		}
	}

	return nil
}
