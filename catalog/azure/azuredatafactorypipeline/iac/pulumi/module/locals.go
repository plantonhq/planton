package module

import (
	azuredatafactorypipelinev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorypipeline/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDataFactoryPipeline *azuredatafactorypipelinev1alpha1.AzureDataFactoryPipeline
}

// A pipeline carries no tags (ARM sub-resources of a factory expose
// none), so there is no tag map to derive.
func initializeLocals(ctx *pulumi.Context, iacInput *azuredatafactorypipelinev1alpha1.AzureDataFactoryPipelineIacInput) *Locals {
	locals := &Locals{}

	locals.AzureDataFactoryPipeline = iacInput.Target

	return locals
}
