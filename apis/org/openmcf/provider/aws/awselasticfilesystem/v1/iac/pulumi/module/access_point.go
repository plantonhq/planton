package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/efs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// AccessPointResults holds the per-name access point outputs.
type AccessPointResults struct {
	AccessPointIds  pulumi.StringMap
	AccessPointArns pulumi.StringMap
}

func accessPoints(ctx *pulumi.Context, locals *Locals, provider *aws.Provider, fs *efs.FileSystem) (*AccessPointResults, error) {
	spec := locals.AwsElasticFileSystem.Spec

	results := &AccessPointResults{
		AccessPointIds:  pulumi.StringMap{},
		AccessPointArns: pulumi.StringMap{},
	}

	if len(spec.AccessPoints) == 0 {
		return results, nil
	}

	for _, apSpec := range spec.AccessPoints {
		args := &efs.AccessPointArgs{
			FileSystemId: fs.ID(),
			Tags: pulumi.StringMap{
				"Name": pulumi.String(apSpec.Name),
			},
		}

		// POSIX user identity enforcement.
		if apSpec.PosixUser != nil {
			posixUser := &efs.AccessPointPosixUserArgs{
				Uid: pulumi.Int(int(apSpec.PosixUser.Uid)),
				Gid: pulumi.Int(int(apSpec.PosixUser.Gid)),
			}
			if len(apSpec.PosixUser.SecondaryGids) > 0 {
				var secondaryGids pulumi.IntArray
				for _, gid := range apSpec.PosixUser.SecondaryGids {
					secondaryGids = append(secondaryGids, pulumi.Int(int(gid)))
				}
				posixUser.SecondaryGids = secondaryGids
			}
			args.PosixUser = posixUser
		}

		// Root directory restriction.
		if apSpec.RootDirectory != nil {
			rootDir := &efs.AccessPointRootDirectoryArgs{
				Path: pulumi.StringPtr(apSpec.RootDirectory.Path),
			}
			if apSpec.RootDirectory.CreationInfo != nil {
				rootDir.CreationInfo = &efs.AccessPointRootDirectoryCreationInfoArgs{
					OwnerUid:    pulumi.Int(int(apSpec.RootDirectory.CreationInfo.OwnerUid)),
					OwnerGid:    pulumi.Int(int(apSpec.RootDirectory.CreationInfo.OwnerGid)),
					Permissions: pulumi.String(apSpec.RootDirectory.CreationInfo.Permissions),
				}
			}
			args.RootDirectory = rootDir
		}

		ap, err := efs.NewAccessPoint(ctx, apSpec.Name, args, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create access point %s", apSpec.Name)
		}

		results.AccessPointIds[apSpec.Name] = ap.ID().ToStringOutput()
		results.AccessPointArns[apSpec.Name] = ap.Arn
	}

	return results, nil
}
