package module

import (
	azuremongoclusteruserv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremongoclusteruser/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMongoClusterUser *azuremongoclusteruserv1alpha1.AzureMongoClusterUser
}

// initializeLocals mirrors the Terraform module's locals. A user grant
// carries no tags (ARM data-plane user entries are untagged -- the
// provider exposes none), so there is no tag map to derive.
func initializeLocals(ctx *pulumi.Context, iacInput *azuremongoclusteruserv1alpha1.AzureMongoClusterUserIacInput) *Locals {
	locals := &Locals{}

	locals.AzureMongoClusterUser = iacInput.Target

	return locals
}
