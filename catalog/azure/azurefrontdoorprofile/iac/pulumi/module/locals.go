package module

import (
	"strings"

	azurefrontdoorprofilev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorProfile *azurefrontdoorprofilev1alpha1.AzureFrontDoorProfile
	ResourceGroupName     string
	AzureTags             map[string]string
	// SkuName is ARM's tier value, materialized from the spec enum with
	// the documented STANDARD default (IaC inputs never carry proto
	// defaults).
	SkuName string
}

// skuStrings maps the spec's sku enum to ARM's tier values.
var skuStrings = map[azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileSku]string{
	azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileSku_STANDARD: "Standard_AzureFrontDoor",
	azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileSku_PREMIUM:  "Premium_AzureFrontDoor",
}

// identityTypeStrings maps the identity-type enum to ARM's values.
var identityTypeStrings = map[azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileIdentityType]string{
	azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

// logScrubbingVariableStrings maps the log-scrubbing enum to ARM's
// match-variable values.
var logScrubbingVariableStrings = map[azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileLogScrubbingVariable]string{
	azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileLogScrubbingVariable_QUERY_STRING_ARG_NAMES: "QueryStringArgNames",
	azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileLogScrubbingVariable_REQUEST_IP_ADDRESS:     "RequestIPAddress",
	azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileLogScrubbingVariable_REQUEST_URI:            "RequestUri",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefrontdoorprofilev1alpha1.AzureFrontDoorProfileIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorProfile = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Materialize the tier default: unspecified deploys STANDARD (the
	// spec's documented default -- IaC inputs never carry proto
	// defaults).
	locals.SkuName = skuStrings[target.Spec.Sku]
	if locals.SkuName == "" {
		locals.SkuName = "Standard_AzureFrontDoor"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureFrontDoorProfile.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
