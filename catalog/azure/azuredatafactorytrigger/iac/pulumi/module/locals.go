package module

import (
	azuredatafactorytriggerv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorytrigger/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDataFactoryTrigger *azuredatafactorytriggerv1alpha1.AzureDataFactoryTrigger
}

// A trigger carries no tags (ARM sub-resources of a factory expose
// none), so there is no tag map to derive.
func initializeLocals(ctx *pulumi.Context, iacInput *azuredatafactorytriggerv1alpha1.AzureDataFactoryTriggerIacInput) *Locals {
	locals := &Locals{}

	locals.AzureDataFactoryTrigger = iacInput.Target

	return locals
}
