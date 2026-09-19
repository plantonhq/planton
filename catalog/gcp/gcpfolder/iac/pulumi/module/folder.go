package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// folder provisions the Resource Manager folder.
//
// The parent is mutable: changing it MOVES the folder (Google's
// folders.move) with everything inside it. The create-time tags are the
// one immutable input -- the provider recreates the folder when they
// change, which Google refuses for a non-empty folder; the spec comment
// steers users to GcpTagBinding for everything but create-time tagging.
//
// deletion_protection is always sent explicitly, both when the spec sets
// it and when it falls back to the default of true: the provider's own
// default is also true, but sending it makes the spec the single source
// of truth on both engines and keeps the plan free of implicit state.
// deletion_policy is sent only when set so the provider's default
// (DELETE) stays the provider's.
func folder(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpFolder.Spec

	// The destroy guard: GCP's default of true when the spec is silent.
	// GetDeletionProtection() alone would return false for an unset
	// optional, which would silently disarm the guard -- so presence is
	// checked first. Matches the Terraform module's coalesce.
	deletionProtection := true
	if spec.DeletionProtection != nil {
		deletionProtection = spec.GetDeletionProtection()
	}

	args := &organizations.FolderArgs{
		Parent:             pulumi.String(locals.Parent),
		DisplayName:        pulumi.String(locals.DisplayName),
		DeletionProtection: pulumi.BoolPtr(deletionProtection),
	}
	if len(spec.Tags) > 0 {
		args.Tags = pulumi.ToStringMap(spec.Tags)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdFolder, err := organizations.NewFolder(ctx, locals.GcpFolder.Metadata.Name, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create folder")
	}

	ctx.Export(OpFolderId, createdFolder.FolderId)
	ctx.Export(OpName, createdFolder.Name)
	ctx.Export(OpLifecycleState, createdFolder.LifecycleState)
	ctx.Export(OpCreateTime, createdFolder.CreateTime)

	return nil
}
