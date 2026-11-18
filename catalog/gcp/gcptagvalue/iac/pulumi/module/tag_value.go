package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/tags"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tagValue provisions the Resource Manager tag value under its key.
//
// The key (the provider's `parent`, tagKeys/{id} -- the GcpTagKey name
// output) and the short name are immutable; only the description updates
// in place. Optional inputs are sent only when set so the provider's
// defaults stay the provider's; deletion_policy likewise.
func tagValue(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpTagValue.Spec

	args := &tags.TagValueArgs{
		Parent:    pulumi.String(spec.TagKey.GetValue()),
		ShortName: pulumi.String(locals.ShortName),
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := tags.NewTagValue(ctx, locals.GcpTagValue.Metadata.Name, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create tag value")
	}

	// The resource's ID is tagValues/{numeric_id}, the resource name every
	// reference takes; the provider's `name` is the bare number Google
	// generated. Both engines export these same three handles.
	ctx.Export(OpName, created.ID().ToStringOutput())
	ctx.Export(OpNamespacedName, created.NamespacedName)
	ctx.Export(OpTagValueId, created.Name)
	ctx.Export(OpCreateTime, created.CreateTime)

	return nil
}
