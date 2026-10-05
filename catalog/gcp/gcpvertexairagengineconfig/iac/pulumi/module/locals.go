package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexairagengineconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexairagengineconfig/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig          *gcpprovider.GcpProviderConfig
	GcpVertexAiRagEngineConfig *gcpvertexairagengineconfigv1alpha1.GcpVertexAiRagEngineConfig
}

// initializeLocals carries the IaC input through; the RAG Engine
// configuration is a per-location singleton Google names, so there is no
// derived name and no label set (the provider resource carries no labels).
func initializeLocals(_ *pulumi.Context, iacInput *gcpvertexairagengineconfigv1alpha1.GcpVertexAiRagEngineConfigIacInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiRagEngineConfig = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
