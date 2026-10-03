package module

import (
	azureeventgridnamespacetopicv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgridnamespacetopic/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureEventgridNamespaceTopic *azureeventgridnamespacetopicv1alpha1.AzureEventgridNamespaceTopic
}

// initializeLocals mirrors the Terraform module's locals. A namespace
// topic carries no tags (the provider exposes none -- it is a pure
// naming-and-retention entry), so there is no tag map to derive.
func initializeLocals(ctx *pulumi.Context, iacInput *azureeventgridnamespacetopicv1alpha1.AzureEventgridNamespaceTopicIacInput) *Locals {
	locals := &Locals{}

	locals.AzureEventgridNamespaceTopic = iacInput.Target

	return locals
}
