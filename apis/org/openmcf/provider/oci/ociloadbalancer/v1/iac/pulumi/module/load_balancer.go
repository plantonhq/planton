package module

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createLoadBalancer(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*loadbalancer.LoadBalancer, error) {
	spec := locals.OciLoadBalancer.Spec

	subnetIds := make(pulumi.StringArray, len(spec.SubnetIds))
	for i, s := range spec.SubnetIds {
		subnetIds[i] = pulumi.String(s.GetValue())
	}

	args := &loadbalancer.LoadBalancerArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:   pulumi.String(locals.DisplayName),
		Shape:         pulumi.String(spec.Shape),
		SubnetIds:     subnetIds,
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
		IsPrivate:     pulumi.Bool(spec.IsPrivate),
	}

	if spec.ShapeDetails != nil {
		args.ShapeDetails = &loadbalancer.LoadBalancerShapeDetailsArgs{
			MinimumBandwidthInMbps: pulumi.Int(int(spec.ShapeDetails.MinimumBandwidthInMbps)),
			MaximumBandwidthInMbps: pulumi.Int(int(spec.ShapeDetails.MaximumBandwidthInMbps)),
		}
	}

	if len(spec.NetworkSecurityGroupIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(spec.NetworkSecurityGroupIds))
		for i, n := range spec.NetworkSecurityGroupIds {
			nsgIds[i] = pulumi.String(n.GetValue())
		}
		args.NetworkSecurityGroupIds = nsgIds
	}

	if spec.IsDeleteProtectionEnabled {
		args.IsDeleteProtectionEnabled = pulumi.Bool(true)
	}

	if spec.IpMode != "" {
		args.IpMode = pulumi.StringPtr(spec.IpMode)
	}

	if len(spec.ReservedIps) > 0 {
		reservedIps := make(loadbalancer.LoadBalancerReservedIpArray, len(spec.ReservedIps))
		for i, rip := range spec.ReservedIps {
			reservedIps[i] = &loadbalancer.LoadBalancerReservedIpArgs{
				Id: pulumi.StringPtr(rip.Id),
			}
		}
		args.ReservedIps = reservedIps
	}

	if spec.IsRequestIdEnabled {
		args.IsRequestIdEnabled = pulumi.Bool(true)
	}
	if spec.RequestIdHeader != "" {
		args.RequestIdHeader = pulumi.StringPtr(spec.RequestIdHeader)
	}

	createdLb, err := loadbalancer.NewLoadBalancer(ctx, "load-balancer", args, pulumiOciOpt(provider))
	if err != nil {
		return nil, fmt.Errorf("failed to create load balancer: %w", err)
	}

	ctx.Export(OpLoadBalancerId, createdLb.ID())

	ctx.Export(OpIpAddresses, createdLb.IpAddressDetails.ApplyT(func(details []loadbalancer.LoadBalancerIpAddressDetail) string {
		var ips []string
		for _, d := range details {
			if d.IpAddress != nil {
				ips = append(ips, *d.IpAddress)
			}
		}
		return strings.Join(ips, ",")
	}).(pulumi.StringOutput))

	return createdLb, nil
}
