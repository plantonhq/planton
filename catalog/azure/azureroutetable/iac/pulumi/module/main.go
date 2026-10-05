package module

import (
	"github.com/pkg/errors"
	azureroutetablev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureroutetable/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureroutetablev1alpha1.AzureRouteTableIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureRouteTable.Spec

	// Routes are managed inline as part of the table (a route has no life of
	// its own in Azure), so the list is ALWAYS set -- an explicitly empty
	// list is what removes the last route; leaving the field unset would
	// make the provider treat existing routes as externally managed.
	routes := make(network.RouteTableRouteArray, 0, len(spec.Routes))
	for _, route := range spec.Routes {
		routeArgs := network.RouteTableRouteArgs{
			Name:          pulumi.String(route.Name),
			AddressPrefix: pulumi.String(route.AddressPrefix),
			NextHopType:   pulumi.String(nextHopTypeToArm[route.NextHopType]),
		}
		// Only VirtualAppliance routes carry a forwarding IP (spec-level
		// validation enforces the pairing); ARM rejects it on any other
		// hop type. The address is a StringValueOrRef -- the platform
		// resolves references (typically an AzureFirewall's
		// private_ip_address) to a literal before the module runs.
		if route.NextHopInIpAddress.GetValue() != "" {
			routeArgs.NextHopInIpAddress = pulumi.String(route.NextHopInIpAddress.GetValue())
		}
		routes = append(routes, routeArgs)
	}

	// Lifecycle notes worth knowing before operating this resource:
	// - Routes, BGP propagation, and tags all update IN PLACE -- and take
	//   effect immediately for EVERY subnet attached to the table. Name,
	//   region, and resource group are the table's ARM identity; changing
	//   any of them replaces the table, detaching it from every subnet
	//   until the replacement is re-attached.
	// - The subnet-side attachment is deliberately not modeled here: a
	//   subnet declares which route table it uses (matching Azure's
	//   model), so one table serves many subnets without listing them.
	createdRouteTable, err := network.NewRouteTable(ctx,
		spec.Name,
		&network.RouteTableArgs{
			Name:              pulumi.String(spec.Name),
			Location:          pulumi.String(spec.Region),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Routes:            routes,
			// Azure defaults to propagating BGP-learned routes; disabling is
			// the forced-tunneling hardening that keeps learned routes from
			// bypassing the user-defined ones.
			BgpRoutePropagationEnabled: pulumi.Bool(locals.BgpRoutePropagationEnabled),
			Tags:                       pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create route table %s", spec.Name)
	}

	// Export outputs from the created resource. route_table_id is the
	// join key subnets use to attach the table's routing policy.
	ctx.Export(OpRouteTableId, createdRouteTable.ID())
	ctx.Export(OpRouteTableName, createdRouteTable.Name)

	return nil
}
