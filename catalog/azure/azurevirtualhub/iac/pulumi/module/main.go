package module

import (
	"github.com/pkg/errors"
	azurevirtualhubv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualhub/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevirtualhubv1alpha1.AzureVirtualHubIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVirtualHub.Spec

	// Create the Virtual Hub -- the managed regional router of a Virtual
	// WAN. A Standard hub bills (~$0.25/hr class) from creation, and ARM
	// takes 15-30 minutes to bring the hub's router to a Provisioned
	// routing state; deleting a hub requires its connections and
	// gateways to be gone first.
	hubArgs := &network.VirtualHubArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// Required by the spec although the provider marks it optional:
		// the provider's schema also serves the standalone Route Server
		// construction (not modeled by this kind), and ARM rejects a WAN
		// hub without an address prefix.
		VirtualWanId:  pulumi.String(spec.VirtualWanId.GetValue()),
		AddressPrefix: pulumi.String(spec.AddressPrefix),
		Sku:           pulumi.String(skuWireValue(spec.Sku)),
		// ExpressRoute is ARM's default preference for prefixes learned
		// over multiple tunnel types.
		HubRoutingPreference: pulumi.String(hubRoutingPreferenceWireValue(spec.HubRoutingPreference)),
		// Off is ARM's default; the WAN's own allow_branch_to_branch_traffic
		// must ALSO be on for branch-to-branch transit to actually flow.
		BranchToBranchTrafficEnabled: pulumi.Bool(spec.BranchToBranchTrafficEnabled),
		// ARM's floor and default is 2 routing infrastructure units.
		VirtualRouterAutoScaleMinCapacity: pulumi.Int(int(optionalInt32(spec.VirtualRouterAutoScaleMinCapacity, 2))),
		Tags:                              pulumi.ToStringMap(locals.AzureTags),
	}

	// The hub resource's classic inline routes (applied to the default
	// route table). The modern per-table form lives in route_tables.
	if len(spec.Routes) > 0 {
		routes := network.VirtualHubRouteArray{}
		for _, route := range spec.Routes {
			routes = append(routes, &network.VirtualHubRouteArgs{
				AddressPrefixes:  pulumi.ToStringArray(route.AddressPrefixes),
				NextHopIpAddress: pulumi.String(route.NextHopIpAddress),
			})
		}
		hubArgs.Routes = routes
	}

	createdHub, err := network.NewVirtualHub(ctx,
		spec.Name,
		hubArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create virtual hub %s", spec.Name)
	}

	// The composed custom route tables: standalone ARM children of the
	// hub, one per spec entry, keyed by name (how spoke isolation and
	// shared-services topologies are built). Routes are managed INLINE
	// on each table -- never mix in the provider's standalone route
	// resource, which fights inline routes over the same ARM collection.
	routeTableIds := pulumi.Map{}
	for _, routeTable := range spec.RouteTables {
		routes := network.VirtualHubRouteTableRouteTypeArray{}
		for _, route := range routeTable.Routes {
			routes = append(routes, &network.VirtualHubRouteTableRouteTypeArgs{
				Name:             pulumi.String(route.Name),
				DestinationsType: pulumi.String(destinationsTypeWireValue(route.DestinationsType)),
				Destinations:     pulumi.ToStringArray(route.Destinations),
				NextHop:          pulumi.String(route.NextHop.GetValue()),
				// NextHopType is left to the provider's default -- ARM's
				// only value is "ResourceId"; there is nothing to configure.
			})
		}
		createdRouteTable, err := network.NewVirtualHubRouteTable(ctx,
			spec.Name+"-"+routeTable.Name,
			&network.VirtualHubRouteTableArgs{
				Name:         pulumi.String(routeTable.Name),
				VirtualHubId: createdHub.ID(),
				Labels:       pulumi.ToStringArray(routeTable.Labels),
				Routes:       routes,
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdHub))
		if err != nil {
			return errors.Wrapf(err, "failed to create route table %s", routeTable.Name)
		}
		routeTableIds[routeTable.Name] = createdRouteTable.ID()
	}

	// The composed route maps: BGP route match/transform policies that
	// connections reference as inbound_route_map_id/outbound_route_map_id.
	// The SDK's constructor is NewRouteMapResource -- the resource lives
	// at a legacy "RouteMapResource" token but creates the SAME ARM
	// object as azurerm_route_map.
	routeMapIds := pulumi.Map{}
	for _, routeMap := range spec.RouteMaps {
		rules := network.RouteMapRuleArray{}
		for _, rule := range routeMap.Rules {
			ruleArgs := &network.RouteMapRuleArgs{
				Name: pulumi.String(rule.Name),
			}
			// Unset leaves the provider's default ("Unknown": evaluation
			// stops after the match) -- the spec deliberately does not
			// model ARM's "Unknown" as a choosable value.
			if rule.NextStepIfMatched != nil {
				if *rule.NextStepIfMatched == azurevirtualhubv1alpha1.AzureVirtualHubRouteMapNextStep_TERMINATE {
					ruleArgs.NextStepIfMatched = pulumi.String("Terminate")
				} else {
					ruleArgs.NextStepIfMatched = pulumi.String("Continue")
				}
			}
			matchCriterions := network.RouteMapRuleMatchCriterionArray{}
			for _, criterion := range rule.MatchCriteria {
				matchCriterions = append(matchCriterions, &network.RouteMapRuleMatchCriterionArgs{
					MatchCondition: pulumi.String(matchConditionWireValue(criterion.MatchCondition)),
					AsPaths:        pulumi.ToStringArray(criterion.AsPath),
					Communities:    pulumi.ToStringArray(criterion.Community),
					RoutePrefixes:  pulumi.ToStringArray(criterion.RoutePrefix),
				})
			}
			ruleArgs.MatchCriterions = matchCriterions
			actions := network.RouteMapRuleActionArray{}
			for _, action := range rule.Actions {
				// The spec guarantees non-DROP actions carry at least one
				// parameter (mirroring the provider's own create-time rule).
				parameters := network.RouteMapRuleActionParameterArray{}
				for _, parameter := range action.Parameters {
					parameters = append(parameters, &network.RouteMapRuleActionParameterArgs{
						AsPaths:       pulumi.ToStringArray(parameter.AsPath),
						Communities:   pulumi.ToStringArray(parameter.Community),
						RoutePrefixes: pulumi.ToStringArray(parameter.RoutePrefix),
					})
				}
				actions = append(actions, &network.RouteMapRuleActionArgs{
					Type:       pulumi.String(actionTypeWireValue(action.Type)),
					Parameters: parameters,
				})
			}
			ruleArgs.Actions = actions
			rules = append(rules, ruleArgs)
		}
		createdRouteMap, err := network.NewRouteMapResource(ctx,
			spec.Name+"-"+routeMap.Name,
			&network.RouteMapResourceArgs{
				Name:         pulumi.String(routeMap.Name),
				VirtualHubId: createdHub.ID(),
				Rules:        rules,
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdHub))
		if err != nil {
			return errors.Wrapf(err, "failed to create route map %s", routeMap.Name)
		}
		routeMapIds[routeMap.Name] = createdRouteMap.ID()
	}

	// The composed BGP peerings between the hub's router and NVAs in
	// connected spokes. All fields are ForceNew on ARM's side; routes
	// only flow once the peer is reachable through a hub connection.
	bgpConnectionIds := pulumi.Map{}
	for _, bgpConnection := range spec.BgpConnections {
		bgpConnectionArgs := &network.BgpConnectionArgs{
			Name:         pulumi.String(bgpConnection.Name),
			VirtualHubId: createdHub.ID(),
			PeerAsn:      pulumi.Int(int(bgpConnection.PeerAsn)),
			PeerIp:       pulumi.String(bgpConnection.PeerIp),
		}
		if bgpConnection.VirtualNetworkConnectionId.GetValue() != "" {
			bgpConnectionArgs.VirtualNetworkConnectionId = pulumi.String(bgpConnection.VirtualNetworkConnectionId.GetValue())
		}
		createdBgpConnection, err := network.NewBgpConnection(ctx,
			spec.Name+"-"+bgpConnection.Name,
			bgpConnectionArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdHub))
		if err != nil {
			return errors.Wrapf(err, "failed to create bgp connection %s", bgpConnection.Name)
		}
		bgpConnectionIds[bgpConnection.Name] = createdBgpConnection.ID()
	}

	// The hub's routing intent (at most one): steers Internet/private
	// traffic through a security appliance in THIS hub. Setting it takes
	// over the hub's routing policy -- per-connection route-table
	// customization and routing intent are mutually exclusive on ARM's
	// side.
	routingIntentId := pulumi.StringInput(pulumi.String(""))
	if spec.RoutingIntent != nil {
		routingPolicies := network.RoutingIntentRoutingPolicyArray{}
		for _, routingPolicy := range spec.RoutingIntent.RoutingPolicies {
			destinations := make([]string, 0, len(routingPolicy.Destinations))
			for _, destination := range routingPolicy.Destinations {
				destinations = append(destinations, routingPolicyDestinationWireValue(destination))
			}
			routingPolicies = append(routingPolicies, &network.RoutingIntentRoutingPolicyArgs{
				Name:         pulumi.String(routingPolicy.Name),
				Destinations: pulumi.ToStringArray(destinations),
				NextHop:      pulumi.String(routingPolicy.NextHop.GetValue()),
			})
		}
		createdRoutingIntent, err := network.NewRoutingIntent(ctx,
			spec.Name+"-"+spec.RoutingIntent.Name,
			&network.RoutingIntentArgs{
				Name:            pulumi.String(spec.RoutingIntent.Name),
				VirtualHubId:    createdHub.ID(),
				RoutingPolicies: routingPolicies,
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdHub))
		if err != nil {
			return errors.Wrapf(err, "failed to create routing intent %s", spec.RoutingIntent.Name)
		}
		routingIntentId = createdRoutingIntent.ID()
	}

	ctx.Export(OpVirtualHubId, createdHub.ID())
	ctx.Export(OpVirtualHubName, createdHub.Name)
	ctx.Export(OpDefaultRouteTableId, createdHub.DefaultRouteTableId)
	ctx.Export(OpVirtualRouterAsn, createdHub.VirtualRouterAsn)
	ctx.Export(OpVirtualRouterIps, createdHub.VirtualRouterIps)
	ctx.Export(OpRouteTableIds, routeTableIds)
	ctx.Export(OpRouteMapIds, routeMapIds)
	ctx.Export(OpBgpConnectionIds, bgpConnectionIds)
	// Empty when no routing intent is configured -- mirrors the
	// Terraform module's try(..., "").
	ctx.Export(OpRoutingIntentId, routingIntentId)

	return nil
}

// optionalInt32 returns the pointed-to value, or the default when the
// optional field is unset -- mirroring the Terraform variable default.
func optionalInt32(value *int32, defaultValue int32) int32 {
	if value == nil {
		return defaultValue
	}
	return *value
}
