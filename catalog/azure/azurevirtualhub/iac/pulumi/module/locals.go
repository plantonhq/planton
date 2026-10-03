package module

import (
	"strings"

	azurevirtualhubv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualhub/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVirtualHub *azurevirtualhubv1alpha1.AzureVirtualHub

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurevirtualhubv1alpha1.AzureVirtualHubIacInput) *Locals {
	locals := &Locals{}

	locals.AzureVirtualHub = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureVirtualHub.String()),
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

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}

// skuWireValue maps the spec's optional sku enum onto ARM's vocabulary,
// applying ARM's default (Standard) when the field is unset -- mirroring
// the Terraform module's null handling.
func skuWireValue(sku *azurevirtualhubv1alpha1.AzureVirtualHubSku) string {
	if sku == nil {
		return "Standard"
	}
	switch *sku {
	case azurevirtualhubv1alpha1.AzureVirtualHubSku_BASIC:
		return "Basic"
	default:
		return "Standard"
	}
}

// hubRoutingPreferenceWireValue maps the spec's optional routing
// preference enum onto ARM's vocabulary, applying ARM's default
// (ExpressRoute) when the field is unset.
func hubRoutingPreferenceWireValue(preference *azurevirtualhubv1alpha1.AzureVirtualHubRoutingPreference) string {
	if preference == nil {
		return "ExpressRoute"
	}
	switch *preference {
	case azurevirtualhubv1alpha1.AzureVirtualHubRoutingPreference_VPN_GATEWAY:
		return "VpnGateway"
	case azurevirtualhubv1alpha1.AzureVirtualHubRoutingPreference_AS_PATH:
		return "ASPath"
	default:
		return "ExpressRoute"
	}
}

// destinationsTypeWireValue maps the route destinations-type enum onto
// ARM's vocabulary (the spec requires an explicit value).
func destinationsTypeWireValue(destinationsType azurevirtualhubv1alpha1.AzureVirtualHubRouteDestinationsType) string {
	switch destinationsType {
	case azurevirtualhubv1alpha1.AzureVirtualHubRouteDestinationsType_RESOURCE_ID:
		return "ResourceId"
	case azurevirtualhubv1alpha1.AzureVirtualHubRouteDestinationsType_SERVICE:
		return "Service"
	default:
		return "CIDR"
	}
}

// matchConditionWireValue maps the route-map match-condition enum onto
// ARM's vocabulary (the spec requires an explicit value).
func matchConditionWireValue(condition azurevirtualhubv1alpha1.AzureVirtualHubRouteMapMatchCondition) string {
	switch condition {
	case azurevirtualhubv1alpha1.AzureVirtualHubRouteMapMatchCondition_EQUALS:
		return "Equals"
	case azurevirtualhubv1alpha1.AzureVirtualHubRouteMapMatchCondition_NOT_CONTAINS:
		return "NotContains"
	case azurevirtualhubv1alpha1.AzureVirtualHubRouteMapMatchCondition_NOT_EQUALS:
		return "NotEquals"
	default:
		return "Contains"
	}
}

// actionTypeWireValue maps the route-map action-type enum onto ARM's
// vocabulary (the spec requires an explicit value).
func actionTypeWireValue(actionType azurevirtualhubv1alpha1.AzureVirtualHubRouteMapActionType) string {
	switch actionType {
	case azurevirtualhubv1alpha1.AzureVirtualHubRouteMapActionType_DROP:
		return "Drop"
	case azurevirtualhubv1alpha1.AzureVirtualHubRouteMapActionType_REMOVE:
		return "Remove"
	case azurevirtualhubv1alpha1.AzureVirtualHubRouteMapActionType_REPLACE:
		return "Replace"
	default:
		return "Add"
	}
}

// routingPolicyDestinationWireValue maps the routing-intent destination
// enum onto ARM's vocabulary (the spec requires an explicit value).
func routingPolicyDestinationWireValue(destination azurevirtualhubv1alpha1.AzureVirtualHubRoutingPolicyDestination) string {
	if destination == azurevirtualhubv1alpha1.AzureVirtualHubRoutingPolicyDestination_PRIVATE_TRAFFIC {
		return "PrivateTraffic"
	}
	return "Internet"
}
