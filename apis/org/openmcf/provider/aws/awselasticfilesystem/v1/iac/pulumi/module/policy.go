package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/efs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func policies(ctx *pulumi.Context, locals *Locals, provider *aws.Provider, fs *efs.FileSystem) error {
	spec := locals.AwsElasticFileSystem.Spec

	// Backup policy — enable or disable automatic daily backups.
	if spec.BackupEnabled {
		_, err := efs.NewBackupPolicy(ctx, "backup-policy", &efs.BackupPolicyArgs{
			FileSystemId: fs.ID(),
			BackupPolicy: &efs.BackupPolicyBackupPolicyArgs{
				Status: pulumi.String("ENABLED"),
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to create efs backup policy")
		}
	}

	// File system resource policy — IAM policy for access control.
	if spec.Policy != nil {
		policyJSON, err := json.Marshal(spec.Policy.AsMap())
		if err != nil {
			return errors.Wrap(err, "failed to serialize efs file system policy to JSON")
		}

		_, err = efs.NewFileSystemPolicy(ctx, "fs-policy", &efs.FileSystemPolicyArgs{
			FileSystemId: fs.ID(),
			Policy:       pulumi.String(string(policyJSON)),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to create efs file system policy")
		}
	}

	return nil
}
