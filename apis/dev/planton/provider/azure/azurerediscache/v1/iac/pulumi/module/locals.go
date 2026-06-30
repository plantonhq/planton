package module

import (
	"strings"

	azurerediscachev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurerediscache/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureRedisCache   *azurerediscachev1.AzureRedisCache
	ResourceGroupName string
	AzureTags         map[string]string
	// Family is auto-derived from sku_name: "C" for Basic/Standard, "P" for Premium.
	Family string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurerediscachev1.AzureRedisCacheStackInput) *Locals {
	locals := &Locals{}

	locals.AzureRedisCache = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Auto-derive the SKU family from the tier name.
	// Azure requires family "C" (C0-C6) for Basic/Standard and "P" (P1-P5) for Premium.
	skuName := target.Spec.GetSkuName()
	if skuName == "Premium" {
		locals.Family = "P"
	} else {
		locals.Family = "C"
	}

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureRedisCache.String()),
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
