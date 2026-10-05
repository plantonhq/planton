package module

import (
	"github.com/pkg/errors"
	azurepublicipprefixv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurepublicipprefix/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurepublicipprefixv1alpha1.AzurePublicIpPrefixIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePublicIpPrefix.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - The prefix is essentially immutable: everything except tags is
	//   fixed at creation, and replacing it changes the ACTUAL reserved
	//   range -- everything partners have allowlisted. Treat replacement as
	//   a coordinated migration, never a casual update.
	// - The prefix cannot be deleted while any of its addresses are in use
	//   by public IPs or NAT gateway associations.
	prefixArgs := &network.PublicIpPrefixArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Azure's default length is 28 (16 addresses); only an explicit choice
	// is ever sent, so an unspecified spec and Azure's default deploy
	// identically on both engines. The same only-send-explicit rule applies
	// to the version/SKU/tier enums below.
	if spec.PrefixLength != nil {
		prefixArgs.PrefixLength = pulumi.Int(int(spec.GetPrefixLength()))
	}
	if locals.IpVersion != "" {
		prefixArgs.IpVersion = pulumi.String(locals.IpVersion)
	}
	if locals.Sku != "" {
		prefixArgs.Sku = pulumi.String(locals.Sku)
	}
	if locals.SkuTier != "" {
		prefixArgs.SkuTier = pulumi.String(locals.SkuTier)
	}
	if len(spec.Zones) > 0 {
		prefixArgs.Zones = pulumi.ToStringArray(spec.Zones)
	}
	if spec.CustomIpPrefixId != "" {
		prefixArgs.CustomIpPrefixId = pulumi.String(spec.CustomIpPrefixId)
	}

	createdPrefix, err := network.NewPublicIpPrefix(ctx,
		spec.Name,
		prefixArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create public ip prefix %s", spec.Name)
	}

	// Export outputs from the created resource. ip_prefix is the
	// actual reserved CIDR -- known only after creation, and the value
	// partners and firewalls allowlist.
	ctx.Export(OpPublicIpPrefixId, createdPrefix.ID())
	ctx.Export(OpIpPrefix, createdPrefix.IpPrefix)
	ctx.Export(OpPublicIpPrefixName, createdPrefix.Name)

	return nil
}
