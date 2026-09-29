package module

import (
	"strings"

	azurerediscachev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurerediscache/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureRedisCache   *azurerediscachev1alpha1.AzureRedisCache
	ResourceGroupName string
	AzureTags         map[string]string
	// SkuName is ARM's tier value, materialized from the spec enum with
	// the documented STANDARD default (stack inputs never carry proto
	// defaults).
	SkuName string
	// Family is Azure's size-family letter, fully determined by the tier:
	// "C" for Basic/Standard, "P" for Premium -- never spelled twice.
	Family string
}

// skuStrings maps the spec's sku enum to ARM's tier values.
var skuStrings = map[azurerediscachev1alpha1.AzureRedisCacheSku]string{
	azurerediscachev1alpha1.AzureRedisCacheSku_BASIC:    "Basic",
	azurerediscachev1alpha1.AzureRedisCacheSku_STANDARD: "Standard",
	azurerediscachev1alpha1.AzureRedisCacheSku_PREMIUM:  "Premium",
}

// dayOfWeekStrings maps the patch-schedule day enum to ARM's capitalized
// English day names.
var dayOfWeekStrings = map[azurerediscachev1alpha1.AzureRedisCachePatchScheduleDay]string{
	azurerediscachev1alpha1.AzureRedisCachePatchScheduleDay_MONDAY:    "Monday",
	azurerediscachev1alpha1.AzureRedisCachePatchScheduleDay_TUESDAY:   "Tuesday",
	azurerediscachev1alpha1.AzureRedisCachePatchScheduleDay_WEDNESDAY: "Wednesday",
	azurerediscachev1alpha1.AzureRedisCachePatchScheduleDay_THURSDAY:  "Thursday",
	azurerediscachev1alpha1.AzureRedisCachePatchScheduleDay_FRIDAY:    "Friday",
	azurerediscachev1alpha1.AzureRedisCachePatchScheduleDay_SATURDAY:  "Saturday",
	azurerediscachev1alpha1.AzureRedisCachePatchScheduleDay_SUNDAY:    "Sunday",
}

// persistenceAuthStrings maps the persistence auth enum to ARM's values.
var persistenceAuthStrings = map[azurerediscachev1alpha1.AzureRedisCachePersistenceAuthMethod]string{
	azurerediscachev1alpha1.AzureRedisCachePersistenceAuthMethod_SAS:              "SAS",
	azurerediscachev1alpha1.AzureRedisCachePersistenceAuthMethod_MANAGED_IDENTITY: "ManagedIdentity",
}

// identityTypeStrings maps the identity-type enum to ARM's values.
var identityTypeStrings = map[azurerediscachev1alpha1.AzureRedisCacheIdentityType]string{
	azurerediscachev1alpha1.AzureRedisCacheIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurerediscachev1alpha1.AzureRedisCacheIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurerediscachev1alpha1.AzureRedisCacheIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurerediscachev1alpha1.AzureRedisCacheStackInput) *Locals {
	locals := &Locals{}

	locals.AzureRedisCache = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Materialize the tier default: unspecified deploys STANDARD (the
	// spec's documented default -- stack inputs never carry proto
	// defaults), then derive the size-family letter from the tier.
	locals.SkuName = skuStrings[target.Spec.SkuName]
	if locals.SkuName == "" {
		locals.SkuName = "Standard"
	}
	locals.Family = "C"
	if locals.SkuName == "Premium" {
		locals.Family = "P"
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureRedisCache.String()),
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
