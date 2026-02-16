package module

import (
	"github.com/plantonhq/openmcf/internal/valuefrom"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nlb creates an AWS Network Load Balancer from the spec's subnet mappings
// and top-level configuration. Returns the created load balancer resource
// for use by listeners and DNS.
func nlb(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*lb.LoadBalancer, error) {
	spec := locals.Nlb.Spec

	// Build subnet mapping arguments from the spec.
	subnetMappings := lb.LoadBalancerSubnetMappingArray{}
	for _, sm := range spec.SubnetMappings {
		mapping := lb.LoadBalancerSubnetMappingArgs{
			SubnetId: pulumi.String(sm.SubnetId.GetValue()),
		}
		if sm.AllocationId != nil && sm.AllocationId.GetValue() != "" {
			mapping.AllocationId = pulumi.StringPtr(sm.AllocationId.GetValue())
		}
		if sm.PrivateIpv4Address != "" {
			mapping.PrivateIpv4Address = pulumi.StringPtr(sm.PrivateIpv4Address)
		}
		subnetMappings = append(subnetMappings, mapping)
	}

	// Determine IP address type. Default to ipv4 when not specified.
	ipAddressType := "ipv4"
	if spec.IpAddressType != "" {
		ipAddressType = spec.IpAddressType
	}

	args := &lb.LoadBalancerArgs{
		Name:                     pulumi.String(locals.Nlb.Metadata.Name),
		LoadBalancerType:         pulumi.String("network"),
		Internal:                 pulumi.Bool(spec.Internal),
		IpAddressType:            pulumi.String(ipAddressType),
		EnableDeletionProtection: pulumi.Bool(spec.DeleteProtectionEnabled),
		SubnetMappings:           subnetMappings,
		Tags:                     pulumi.ToStringMap(locals.AwsTags),
	}

	// Security groups are optional for NLB.
	if len(spec.SecurityGroups) > 0 {
		args.SecurityGroups = pulumi.ToStringArray(valuefrom.ToStringArray(spec.SecurityGroups))
	}

	// Cross-zone load balancing (default false for NLB).
	if spec.CrossZoneLoadBalancingEnabled {
		args.EnableCrossZoneLoadBalancing = pulumi.Bool(true)
	}

	// DNS record client routing policy (NLB-specific).
	if spec.DnsRecordClientRoutingPolicy != "" {
		args.DnsRecordClientRoutingPolicy = pulumi.StringPtr(spec.DnsRecordClientRoutingPolicy)
	}

	createdNlb, err := lb.NewLoadBalancer(ctx, locals.Nlb.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Network Load Balancer")
	}

	// Export NLB-level outputs.
	ctx.Export(OpLoadBalancerArn, createdNlb.Arn)
	ctx.Export(OpLoadBalancerName, createdNlb.Name)
	ctx.Export(OpLoadBalancerDnsName, createdNlb.DnsName)
	ctx.Export(OpLoadBalancerHostedZoneId, createdNlb.ZoneId)

	return createdNlb, nil
}
