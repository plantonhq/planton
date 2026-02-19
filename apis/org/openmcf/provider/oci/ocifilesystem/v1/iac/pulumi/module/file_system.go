package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/filestorage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func fileSystem(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*filestorage.FileSystem, error) {
	spec := locals.OciFileSystem.Spec

	fsArgs := &filestorage.FileSystemArgs{
		CompartmentId:      pulumi.String(spec.CompartmentId.GetValue()),
		AvailabilityDomain: pulumi.String(spec.AvailabilityDomain),
		FreeformTags:       pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.DisplayName != "" {
		fsArgs.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}

	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		fsArgs.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	if spec.FilesystemSnapshotPolicyId != nil && spec.FilesystemSnapshotPolicyId.GetValue() != "" {
		fsArgs.FilesystemSnapshotPolicyId = pulumi.StringPtr(spec.FilesystemSnapshotPolicyId.GetValue())
	}

	createdFs, err := filestorage.NewFileSystem(ctx, locals.DisplayName, fsArgs, pulumiOciOpt(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file system")
	}

	ctx.Export(OpFileSystemId, createdFs.ID())

	return createdFs, nil
}
