package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// route advertises a private CIDR as reachable through a tunnel, within a virtual
// network. The tunnel and virtual network are resolved from StringValueOrRef inputs.
func route(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareZeroTrustTunnelRoute.Spec

	routeArgs := &cloudflare.ZeroTrustTunnelCloudflaredRouteArgs{
		AccountId: pulumi.String(spec.AccountId),
		Network:   pulumi.String(spec.Network),
		TunnelId:  pulumi.String(spec.TunnelId.GetValue()),
	}
	if spec.VirtualNetworkId != nil && spec.VirtualNetworkId.GetValue() != "" {
		routeArgs.VirtualNetworkId = pulumi.String(spec.VirtualNetworkId.GetValue())
	}
	if spec.Comment != "" {
		routeArgs.Comment = pulumi.String(spec.Comment)
	}

	createdRoute, err := cloudflare.NewZeroTrustTunnelCloudflaredRoute(
		ctx,
		"route",
		routeArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare tunnel route")
	}

	ctx.Export(OpRouteId, createdRoute.ID())
	ctx.Export(OpNetwork, createdRoute.Network)

	return nil
}
