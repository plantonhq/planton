package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// networkPeering is the CREATE form: this side's peering entry on
// `network`, pointing at `peer_network`, with the route exchange it allows.
//
// The name, both networks, and the two public-IP subnet-route flags are
// immutable (a change recreates the peering); the custom-route flags,
// stack_type, and update_strategy change in place. All four route-exchange
// flags are always sent (false is a real choice; the public-IP pair carries
// the spec's defaults when unset, which the manifest loader applies before
// either engine runs). Optional enums are sent only when set so the
// provider's defaults stay the provider's.
func networkPeering(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVpcPeering.Spec

	args := &compute.NetworkPeeringArgs{
		Name:               pulumi.String(locals.PeeringName),
		Network:            pulumi.String(spec.Network.GetValue()),
		PeerNetwork:        pulumi.String(spec.PeerNetwork.GetValue()),
		ExportCustomRoutes: pulumi.Bool(spec.ExportCustomRoutes),
		ImportCustomRoutes: pulumi.Bool(spec.ImportCustomRoutes),
		// Explicit send of the provider defaults (true / false) when unset:
		// both flags are ForceNew, so the value the spec implies must be
		// the value on the wire from the first apply.
		ExportSubnetRoutesWithPublicIp: pulumi.Bool(exportSubnetRoutesWithPublicIp(spec.ExportSubnetRoutesWithPublicIp)),
		ImportSubnetRoutesWithPublicIp: pulumi.Bool(spec.GetImportSubnetRoutesWithPublicIp()),
	}
	if spec.StackType != "" {
		args.StackType = pulumi.StringPtr(spec.StackType)
	}
	if spec.UpdateStrategy != "" {
		args.UpdateStrategy = pulumi.StringPtr(spec.UpdateStrategy)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := compute.NewNetworkPeering(ctx, locals.GcpVpcPeering.Metadata.Name, args,
		pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create network peering")
	}

	ctx.Export(OpPeeringName, created.Name)
	ctx.Export(OpNetwork, created.Network)
	ctx.Export(OpState, created.State)
	ctx.Export(OpStateDetails, created.StateDetails)

	return nil
}

// exportSubnetRoutesWithPublicIp resolves the tri-state optional to the
// value on the wire: the spec's value when set, the provider's default
// (true) when not.
func exportSubnetRoutesWithPublicIp(value *bool) bool {
	if value == nil {
		return true
	}
	return *value
}
