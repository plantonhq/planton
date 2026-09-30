package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// rdsInstance provisions the aws_db_instance. The instance composes onto
// its neighbors instead of embedding them: subnets, security groups, KMS
// keys, and the monitoring role attach by reference, and database ingress
// rules live on the referenced AwsSecurityGroup nodes -- this module
// never creates or mutates a resource that deserves to be its own node.
//
// Create-only in AWS: the engine, username, db_name, character sets,
// timezone, availability-zone pin, storage encryption + KMS key, and the
// restore sources. Everything else updates in place (immediately or at
// the next maintenance window, per apply_immediately). Growing storage
// applies in place; shrinking requires a new instance.
func rdsInstance(ctx *pulumi.Context, locals *Locals, provider *aws.Provider,
	createdSubnetGroup *rds.SubnetGroup, createdParameterGroup *rds.ParameterGroup,
	createdOptionGroup *rds.OptionGroup) (*rds.Instance, error) {
	spec := locals.AwsRdsInstance.Spec

	args := &rds.InstanceArgs{
		Identifier:    pulumi.String(locals.InstanceIdentifier),
		InstanceClass: pulumi.String(spec.InstanceClass),
		Tags:          pulumi.ToStringMap(locals.AwsTags),

		// Deletion safety: the spec's CEL contract requires a
		// final-snapshot name unless skipping is explicit, so a delete can
		// never fail late on a missing snapshot identifier.
		SkipFinalSnapshot:  pulumi.Bool(spec.SkipFinalSnapshot),
		DeletionProtection: pulumi.Bool(spec.DeletionProtection),

		// Storage encryption is a one-way door: it can only be chosen at
		// create time, which is why the spec recommends it on by default.
		StorageEncrypted: pulumi.Bool(spec.StorageEncrypted),

		// Availability: a synchronous standby with automatic failover, or
		// a single-AZ instance optionally pinned below.
		MultiAz:            pulumi.Bool(spec.MultiAz),
		PubliclyAccessible: pulumi.Bool(spec.PubliclyAccessible),

		CopyTagsToSnapshot:               pulumi.Bool(spec.CopyTagsToSnapshot),
		IamDatabaseAuthenticationEnabled: pulumi.Bool(spec.IamDatabaseAuthenticationEnabled),
		AllowMajorVersionUpgrade:         pulumi.Bool(spec.AllowMajorVersionUpgrade),
		ApplyImmediately:                 pulumi.Bool(spec.ApplyImmediately),
	}

	// Empty on a replica or restore: AWS derives the engine from the
	// source (the CEL contract requires it otherwise). Empty version pins
	// nothing -- AWS picks the engine's current default, never going stale.
	if spec.Engine != "" {
		args.Engine = pulumi.String(spec.Engine)
	}
	if spec.EngineVersion != "" {
		args.EngineVersion = pulumi.String(spec.EngineVersion)
	}

	// Storage: allocated is inherited from the source on replicas and
	// restores; the autoscaling ceiling is the cheap insurance against
	// disk-full outages.
	if spec.AllocatedStorageGb != 0 {
		args.AllocatedStorage = pulumi.Int(int(spec.AllocatedStorageGb))
	}
	if spec.MaxAllocatedStorageGb != 0 {
		args.MaxAllocatedStorage = pulumi.Int(int(spec.MaxAllocatedStorageGb))
	}
	if spec.StorageType != "" {
		args.StorageType = pulumi.String(spec.StorageType)
	}
	if spec.Iops != 0 {
		args.Iops = pulumi.Int(int(spec.Iops))
	}
	if spec.StorageThroughput != 0 {
		args.StorageThroughput = pulumi.Int(int(spec.StorageThroughput))
	}
	// A dedicated EBS volume for logs -- steadier I/O for audit-heavy or
	// WAL-heavy workloads. Forwarded only when true.
	if spec.DedicatedLogVolume {
		args.DedicatedLogVolume = pulumi.Bool(true)
	}
	if spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
	}

	if spec.DbName != "" {
		args.DbName = pulumi.String(spec.DbName)
	}
	if spec.Username != "" {
		args.Username = pulumi.String(spec.Username)
	}

	// The three-way password contract (CEL enforces exactly one strategy):
	// AWS-managed secret (recommended -- no secret in manifest or state)
	// or a directly supplied password. manage_master_user_password is
	// forwarded ONLY when true: an explicit false conflicts with password
	// in the provider's validation.
	if spec.ManageMasterUserPassword {
		args.ManageMasterUserPassword = pulumi.Bool(true)
	}
	if spec.MasterUserSecretKmsKeyId.GetValue() != "" {
		args.MasterUserSecretKmsKeyId = pulumi.String(spec.MasterUserSecretKmsKeyId.GetValue())
	}
	if spec.Password != "" {
		args.Password = pulumi.String(spec.Password)
	}

	// Networking: the subnet group managed here (or referenced), the VPC
	// default SG when no groups are given (AWS's own default).
	if createdSubnetGroup != nil {
		args.DbSubnetGroupName = createdSubnetGroup.Name
	} else if spec.DbSubnetGroupName != "" {
		args.DbSubnetGroupName = pulumi.String(spec.DbSubnetGroupName)
	}
	if len(spec.SecurityGroupIds) > 0 {
		securityGroupIds := pulumi.StringArray{}
		for _, securityGroupId := range spec.SecurityGroupIds {
			securityGroupIds = append(securityGroupIds, pulumi.String(securityGroupId.GetValue()))
		}
		args.VpcSecurityGroupIds = securityGroupIds
	}
	if spec.NetworkType != "" {
		args.NetworkType = pulumi.String(spec.NetworkType)
	}
	if spec.Port != 0 {
		args.Port = pulumi.Int(int(spec.Port))
	}
	// Empty lets AWS place the instance; a pin is create-only and is
	// mutually exclusive with multi_az (CEL).
	if spec.AvailabilityZone != "" {
		args.AvailabilityZone = pulumi.String(spec.AvailabilityZone)
	}

	// Read replica: engine, storage, and credentials are inherited from
	// the source (CEL keeps them empty here). replica_mode is Oracle's
	// queryable-vs-mounted choice.
	if spec.ReplicateSourceDb != "" {
		args.ReplicateSourceDb = pulumi.String(spec.ReplicateSourceDb)
	}
	if spec.ReplicaMode != "" {
		args.ReplicaMode = pulumi.String(spec.ReplicaMode)
	}

	// Create-time restore sources (mutually exclusive with each other and
	// with replicate_source_db, CEL-enforced). s3_import is the MySQL
	// migration on-ramp: restore a Percona XtraBackup straight from S3.
	if spec.SnapshotIdentifier != "" {
		args.SnapshotIdentifier = pulumi.String(spec.SnapshotIdentifier)
	}
	if spec.S3Import != nil {
		s3ImportArgs := &rds.InstanceS3ImportArgs{
			BucketName:          pulumi.String(spec.S3Import.BucketName),
			IngestionRole:       pulumi.String(spec.S3Import.IngestionRole.GetValue()),
			SourceEngine:        pulumi.String(spec.S3Import.SourceEngine),
			SourceEngineVersion: pulumi.String(spec.S3Import.SourceEngineVersion),
		}
		if spec.S3Import.BucketPrefix != "" {
			s3ImportArgs.BucketPrefix = pulumi.String(spec.S3Import.BucketPrefix)
		}
		args.S3Import = s3ImportArgs
	}

	// One-way storage file-system upgrade, consulted on replicas and
	// snapshot restores whose source still runs the older configuration.
	if spec.UpgradeStorageConfig {
		args.UpgradeStorageConfig = pulumi.Bool(true)
	}
	if spec.RestoreToPointInTime != nil {
		restoreArgs := &rds.InstanceRestoreToPointInTimeArgs{}
		if spec.RestoreToPointInTime.SourceDbInstanceIdentifier != "" {
			restoreArgs.SourceDbInstanceIdentifier = pulumi.String(spec.RestoreToPointInTime.SourceDbInstanceIdentifier)
		}
		if spec.RestoreToPointInTime.SourceDbiResourceId != "" {
			restoreArgs.SourceDbiResourceId = pulumi.String(spec.RestoreToPointInTime.SourceDbiResourceId)
		}
		if spec.RestoreToPointInTime.SourceDbInstanceAutomatedBackupsArn != "" {
			restoreArgs.SourceDbInstanceAutomatedBackupsArn = pulumi.String(spec.RestoreToPointInTime.SourceDbInstanceAutomatedBackupsArn)
		}
		if spec.RestoreToPointInTime.RestoreTime != "" {
			restoreArgs.RestoreTime = pulumi.String(spec.RestoreToPointInTime.RestoreTime)
		}
		if spec.RestoreToPointInTime.UseLatestRestorableTime {
			restoreArgs.UseLatestRestorableTime = pulumi.Bool(true)
		}
		args.RestoreToPointInTime = restoreArgs
	}

	// Backups: 0 disables automated backups (and point-in-time recovery).
	if spec.BackupRetentionPeriod != 0 {
		args.BackupRetentionPeriod = pulumi.Int(int(spec.BackupRetentionPeriod))
	}
	if spec.BackupWindow != "" {
		args.BackupWindow = pulumi.String(spec.BackupWindow)
	}
	if spec.MaintenanceWindow != "" {
		args.MaintenanceWindow = pulumi.String(spec.MaintenanceWindow)
	}
	// Tri-state: unset keeps the AWS default (true). An explicit false
	// retains automated backups after deletion -- the last line of defense
	// against a mistaken teardown.
	if spec.DeleteAutomatedBackups != nil {
		args.DeleteAutomatedBackups = pulumi.Bool(spec.GetDeleteAutomatedBackups())
	}
	if spec.FinalSnapshotIdentifier != "" {
		args.FinalSnapshotIdentifier = pulumi.String(spec.FinalSnapshotIdentifier)
	}

	if len(spec.EnabledCloudwatchLogsExports) > 0 {
		args.EnabledCloudwatchLogsExports = pulumi.ToStringArray(spec.EnabledCloudwatchLogsExports)
	}

	// Observability: Performance Insights (per-query telemetry, free at
	// 7-day retention), Enhanced Monitoring (OS-level metrics through the
	// referenced role), and Database Insights.
	if spec.PerformanceInsightsEnabled {
		args.PerformanceInsightsEnabled = pulumi.Bool(true)
	}
	if spec.PerformanceInsightsKmsKeyId.GetValue() != "" {
		args.PerformanceInsightsKmsKeyId = pulumi.String(spec.PerformanceInsightsKmsKeyId.GetValue())
	}
	if spec.PerformanceInsightsRetentionPeriod != 0 {
		args.PerformanceInsightsRetentionPeriod = pulumi.Int(int(spec.PerformanceInsightsRetentionPeriod))
	}
	if spec.MonitoringInterval != 0 {
		args.MonitoringInterval = pulumi.Int(int(spec.MonitoringInterval))
	}
	if spec.MonitoringRoleArn.GetValue() != "" {
		args.MonitoringRoleArn = pulumi.String(spec.MonitoringRoleArn.GetValue())
	}
	if spec.DatabaseInsightsMode != "" {
		args.DatabaseInsightsMode = pulumi.String(spec.DatabaseInsightsMode)
	}

	// Engine-configuration attachments: the managed inline group, an
	// existing referenced group, or the engine default -- parameter and
	// option groups are configuration lists, not composable nodes.
	if createdParameterGroup != nil {
		args.ParameterGroupName = createdParameterGroup.Name
	} else if spec.ParameterGroupName != "" {
		args.ParameterGroupName = pulumi.String(spec.ParameterGroupName)
	}
	if createdOptionGroup != nil {
		args.OptionGroupName = createdOptionGroup.Name
	} else if spec.OptionGroupName != "" {
		args.OptionGroupName = pulumi.String(spec.OptionGroupName)
	}

	// Active Directory join -- either the AWS-managed directory shape or
	// the self-managed shape (never both; CEL-enforced).
	if spec.ActiveDirectory != nil {
		if spec.ActiveDirectory.Domain != "" {
			args.Domain = pulumi.String(spec.ActiveDirectory.Domain)
		}
		if spec.ActiveDirectory.DomainIamRoleName != "" {
			args.DomainIamRoleName = pulumi.String(spec.ActiveDirectory.DomainIamRoleName)
		}
		if spec.ActiveDirectory.DomainFqdn != "" {
			args.DomainFqdn = pulumi.String(spec.ActiveDirectory.DomainFqdn)
		}
		if spec.ActiveDirectory.DomainOu != "" {
			args.DomainOu = pulumi.String(spec.ActiveDirectory.DomainOu)
		}
		if spec.ActiveDirectory.DomainAuthSecretArn != "" {
			args.DomainAuthSecretArn = pulumi.String(spec.ActiveDirectory.DomainAuthSecretArn)
		}
		if len(spec.ActiveDirectory.DomainDnsIps) > 0 {
			args.DomainDnsIps = pulumi.ToStringArray(spec.ActiveDirectory.DomainDnsIps)
		}
	}

	if spec.LicenseModel != "" {
		args.LicenseModel = pulumi.String(spec.LicenseModel)
	}
	// Extended support posture -- empty keeps the AWS default (paid
	// extended support engages automatically when the version leaves
	// standard support).
	if spec.EngineLifecycleSupport != "" {
		args.EngineLifecycleSupport = pulumi.String(spec.EngineLifecycleSupport)
	}
	if spec.CharacterSetName != "" {
		args.CharacterSetName = pulumi.String(spec.CharacterSetName)
	}
	if spec.NcharCharacterSetName != "" {
		args.NcharCharacterSetName = pulumi.String(spec.NcharCharacterSetName)
	}
	if spec.Timezone != "" {
		args.Timezone = pulumi.String(spec.Timezone)
	}
	if spec.CaCertIdentifier != "" {
		args.CaCertIdentifier = pulumi.String(spec.CaCertIdentifier)
	}

	// RDS Blue/Green Deployments: a synchronized green copy takes the
	// change and switches over in under a minute -- near-zero-downtime
	// engine upgrades and parameter changes.
	if spec.BlueGreenUpdateEnabled {
		args.BlueGreenUpdate = &rds.InstanceBlueGreenUpdateArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	// Tri-state: unset keeps the AWS default (true).
	if spec.AutoMinorVersionUpgrade != nil {
		args.AutoMinorVersionUpgrade = pulumi.Bool(spec.GetAutoMinorVersionUpgrade())
	}

	createdInstance, err := rds.NewInstance(ctx, "rds-instance", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create RDS instance")
	}
	return createdInstance, nil
}
