package module

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/networkloadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createNetworkLoadBalancer(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*networkloadbalancer.NetworkLoadBalancer, error) {
	spec := locals.OciNetworkLoadBalancer.Spec

	args := &networkloadbalancer.NetworkLoadBalancerArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:   pulumi.String(locals.DisplayName),
		SubnetId:      pulumi.String(spec.SubnetId.GetValue()),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.IsPrivate {
		args.IsPrivate = pulumi.BoolPtr(true)
	}

	if spec.IsPreserveSourceDestination {
		args.IsPreserveSourceDestination = pulumi.BoolPtr(true)
	}

	if spec.IsSymmetricHashEnabled {
		args.IsSymmetricHashEnabled = pulumi.BoolPtr(true)
	}

	if len(spec.NetworkSecurityGroupIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(spec.NetworkSecurityGroupIds))
		for i, n := range spec.NetworkSecurityGroupIds {
			nsgIds[i] = pulumi.String(n.GetValue())
		}
		args.NetworkSecurityGroupIds = nsgIds
	}

	if spec.NlbIpVersion != "" {
		args.NlbIpVersion = pulumi.StringPtr(spec.NlbIpVersion)
	}

	if len(spec.ReservedIps) > 0 {
		reservedIps := make(networkloadbalancer.NetworkLoadBalancerReservedIpArray, len(spec.ReservedIps))
		for i, rip := range spec.ReservedIps {
			reservedIps[i] = &networkloadbalancer.NetworkLoadBalancerReservedIpArgs{
				Id: pulumi.StringPtr(rip.Id),
			}
		}
		args.ReservedIps = reservedIps
	}

	if spec.AssignedIpv6 != "" {
		args.AssignedIpv6 = pulumi.StringPtr(spec.AssignedIpv6)
	}

	if spec.AssignedPrivateIpv4 != "" {
		args.AssignedPrivateIpv4 = pulumi.StringPtr(spec.AssignedPrivateIpv4)
	}

	if spec.SubnetIpv6Cidr != "" {
		args.SubnetIpv6cidr = pulumi.StringPtr(spec.SubnetIpv6Cidr)
	}

	createdNlb, err := networkloadbalancer.NewNetworkLoadBalancer(ctx, "network-load-balancer", args, pulumiOciOpt(provider))
	if err != nil {
		return nil, fmt.Errorf("failed to create network load balancer: %w", err)
	}

	ctx.Export(OpNetworkLoadBalancerId, createdNlb.ID())

	ctx.Export(OpIpAddresses, createdNlb.IpAddresses.ApplyT(func(addresses []networkloadbalancer.NetworkLoadBalancerIpAddress) string {
		var ips []string
		for _, addr := range addresses {
			if addr.IpAddress != nil {
				ips = append(ips, *addr.IpAddress)
			}
		}
		return strings.Join(ips, ",")
	}).(pulumi.StringOutput))

	return createdNlb, nil
}
