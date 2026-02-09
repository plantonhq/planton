package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// network provisions the OpenStack Neutron network and exports outputs.
func network(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackNetwork.Spec
	networkName := locals.OpenStackNetwork.Metadata.Name

	networkArgs := &networking.NetworkArgs{
		Name: pulumi.String(networkName),
	}

	// Set description if provided.
	if spec.Description != "" {
		networkArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set admin_state_up. The middleware guarantees the default (true) is applied,
	// so GetAdminStateUp() always returns a usable value.
	if spec.AdminStateUp != nil {
		networkArgs.AdminStateUp = pulumi.BoolPtr(spec.GetAdminStateUp())
	}

	// Set shared if true (default false is the proto3 zero value).
	if spec.Shared {
		networkArgs.Shared = pulumi.BoolPtr(true)
	}

	// Set external if true (default false is the proto3 zero value).
	if spec.External {
		networkArgs.External = pulumi.BoolPtr(true)
	}

	// Set MTU if provided (0 means unset).
	if spec.Mtu > 0 {
		networkArgs.Mtu = pulumi.IntPtr(int(spec.Mtu))
	}

	// Set DNS domain if provided.
	if spec.DnsDomain != "" {
		networkArgs.DnsDomain = pulumi.StringPtr(spec.DnsDomain)
	}

	// Set port_security_enabled if explicitly provided.
	if spec.PortSecurityEnabled != nil {
		networkArgs.PortSecurityEnabled = pulumi.BoolPtr(spec.GetPortSecurityEnabled())
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, len(spec.Tags))
		for i, tag := range spec.Tags {
			tags[i] = pulumi.String(tag)
		}
		networkArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		networkArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdNetwork, err := networking.NewNetwork(
		ctx,
		strings.ToLower(networkName),
		networkArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack network")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpNetworkId, createdNetwork.ID())
	ctx.Export(OpName, createdNetwork.Name)
	ctx.Export(OpRegion, createdNetwork.Region)

	return nil
}
