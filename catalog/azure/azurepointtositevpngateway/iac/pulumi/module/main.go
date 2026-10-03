package module

import (
	"github.com/pkg/errors"
	azurepointtositevpngatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurepointtositevpngateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurepointtositevpngatewayv1alpha1.AzurePointToSiteVpnGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePointToSiteVpnGateway.Spec

	// One block per named client address pool. Most gateways carry
	// exactly one; multiple pools require OpenVPN on the server
	// configuration and are matched to users via its policy groups.
	connectionConfigurations := network.PointToPointVpnGatewayConnectionConfigurationArray{}
	for _, connectionConfiguration := range spec.ConnectionConfigurations {
		configurationArgs := &network.PointToPointVpnGatewayConnectionConfigurationArgs{
			Name: pulumi.String(connectionConfiguration.Name),
			// The provider nests the pool under vpn_client_address_pool;
			// the spec flattens it to address_prefixes (recorded in the
			// parity manifest).
			VpnClientAddressPool: &network.PointToPointVpnGatewayConnectionConfigurationVpnClientAddressPoolArgs{
				AddressPrefixes: pulumi.ToStringArray(connectionConfiguration.AddressPrefixes),
			},
			// Off is ARM's default: clients keep their local internet
			// egress. On, the hub advertises 0.0.0.0/0 into the tunnel.
			InternetSecurityEnabled: pulumi.Bool(connectionConfiguration.InternetSecurityEnabled),
		}

		// Unset routing applies ARM's default behavior: associate with
		// and propagate to the hub's built-in default route table. A
		// configured block carries its association (the spec requires it
		// -- the provider's own contract). The provider names the
		// propagation targets `ids`; the spec says what they ARE
		// (route_table_ids -- recorded in the parity manifest).
		if connectionConfiguration.Route != nil {
			routeArgs := &network.PointToPointVpnGatewayConnectionConfigurationRouteArgs{
				AssociatedRouteTableId: pulumi.String(connectionConfiguration.Route.AssociatedRouteTableId.GetValue()),
			}
			if connectionConfiguration.Route.InboundRouteMapId.GetValue() != "" {
				routeArgs.InboundRouteMapId = pulumi.String(connectionConfiguration.Route.InboundRouteMapId.GetValue())
			}
			if connectionConfiguration.Route.OutboundRouteMapId.GetValue() != "" {
				routeArgs.OutboundRouteMapId = pulumi.String(connectionConfiguration.Route.OutboundRouteMapId.GetValue())
			}
			if connectionConfiguration.Route.PropagatedRouteTable != nil {
				routeTableIds := make([]string, 0, len(connectionConfiguration.Route.PropagatedRouteTable.RouteTableIds))
				for _, routeTableId := range connectionConfiguration.Route.PropagatedRouteTable.RouteTableIds {
					routeTableIds = append(routeTableIds, routeTableId.GetValue())
				}
				routeArgs.PropagatedRouteTable = &network.PointToPointVpnGatewayConnectionConfigurationRoutePropagatedRouteTableArgs{
					Ids:    pulumi.ToStringArray(routeTableIds),
					Labels: pulumi.ToStringArray(connectionConfiguration.Route.PropagatedRouteTable.Labels),
				}
			}
			configurationArgs.Route = routeArgs
		}

		connectionConfigurations = append(connectionConfigurations, configurationArgs)
	}

	// Create the point-to-site VPN gateway -- the managed receiver
	// inside a virtual hub that individual devices dial into (ARM
	// allows one per hub, a slot separate from the hub's site-to-site
	// VPN gateway). The gateway bills from creation and is a SLOW
	// resource: creates run 30-45 minutes -- the provider's own timeout
	// class is 90 minutes. The SDK's constructor is
	// NewPointToPointVpnGateway -- the resource lives at a legacy,
	// misnamed "PointToPointVpnGateway" token but creates the SAME ARM
	// p2sVpnGateways object as azurerm_point_to_site_vpn_gateway; do
	// not "fix" the name.
	gatewayArgs := &network.PointToPointVpnGatewayArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// Both ForceNew: the gateway is born in its hub, pointing at
		// its authentication policy.
		VirtualHubId:             pulumi.String(spec.VirtualHubId.GetValue()),
		VpnServerConfigurationId: pulumi.String(spec.VpnServerConfigurationId.GetValue()),
		// 500 concurrent connections per unit across the managed
		// instance pair. The provider REQUIRES an explicit value (no
		// provider default); the spec's unset applies 1 -- rendered
		// explicitly, mirroring the Terraform module's null handling.
		ScaleUnit:                pulumi.Int(int(optionalInt32(spec.ScaleUnit, 1))),
		ConnectionConfigurations: connectionConfigurations,
		// Off is ARM's default (client internet traffic rides
		// Microsoft's backbone). ForceNew: changing it replaces the
		// gateway.
		RoutingPreferenceInternetEnabled: pulumi.Bool(spec.RoutingPreferenceInternetEnabled),
		Tags:                             pulumi.ToStringMap(locals.AzureTags),
	}

	// Pushed to connecting clients. Emitted only when configured;
	// NOTE: the provider cannot CLEAR a previously set list (its
	// update path skips empty lists) -- removing servers requires
	// replacing the gateway, which the spec documents on the field.
	if len(spec.DnsServers) > 0 {
		gatewayArgs.DnsServers = pulumi.ToStringArray(spec.DnsServers)
	}

	createdGateway, err := network.NewPointToPointVpnGateway(ctx,
		spec.Name,
		gatewayArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create point-to-site vpn gateway %s", spec.Name)
	}

	ctx.Export(OpPointToSiteVpnGatewayId, createdGateway.ID())
	ctx.Export(OpPointToSiteVpnGatewayName, createdGateway.Name)

	return nil
}
