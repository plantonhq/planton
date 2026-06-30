package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serverGroup provisions the OpenStack Compute server group and exports outputs.
func serverGroup(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackServerGroup.Spec
	resourceName := locals.OpenStackServerGroup.Metadata.Name

	// The Pulumi SDK models policies as a single string (matching the
	// reality that only one policy is allowed per server group).
	serverGroupArgs := &compute.ServerGroupArgs{
		Name:     pulumi.String(resourceName),
		Policies: pulumi.String(spec.Policy),
	}

	// Set region override if provided.
	if spec.Region != "" {
		serverGroupArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdServerGroup, err := compute.NewServerGroup(
		ctx,
		strings.ToLower(resourceName),
		serverGroupArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack server group")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpServerGroupId, createdServerGroup.ID())
	ctx.Export(OpName, createdServerGroup.Name)
	ctx.Export(OpMembers, createdServerGroup.Members)
	ctx.Export(OpRegion, createdServerGroup.Region)

	return nil
}
