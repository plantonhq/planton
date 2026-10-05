package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaideployedindexv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaideployedindex/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig        *gcpprovider.GcpProviderConfig
	GcpVertexAiDeployedIndex *gcpvertexaideployedindexv1alpha1.GcpVertexAiDeployedIndex
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpvertexaideployedindexv1alpha1.GcpVertexAiDeployedIndexIacInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiDeployedIndex = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig

	// This resource class carries NO labels and NO project field in the
	// GCP API — the deployment lives inside the index endpoint resource
	// and inherits its project — so there is no label merge here:
	// platform attribution is impossible on a DeployedIndex and none is
	// faked.

	return locals
}
