package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/fsx"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func fileSystem(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*fsx.OntapFileSystem, error) {
	spec := locals.AwsFsxOntapFileSystem.Spec
	name := locals.AwsFsxOntapFileSystem.Metadata.Name

	subnetIds := make(pulumi.StringArray, 0, len(spec.SubnetIds))
	for _, s := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
	}

	args := &fsx.OntapFileSystemArgs{
		SubnetIds:                    subnetIds,
		ThroughputCapacityPerHaPair: pulumi.IntPtr(int(spec.ThroughputCapacityPerHaPair)),
		StorageCapacity:              pulumi.Int(int(spec.StorageCapacityGib)),
		Tags:                         pulumi.ToStringMap(locals.AwsTags),
	}

	// Deployment type (optional, default SINGLE_AZ_2 via OpenMCF middleware).
	if spec.GetDeploymentType() != "" {
		args.DeploymentType = pulumi.String(spec.GetDeploymentType())
	}

	// Storage type (optional, default SSD via OpenMCF middleware).
	if spec.GetStorageType() != "" {
		args.StorageType = pulumi.StringPtr(spec.GetStorageType())
	}

	// HA pairs (optional, default 1 via OpenMCF middleware).
	if spec.GetHaPairs() > 1 {
		args.HaPairs = pulumi.IntPtr(int(spec.GetHaPairs()))
	}

	// Security groups (ForceNew).
	if len(spec.SecurityGroupIds) > 0 {
		sgIds := make(pulumi.StringArray, 0, len(spec.SecurityGroupIds))
		for _, sg := range spec.SecurityGroupIds {
			sgIds = append(sgIds, pulumi.String(sg.GetValue()))
		}
		args.SecurityGroupIds = sgIds
	}

	// Preferred subnet (multi-AZ only).
	if spec.PreferredSubnetId != nil && spec.PreferredSubnetId.GetValue() != "" {
		args.PreferredSubnetId = pulumi.String(spec.PreferredSubnetId.GetValue())
	}

	// Endpoint IP address range (multi-AZ only).
	if spec.EndpointIpAddressRange != "" {
		args.EndpointIpAddressRange = pulumi.StringPtr(spec.EndpointIpAddressRange)
	}

	// Route table IDs (multi-AZ only).
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

	// ONTAP admin password (sensitive).
	if spec.FsxAdminPassword != "" {
		args.FsxAdminPassword = pulumi.StringPtr(spec.FsxAdminPassword)
	}

	// Disk IOPS configuration.
	if spec.DiskIopsConfiguration != nil {
		iopsArgs := &fsx.OntapFileSystemDiskIopsConfigurationArgs{}
		if spec.DiskIopsConfiguration.GetMode() != "" {
			iopsArgs.Mode = pulumi.StringPtr(spec.DiskIopsConfiguration.GetMode())
		}
		if spec.DiskIopsConfiguration.Iops > 0 {
			iopsArgs.Iops = pulumi.IntPtr(int(spec.DiskIopsConfiguration.Iops))
		}
		args.DiskIopsConfiguration = iopsArgs
	}

	// Automatic backup retention days.
	if spec.GetAutomaticBackupRetentionDays() > 0 {
		args.AutomaticBackupRetentionDays = pulumi.IntPtr(int(spec.GetAutomaticBackupRetentionDays()))
	}

	// Daily automatic backup start time (HH:MM format).
	if spec.DailyAutomaticBackupStartTime != "" {
		args.DailyAutomaticBackupStartTime = pulumi.StringPtr(spec.DailyAutomaticBackupStartTime)
	}

	// Weekly maintenance start time (d:HH:MM format).
	if spec.WeeklyMaintenanceStartTime != "" {
		args.WeeklyMaintenanceStartTime = pulumi.StringPtr(spec.WeeklyMaintenanceStartTime)
	}

	createdFs, err := fsx.NewOntapFileSystem(ctx, name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create fsx ontap file system")
	}

	return createdFs, nil
}
