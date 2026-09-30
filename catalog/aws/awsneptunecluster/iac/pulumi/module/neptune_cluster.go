package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/neptune"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// neptuneCluster provisions the aws_neptune_cluster itself -- the
// shared-storage brain: endpoints, backups, encryption, and engine
// lifecycle. The cluster composes onto its neighbors instead of embedding
// them: subnets, security groups, KMS keys, and IAM roles attach by
// reference, and database ingress rules live on the referenced
// AwsSecurityGroup nodes -- this module never creates or mutates a
// resource that deserves to be its own node.
//
// Neptune has no master username or password -- access is network
// reachability plus (optionally) IAM database authentication.
//
// Create-only in AWS: the identifier, port, subnet group, availability
// zones, storage encryption + KMS key, and snapshot_identifier.
// Everything else updates in place (immediately or at the next
// maintenance window, per apply_immediately).
func neptuneCluster(ctx *pulumi.Context, locals *Locals, provider *aws.Provider,
	createdSubnetGroup *neptune.SubnetGroup, createdParameterGroup *neptune.ClusterParameterGroup) (*neptune.Cluster, error) {
	spec := locals.AwsNeptuneCluster.Spec

	args := &neptune.ClusterArgs{
		ClusterIdentifier: pulumi.String(locals.ClusterIdentifier),
		Engine:            pulumi.String("neptune"),
		Tags:              pulumi.ToStringMap(locals.AwsTags),

		// Deletion safety: the spec's CEL contract requires a
		// final-snapshot name unless skipping is explicit, so a delete can
		// never fail late on a missing snapshot identifier.
		SkipFinalSnapshot:  pulumi.Bool(spec.SkipFinalSnapshot),
		DeletionProtection: pulumi.Bool(spec.DeletionProtection),

		// Storage encryption is a one-way door: it can only be chosen at
		// create time, which is why the spec defaults it to TRUE (the
		// platform materializes the default; an explicit false is for
		// restore/global-join shapes whose source is unencrypted).
		StorageEncrypted: pulumi.Bool(spec.GetStorageEncrypted()),

		// SigV4-signed requests from IAM identities -- Neptune's only
		// credential mechanism.
		IamDatabaseAuthenticationEnabled: pulumi.Bool(spec.IamDatabaseAuthenticationEnabled),

		CopyTagsToSnapshot:       pulumi.Bool(spec.CopyTagsToSnapshot),
		ApplyImmediately:         pulumi.Bool(spec.ApplyImmediately),
		AllowMajorVersionUpgrade: pulumi.Bool(spec.AllowMajorVersionUpgrade),
	}

	// Empty pins nothing: AWS picks the current default Neptune version,
	// so an unpinned manifest never goes stale.
	if spec.EngineVersion != "" {
		args.EngineVersion = pulumi.String(spec.EngineVersion)
	}

	// Networking: the subnet group managed here (or referenced), the VPC
	// default SG when no groups are given (AWS's own default), and
	// AWS-picked AZ spread unless explicitly pinned (the list is
	// create-only -- letting AWS choose is almost always right).
	if createdSubnetGroup != nil {
		args.NeptuneSubnetGroupName = createdSubnetGroup.Name
	} else if spec.NeptuneSubnetGroupName != "" {
		args.NeptuneSubnetGroupName = pulumi.String(spec.NeptuneSubnetGroupName)
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
	// 0 keeps the AWS default (8182). Create-only -- a port change
	// replaces the cluster.
	if spec.Port != 0 {
		args.Port = pulumi.Int(int(spec.Port))
	}

	// "iopt1" opts into I/O-Optimized storage (engine 1.3+); empty keeps
	// standard per-I/O billing (the AWS default).
	if spec.StorageType != "" {
		args.StorageType = pulumi.String(spec.StorageType)
	}

	if spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyArn = pulumi.String(spec.KmsKeyId.GetValue())
	}

	// Roles the ENGINE assumes for the S3 bulk loader and Neptune ML.
	// The roles own their policies -- this cluster only associates them
	// (a module never mutates a resource it references).
	if len(spec.IamRoles) > 0 {
		iamRoles := pulumi.StringArray{}
		for _, iamRole := range spec.IamRoles {
			iamRoles = append(iamRoles, pulumi.String(iamRole.GetValue()))
		}
		args.IamRoles = iamRoles
	}

	// Backups: Neptune backups are continuous; retention bounds the
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

	// "audit" and "slowquery" -- both also need their matching cluster
	// parameters (neptune_enable_audit_log / the slow-query threshold
	// parameters) before Neptune emits anything.
	if len(spec.EnabledCloudwatchLogsExports) > 0 {
		args.EnableCloudwatchLogsExports = pulumi.ToStringArray(spec.EnabledCloudwatchLogsExports)
	}

	// Neptune Serverless: this block + db.serverless instances. NCU
	// bounds are 1-128 on both ends.
	if spec.ServerlessV2Scaling != nil {
		args.ServerlessV2ScalingConfiguration = &neptune.ClusterServerlessV2ScalingConfigurationArgs{
			MinCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MinCapacity),
			MaxCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MaxCapacity),
		}
	}

	// Create-time restore source: a manual or automated cluster
	// snapshot.
	if spec.SnapshotIdentifier != "" {
		args.SnapshotIdentifier = pulumi.String(spec.SnapshotIdentifier)
	}

	// Cross-cluster replication and Neptune global database membership.
	if spec.ReplicationSourceIdentifier != "" {
		args.ReplicationSourceIdentifier = pulumi.String(spec.ReplicationSourceIdentifier)
	}
	if spec.GlobalClusterIdentifier != "" {
		args.GlobalClusterIdentifier = pulumi.String(spec.GlobalClusterIdentifier)
	}

	// Parameter groups: the managed inline group, an existing referenced
	// group, or the engine default. neptune_instance_parameter_group_name
	// is only consulted by AWS during a major version upgrade (the
	// spec's CEL requires it alongside allow_major_version_upgrade).
	if createdParameterGroup != nil {
		args.NeptuneClusterParameterGroupName = createdParameterGroup.Name
	} else if spec.NeptuneClusterParameterGroupName != "" {
		args.NeptuneClusterParameterGroupName = pulumi.String(spec.NeptuneClusterParameterGroupName)
	}
	if spec.NeptuneInstanceParameterGroupName != "" {
		args.NeptuneInstanceParameterGroupName = pulumi.String(spec.NeptuneInstanceParameterGroupName)
	}

	createdCluster, err := neptune.NewCluster(ctx, "neptune-cluster", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Neptune cluster")
	}
	return createdCluster, nil
}
