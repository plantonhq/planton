package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func backupPolicyAssignment(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, createdVolume *core.Volume) error {
	spec := locals.OciBlockVolume.Spec

	if spec.BackupPolicyId == nil || spec.BackupPolicyId.GetValue() == "" {
		return nil
	}

	assignmentArgs := &core.VolumeBackupPolicyAssignmentArgs{
		AssetId:  createdVolume.ID(),
		PolicyId: pulumi.String(spec.BackupPolicyId.GetValue()),
	}

	_, err := core.NewVolumeBackupPolicyAssignment(
		ctx,
		locals.DisplayName+"-backup-policy",
		assignmentArgs,
		pulumiOciOpt(provider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create backup policy assignment")
	}

	return nil
}
