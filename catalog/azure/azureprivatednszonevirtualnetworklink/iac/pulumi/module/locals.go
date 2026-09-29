package module

import (
	"strings"

	"github.com/pkg/errors"
	azureprivatednszonevirtualnetworklinkv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednszonevirtualnetworklink/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzurePrivateDnsZoneVirtualNetworkLink *azureprivatednszonevirtualnetworklinkv1alpha1.AzurePrivateDnsZoneVirtualNetworkLink

	// PrivateDnsZoneId and VirtualNetworkId are StringValueOrRef fields; the
	// platform middleware resolves valueFrom references before IaC modules
	// run, so GetValue() always returns the resolved literal ARM ID.
	PrivateDnsZoneId string
	VirtualNetworkId string

	// ResolutionPolicy is the ARM string for the spec's enum
	// ("Default"/"NxDomainRedirect"), or empty when unspecified so Azure
	// applies its own per-zone-type default.
	ResolutionPolicy string

	// RegistrationEnabled carries Azure's default (false) when the spec
	// leaves it unset, so both engines send the same effective value.
	RegistrationEnabled bool

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureprivatednszonevirtualnetworklinkv1alpha1.AzurePrivateDnsZoneVirtualNetworkLinkStackInput) *Locals {
	locals := &Locals{}

	locals.AzurePrivateDnsZoneVirtualNetworkLink = stackInput.Target
	target := stackInput.Target

	locals.PrivateDnsZoneId = target.Spec.PrivateDnsZoneId.GetValue()
	locals.VirtualNetworkId = target.Spec.VirtualNetworkId.GetValue()

	switch target.Spec.ResolutionPolicy {
	case azureprivatednszonevirtualnetworklinkv1alpha1.AzurePrivateDnsZoneVirtualNetworkLinkResolutionPolicy_DEFAULT:
		locals.ResolutionPolicy = "Default"
	case azureprivatednszonevirtualnetworklinkv1alpha1.AzurePrivateDnsZoneVirtualNetworkLinkResolutionPolicy_NX_DOMAIN_REDIRECT:
		locals.ResolutionPolicy = "NxDomainRedirect"
	}

	if target.Spec.RegistrationEnabled != nil {
		locals.RegistrationEnabled = *target.Spec.RegistrationEnabled
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzurePrivateDnsZoneVirtualNetworkLink.String()),
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

// parseZoneId extracts the resource-group name and zone name embedded in a
// private DNS zone's ARM ID
// (/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/privateDnsZones/{zone}).
// The link is an ARM child of the zone, but the provider SDK takes the zone
// name and resource group as separate arguments -- the module parses rather
// than asking the user to restate derivable state that could then disagree
// with the referenced zone.
func parseZoneId(zoneId string) (resourceGroupName string, zoneName string, err error) {
	segments := strings.Split(strings.Trim(zoneId, "/"), "/")
	for i := 0; i < len(segments)-1; i++ {
		if strings.EqualFold(segments[i], "resourceGroups") {
			resourceGroupName = segments[i+1]
		}
		if strings.EqualFold(segments[i], "privateDnsZones") {
			zoneName = segments[i+1]
		}
	}
	if resourceGroupName == "" || zoneName == "" {
		return "", "", errors.Errorf(
			"private_dns_zone_id %q is not a full private-DNS-zone ARM ID "+
				"(expected /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/privateDnsZones/{zone})",
			zoneId)
	}
	return resourceGroupName, zoneName, nil
}
