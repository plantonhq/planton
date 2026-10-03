package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaimodelgardendeploymentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaimodelgardendeployment/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig                *gcpprovider.GcpProviderConfig
	GcpVertexAiModelGardenDeployment *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeployment
}

// initializeLocals carries the IaC input through. Google's one-step
// deployment resource carries no labels of its own (the endpoint and model
// it creates are Google-named), so there is no derived name and no label
// set here.
func initializeLocals(_ *pulumi.Context, iacInput *gcpvertexaimodelgardendeploymentv1alpha1.GcpVertexAiModelGardenDeploymentIacInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiModelGardenDeployment = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
