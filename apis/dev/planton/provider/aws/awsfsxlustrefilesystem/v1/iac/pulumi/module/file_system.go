package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/fsx"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func fileSystem(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*fsx.LustreFileSystem, error) {
	spec := locals.AwsFsxLustreFileSystem.Spec
	name := locals.AwsFsxLustreFileSystem.Metadata.Name

	args := &fsx.LustreFileSystemArgs{
		SubnetIds: pulumi.String(spec.SubnetId.GetValue()),
		Tags:      pulumi.ToStringMap(locals.AwsTags),
	}

	// Storage capacity (required by spec, minimum 1200 GiB).
	args.StorageCapacity = pulumi.Int(int(spec.StorageCapacityGib))

	// Deployment type (optional, default SCRATCH_2 via Planton middleware).
	if spec.GetDeploymentType() != "" {
		args.DeploymentType = pulumi.StringPtr(spec.GetDeploymentType())
	}

	// Storage type (optional, default SSD via Planton middleware).
	if spec.GetStorageType() != "" {
		args.StorageType = pulumi.StringPtr(spec.GetStorageType())
	}

	// Per-unit storage throughput (PERSISTENT_1 and PERSISTENT_2 only).
	if spec.PerUnitStorageThroughput > 0 {
		args.PerUnitStorageThroughput = pulumi.IntPtr(int(spec.PerUnitStorageThroughput))
	}

	// Data compression type (optional, default NONE via Planton middleware).
	if spec.GetDataCompressionType() != "" {
		args.DataCompressionType = pulumi.StringPtr(spec.GetDataCompressionType())
	}

	// File system type version (e.g., "2.12", "2.15").
	if spec.FileSystemTypeVersion != "" {
		args.FileSystemTypeVersion = pulumi.StringPtr(spec.FileSystemTypeVersion)
	}

	// Security groups.
	if len(spec.SecurityGroupIds) > 0 {
		sgIds := make(pulumi.StringArray, 0, len(spec.SecurityGroupIds))
		for _, sg := range spec.SecurityGroupIds {
			sgIds = append(sgIds, pulumi.String(sg.GetValue()))
		}
		args.SecurityGroupIds = sgIds
	}

	// Customer-managed KMS key for encryption at rest.
	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	// S3 import path (legacy, SCRATCH only).
	if spec.ImportPath != "" {
		args.ImportPath = pulumi.StringPtr(spec.ImportPath)
	}

	// S3 export path (legacy, requires import_path).
	if spec.ExportPath != "" {
		args.ExportPath = pulumi.StringPtr(spec.ExportPath)
	}

	// Log configuration.
	if spec.LogConfiguration != nil {
		logArgs := &fsx.LustreFileSystemLogConfigurationArgs{}
		if spec.LogConfiguration.Destination != nil && spec.LogConfiguration.Destination.GetValue() != "" {
			logArgs.Destination = pulumi.StringPtr(spec.LogConfiguration.Destination.GetValue())
		}
		if spec.LogConfiguration.GetLevel() != "" {
			logArgs.Level = pulumi.StringPtr(spec.LogConfiguration.GetLevel())
		}
		args.LogConfiguration = logArgs
	}

	// Automatic backup retention days (PERSISTENT only).
	if spec.GetAutomaticBackupRetentionDays() > 0 {
		args.AutomaticBackupRetentionDays = pulumi.IntPtr(int(spec.GetAutomaticBackupRetentionDays()))
	}

	// Daily automatic backup start time (HH:MM format).
	if spec.DailyAutomaticBackupStartTime != "" {
		args.DailyAutomaticBackupStartTime = pulumi.StringPtr(spec.DailyAutomaticBackupStartTime)
	}

	// Copy tags to backups.
	if spec.CopyTagsToBackups {
		args.CopyTagsToBackups = pulumi.BoolPtr(true)
	}

	// Weekly maintenance start time (d:HH:MM format).
	if spec.WeeklyMaintenanceStartTime != "" {
		args.WeeklyMaintenanceStartTime = pulumi.StringPtr(spec.WeeklyMaintenanceStartTime)
	}

	// Metadata configuration (PERSISTENT_2 only).
	if spec.MetadataConfiguration != nil {
		metaArgs := &fsx.LustreFileSystemMetadataConfigurationArgs{}
		if spec.MetadataConfiguration.GetMode() != "" {
			metaArgs.Mode = pulumi.StringPtr(spec.MetadataConfiguration.GetMode())
		}
		if spec.MetadataConfiguration.Iops > 0 {
			metaArgs.Iops = pulumi.IntPtr(int(spec.MetadataConfiguration.Iops))
		}
		args.MetadataConfiguration = metaArgs
	}

	createdFs, err := fsx.NewLustreFileSystem(ctx, name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create fsx lustre file system")
	}

	return createdFs, nil
}
