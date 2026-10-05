package module

import (
	"github.com/pkg/errors"
	azureprivatednsresolvervirtualnetworklinkv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednsresolvervirtualnetworklink/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureprivatednsresolvervirtualnetworklinkv1alpha1.AzurePrivateDnsResolverVirtualNetworkLinkIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateDnsResolverVirtualNetworkLink.Spec

	// The attachment that makes a DNS forwarding ruleset take effect in
	// ONE virtual network. The linked network must be in the ruleset's
	// region but does NOT need to be peered with the resolver's network
	// (hub-and-spoke: spokes link to the hub's ruleset). Everything
	// except metadata is create-only.
	linkArgs := &privatedns.ResolverVirtualNetworkLinkArgs{
		Name:                   pulumi.String(spec.Name),
		DnsForwardingRulesetId: pulumi.String(spec.DnsForwardingRulesetId.GetValue()),
		VirtualNetworkId:       pulumi.String(spec.VirtualNetworkId.GetValue()),
	}
	// ARM's free-form annotation map on the link itself (links carry no
	// tags) -- the only surface updatable in place.
	if len(spec.Metadata) > 0 {
		linkArgs.Metadata = pulumi.ToStringMap(spec.Metadata)
	}

	createdLink, err := privatedns.NewResolverVirtualNetworkLink(ctx,
		spec.Name,
		linkArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create virtual network link %s", spec.Name)
	}

	ctx.Export(OpVirtualNetworkLinkId, createdLink.ID())
	ctx.Export(OpVirtualNetworkLinkName, createdLink.Name)

	return nil
}
