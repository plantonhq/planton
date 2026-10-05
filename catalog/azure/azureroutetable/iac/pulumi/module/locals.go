package module

import (
	"strings"

	azureroutetablev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureroutetable/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nextHopTypeToArm maps the spec enum's values to ARM's RouteNextHopType
// strings. The spec enum requires a defined, non-unspecified value, so
// every route resolves to exactly one ARM hop type.
var nextHopTypeToArm = map[azureroutetablev1alpha1.AzureRouteTableNextHopType]string{
	azureroutetablev1alpha1.AzureRouteTableNextHopType_VIRTUAL_NETWORK_GATEWAY: "VirtualNetworkGateway",
	azureroutetablev1alpha1.AzureRouteTableNextHopType_VNET_LOCAL:              "VnetLocal",
	azureroutetablev1alpha1.AzureRouteTableNextHopType_INTERNET:                "Internet",
	azureroutetablev1alpha1.AzureRouteTableNextHopType_VIRTUAL_APPLIANCE:       "VirtualAppliance",
	azureroutetablev1alpha1.AzureRouteTableNextHopType_NONE:                    "None",
}

type Locals struct {
	AzureRouteTable *azureroutetablev1alpha1.AzureRouteTable

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal.
	ResourceGroupName string

	// BgpRoutePropagationEnabled carries Azure's default (true) when the
	// spec leaves it unset, so both engines send the same effective value.
	BgpRoutePropagationEnabled bool

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureroutetablev1alpha1.AzureRouteTableIacInput) *Locals {
	locals := &Locals{}

	locals.AzureRouteTable = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Azure's default is to propagate BGP-learned routes; only an explicit
	// false (the forced-tunneling hardening) turns it off.
	locals.BgpRoutePropagationEnabled = true
	if target.Spec.BgpRoutePropagationEnabled != nil {
		locals.BgpRoutePropagationEnabled = *target.Spec.BgpRoutePropagationEnabled
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureRouteTable.String()),
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
