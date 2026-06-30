package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/network"
)

// gateway provisions the complete Public Gateway composite:
//  1. A dedicated Flexible IP (public IPv4 address).
//  2. The Public Gateway appliance (attached to the IP).
//  3. A GatewayNetwork binding (connecting the gateway to the Private Network).
//  4. Optional PAT (port forwarding) rules.
//
// All outputs are exported for downstream resource references.
func gateway(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) error {
	spec := locals.ScalewayPublicGateway.Spec

	// ── 1. Create the Flexible IP ──────────────────────────────────────────
	//
	// A dedicated public IPv4 address for the gateway. Creating it as a
	// separate resource (rather than letting the gateway auto-create one)
	// gives us explicit control over the IP lifecycle -- the IP can be
	// preserved and reassigned if the gateway is replaced.

	ipArgs := &network.PublicGatewayIpArgs{
		Tags: pulumi.ToStringArray(locals.ScalewayTags),
		Zone: pulumi.String(spec.Zone),
	}

	// Set reverse DNS if specified.
	if spec.ReverseDns != "" {
		ipArgs.Reverse = pulumi.String(spec.ReverseDns)
	}

	createdIp, err := network.NewPublicGatewayIp(
		ctx,
		"gateway-ip",
		ipArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create public gateway ip")
	}

	// ── 2. Create the Public Gateway ───────────────────────────────────────
	//
	// The managed network appliance that provides NAT, SSH bastion, and
	// port forwarding for the attached Private Network.

	gwArgs := &network.PublicGatewayArgs{
		Name: pulumi.String(locals.ScalewayPublicGateway.Metadata.Name),
		Type: pulumi.String(spec.Type),
		IpId: createdIp.ID(),
		Tags: pulumi.ToStringArray(locals.ScalewayTags),
		Zone: pulumi.String(spec.Zone),
	}

	// Enable SMTP if requested.
	if spec.EnableSmtp {
		gwArgs.EnableSmtp = pulumi.Bool(true)
	}

	// Configure SSH bastion if specified.
	if spec.Bastion != nil && spec.Bastion.Enabled {
		gwArgs.BastionEnabled = pulumi.Bool(true)

		if spec.Bastion.Port > 0 {
			gwArgs.BastionPort = pulumi.Int(int(spec.Bastion.Port))
		}

		if len(spec.Bastion.AllowedIpRanges) > 0 {
			gwArgs.AllowedIpRanges = pulumi.ToStringArray(spec.Bastion.AllowedIpRanges)
		}
	}

	createdGateway, err := network.NewPublicGateway(
		ctx,
		"gateway",
		gwArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create public gateway")
	}

	// ── 3. Create the GatewayNetwork attachment ────────────────────────────
	//
	// This is the glue resource that connects the gateway to the Private
	// Network. Without it, the gateway exists but serves no network.
	// The enable_masquerade flag controls whether NAT is active.

	gnArgs := &network.GatewayNetworkArgs{
		GatewayId:        createdGateway.ID(),
		PrivateNetworkId: pulumi.String(locals.PrivateNetworkId),
		EnableMasquerade: pulumi.Bool(spec.EnableMasquerade),
		Zone:             pulumi.String(spec.Zone),
	}

	createdGatewayNetwork, err := network.NewGatewayNetwork(
		ctx,
		"gateway-network",
		gnArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create gateway network attachment")
	}

	// ── 4. Create PAT rules (optional) ─────────────────────────────────────
	//
	// Port forwarding rules that map public ports on the gateway's IP to
	// private IP:port pairs inside the attached Private Network.

	for i, rule := range spec.PatRules {
		protocol := rule.Protocol
		if protocol == "" {
			protocol = "both"
		}

		_, err := network.NewPublicGatewayPatRule(
			ctx,
			fmt.Sprintf("pat-rule-%d", i),
			&network.PublicGatewayPatRuleArgs{
				GatewayId:   createdGateway.ID(),
				PrivateIp:   pulumi.String(rule.PrivateIp),
				PrivatePort: pulumi.Int(int(rule.PrivatePort)),
				PublicPort:  pulumi.Int(int(rule.PublicPort)),
				Protocol:    pulumi.String(protocol),
				Zone:        pulumi.String(spec.Zone),
			},
			pulumi.Provider(scalewayProvider),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create pat rule %d (public port %d -> %s:%d)",
				i, rule.PublicPort, rule.PrivateIp, rule.PrivatePort)
		}
	}

	// ── 5. Export stack outputs ─────────────────────────────────────────────

	ctx.Export(OpGatewayId, createdGateway.ID())
	ctx.Export(OpPublicIpAddress, createdIp.Address)
	ctx.Export(OpPublicIpId, createdIp.ID())
	ctx.Export(OpGatewayNetworkId, createdGatewayNetwork.ID())

	return nil
}
