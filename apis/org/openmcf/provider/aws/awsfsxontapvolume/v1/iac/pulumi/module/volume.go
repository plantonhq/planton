package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/fsx"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func volume(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*fsx.OntapVolume, error) {
	spec := locals.AwsFsxOntapVolume.Spec
	name := locals.AwsFsxOntapVolume.Metadata.Name

	args := &fsx.OntapVolumeArgs{
		StorageVirtualMachineId: pulumi.String(spec.StorageVirtualMachineId.GetValue()),
		Name:                    pulumi.StringPtr(spec.Name),
		SizeInMegabytes:         pulumi.Int(int(spec.SizeInMegabytes)),
		Tags:                    pulumi.ToStringMap(locals.AwsTags),
	}

	// Junction path (optional — unmounted volume if omitted).
	if spec.JunctionPath != "" {
		args.JunctionPath = pulumi.StringPtr(spec.JunctionPath)
	}

	// ONTAP volume type (ForceNew, default RW via OpenMCF middleware).
	if spec.GetOntapVolumeType() != "" {
		args.OntapVolumeType = pulumi.StringPtr(spec.GetOntapVolumeType())
	}

	// Volume style (ForceNew, default FLEXVOL via OpenMCF middleware).
	if spec.GetVolumeStyle() != "" {
		args.VolumeStyle = pulumi.StringPtr(spec.GetVolumeStyle())
	}

	// Security style (optional — inherits from SVM if omitted).
	if spec.SecurityStyle != "" {
		args.SecurityStyle = pulumi.StringPtr(spec.SecurityStyle)
	}

	// Snapshot policy (optional).
	if spec.SnapshotPolicy != "" {
		args.SnapshotPolicy = pulumi.StringPtr(spec.SnapshotPolicy)
	}

	// Storage efficiency (deduplication, compression, compaction).
	if spec.StorageEfficiencyEnabled {
		args.StorageEfficiencyEnabled = pulumi.BoolPtr(spec.StorageEfficiencyEnabled)
	}

	// Copy tags to backups.
	if spec.GetCopyTagsToBackups() {
		args.CopyTagsToBackups = pulumi.BoolPtr(spec.GetCopyTagsToBackups())
	}

	// Skip final backup on deletion.
	if spec.GetSkipFinalBackup() {
		args.SkipFinalBackup = pulumi.BoolPtr(spec.GetSkipFinalBackup())
	}

	// Bypass SnapLock Enterprise retention on deletion.
	if spec.GetBypassSnaplockEnterpriseRetention() {
		args.BypassSnaplockEnterpriseRetention = pulumi.BoolPtr(spec.GetBypassSnaplockEnterpriseRetention())
	}

	// Tiering policy (optional).
	if spec.TieringPolicy != nil && spec.TieringPolicy.Name != "" {
		tp := &fsx.OntapVolumeTieringPolicyArgs{
			Name: pulumi.StringPtr(spec.TieringPolicy.Name),
		}
		if spec.TieringPolicy.CoolingPeriod > 0 {
			tp.CoolingPeriod = pulumi.IntPtr(int(spec.TieringPolicy.CoolingPeriod))
		}
		args.TieringPolicy = tp
	}

	// SnapLock configuration (optional).
	if spec.SnaplockConfiguration != nil {
		sl := spec.SnaplockConfiguration

		slArgs := &fsx.OntapVolumeSnaplockConfigurationArgs{
			SnaplockType: pulumi.String(sl.SnaplockType),
		}

		if sl.GetAuditLogVolume() {
			slArgs.AuditLogVolume = pulumi.BoolPtr(sl.GetAuditLogVolume())
		}

		if sl.GetPrivilegedDelete() != "" {
			slArgs.PrivilegedDelete = pulumi.StringPtr(sl.GetPrivilegedDelete())
		}

		if sl.GetVolumeAppendModeEnabled() {
			slArgs.VolumeAppendModeEnabled = pulumi.BoolPtr(sl.GetVolumeAppendModeEnabled())
		}

		// Autocommit period.
		if sl.AutocommitPeriod != nil && sl.AutocommitPeriod.Type != "" {
			acArgs := &fsx.OntapVolumeSnaplockConfigurationAutocommitPeriodArgs{
				Type: pulumi.StringPtr(sl.AutocommitPeriod.Type),
			}
			if sl.AutocommitPeriod.Value > 0 {
				acArgs.Value = pulumi.IntPtr(int(sl.AutocommitPeriod.Value))
			}
			slArgs.AutocommitPeriod = acArgs
		}

		// Retention period (default, minimum, maximum).
		if sl.RetentionPeriod != nil {
			rp := &fsx.OntapVolumeSnaplockConfigurationRetentionPeriodArgs{}

			if sl.RetentionPeriod.DefaultRetention != nil && sl.RetentionPeriod.DefaultRetention.Type != "" {
				drArgs := &fsx.OntapVolumeSnaplockConfigurationRetentionPeriodDefaultRetentionArgs{
					Type: pulumi.StringPtr(sl.RetentionPeriod.DefaultRetention.Type),
				}
				if sl.RetentionPeriod.DefaultRetention.Value > 0 {
					drArgs.Value = pulumi.IntPtr(int(sl.RetentionPeriod.DefaultRetention.Value))
				}
				rp.DefaultRetention = drArgs
			}

			if sl.RetentionPeriod.MinimumRetention != nil && sl.RetentionPeriod.MinimumRetention.Type != "" {
				mnArgs := &fsx.OntapVolumeSnaplockConfigurationRetentionPeriodMinimumRetentionArgs{
					Type: pulumi.StringPtr(sl.RetentionPeriod.MinimumRetention.Type),
				}
				if sl.RetentionPeriod.MinimumRetention.Value > 0 {
					mnArgs.Value = pulumi.IntPtr(int(sl.RetentionPeriod.MinimumRetention.Value))
				}
				rp.MinimumRetention = mnArgs
			}

			if sl.RetentionPeriod.MaximumRetention != nil && sl.RetentionPeriod.MaximumRetention.Type != "" {
				mxArgs := &fsx.OntapVolumeSnaplockConfigurationRetentionPeriodMaximumRetentionArgs{
					Type: pulumi.StringPtr(sl.RetentionPeriod.MaximumRetention.Type),
				}
				if sl.RetentionPeriod.MaximumRetention.Value > 0 {
					mxArgs.Value = pulumi.IntPtr(int(sl.RetentionPeriod.MaximumRetention.Value))
				}
				rp.MaximumRetention = mxArgs
			}

			slArgs.RetentionPeriod = rp
		}

		args.SnaplockConfiguration = slArgs
	}

	// Aggregate configuration (for FLEXGROUP volumes).
	if spec.AggregateConfiguration != nil && len(spec.AggregateConfiguration.Aggregates) > 0 {
		aggrs := make(pulumi.StringArray, 0, len(spec.AggregateConfiguration.Aggregates))
		for _, a := range spec.AggregateConfiguration.Aggregates {
			aggrs = append(aggrs, pulumi.String(a))
		}

		aggrArgs := &fsx.OntapVolumeAggregateConfigurationArgs{
			Aggregates: aggrs,
		}

		if spec.AggregateConfiguration.ConstituentsPerAggregate > 0 {
			aggrArgs.ConstituentsPerAggregate = pulumi.IntPtr(int(spec.AggregateConfiguration.ConstituentsPerAggregate))
		}

		args.AggregateConfiguration = aggrArgs
	}

	createdVolume, err := fsx.NewOntapVolume(ctx, name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create fsx ontap volume")
	}

	return createdVolume, nil
}
