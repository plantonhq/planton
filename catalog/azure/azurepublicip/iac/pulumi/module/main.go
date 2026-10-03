package module

import (
	"github.com/pkg/errors"
	azurepublicipv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurepublicip/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurepublicipv1alpha1.AzurePublicIpIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePublicIp.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - reverse_fqdn, ddos settings, idle timeout, and tags update IN
	//   PLACE. Name, SKU/tier, version, zones, prefix membership, ip_tags,
	//   and edge zone are fixed at creation -- changing any of them
	//   replaces the resource and with it the ACTUAL ADDRESS, so treat
	//   replacement as a coordinated migration (DNS, allowlists).
	// - Allocation is always Static: dynamic allocation existed only for
	//   the Basic SKU, whose creation Azure discontinued in 2025, and every
	//   current SKU requires static.
	publicIpArgs := &network.PublicIpArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		AllocationMethod:  pulumi.String("Static"),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Only explicit enum choices are sent, so an unspecified spec and
	// Azure's defaults (Standard / Regional / IPv4 / region-unique label /
	// VirtualNetworkInherited DDoS stance) deploy identically on both
	// engines.
	if locals.Sku != "" {
		publicIpArgs.Sku = pulumi.String(locals.Sku)
	}
	if locals.SkuTier != "" {
		publicIpArgs.SkuTier = pulumi.String(locals.SkuTier)
	}
	if locals.IpVersion != "" {
		publicIpArgs.IpVersion = pulumi.String(locals.IpVersion)
	}
	if len(spec.Zones) > 0 {
		publicIpArgs.Zones = pulumi.ToStringArray(spec.Zones)
	}
	if locals.PublicIpPrefixId != "" {
		publicIpArgs.PublicIpPrefixId = pulumi.String(locals.PublicIpPrefixId)
	}
	if spec.DomainNameLabel != "" {
		publicIpArgs.DomainNameLabel = pulumi.String(spec.DomainNameLabel)
	}
	if locals.DomainNameLabelScope != "" {
		publicIpArgs.DomainNameLabelScope = pulumi.String(locals.DomainNameLabelScope)
	}
	if spec.ReverseFqdn != "" {
		publicIpArgs.ReverseFqdn = pulumi.String(spec.ReverseFqdn)
	}
	if len(spec.IpTags) > 0 {
		publicIpArgs.IpTags = pulumi.ToStringMap(spec.IpTags)
	}
	if locals.DdosProtectionMode != "" {
		publicIpArgs.DdosProtectionMode = pulumi.String(locals.DdosProtectionMode)
	}
	// Only valid alongside the ENABLED mode (spec-level validation enforces
	// the pairing).
	if spec.DdosProtectionPlanId != "" {
		publicIpArgs.DdosProtectionPlanId = pulumi.String(spec.DdosProtectionPlanId)
	}
	if spec.EdgeZone != "" {
		publicIpArgs.EdgeZone = pulumi.String(spec.EdgeZone)
	}

	// Presence-guarded: the getter returns 0 when the optional field is
	// unset, which azurerm's 4-30 validation rejects. An absent spec value
	// falls back to Azure's default (4) -- the same value the Terraform
	// module's optional(number, 4) encodes.
	if spec.IdleTimeoutInMinutes != nil {
		publicIpArgs.IdleTimeoutInMinutes = pulumi.Int(int(spec.GetIdleTimeoutInMinutes()))
	} else {
		publicIpArgs.IdleTimeoutInMinutes = pulumi.Int(4)
	}

	createdPublicIp, err := network.NewPublicIp(ctx,
		spec.Name,
		publicIpArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create public ip %s", spec.Name)
	}

	// Export outputs from the created resource.
	ctx.Export(OpPublicIpId, createdPublicIp.ID())
	ctx.Export(OpIpAddress, createdPublicIp.IpAddress)
	ctx.Export(OpFqdn, createdPublicIp.Fqdn)
	ctx.Export(OpPublicIpName, createdPublicIp.Name)

	return nil
}
