package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/efs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// AccessPointResult carries the created access point for the exports.
type AccessPointResult struct {
	AccessPoint *efs.AccessPoint
}

func accessPoint(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*AccessPointResult, error) {
	spec := locals.AwsEfsAccessPoint.Spec

	// The ENTIRE access point is create-time immutable (only tags mutate):
	// changing the file system, POSIX user, or root directory replaces it.
	// AWS assigns the fsap- identity; the Name tag carries the human name.
	args := &efs.AccessPointArgs{
		FileSystemId: pulumi.String(spec.FileSystemId.GetValue()),
		Tags:         pulumi.ToStringMap(locals.AwsTags),
	}

	// POSIX identity enforcement: when set, every file operation through this
	// access point uses this UID/GID regardless of what the NFS client claims
	// — the core least-privilege mechanism of access points.
	if spec.PosixUser != nil {
		posixUser := &efs.AccessPointPosixUserArgs{
			Uid: pulumi.Int(int(spec.PosixUser.GetUid())),
			Gid: pulumi.Int(int(spec.PosixUser.GetGid())),
		}
		if len(spec.PosixUser.SecondaryGids) > 0 {
			var secondaryGids pulumi.IntArray
			for _, gid := range spec.PosixUser.SecondaryGids {
				secondaryGids = append(secondaryGids, pulumi.Int(int(gid)))
			}
			posixUser.SecondaryGids = secondaryGids
		}
		args.PosixUser = posixUser
	}

	// Root directory restriction: the path is exposed as "/" to clients.
	// creation_info lets EFS create a not-yet-existing path with the right
	// ownership on first mount — without it, mounting a missing path fails.
	if spec.RootDirectory != nil {
		rootDir := &efs.AccessPointRootDirectoryArgs{
			Path: pulumi.StringPtr(spec.RootDirectory.Path),
		}
		if spec.RootDirectory.CreationInfo != nil {
			rootDir.CreationInfo = &efs.AccessPointRootDirectoryCreationInfoArgs{
				OwnerUid:    pulumi.Int(int(spec.RootDirectory.CreationInfo.OwnerUid)),
				OwnerGid:    pulumi.Int(int(spec.RootDirectory.CreationInfo.OwnerGid)),
				Permissions: pulumi.String(spec.RootDirectory.CreationInfo.Permissions),
			}
		}
		args.RootDirectory = rootDir
	}

	// Stable logical name; the human identity travels through the Name tag.
	createdAccessPoint, err := efs.NewAccessPoint(ctx, "access-point", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create efs access point")
	}

	return &AccessPointResult{AccessPoint: createdAccessPoint}, nil
}
