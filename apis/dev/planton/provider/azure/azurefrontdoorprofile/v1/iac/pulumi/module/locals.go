package module

import (
	"strings"

	azurefrontdoorprofilev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurefrontdoorprofile/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorProfile *azurefrontdoorprofilev1.AzureFrontDoorProfile
	ResourceGroupName     string
	AzureTags             map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurefrontdoorprofilev1.AzureFrontDoorProfileStackInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorProfile = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Create Azure tags for resource tagging.
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureFrontDoorProfile.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags["environment"] = target.Metadata.Env
	}

	return locals
}
