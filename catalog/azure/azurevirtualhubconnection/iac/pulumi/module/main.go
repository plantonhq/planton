package module

import (
	"github.com/pkg/errors"
	azurevirtualhubconnectionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualhubconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevirtualhubconnectionv1alpha1.AzureVirtualHubConnectionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVirtualHubConnection.Spec

	// Create the Virtual Hub Connection -- the attachment that joins one
	// spoke VNet to a Virtual WAN hub. The connection is free; what it
	// unlocks is the hub's routing. Deleting the hub requires this
	// connection to be gone first (the runner's reverse teardown handles
	// the ordering).
	connectionArgs := &network.VirtualHubConnectionArgs{
		Name:                   pulumi.String(spec.Name),
		VirtualHubId:           pulumi.String(spec.VirtualHubId.GetValue()),
		RemoteVirtualNetworkId: pulumi.String(spec.RemoteVirtualNetworkId.GetValue()),
		// Off is ARM's default: the spoke keeps its own internet egress.
		// On, the hub advertises 0.0.0.0/0 to this connection (typically
		// paired with a hub firewall via routing intent).
		InternetSecurityEnabled: pulumi.Bool(spec.InternetSecurityEnabled),
	}

	// Unset routing applies ARM's default behavior: associate with and
	// propagate to the hub's built-in default route table (any-to-any).
	// The spec guarantees a configured block carries at least one of the
	// provider's at-least-one-of trio (association, propagation, static
	// routes).
	if spec.Routing != nil {
		routingArgs := &network.VirtualHubConnectionRoutingArgs{
			// ARM fixes the criteria once the connection is created
			// (ForceNew); Contains is ARM's default.
			StaticVnetLocalRouteOverrideCriteria: pulumi.String(overrideCriteriaWireValue(spec.Routing.StaticVnetLocalRouteOverrideCriteria)),
			// ARM defaults propagation of static routes ON; the optional
			// field's default mirrors it.
			StaticVnetPropagateStaticRoutesEnabled: pulumi.Bool(optionalBool(spec.Routing.StaticVnetPropagateStaticRoutesEnabled, true)),
		}
		if spec.Routing.AssociatedRouteTableId.GetValue() != "" {
			routingArgs.AssociatedRouteTableId = pulumi.String(spec.Routing.AssociatedRouteTableId.GetValue())
		}
		if spec.Routing.InboundRouteMapId.GetValue() != "" {
			routingArgs.InboundRouteMapId = pulumi.String(spec.Routing.InboundRouteMapId.GetValue())
		}
		if spec.Routing.OutboundRouteMapId.GetValue() != "" {
			routingArgs.OutboundRouteMapId = pulumi.String(spec.Routing.OutboundRouteMapId.GetValue())
		}
		if spec.Routing.PropagatedRouteTable != nil {
			routingArgs.PropagatedRouteTable = &network.VirtualHubConnectionRoutingPropagatedRouteTableArgs{
				Labels:        pulumi.ToStringArray(spec.Routing.PropagatedRouteTable.Labels),
				RouteTableIds: pulumi.ToStringArray(resolveRefValues(spec.Routing.PropagatedRouteTable.RouteTableIds)),
			}
		}
		if len(spec.Routing.StaticVnetRoutes) > 0 {
			staticVnetRoutes := network.VirtualHubConnectionRoutingStaticVnetRouteArray{}
			for _, staticVnetRoute := range spec.Routing.StaticVnetRoutes {
				staticVnetRoutes = append(staticVnetRoutes, &network.VirtualHubConnectionRoutingStaticVnetRouteArgs{
					Name:             pulumi.String(staticVnetRoute.Name),
					AddressPrefixes:  pulumi.ToStringArray(staticVnetRoute.AddressPrefixes),
					NextHopIpAddress: pulumi.String(staticVnetRoute.NextHopIpAddress),
				})
			}
			routingArgs.StaticVnetRoutes = staticVnetRoutes
		}
		connectionArgs.Routing = routingArgs
	}

	createdConnection, err := network.NewVirtualHubConnection(ctx,
		spec.Name,
		connectionArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create virtual hub connection %s", spec.Name)
	}

	ctx.Export(OpVirtualHubConnectionId, createdConnection.ID())
	ctx.Export(OpVirtualHubConnectionName, createdConnection.Name)

	return nil
}

// resolveRefValues extracts the resolved literal from each
// StringValueOrRef (the platform middleware resolves valueFrom
// references before IaC modules run).
func resolveRefValues(refs []*foreignkeyv1.StringValueOrRef) []string {
	values := make([]string, 0, len(refs))
	for _, ref := range refs {
		values = append(values, ref.GetValue())
	}
	return values
}
