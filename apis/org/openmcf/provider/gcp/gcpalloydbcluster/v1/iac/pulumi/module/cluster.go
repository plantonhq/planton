package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/alloydb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func cluster(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*alloydb.Cluster, error) {
	spec := locals.GcpAlloydbCluster.Spec

	args := &alloydb.ClusterArgs{
		ClusterId:          pulumi.String(spec.ClusterName),
		Location:           pulumi.String(spec.Location),
		Project:            pulumi.StringPtr(spec.ProjectId.GetValue()),
		Labels:             pulumi.ToStringMap(locals.GcpLabels),
		DeletionProtection: pulumi.BoolPtr(spec.DeletionProtection),
	}

	// Network configuration.
	networkConfig := &alloydb.ClusterNetworkConfigArgs{
		Network: pulumi.StringPtr(spec.Network.GetValue()),
	}
	if spec.AllocatedIpRange != "" {
		networkConfig.AllocatedIpRange = pulumi.StringPtr(spec.AllocatedIpRange)
	}
	args.NetworkConfig = networkConfig

	// Database version.
	if spec.DatabaseVersion != "" {
		args.DatabaseVersion = pulumi.StringPtr(spec.DatabaseVersion)
	}

	// Display name.
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}

	// Initial user.
	if spec.InitialUser != nil {
		initialUserArgs := &alloydb.ClusterInitialUserArgs{
			Password: pulumi.String(spec.InitialUser.Password),
		}
		if spec.InitialUser.User != "" {
			initialUserArgs.User = pulumi.StringPtr(spec.InitialUser.User)
		}
		args.InitialUser = initialUserArgs
	}

	// Automated backup policy.
	if spec.AutomatedBackupPolicy != nil {
		backupArgs := &alloydb.ClusterAutomatedBackupPolicyArgs{}

		if spec.AutomatedBackupPolicy.Enabled {
			backupArgs.Enabled = pulumi.BoolPtr(true)
		}
		if spec.AutomatedBackupPolicy.BackupWindow != "" {
			backupArgs.BackupWindow = pulumi.StringPtr(spec.AutomatedBackupPolicy.BackupWindow)
		}
		if spec.AutomatedBackupPolicy.Location != "" {
			backupArgs.Location = pulumi.StringPtr(spec.AutomatedBackupPolicy.Location)
		}

		// Retention: quantity-based or time-based (mutually exclusive, validated by proto).
		if spec.AutomatedBackupPolicy.QuantityBasedRetentionCount > 0 {
			backupArgs.QuantityBasedRetention = &alloydb.ClusterAutomatedBackupPolicyQuantityBasedRetentionArgs{
				Count: pulumi.IntPtr(int(spec.AutomatedBackupPolicy.QuantityBasedRetentionCount)),
			}
		}
		if spec.AutomatedBackupPolicy.TimeBasedRetentionPeriod != "" {
			backupArgs.TimeBasedRetention = &alloydb.ClusterAutomatedBackupPolicyTimeBasedRetentionArgs{
				RetentionPeriod: pulumi.StringPtr(spec.AutomatedBackupPolicy.TimeBasedRetentionPeriod),
			}
		}

		// Weekly schedule.
		if spec.AutomatedBackupPolicy.WeeklySchedule != nil {
			schedule := spec.AutomatedBackupPolicy.WeeklySchedule
			scheduleArgs := &alloydb.ClusterAutomatedBackupPolicyWeeklyScheduleArgs{}

			if len(schedule.DaysOfWeek) > 0 {
				scheduleArgs.DaysOfWeeks = pulumi.ToStringArray(schedule.DaysOfWeek)
			}

			// Convert start_hour to the TimeOfDay structure GCP expects.
			scheduleArgs.StartTimes = alloydb.ClusterAutomatedBackupPolicyWeeklyScheduleStartTimeArray{
				&alloydb.ClusterAutomatedBackupPolicyWeeklyScheduleStartTimeArgs{
					Hours: pulumi.IntPtr(int(schedule.StartHour)),
				},
			}

			backupArgs.WeeklySchedule = scheduleArgs
		}

		// Backup encryption.
		if spec.AutomatedBackupPolicy.EncryptionKmsKeyName != nil && spec.AutomatedBackupPolicy.EncryptionKmsKeyName.GetValue() != "" {
			backupArgs.EncryptionConfig = &alloydb.ClusterAutomatedBackupPolicyEncryptionConfigArgs{
				KmsKeyName: pulumi.StringPtr(spec.AutomatedBackupPolicy.EncryptionKmsKeyName.GetValue()),
			}
		}

		args.AutomatedBackupPolicy = backupArgs
	}

	// Continuous backup config.
	if spec.ContinuousBackupConfig != nil {
		continuousArgs := &alloydb.ClusterContinuousBackupConfigArgs{}

		// The enabled field: proto bool defaults to false, so we need to handle
		// the case where the block is present but enabled is not explicitly set.
		// In GCP, continuous backup defaults to enabled=true when not specified.
		continuousArgs.Enabled = pulumi.BoolPtr(spec.ContinuousBackupConfig.Enabled)

		if spec.ContinuousBackupConfig.RecoveryWindowDays > 0 {
			continuousArgs.RecoveryWindowDays = pulumi.IntPtr(int(spec.ContinuousBackupConfig.RecoveryWindowDays))
		}

		if spec.ContinuousBackupConfig.EncryptionKmsKeyName != nil && spec.ContinuousBackupConfig.EncryptionKmsKeyName.GetValue() != "" {
			continuousArgs.EncryptionConfig = &alloydb.ClusterContinuousBackupConfigEncryptionConfigArgs{
				KmsKeyName: pulumi.StringPtr(spec.ContinuousBackupConfig.EncryptionKmsKeyName.GetValue()),
			}
		}

		args.ContinuousBackupConfig = continuousArgs
	}

	// Cluster-level CMEK encryption.
	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		args.EncryptionConfig = &alloydb.ClusterEncryptionConfigArgs{
			KmsKeyName: pulumi.StringPtr(spec.KmsKeyName.GetValue()),
		}
	}

	// Maintenance window.
	if spec.MaintenanceWindow != nil {
		args.MaintenanceUpdatePolicy = &alloydb.ClusterMaintenanceUpdatePolicyArgs{
			MaintenanceWindows: alloydb.ClusterMaintenanceUpdatePolicyMaintenanceWindowArray{
				&alloydb.ClusterMaintenanceUpdatePolicyMaintenanceWindowArgs{
					Day: pulumi.String(spec.MaintenanceWindow.Day),
					StartTime: &alloydb.ClusterMaintenanceUpdatePolicyMaintenanceWindowStartTimeArgs{
						Hours: pulumi.Int(int(spec.MaintenanceWindow.StartHour)),
					},
				},
			},
		}
	}

	createdCluster, err := alloydb.NewCluster(ctx, "alloydb-cluster", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create alloydb cluster")
	}

	// Export cluster outputs.
	ctx.Export(OpClusterId, createdCluster.Name)
	ctx.Export(OpClusterName, pulumi.String(spec.ClusterName))
	ctx.Export(OpDatabaseVersion, createdCluster.DatabaseVersion)
	ctx.Export(OpState, createdCluster.State)

	return createdCluster, nil
}
