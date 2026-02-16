package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/fsx"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func fileSystem(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*fsx.OpenZfsFileSystem, error) {
	spec := locals.AwsFsxOpenzfsFileSystem.Spec
	name := locals.AwsFsxOpenzfsFileSystem.Metadata.Name

	// Subnet IDs (required, at least 1).
	subnetIds := make(pulumi.StringArray, 0, len(spec.SubnetIds))
	for _, s := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
	}

	args := &fsx.OpenZfsFileSystemArgs{
		SubnetIds:          subnetIds,
		ThroughputCapacity: pulumi.Int(int(spec.ThroughputCapacity)),
		StorageCapacity:    pulumi.IntPtr(int(spec.StorageCapacityGib)),
		Tags:               pulumi.ToStringMap(locals.AwsTags),
	}

	// Deployment type (optional, default SINGLE_AZ_2 via OpenMCF middleware).
	if spec.GetDeploymentType() != "" {
		args.DeploymentType = pulumi.String(spec.GetDeploymentType())
	}

	// Security groups.
	if len(spec.SecurityGroupIds) > 0 {
		sgIds := make(pulumi.StringArray, 0, len(spec.SecurityGroupIds))
		for _, sg := range spec.SecurityGroupIds {
			sgIds = append(sgIds, pulumi.String(sg.GetValue()))
		}
		args.SecurityGroupIds = sgIds
	}

	// Preferred subnet (MULTI_AZ_1 only).
	if spec.PreferredSubnetId != nil && spec.PreferredSubnetId.GetValue() != "" {
		args.PreferredSubnetId = pulumi.StringPtr(spec.PreferredSubnetId.GetValue())
	}

	// Endpoint IP address range (MULTI_AZ_1 only).
	if spec.EndpointIpAddressRange != "" {
		args.EndpointIpAddressRange = pulumi.StringPtr(spec.EndpointIpAddressRange)
	}

	// Route table IDs (MULTI_AZ_1 only).
	if len(spec.RouteTableIds) > 0 {
		rtIds := make(pulumi.StringArray, 0, len(spec.RouteTableIds))
		for _, rt := range spec.RouteTableIds {
			rtIds = append(rtIds, pulumi.String(rt.GetValue()))
		}
		args.RouteTableIds = rtIds
	}

	// Customer-managed KMS key for encryption at rest.
	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	// Disk IOPS configuration.
	if spec.DiskIopsConfiguration != nil {
		iopsArgs := &fsx.OpenZfsFileSystemDiskIopsConfigurationArgs{}
		if spec.DiskIopsConfiguration.GetMode() != "" {
			iopsArgs.Mode = pulumi.StringPtr(spec.DiskIopsConfiguration.GetMode())
		}
		if spec.DiskIopsConfiguration.Iops > 0 {
			iopsArgs.Iops = pulumi.IntPtr(int(spec.DiskIopsConfiguration.Iops))
		}
		args.DiskIopsConfiguration = iopsArgs
	}

	// Root volume configuration.
	if spec.RootVolumeConfiguration != nil {
		rootVolArgs := &fsx.OpenZfsFileSystemRootVolumeConfigurationArgs{}

		if spec.RootVolumeConfiguration.GetDataCompressionType() != "" {
			rootVolArgs.DataCompressionType = pulumi.StringPtr(spec.RootVolumeConfiguration.GetDataCompressionType())
		}

		if spec.RootVolumeConfiguration.ReadOnly {
			rootVolArgs.ReadOnly = pulumi.BoolPtr(true)
		}

		if spec.RootVolumeConfiguration.GetRecordSizeKib() > 0 {
			rootVolArgs.RecordSizeKib = pulumi.IntPtr(int(spec.RootVolumeConfiguration.GetRecordSizeKib()))
		}

		if spec.RootVolumeConfiguration.CopyTagsToSnapshots {
			rootVolArgs.CopyTagsToSnapshots = pulumi.BoolPtr(true)
		}

		// NFS exports.
		if spec.RootVolumeConfiguration.NfsExports != nil &&
			len(spec.RootVolumeConfiguration.NfsExports.ClientConfigurations) > 0 {

			clientConfigs := make(fsx.OpenZfsFileSystemRootVolumeConfigurationNfsExportsClientConfigurationArray, 0,
				len(spec.RootVolumeConfiguration.NfsExports.ClientConfigurations))

			for _, cc := range spec.RootVolumeConfiguration.NfsExports.ClientConfigurations {
				opts := make(pulumi.StringArray, 0, len(cc.Options))
				for _, o := range cc.Options {
					opts = append(opts, pulumi.String(o))
				}
				clientConfigs = append(clientConfigs, &fsx.OpenZfsFileSystemRootVolumeConfigurationNfsExportsClientConfigurationArgs{
					Clients: pulumi.String(cc.Clients),
					Options: opts,
				})
			}

			rootVolArgs.NfsExports = &fsx.OpenZfsFileSystemRootVolumeConfigurationNfsExportsArgs{
				ClientConfigurations: clientConfigs,
			}
		}

		// User and group quotas.
		if len(spec.RootVolumeConfiguration.UserAndGroupQuotas) > 0 {
			quotas := make(fsx.OpenZfsFileSystemRootVolumeConfigurationUserAndGroupQuotaArray, 0,
				len(spec.RootVolumeConfiguration.UserAndGroupQuotas))

			for _, q := range spec.RootVolumeConfiguration.UserAndGroupQuotas {
				quotas = append(quotas, &fsx.OpenZfsFileSystemRootVolumeConfigurationUserAndGroupQuotaArgs{
					Id:                      pulumi.Int(int(q.Id)),
					StorageCapacityQuotaGib: pulumi.Int(int(q.StorageCapacityQuotaGib)),
					Type:                    pulumi.String(q.Type),
				})
			}

			rootVolArgs.UserAndGroupQuotas = quotas
		}

		args.RootVolumeConfiguration = rootVolArgs
	}

	// Automatic backup retention days.
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

	// Copy tags to volumes.
	if spec.CopyTagsToVolumes {
		args.CopyTagsToVolumes = pulumi.BoolPtr(true)
	}

	// Weekly maintenance start time (d:HH:MM format).
	if spec.WeeklyMaintenanceStartTime != "" {
		args.WeeklyMaintenanceStartTime = pulumi.StringPtr(spec.WeeklyMaintenanceStartTime)
	}

	createdFs, err := fsx.NewOpenZfsFileSystem(ctx, name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create fsx openzfs file system")
	}

	return createdFs, nil
}
