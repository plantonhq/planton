package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/docdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// docdbCluster provisions the aws_docdb_cluster itself -- the
// shared-storage brain: endpoints, credentials, backups, encryption, and
// engine lifecycle. The cluster composes onto its neighbors instead of
// embedding them: subnets, security groups, and KMS keys attach by
// reference, and database ingress rules live on the referenced
// AwsSecurityGroup nodes -- this module never creates or mutates a
// resource that deserves to be its own node.
//
// Create-only in AWS: the identifier, port, subnet group, availability
// zones, master username, storage encryption + KMS key, and restore
// sources. Everything else updates in place (immediately or at the next
// maintenance window, per apply_immediately).
func docdbCluster(ctx *pulumi.Context, locals *Locals, provider *aws.Provider,
	createdSubnetGroup *docdb.SubnetGroup, createdParameterGroup *docdb.ClusterParameterGroup) (*docdb.Cluster, error) {
	spec := locals.AwsDocumentDb.Spec

	args := &docdb.ClusterArgs{
		ClusterIdentifier: pulumi.String(locals.ClusterIdentifier),
		Tags:              pulumi.ToStringMap(locals.AwsTags),

		// Deletion safety: the spec's CEL contract requires a
		// final-snapshot name unless skipping is explicit, so a delete can
		// never fail late on a missing snapshot identifier.
		SkipFinalSnapshot:  pulumi.Bool(spec.SkipFinalSnapshot),
		DeletionProtection: pulumi.Bool(spec.DeletionProtection),

		// Storage encryption is a one-way door: it can only be chosen at
		// create time, which is why the spec recommends it on by default.
		StorageEncrypted: pulumi.Bool(spec.StorageEncrypted),

		ApplyImmediately:         pulumi.Bool(spec.ApplyImmediately),
		AllowMajorVersionUpgrade: pulumi.Bool(spec.AllowMajorVersionUpgrade),
	}

	// Empty pins nothing: AWS picks the current default DocumentDB
	// version, so an unpinned manifest never goes stale.
	if spec.EngineVersion != "" {
		args.EngineVersion = pulumi.String(spec.EngineVersion)
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
	// 0 keeps the AWS default (27017). Create-only -- a port change
	// replaces the cluster.
	if spec.Port != 0 {
		args.Port = pulumi.Int(int(spec.Port))
	}

	// "iopt1" opts into I/O-Optimized storage; empty keeps standard
	// per-I/O billing (the AWS default).
	if spec.StorageType != "" {
		args.StorageType = pulumi.String(spec.StorageType)
	}

	if spec.MasterUsername != "" {
		args.MasterUsername = pulumi.String(spec.MasterUsername)
	}

	// The password contract (CEL enforces exactly one strategy): the
	// AWS-managed Secrets Manager secret (recommended -- no secret in
	// manifest or state) or a directly supplied password.
	// manage_master_user_password is forwarded ONLY when true: an
	// explicit false conflicts with master_password in the provider's
	// validation.
	if spec.ManageMasterUserPassword {
		args.ManageMasterUserPassword = pulumi.Bool(true)
	}
	if spec.MasterPassword != "" {
		args.MasterPassword = pulumi.String(spec.MasterPassword)
	}

	if spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
	}

	// Backups: DocumentDB backups are continuous; retention bounds the
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
	if spec.FinalSnapshotIdentifier != "" {
		args.FinalSnapshotIdentifier = pulumi.String(spec.FinalSnapshotIdentifier)
	}

	// "audit" and "profiler" -- both also need their matching cluster
	// parameters (audit_logs / profiler) before DocumentDB emits
	// anything.
	if len(spec.EnabledCloudwatchLogsExports) > 0 {
		args.EnabledCloudwatchLogsExports = pulumi.ToStringArray(spec.EnabledCloudwatchLogsExports)
	}

	// DocumentDB Serverless: this block + db.serverless instances.
	// Adding or modifying scales in place; REMOVING the block from a
	// live cluster replaces it (AWS cannot switch a cluster off
	// serverless).
	if spec.ServerlessV2Scaling != nil {
		args.ServerlessV2ScalingConfiguration = &docdb.ClusterServerlessV2ScalingConfigurationArgs{
			MinCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MinCapacity),
			MaxCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MaxCapacity),
		}
	}

	// Create-time restore sources (mutually exclusive, CEL-enforced):
	// from a snapshot, or from another cluster's continuous backup
	// (point-in-time restore / copy-on-write fast clone).
	if spec.SnapshotIdentifier != "" {
		args.SnapshotIdentifier = pulumi.String(spec.SnapshotIdentifier)
	}
	if spec.RestoreToPointInTime != nil {
		restoreArgs := &docdb.ClusterRestoreToPointInTimeArgs{
			SourceClusterIdentifier: pulumi.String(spec.RestoreToPointInTime.SourceClusterIdentifier),
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

	// DocumentDB global cluster membership. The first cluster joined
	// becomes the global writer; later joiners are read-only
	// secondaries.
	if spec.GlobalClusterIdentifier != "" {
		args.GlobalClusterIdentifier = pulumi.String(spec.GlobalClusterIdentifier)
	}

	// Parameter groups: the managed inline group, an existing referenced
	// group, or the engine default.
	if createdParameterGroup != nil {
		args.DbClusterParameterGroupName = createdParameterGroup.Name
	} else if spec.DbClusterParameterGroupName != "" {
		args.DbClusterParameterGroupName = pulumi.String(spec.DbClusterParameterGroupName)
	}

	createdCluster, err := docdb.NewCluster(ctx, "docdb-cluster", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create DocumentDB cluster")
	}
	return createdCluster, nil
}
