package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/tags"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tagKey provisions the Resource Manager tag key.
//
// The owner, short name, purpose, and purpose data are immutable -- a
// change to any recreates the key, which Google refuses while the key has
// values -- while the description and allowed-values regex update in
// place. Optional inputs are sent only when set so the provider's defaults
// stay the provider's; deletion_policy likewise.
func tagKey(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpTagKey.Spec

	args := &tags.TagKeyArgs{
		Parent:    pulumi.String(locals.Parent),
		ShortName: pulumi.String(locals.ShortName),
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Purpose != "" {
		args.Purpose = pulumi.StringPtr(spec.Purpose)
	}
	if len(spec.PurposeData) > 0 {
		args.PurposeData = pulumi.ToStringMap(spec.PurposeData)
	}
	if spec.AllowedValuesRegex != "" {
		args.AllowedValuesRegex = pulumi.StringPtr(spec.AllowedValuesRegex)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := tags.NewTagKey(ctx, locals.GcpTagKey.Metadata.Name, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create tag key")
	}

	// The resource's ID is tagKeys/{numeric_id}, the resource name every
	// reference takes; the provider's `name` is the bare number Google
	// generated. Both engines export these same three handles.
	ctx.Export(OpName, created.ID().ToStringOutput())
	ctx.Export(OpNamespacedName, created.NamespacedName)
	ctx.Export(OpTagKeyId, created.Name)
	ctx.Export(OpCreateTime, created.CreateTime)

	return nil
}
