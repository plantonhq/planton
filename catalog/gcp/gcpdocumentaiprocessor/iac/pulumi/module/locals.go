package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdocumentaiprocessorv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdocumentaiprocessor/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig      *gcpprovider.GcpProviderConfig
	GcpDocumentAiProcessor *gcpdocumentaiprocessorv1alpha1.GcpDocumentAiProcessor
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpdocumentaiprocessorv1alpha1.GcpDocumentAiProcessorStackInput) *Locals {
	locals := &Locals{}
	locals.GcpDocumentAiProcessor = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
