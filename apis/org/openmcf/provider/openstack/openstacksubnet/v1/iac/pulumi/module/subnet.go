package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// subnet provisions the OpenStack Neutron subnet and exports outputs.
func subnet(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackSubnet.Spec
	subnetName := locals.OpenStackSubnet.Metadata.Name

	subnetArgs := &networking.SubnetArgs{
		Name:      pulumi.String(subnetName),
		NetworkId: pulumi.String(locals.NetworkId),
		Cidr:      pulumi.StringPtr(spec.Cidr),
	}

	// Set IP version if explicitly provided (default is handled by middleware).
	if spec.IpVersion != nil {
		subnetArgs.IpVersion = pulumi.IntPtr(int(spec.GetIpVersion()))
	}

	// Set gateway_ip if provided.
	if spec.GatewayIp != "" {
		subnetArgs.GatewayIp = pulumi.StringPtr(spec.GatewayIp)
	}

	// Set no_gateway if true.
	if spec.NoGateway {
		subnetArgs.NoGateway = pulumi.BoolPtr(true)
	}

	// Set enable_dhcp. The middleware guarantees the default (true) is applied,
	// so GetEnableDhcp() always returns a usable value.
	if spec.EnableDhcp != nil {
		subnetArgs.EnableDhcp = pulumi.BoolPtr(spec.GetEnableDhcp())
	}

	// Set DNS nameservers if provided.
	if len(spec.DnsNameservers) > 0 {
		nameservers := make(pulumi.StringArray, len(spec.DnsNameservers))
		for i, ns := range spec.DnsNameservers {
			nameservers[i] = pulumi.String(ns)
		}
		subnetArgs.DnsNameservers = nameservers
	}

	// Set allocation pools if provided.
	if len(spec.AllocationPools) > 0 {
		pools := make(networking.SubnetAllocationPoolArray, len(spec.AllocationPools))
		for i, pool := range spec.AllocationPools {
			pools[i] = &networking.SubnetAllocationPoolArgs{
				Start: pulumi.String(pool.Start),
				End:   pulumi.String(pool.End),
			}
		}
		subnetArgs.AllocationPools = pools
	}

	// Set description if provided.
	if spec.Description != "" {
		subnetArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		subnetArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		subnetArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdSubnet, err := networking.NewSubnet(
		ctx,
		strings.ToLower(subnetName),
		subnetArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack subnet")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpSubnetId, createdSubnet.ID())
	ctx.Export(OpName, createdSubnet.Name)
	ctx.Export(OpCidr, createdSubnet.Cidr)
	ctx.Export(OpGatewayIp, createdSubnet.GatewayIp)
	ctx.Export(OpNetworkId, createdSubnet.NetworkId)
	ctx.Export(OpRegion, createdSubnet.Region)

	return nil
}
