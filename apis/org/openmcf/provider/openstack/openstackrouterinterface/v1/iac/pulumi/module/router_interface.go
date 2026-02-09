package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// routerInterface provisions the OpenStack Neutron router interface and exports outputs.
func routerInterface(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackRouterInterface.Spec
	resourceName := locals.OpenStackRouterInterface.Metadata.Name

	routerInterfaceArgs := &networking.RouterInterfaceArgs{
		RouterId: pulumi.String(locals.RouterId),
		SubnetId: pulumi.StringPtr(locals.SubnetId),
	}

	// Set region override if provided.
	if spec.Region != "" {
		routerInterfaceArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdRouterInterface, err := networking.NewRouterInterface(
		ctx,
		strings.ToLower(resourceName),
		routerInterfaceArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack router interface")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	// The Pulumi resource ID for a router interface is the port_id.
	ctx.Export(OpPortId, createdRouterInterface.PortId)
	ctx.Export(OpRouterId, createdRouterInterface.RouterId)
	ctx.Export(OpSubnetId, createdRouterInterface.SubnetId)
	ctx.Export(OpRegion, createdRouterInterface.Region)

	return nil
}
