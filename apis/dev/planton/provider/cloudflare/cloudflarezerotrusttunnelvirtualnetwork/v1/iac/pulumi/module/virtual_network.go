package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// virtualNetwork provisions a Cloudflare Tunnel virtual network: an isolated routing
// segment that lets overlapping private CIDRs be reached through separate tunnels.
func virtualNetwork(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareZeroTrustTunnelVirtualNetwork.Spec

	vnArgs := &cloudflare.ZeroTrustTunnelCloudflaredVirtualNetworkArgs{
		AccountId:        pulumi.String(spec.AccountId),
		Name:             pulumi.String(spec.Name),
		IsDefaultNetwork: pulumi.Bool(spec.GetIsDefaultNetwork()),
	}
	if spec.Comment != "" {
		vnArgs.Comment = pulumi.String(spec.Comment)
	}

	createdVirtualNetwork, err := cloudflare.NewZeroTrustTunnelCloudflaredVirtualNetwork(
		ctx,
		"virtual-network",
		vnArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare tunnel virtual network")
	}

	ctx.Export(OpVirtualNetworkId, createdVirtualNetwork.ID())
	ctx.Export(OpVirtualNetworkName, createdVirtualNetwork.Name)

	return nil
}
