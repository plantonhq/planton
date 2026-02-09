package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// port provisions the OpenStack Neutron port and exports outputs.
func port(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackNetworkPort.Spec
	portName := locals.OpenStackNetworkPort.Metadata.Name

	portArgs := &networking.PortArgs{
		Name:      pulumi.String(portName),
		NetworkId: pulumi.String(locals.NetworkId),
	}

	// Build fixed_ips from spec.
	// Each FixedIp entry assigns an IP from a subnet on the port's network.
	if len(spec.FixedIps) > 0 {
		fixedIps := make(networking.PortFixedIpArray, 0, len(spec.FixedIps))
		for _, fip := range spec.FixedIps {
			entry := networking.PortFixedIpArgs{}

			// Extract subnet_id from StringValueOrRef if present.
			if fip.SubnetId != nil {
				entry.SubnetId = pulumi.StringPtr(fip.SubnetId.GetValue())
			}

			// Set specific IP address if requested.
			if fip.IpAddress != "" {
				entry.IpAddress = pulumi.StringPtr(fip.IpAddress)
			}

			fixedIps = append(fixedIps, entry)
		}
		portArgs.FixedIps = fixedIps
	}

	// Set security group IDs from resolved repeated StringValueOrRef.
	if len(locals.SecurityGroupIds) > 0 {
		sgIds := make(pulumi.StringArray, len(locals.SecurityGroupIds))
		for i, sgId := range locals.SecurityGroupIds {
			sgIds[i] = pulumi.String(sgId)
		}
		portArgs.SecurityGroupIds = sgIds
	}

	// Set no_security_groups to explicitly remove all SGs including the default.
	if spec.NoSecurityGroups {
		portArgs.NoSecurityGroups = pulumi.BoolPtr(true)
	}

	// Set admin_state_up. The middleware guarantees the default (true) is applied,
	// so GetAdminStateUp() always returns a usable value.
	if spec.AdminStateUp != nil {
		portArgs.AdminStateUp = pulumi.BoolPtr(spec.GetAdminStateUp())
	}

	// Set specific MAC address if requested (ForceNew).
	if spec.MacAddress != "" {
		portArgs.MacAddress = pulumi.StringPtr(spec.MacAddress)
	}

	// Set port_security_enabled if explicitly provided.
	if spec.PortSecurityEnabled != nil {
		portArgs.PortSecurityEnabled = pulumi.BoolPtr(spec.GetPortSecurityEnabled())
	}

	// Set description if provided.
	if spec.Description != "" {
		portArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		portArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		portArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdPort, err := networking.NewPort(
		ctx,
		strings.ToLower(portName),
		portArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack port")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpPortId, createdPort.ID())
	ctx.Export(OpMacAddress, createdPort.MacAddress)
	ctx.Export(OpAllFixedIps, createdPort.AllFixedIps)
	ctx.Export(OpAllSecurityGroupIds, createdPort.AllSecurityGroupIds)
	ctx.Export(OpRegion, createdPort.Region)

	return nil
}
