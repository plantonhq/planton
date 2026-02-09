package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// floatingIp provisions the OpenStack Neutron floating IP and exports outputs.
func floatingIp(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackFloatingIp.Spec
	fipName := locals.OpenStackFloatingIp.Metadata.Name

	fipArgs := &networking.FloatingIpArgs{
		// "Pool" is the Pulumi/TF name for the external network ID.
		Pool: pulumi.String(locals.FloatingNetworkId),
	}

	// Set port_id for built-in association (optional FK).
	if locals.PortId != "" {
		fipArgs.PortId = pulumi.StringPtr(locals.PortId)
	}

	// Set fixed_ip if provided.
	// Only relevant when port_id is set and the port has multiple IPs.
	if spec.FixedIp != "" {
		fipArgs.FixedIp = pulumi.StringPtr(spec.FixedIp)
	}

	// Set subnet_id for allocation from a specific external subnet.
	if spec.SubnetId != "" {
		fipArgs.SubnetId = pulumi.StringPtr(spec.SubnetId)
	}

	// Set address for requesting a specific floating IP.
	if spec.Address != "" {
		fipArgs.Address = pulumi.StringPtr(spec.Address)
	}

	// Set description if provided.
	if spec.Description != "" {
		fipArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		fipArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		fipArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdFip, err := networking.NewFloatingIp(
		ctx,
		strings.ToLower(fipName),
		fipArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack floating ip")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpFloatingIpId, createdFip.ID())
	ctx.Export(OpAddress, createdFip.Address)
	ctx.Export(OpFloatingNetworkId, createdFip.Pool)
	ctx.Export(OpPortId, createdFip.PortId)
	ctx.Export(OpFixedIp, createdFip.FixedIp)
	ctx.Export(OpRegion, createdFip.Region)

	return nil
}
