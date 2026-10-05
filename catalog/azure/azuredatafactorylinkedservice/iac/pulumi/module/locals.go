package module

import (
	azuredatafactorylinkedservicev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorylinkedservice/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDataFactoryLinkedService *azuredatafactorylinkedservicev1alpha1.AzureDataFactoryLinkedService
}

// A linked service carries no tags (ARM sub-resources of a factory
// expose none), so there is no tag map to derive.
func initializeLocals(ctx *pulumi.Context, iacInput *azuredatafactorylinkedservicev1alpha1.AzureDataFactoryLinkedServiceIacInput) *Locals {
	locals := &Locals{}

	locals.AzureDataFactoryLinkedService = iacInput.Target

	return locals
}
