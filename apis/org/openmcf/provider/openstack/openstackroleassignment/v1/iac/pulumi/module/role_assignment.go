package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/identity"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// roleAssignment provisions the OpenStack Identity role assignment and exports outputs.
func roleAssignment(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackRoleAssignment.Spec
	resourceName := locals.OpenStackRoleAssignment.Metadata.Name

	raArgs := &identity.RoleAssignmentArgs{
		RoleId: pulumi.String(spec.RoleId),
	}

	// Scope: project_id (FK) or domain_id (plain string) -- exactly one.
	if spec.ProjectId != nil {
		// The platform middleware resolves valueFrom references before IaC modules run,
		// so .GetValue() always returns the resolved literal string.
		raArgs.ProjectId = pulumi.StringPtr(spec.ProjectId.GetValue())
	} else if spec.DomainId != "" {
		raArgs.DomainId = pulumi.StringPtr(spec.DomainId)
	}

	// Principal: user_id or group_id -- exactly one.
	if spec.UserId != "" {
		raArgs.UserId = pulumi.StringPtr(spec.UserId)
	} else if spec.GroupId != "" {
		raArgs.GroupId = pulumi.StringPtr(spec.GroupId)
	}

	// Set region override if provided.
	if spec.Region != "" {
		raArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdRA, err := identity.NewRoleAssignment(
		ctx,
		strings.ToLower(resourceName),
		raArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack role assignment")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpId, createdRA.ID())
	ctx.Export(OpRoleId, createdRA.RoleId)
	ctx.Export(OpProjectId, createdRA.ProjectId)
	ctx.Export(OpDomainId, createdRA.DomainId)
	ctx.Export(OpUserId, createdRA.UserId)
	ctx.Export(OpGroupId, createdRA.GroupId)
	ctx.Export(OpRegion, createdRA.Region)

	return nil
}
