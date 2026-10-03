package module

import (
	azurevirtualhubconnectionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualhubconnection/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The hub connection carries no tags and no resource group of its own:
// ARM addresses it as a child of the hub, and the provider's schema has
// no tags argument -- so these locals skip the family's usual
// tag-merging machinery.
type Locals struct {
	AzureVirtualHubConnection *azurevirtualhubconnectionv1alpha1.AzureVirtualHubConnection
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurevirtualhubconnectionv1alpha1.AzureVirtualHubConnectionIacInput) *Locals {
	locals := &Locals{}

	locals.AzureVirtualHubConnection = iacInput.Target

	return locals
}

// overrideCriteriaWireValue maps the spec's optional override-criteria
// enum onto ARM's vocabulary, applying ARM's default (Contains) when
// the field is unset -- mirroring the Terraform module's null handling.
// ARM fixes the criteria once the connection is created.
func overrideCriteriaWireValue(criteria *azurevirtualhubconnectionv1alpha1.AzureVirtualHubConnectionStaticVnetLocalRouteOverrideCriteria) string {
	if criteria == nil {
		return "Contains"
	}
	if *criteria == azurevirtualhubconnectionv1alpha1.AzureVirtualHubConnectionStaticVnetLocalRouteOverrideCriteria_EQUAL {
		return "Equal"
	}
	return "Contains"
}

// optionalBool returns the pointed-to value, or the default when the
// optional field is unset -- mirroring the Terraform variable default.
func optionalBool(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}
