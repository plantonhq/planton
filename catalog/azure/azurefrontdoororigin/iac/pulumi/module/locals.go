package module

import (
	azurefrontdoororiginv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoororigin/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorOrigin *azurefrontdoororiginv1alpha1.AzureFrontDoorOrigin
	OriginGroupId        string
}

// privateLinkTargetTypeStrings maps the target-type enum to the pulumi
// provider's vocabulary. Note the camelCase secondary values: the pulumi
// bridge expects `blobSecondary`/`webSecondary` where the Terraform
// provider spells `blob_secondary`/`web_secondary` -- both engines land
// on the same ARM group id, each speaking its own provider's dialect.
var privateLinkTargetTypeStrings = map[azurefrontdoororiginv1alpha1.AzureFrontDoorOriginPrivateLinkTargetType]string{
	azurefrontdoororiginv1alpha1.AzureFrontDoorOriginPrivateLinkTargetType_SITES:                "sites",
	azurefrontdoororiginv1alpha1.AzureFrontDoorOriginPrivateLinkTargetType_BLOB:                 "blob",
	azurefrontdoororiginv1alpha1.AzureFrontDoorOriginPrivateLinkTargetType_BLOB_SECONDARY:       "blobSecondary",
	azurefrontdoororiginv1alpha1.AzureFrontDoorOriginPrivateLinkTargetType_WEB:                  "web",
	azurefrontdoororiginv1alpha1.AzureFrontDoorOriginPrivateLinkTargetType_WEB_SECONDARY:        "webSecondary",
	azurefrontdoororiginv1alpha1.AzureFrontDoorOriginPrivateLinkTargetType_MANAGED_ENVIRONMENTS: "managedEnvironments",
	azurefrontdoororiginv1alpha1.AzureFrontDoorOriginPrivateLinkTargetType_GATEWAY:              "Gateway",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefrontdoororiginv1alpha1.AzureFrontDoorOriginIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorOrigin = iacInput.Target
	locals.OriginGroupId = iacInput.Target.Spec.OriginGroupId.GetValue()

	// No Azure tags: ARM does not support tags on Front Door origins,
	// so the platform's identity tags live on the profile.

	return locals
}
