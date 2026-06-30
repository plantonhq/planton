package module

import (
	"strings"

	azurednszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurednszone/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDnsZone      *azurednszonev1.AzureDnsZone
	ResourceGroupName string
	AzureTags         map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurednszonev1.AzureDnsZoneStackInput) *Locals {
	locals := &Locals{}

	locals.AzureDnsZone = stackInput.Target

	target := stackInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware resolves
	// valueFrom references before IaC modules run, so .GetValue() always returns the
	// resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureDnsZone.String()),
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
