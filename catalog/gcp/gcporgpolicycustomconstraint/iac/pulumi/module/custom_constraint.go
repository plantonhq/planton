package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/orgpolicy"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// customConstraint provisions the organization's custom constraint.
//
// The name (with its `custom.` prefix), the parent organization, and the
// resource types are immutable -- a change to any recreates the
// constraint, and every policy enforcing the old name lapses -- while the
// condition, action, methods, display name, and description update in
// place. Optional strings are sent only when set; deletion_policy is sent
// only when set so the provider's default (DELETE) stays the provider's.
func customConstraint(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpOrgPolicyCustomConstraint.Spec

	args := &orgpolicy.CustomConstraintArgs{
		Name:          pulumi.String(locals.ConstraintName),
		Parent:        pulumi.String(locals.Parent),
		ResourceTypes: pulumi.ToStringArray(spec.ResourceTypes),
		MethodTypes:   pulumi.ToStringArray(spec.MethodTypes),
		Condition:     pulumi.String(spec.Condition),
		ActionType:    pulumi.String(spec.ActionType),
	}
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := orgpolicy.NewCustomConstraint(ctx, locals.GcpOrgPolicyCustomConstraint.Metadata.Name, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create custom constraint")
	}

	// The resource ID is the full name Google addresses the constraint by
	// ({parent}/customConstraints/custom.{name}); the provider's `name`
	// attribute is the `custom.{name}` handle a policy enforces.
	ctx.Export(OpName, created.ID().ToStringOutput())
	ctx.Export(OpConstraint, created.Name)
	ctx.Export(OpUpdateTime, created.UpdateTime)

	return nil
}
