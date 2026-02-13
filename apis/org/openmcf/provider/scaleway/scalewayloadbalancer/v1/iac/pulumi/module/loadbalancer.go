package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/loadbalancers"
)

// lbResources holds references to the core LB resources for downstream use.
type lbResources struct {
	lb *loadbalancers.LoadBalancer
	ip *loadbalancers.Ip
}

// loadBalancer creates the core Load Balancer infrastructure:
//  1. A dedicated Flexible IP (public IPv4 address).
//  2. The Load Balancer appliance (with optional Private Network attachment).
//
// The Flexible IP is created as a separate resource (rather than using
// assign_flexible_ip on the LB) to give explicit lifecycle control. The
// IP survives LB replacement, preserving DNS records and firewall rules.
func loadBalancer(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) (*lbResources, error) {
	spec := locals.ScalewayLoadBalancer.Spec

	// ── 1. Create the Flexible IP ──────────────────────────────────────────
	createdIp, err := loadbalancers.NewIp(
		ctx,
		"lb-ip",
		&loadbalancers.IpArgs{
			Tags: pulumi.ToStringArray(locals.ScalewayTags),
			Zone: pulumi.String(spec.Zone),
		},
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create load balancer flexible ip")
	}

	// ── 2. Create the Load Balancer ────────────────────────────────────────
	lbArgs := &loadbalancers.LoadBalancerArgs{
		Name:   pulumi.String(locals.ScalewayLoadBalancer.Metadata.Name),
		Type:   pulumi.String(spec.Type),
		IpIds:  pulumi.StringArray{createdIp.ID()},
		Tags:   pulumi.ToStringArray(locals.ScalewayTags),
		Zone:   pulumi.String(spec.Zone),
	}

	if spec.Description != "" {
		lbArgs.Description = pulumi.String(spec.Description)
	}

	if spec.SslCompatibilityLevel != "" {
		lbArgs.SslCompatibilityLevel = pulumi.String(spec.SslCompatibilityLevel)
	}

	// Attach to Private Network if specified.
	if locals.PrivateNetworkId != "" {
		lbArgs.PrivateNetworks = loadbalancers.LoadBalancerPrivateNetworkArray{
			&loadbalancers.LoadBalancerPrivateNetworkArgs{
				PrivateNetworkId: pulumi.String(locals.PrivateNetworkId),
			},
		}
	}

	createdLb, err := loadbalancers.NewLoadBalancer(
		ctx,
		"lb",
		lbArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create load balancer")
	}

	// ── 3. Export stack outputs ─────────────────────────────────────────────
	ctx.Export(OpLbId, createdLb.ID())
	ctx.Export(OpLbIpAddress, createdIp.IpAddress)
	ctx.Export(OpLbIpId, createdIp.ID())

	return &lbResources{
		lb: createdLb,
		ip: createdIp,
	}, nil
}
