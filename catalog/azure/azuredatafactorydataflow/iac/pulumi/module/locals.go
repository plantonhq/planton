package module

import (
	azuredatafactorydataflowv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorydataflow/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDataFactoryDataFlow *azuredatafactorydataflowv1alpha1.AzureDataFactoryDataFlow
}

// A data flow carries no tags (ARM sub-resources of a factory expose
// none), so there is no tag map to derive.
func initializeLocals(ctx *pulumi.Context, iacInput *azuredatafactorydataflowv1alpha1.AzureDataFactoryDataFlowIacInput) *Locals {
	locals := &Locals{}

	locals.AzureDataFactoryDataFlow = iacInput.Target

	return locals
}
