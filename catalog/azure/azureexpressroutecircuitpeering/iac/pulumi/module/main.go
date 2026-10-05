package module

import (
	"github.com/pkg/errors"
	azureexpressroutecircuitpeeringv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureexpressroutecircuitpeering/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureexpressroutecircuitpeeringv1alpha1.AzureExpressRouteCircuitPeeringIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureExpressRouteCircuitPeering.Spec

	peeringArgs := &network.ExpressRouteCircuitPeeringArgs{
		PeeringType:             pulumi.String(peeringTypeWireValue(spec.PeeringType)),
		ExpressRouteCircuitName: pulumi.String(locals.ExpressRouteCircuitName),
		ResourceGroupName:       pulumi.String(locals.ResourceGroupName),
		VlanId:                  pulumi.Int(int(spec.VlanId)),
		Ipv4Enabled:             pulumi.Bool(spec.GetIpv4Enabled()),
	}

	// The /30 pair travels together (spec-validated); empty means the
	// session addressing is configured later (private peering allows it).
	if spec.PrimaryPeerAddressPrefix != "" {
		peeringArgs.PrimaryPeerAddressPrefix = pulumi.String(spec.PrimaryPeerAddressPrefix)
		peeringArgs.SecondaryPeerAddressPrefix = pulumi.String(spec.SecondaryPeerAddressPrefix)
	}

	if spec.PeerAsn > 0 {
		peeringArgs.PeerAsn = pulumi.Int(int(spec.PeerAsn))
	}

	if spec.SharedKey.GetValue() != "" {
		peeringArgs.SharedKey = pulumi.String(spec.SharedKey.GetValue())
	}

	if spec.RouteFilterId != "" {
		peeringArgs.RouteFilterId = pulumi.String(spec.RouteFilterId)
	}

	if spec.MicrosoftPeeringConfig != nil {
		peeringArgs.MicrosoftPeeringConfig = expandMicrosoftPeeringConfig(spec.MicrosoftPeeringConfig)
	}

	if spec.Ipv6 != nil {
		ipv6Args := &network.ExpressRouteCircuitPeeringIpv6Args{
			PrimaryPeerAddressPrefix:   pulumi.String(spec.Ipv6.PrimaryPeerAddressPrefix),
			SecondaryPeerAddressPrefix: pulumi.String(spec.Ipv6.SecondaryPeerAddressPrefix),
			Enabled:                    pulumi.Bool(spec.Ipv6.GetEnabled()),
		}
		if spec.Ipv6.RouteFilterId != "" {
			ipv6Args.RouteFilterId = pulumi.String(spec.Ipv6.RouteFilterId)
		}
		if spec.Ipv6.MicrosoftPeering != nil {
			ipv6Args.MicrosoftPeering = &network.ExpressRouteCircuitPeeringIpv6MicrosoftPeeringArgs{
				AdvertisedPublicPrefixes: pulumi.ToStringArray(spec.Ipv6.MicrosoftPeering.AdvertisedPublicPrefixes),
				CustomerAsn:              pulumi.Int(int(spec.Ipv6.MicrosoftPeering.CustomerAsn)),
				RoutingRegistryName:      pulumi.String(spec.Ipv6.MicrosoftPeering.GetRoutingRegistryName()),
				AdvertisedCommunities:    pulumi.ToStringArray(spec.Ipv6.MicrosoftPeering.AdvertisedCommunities),
			}
		}
		peeringArgs.Ipv6 = ipv6Args
	}

	// ARM accepts and STORES peering configuration even while the circuit's
	// provider state is "NotProvisioned" (live-verified for private
	// peering); the BGP session only establishes after the connectivity
	// provider completes the cross-connect. Microsoft peering's
	// server-side public-prefix validation has not been exercised on an
	// unprovisioned circuit.
	createdPeering, err := network.NewExpressRouteCircuitPeering(ctx,
		locals.AzureExpressRouteCircuitPeering.Metadata.Name,
		peeringArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create express route circuit peering on %s", locals.ExpressRouteCircuitName)
	}

	// The composed Global Reach connections: ARM children of this
	// peering, joining it to other circuits' private peerings. The
	// near-side peering_id is wired to the peering just created -- only
	// the far side is user surface.
	connectionIds := pulumi.Map{}
	for _, connection := range spec.Connections {
		connectionArgs := &network.ExpressRouteCircuitConnectionArgs{
			Name:              pulumi.String(connection.Name),
			PeeringId:         createdPeering.ID(),
			PeerPeeringId:     pulumi.String(connection.PeerPeeringId.GetValue()),
			AddressPrefixIpv4: pulumi.String(connection.AddressPrefixIpv4),
		}
		if connection.AddressPrefixIpv6 != "" {
			connectionArgs.AddressPrefixIpv6 = pulumi.String(connection.AddressPrefixIpv6)
		}
		if connection.AuthorizationKey.GetValue() != "" {
			connectionArgs.AuthorizationKey = pulumi.String(connection.AuthorizationKey.GetValue())
		}

		createdConnection, err := network.NewExpressRouteCircuitConnection(ctx,
			locals.AzureExpressRouteCircuitPeering.Metadata.Name+"-"+connection.Name,
			connectionArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdPeering))
		if err != nil {
			return errors.Wrapf(err, "failed to create global reach connection %s", connection.Name)
		}
		connectionIds[connection.Name] = createdConnection.ID()
	}

	ctx.Export(OpExpressRouteCircuitPeeringId, createdPeering.ID())
	ctx.Export(OpAzureAsn, createdPeering.AzureAsn)
	ctx.Export(OpPrimaryAzurePort, createdPeering.PrimaryAzurePort)
	ctx.Export(OpSecondaryAzurePort, createdPeering.SecondaryAzurePort)
	ctx.Export(OpConnectionIds, connectionIds)

	return nil
}

// expandMicrosoftPeeringConfig maps the spec's advertisement contract
// onto the SDK block. RoutingRegistryName defaults to "NONE" via the
// platform's default middleware, so the getter is always populated.
func expandMicrosoftPeeringConfig(config *azureexpressroutecircuitpeeringv1alpha1.AzureExpressRouteCircuitPeeringMicrosoftConfig) *network.ExpressRouteCircuitPeeringMicrosoftPeeringConfigArgs {
	return &network.ExpressRouteCircuitPeeringMicrosoftPeeringConfigArgs{
		AdvertisedPublicPrefixes: pulumi.ToStringArray(config.AdvertisedPublicPrefixes),
		CustomerAsn:              pulumi.Int(int(config.CustomerAsn)),
		RoutingRegistryName:      pulumi.String(config.GetRoutingRegistryName()),
		AdvertisedCommunities:    pulumi.ToStringArray(config.AdvertisedCommunities),
	}
}
