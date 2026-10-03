package module

import (
	"github.com/pkg/errors"
	azureexpressroutegatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureexpressroutegateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureexpressroutegatewayv1alpha1.AzureExpressRouteGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureExpressRouteGateway.Spec

	// Create the ExpressRoute Gateway -- the Virtual WAN on-ramp for
	// ExpressRoute circuits. The gateway bills per scale unit
	// (~$0.42/hr per unit) FROM CREATION, and ARM takes roughly 30
	// minutes to provision one. A hub holds at most one ExpressRoute
	// gateway.
	createdGateway, err := network.NewExpressRouteGateway(ctx,
		spec.Name,
		&network.ExpressRouteGatewayArgs{
			Name:              pulumi.String(spec.Name),
			Location:          pulumi.String(spec.Region),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			VirtualHubId:      pulumi.String(spec.VirtualHubId.GetValue()),
			// The MINIMUM scale units (1-10, spec-validated); each unit
			// carries ~2 Gbps and ARM auto-scales above this floor.
			ScaleUnits: pulumi.Int(int(spec.ScaleUnits)),
			// Off is ARM's default: only Virtual WAN networks may ride the
			// gateway. On, classic VNets connected to the same circuit may
			// too.
			AllowNonVirtualWanTraffic: pulumi.Bool(spec.AllowNonVirtualWanTraffic),
			Tags:                      pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create express route gateway %s", spec.Name)
	}

	// The composed connections: standalone ARM children of the gateway,
	// one per spec entry, keyed by name. Each joins one ExpressRoute
	// circuit PEERING to the hub -- and ARM accepts it only when the
	// circuit's provider side is PROVISIONED (a live carrier or an
	// ExpressRoute Direct port behind it).
	connectionIds := pulumi.Map{}
	for _, connection := range spec.Connections {
		connectionArgs := &network.ExpressRouteConnectionArgs{
			Name:                             pulumi.String(connection.Name),
			ExpressRouteGatewayId:            createdGateway.ID(),
			ExpressRouteCircuitPeeringId:     pulumi.String(connection.ExpressRouteCircuitPeeringId.GetValue()),
			InternetSecurityEnabled:          pulumi.Bool(connection.InternetSecurityEnabled),
			ExpressRouteGatewayBypassEnabled: pulumi.Bool(connection.ExpressRouteGatewayBypassEnabled),
			// 0 is ARM's default; higher weight wins when the same prefix
			// is reachable over multiple connections.
			RoutingWeight: pulumi.Int(int(connection.RoutingWeight)),
		}

		// The authorization key (a UUID, sensitive) for a circuit in
		// ANOTHER subscription; empty means the circuit is in this
		// subscription.
		if connection.AuthorizationKey != "" {
			connectionArgs.AuthorizationKey = pulumi.String(connection.AuthorizationKey)
		}

		// Unset routing applies ARM's default behavior: associate with
		// and propagate to the hub's built-in default route table. The
		// spec guarantees a configured block carries an association or a
		// propagation (the provider's at-least-one-of pair).
		if connection.Routing != nil {
			routingArgs := &network.ExpressRouteConnectionRoutingArgs{}
			if connection.Routing.AssociatedRouteTableId.GetValue() != "" {
				routingArgs.AssociatedRouteTableId = pulumi.String(connection.Routing.AssociatedRouteTableId.GetValue())
			}
			if connection.Routing.InboundRouteMapId.GetValue() != "" {
				routingArgs.InboundRouteMapId = pulumi.String(connection.Routing.InboundRouteMapId.GetValue())
			}
			if connection.Routing.OutboundRouteMapId.GetValue() != "" {
				routingArgs.OutboundRouteMapId = pulumi.String(connection.Routing.OutboundRouteMapId.GetValue())
			}
			if connection.Routing.PropagatedRouteTable != nil {
				routingArgs.PropagatedRouteTable = &network.ExpressRouteConnectionRoutingPropagatedRouteTableArgs{
					Labels:        pulumi.ToStringArray(connection.Routing.PropagatedRouteTable.Labels),
					RouteTableIds: pulumi.ToStringArray(resolveRefValues(connection.Routing.PropagatedRouteTable.RouteTableIds)),
				}
			}
			connectionArgs.Routing = routingArgs
		}

		createdConnection, err := network.NewExpressRouteConnection(ctx,
			spec.Name+"-"+connection.Name,
			connectionArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdGateway))
		if err != nil {
			return errors.Wrapf(err, "failed to create express route connection %s", connection.Name)
		}
		connectionIds[connection.Name] = createdConnection.ID()
	}

	ctx.Export(OpExpressRouteGatewayId, createdGateway.ID())
	ctx.Export(OpExpressRouteGatewayName, createdGateway.Name)
	ctx.Export(OpConnectionIds, connectionIds)

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
