package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// rdsCluster provisions the aws_rds_cluster itself -- the shared-storage
// brain: endpoints, credentials, backups, encryption, and engine
// lifecycle. The cluster composes onto its neighbors instead of embedding
// them: subnets, security groups, KMS keys, and IAM roles attach by
// reference, and database ingress rules live on the referenced
// AwsSecurityGroup nodes -- this module never creates or mutates a
// resource that deserves to be its own node.
//
// Create-only in AWS: the identifier, engine, engine_mode, subnet group,
// availability zones, master username, database name, storage
// encryption + KMS key, restore sources, and source_region. Everything
// else updates in place (immediately or at the next maintenance window,
// per apply_immediately).
func rdsCluster(ctx *pulumi.Context, locals *Locals, provider *aws.Provider,
	createdSubnetGroup *rds.SubnetGroup, createdParameterGroup *rds.ClusterParameterGroup) (*rds.Cluster, error) {
	spec := locals.AwsRdsCluster.Spec

	args := &rds.ClusterArgs{
		ClusterIdentifier: pulumi.String(locals.ClusterIdentifier),
		Engine:            pulumi.String(spec.Engine),
		Tags:              pulumi.ToStringMap(locals.AwsTags),

		// Deletion safety: the spec's CEL contract requires a
		// final-snapshot name unless skipping is explicit, so a delete can
		// never fail late on a missing snapshot identifier.
		SkipFinalSnapshot:  pulumi.Bool(spec.SkipFinalSnapshot),
		DeletionProtection: pulumi.Bool(spec.DeletionProtection),

		// Storage encryption is a one-way door: it can only be chosen at
		// create time, which is why the spec recommends it on by default.
		StorageEncrypted: pulumi.Bool(spec.StorageEncrypted),

		CopyTagsToSnapshot:               pulumi.Bool(spec.CopyTagsToSnapshot),
		IamDatabaseAuthenticationEnabled: pulumi.Bool(spec.IamDatabaseAuthenticationEnabled),
		// The Data API: SQL over HTTPS with IAM auth -- the natural fit
		// for Lambda and other connection-averse callers.
		EnableHttpEndpoint:       pulumi.Bool(spec.EnableHttpEndpoint),
		ApplyImmediately:         pulumi.Bool(spec.ApplyImmediately),
		AllowMajorVersionUpgrade: pulumi.Bool(spec.AllowMajorVersionUpgrade),
	}

	// Empty pins nothing: AWS picks the engine's current default version,
	// so an unpinned manifest never goes stale.
	if spec.EngineVersion != "" {
		args.EngineVersion = pulumi.String(spec.EngineVersion)
	}
	// Empty and "provisioned" are the same thing to AWS. Serverless v2 is
	// provisioned mode + a serverlessv2 block -- only legacy Serverless v1
	// sets "serverless".
	if spec.EngineMode != "" {
		args.EngineMode = pulumi.String(spec.EngineMode)
	}
	if spec.EngineLifecycleSupport != "" {
		args.EngineLifecycleSupport = pulumi.String(spec.EngineLifecycleSupport)
	}

	// Networking: the subnet group managed here (or referenced), the VPC
	// default SG when no groups are given (AWS's own default), and
	// AWS-picked AZ spread unless explicitly pinned (the list is
	// create-only -- letting AWS choose is almost always right).
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
	if len(spec.AvailabilityZones) > 0 {
		args.AvailabilityZones = pulumi.ToStringArray(spec.AvailabilityZones)
	}
	if spec.NetworkType != "" {
		args.NetworkType = pulumi.String(spec.NetworkType)
	}
	if spec.Port != 0 {
		args.Port = pulumi.Int(int(spec.Port))
	}

	// Multi-AZ RDS cluster shape (community mysql/postgres engines): AWS
	// manages one writer + two readers internally, sized here. The CEL
	// rules guarantee these are set exactly when the engine calls for
	// them, and that the folded instances list stays empty.
	if spec.DbClusterInstanceClass != "" {
		args.DbClusterInstanceClass = pulumi.String(spec.DbClusterInstanceClass)
	}
	if spec.AllocatedStorageGb != 0 {
		args.AllocatedStorage = pulumi.Int(int(spec.AllocatedStorageGb))
	}
	if spec.Iops != 0 {
		args.Iops = pulumi.Int(int(spec.Iops))
	}
	if spec.StorageType != "" {
		args.StorageType = pulumi.String(spec.StorageType)
	}

	if spec.DatabaseName != "" {
		args.DatabaseName = pulumi.String(spec.DatabaseName)
	}
	if spec.MasterUsername != "" {
		args.MasterUsername = pulumi.String(spec.MasterUsername)
	}

	// The three-way password contract (CEL enforces exactly one strategy):
	// AWS-managed secret (recommended -- no secret in manifest or state)
	// or a directly supplied password. manage_master_user_password is
	// forwarded ONLY when true: an explicit false conflicts with
	// master_password in the provider's validation.
	if spec.ManageMasterUserPassword {
		args.ManageMasterUserPassword = pulumi.Bool(true)
	}
	if spec.MasterUserSecretKmsKeyId.GetValue() != "" {
		args.MasterUserSecretKmsKeyId = pulumi.String(spec.MasterUserSecretKmsKeyId.GetValue())
	}
	if spec.MasterPassword != "" {
		args.MasterPassword = pulumi.String(spec.MasterPassword)
	}

	if spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
	}

	// Backups: Aurora backups are continuous; retention bounds the
	// point-in-time recovery window. 0 keeps the AWS default (1 day).
	if spec.BackupRetentionPeriod != 0 {
		args.BackupRetentionPeriod = pulumi.Int(int(spec.BackupRetentionPeriod))
	}
	if spec.PreferredBackupWindow != "" {
		args.PreferredBackupWindow = pulumi.String(spec.PreferredBackupWindow)
	}
	if spec.PreferredMaintenanceWindow != "" {
		args.PreferredMaintenanceWindow = pulumi.String(spec.PreferredMaintenanceWindow)
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

	// Aurora MySQL in-place rewind. Enabling later is not supported by
	// AWS, so production Aurora MySQL wants this at create.
	if spec.BacktrackWindowSeconds != 0 {
		args.BacktrackWindow = pulumi.Int(int(spec.BacktrackWindowSeconds))
	}

	// Engine feature roles are managed as one association resource per
	// spec.iam_roles entry (role_associations.go) -- never this
	// resource's inline IamRoles argument, which cannot carry feature
	// names and, per the provider's own warning, overwrites association
	// resources when the two are mixed.

	if len(spec.EnabledCloudwatchLogsExports) > 0 {
		args.EnabledCloudwatchLogsExports = pulumi.ToStringArray(spec.EnabledCloudwatchLogsExports)
	}

	// Cluster-level observability; per-instance overrides live on the
	// folded instance entries.
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

	// Aurora Serverless v2: provisioned mode + this block + db.serverless
	// instances. min_capacity 0 enables automatic pause -- compute cost
	// drops to zero while idle, resumed on the next connection.
	if spec.ServerlessV2Scaling != nil {
		scalingArgs := &rds.ClusterServerlessv2ScalingConfigurationArgs{
			MinCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MinCapacity),
			MaxCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MaxCapacity),
		}
		if spec.ServerlessV2Scaling.SecondsUntilAutoPause != 0 {
			scalingArgs.SecondsUntilAutoPause = pulumi.Int(int(spec.ServerlessV2Scaling.SecondsUntilAutoPause))
		}
		args.Serverlessv2ScalingConfiguration = scalingArgs
	}

	// Legacy Aurora Serverless v1 (engine_mode "serverless") -- AWS owns
	// the compute entirely; the folded instances list stays empty (CEL).
	if spec.ServerlessV1Scaling != nil {
		scalingArgs := &rds.ClusterScalingConfigurationArgs{}
		if spec.ServerlessV1Scaling.AutoPause != nil {
			scalingArgs.AutoPause = pulumi.Bool(spec.ServerlessV1Scaling.GetAutoPause())
		}
		if spec.ServerlessV1Scaling.MinCapacity != 0 {
			scalingArgs.MinCapacity = pulumi.Int(int(spec.ServerlessV1Scaling.MinCapacity))
		}
		if spec.ServerlessV1Scaling.MaxCapacity != 0 {
			scalingArgs.MaxCapacity = pulumi.Int(int(spec.ServerlessV1Scaling.MaxCapacity))
		}
		if spec.ServerlessV1Scaling.SecondsUntilAutoPause != 0 {
			scalingArgs.SecondsUntilAutoPause = pulumi.Int(int(spec.ServerlessV1Scaling.SecondsUntilAutoPause))
		}
		if spec.ServerlessV1Scaling.SecondsBeforeTimeout != 0 {
			scalingArgs.SecondsBeforeTimeout = pulumi.Int(int(spec.ServerlessV1Scaling.SecondsBeforeTimeout))
		}
		if spec.ServerlessV1Scaling.TimeoutAction != "" {
			scalingArgs.TimeoutAction = pulumi.String(spec.ServerlessV1Scaling.TimeoutAction)
		}
		args.ScalingConfiguration = scalingArgs
	}

	// Kerberos authentication through an AWS Managed Microsoft AD --
	// clusters only support the managed-directory shape (the pair is
	// CEL-coupled; self-managed AD is an instance-kind capability).
	if spec.Domain != "" {
		args.Domain = pulumi.String(spec.Domain)
	}
	if spec.DomainIamRoleName != "" {
		args.DomainIamRoleName = pulumi.String(spec.DomainIamRoleName)
	}

	// Tri-state: unset keeps the AWS default (true). Per-instance
	// auto_minor_version_upgrade on the folded instances overrides this
	// for that instance.
	if spec.AutoMinorVersionUpgrade != nil {
		args.AutoMinorVersionUpgrade = pulumi.Bool(spec.GetAutoMinorVersionUpgrade())
	}

	// Create-time restore sources (mutually exclusive, CEL-enforced):
	// from a snapshot, from another cluster's continuous backup
	// (point-in-time restore / copy-on-write fast clone), or from a
	// Percona XtraBackup in S3 (aurora-mysql migration on-ramp).
	if spec.SnapshotIdentifier != "" {
		args.SnapshotIdentifier = pulumi.String(spec.SnapshotIdentifier)
	}
	if spec.S3Import != nil {
		s3ImportArgs := &rds.ClusterS3ImportArgs{
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
	if spec.RestoreToPointInTime != nil {
		restoreArgs := &rds.ClusterRestoreToPointInTimeArgs{}
		if spec.RestoreToPointInTime.SourceClusterIdentifier != "" {
			restoreArgs.SourceClusterIdentifier = pulumi.String(spec.RestoreToPointInTime.SourceClusterIdentifier)
		}
		if spec.RestoreToPointInTime.SourceClusterResourceId != "" {
			restoreArgs.SourceClusterResourceId = pulumi.String(spec.RestoreToPointInTime.SourceClusterResourceId)
		}
		if spec.RestoreToPointInTime.RestoreToTime != "" {
			restoreArgs.RestoreToTime = pulumi.String(spec.RestoreToPointInTime.RestoreToTime)
		}
		if spec.RestoreToPointInTime.UseLatestRestorableTime {
			restoreArgs.UseLatestRestorableTime = pulumi.Bool(true)
		}
		if spec.RestoreToPointInTime.RestoreType != "" {
			restoreArgs.RestoreType = pulumi.String(spec.RestoreToPointInTime.RestoreType)
		}
		args.RestoreToPointInTime = restoreArgs
	}

	// Cross-region replication and Aurora Global Database membership.
	if spec.ReplicationSourceIdentifier != "" {
		args.ReplicationSourceIdentifier = pulumi.String(spec.ReplicationSourceIdentifier)
	}
	if spec.SourceRegion != "" {
		args.SourceRegion = pulumi.String(spec.SourceRegion)
	}
	if spec.GlobalClusterIdentifier != "" {
		args.GlobalClusterIdentifier = pulumi.String(spec.GlobalClusterIdentifier)
	}
	// Forwarded only when true: the provider defaults both to false, and
	// a bare false on a non-global cluster is a pointless diff.
	if spec.EnableGlobalWriteForwarding {
		args.EnableGlobalWriteForwarding = pulumi.Bool(true)
	}
	if spec.EnableLocalWriteForwarding {
		args.EnableLocalWriteForwarding = pulumi.Bool(true)
	}

	// Parameter groups: the managed inline group, an existing referenced
	// group, or the engine default. db_instance_parameter_group_name is
	// only consulted by AWS during a major version upgrade.
	if createdParameterGroup != nil {
		args.DbClusterParameterGroupName = createdParameterGroup.Name
	} else if spec.DbClusterParameterGroupName != "" {
		args.DbClusterParameterGroupName = pulumi.String(spec.DbClusterParameterGroupName)
	}
	if spec.DbInstanceParameterGroupName != "" {
		args.DbInstanceParameterGroupName = pulumi.String(spec.DbInstanceParameterGroupName)
	}

	if spec.CaCertificateIdentifier != "" {
		args.CaCertificateIdentifier = pulumi.String(spec.CaCertificateIdentifier)
	}

	createdCluster, err := rds.NewCluster(ctx, "rds-cluster", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create RDS cluster")
	}
	return createdCluster, nil
}
