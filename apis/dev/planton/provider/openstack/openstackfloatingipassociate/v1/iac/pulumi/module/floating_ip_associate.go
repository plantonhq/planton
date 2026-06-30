package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// floatingIpAssociate provisions the OpenStack Neutron floating IP association
// and exports outputs.
func floatingIpAssociate(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackFloatingIpAssociate.Spec
	resourceName := locals.OpenStackFloatingIpAssociate.Metadata.Name

	associateArgs := &networking.FloatingIpAssociateArgs{
		FloatingIp: pulumi.String(locals.FloatingIp),
		PortId:     pulumi.String(locals.PortId),
	}

	// Set fixed_ip if provided (for multi-IP ports).
	if spec.FixedIp != "" {
		associateArgs.FixedIp = pulumi.StringPtr(spec.FixedIp)
	}

	// Set region override if provided (ForceNew).
	if spec.Region != "" {
		associateArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdAssociate, err := networking.NewFloatingIpAssociate(
		ctx,
		strings.ToLower(resourceName),
		associateArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack floating ip association")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpId, createdAssociate.ID())
	ctx.Export(OpFloatingIp, createdAssociate.FloatingIp)
	ctx.Export(OpPortId, createdAssociate.PortId)
	ctx.Export(OpFixedIp, createdAssociate.FixedIp)
	ctx.Export(OpRegion, createdAssociate.Region)

	return nil
}
