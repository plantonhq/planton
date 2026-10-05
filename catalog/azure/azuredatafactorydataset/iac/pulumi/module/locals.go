package module

import (
	azuredatafactorydatasetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorydataset/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDataFactoryDataset *azuredatafactorydatasetv1alpha1.AzureDataFactoryDataset
}

// A dataset carries no tags (ARM sub-resources of a factory expose
// none), so there is no tag map to derive.
func initializeLocals(ctx *pulumi.Context, iacInput *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetIacInput) *Locals {
	locals := &Locals{}

	locals.AzureDataFactoryDataset = iacInput.Target

	return locals
}
