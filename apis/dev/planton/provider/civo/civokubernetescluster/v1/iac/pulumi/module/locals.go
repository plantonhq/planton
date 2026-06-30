package module

import (
	civoprovider "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	civokubernetesclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civokubernetescluster/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CivoProviderConfig    *civoprovider.CivoProviderConfig
	CivoKubernetesCluster *civokubernetesclusterv1.CivoKubernetesCluster
}

// initializeLocals copies stack‑input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, stackInput *civokubernetesclusterv1.CivoKubernetesClusterStackInput) *Locals {
	return &Locals{
		CivoProviderConfig:    stackInput.ProviderConfig,
		CivoKubernetesCluster: stackInput.Target,
	}
}
