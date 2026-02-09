package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/identity"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// project provisions the OpenStack Identity project and exports outputs.
func project(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackProject.Spec
	resourceName := locals.OpenStackProject.Metadata.Name

	projectArgs := &identity.ProjectArgs{
		Name:        pulumi.String(resourceName),
		Description: pulumi.StringPtr(spec.Description),
	}

	// Set domain_id if provided.
	if spec.DomainId != "" {
		projectArgs.DomainId = pulumi.StringPtr(spec.DomainId)
	}

	// Set enabled. Middleware guarantees the default (true) is applied,
	// so GetEnabled() always returns a valid value.
	projectArgs.Enabled = pulumi.BoolPtr(spec.GetEnabled())

	// Set parent_id if provided.
	if spec.ParentId != "" {
		projectArgs.ParentId = pulumi.StringPtr(spec.ParentId)
	}

	// Set tags if provided.
	if len(spec.Tags) > 0 {
		tags := pulumi.StringArray{}
		for _, tag := range spec.Tags {
			tags = append(tags, pulumi.String(tag))
		}
		projectArgs.Tags = tags
	}

	// Set region override if provided.
	if spec.Region != "" {
		projectArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdProject, err := identity.NewProject(
		ctx,
		strings.ToLower(resourceName),
		projectArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack identity project")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpProjectId, createdProject.ID())
	ctx.Export(OpName, createdProject.Name)
	ctx.Export(OpDomainId, createdProject.DomainId)
	ctx.Export(OpEnabled, createdProject.Enabled)
	ctx.Export(OpRegion, createdProject.Region)

	return nil
}
