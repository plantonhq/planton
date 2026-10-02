package module

import (
	"strings"

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

	// The provider's `name` is tagValues/{numeric_id}; the bare numeric id
	// is derived from it so both engines export the same three handles.
	ctx.Export(OpName, created.Name)
	ctx.Export(OpNamespacedName, created.NamespacedName)
	ctx.Export(OpTagValueId, created.Name.ApplyT(func(name string) string {
		return strings.TrimPrefix(name, "tagValues/")
	}).(pulumi.StringOutput))
	ctx.Export(OpCreateTime, created.CreateTime)

	return nil
}
