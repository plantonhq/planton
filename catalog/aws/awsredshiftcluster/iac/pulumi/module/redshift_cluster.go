package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/redshift"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// redshiftCluster provisions the cluster itself -- the warehouse brain:
// compute topology, credentials, encryption, snapshots, maintenance, and
// restore shapes. Create-only in AWS: the identifier, the subnet group,
// the master username, and the snapshot restore sources. Node type and
// node count are NOT create-only -- changing them triggers an in-place
// (but access-interrupting) elastic/classic resize, never a replacement.
func redshiftCluster(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSubnetGroup *redshift.SubnetGroup,
	createdParameterGroup *redshift.ParameterGroup,
) (*redshift.Cluster, error) {
	spec := locals.AwsRedshiftCluster.Spec

	args := &redshift.ClusterArgs{
		ClusterIdentifier: pulumi.String(locals.ClusterIdentifier),

		// Compute topology. 0 nodes keeps the AWS default (1,
		// single-node); the provider derives cluster_type from the count,
		// so it is never set here.
		NodeType: pulumi.String(spec.NodeType),

		// Zone relocation moves the single cluster between AZs on outage
		// or demand (RA3-only; the port must sit in 5431-5455 or
		// 8191-8215 -- the default 5439 qualifies). Mutually exclusive
		// with multi_az (CEL-enforced): relocation moves, Multi-AZ fails
		// over to a standby.
		AvailabilityZoneRelocationEnabled: pulumi.Bool(spec.AvailabilityZoneRelocationEnabled),
		MultiAz:                           pulumi.Bool(spec.MultiAz),

		// Public reachability is opt-in; CEL ties elastic_ip to
		// publicly_accessible so the constraint fails at validate, not at
		// deploy.
		PubliclyAccessible: pulumi.Bool(spec.PubliclyAccessible),

		// Enhanced VPC routing forces COPY/UNLOAD data movement through
		// the VPC where flow logs and endpoints can see and govern it.
		EnhancedVpcRouting: pulumi.Bool(spec.EnhancedVpcRouting),

		// Deletion safety: the CEL contract requires a final-snapshot
		// name unless skipping is explicit, so a delete can never fail
		// late on a missing snapshot identifier.
		SkipFinalSnapshot: pulumi.Bool(spec.SkipFinalSnapshot),

		ApplyImmediately: pulumi.Bool(spec.ApplyImmediately),

		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.NumberOfNodes != 0 {
		args.NumberOfNodes = pulumi.Int(int(spec.NumberOfNodes))
	}

	// Empty pins nothing: "1.0" is the only version family Redshift has
	// ever shipped; actual engine patches ride the maintenance track.
	if spec.ClusterVersion != "" {
		args.ClusterVersion = pulumi.String(spec.ClusterVersion)
	}

	// Empty keeps the AWS default first database ("dev").
	if spec.DatabaseName != "" {
		args.DatabaseName = pulumi.String(spec.DatabaseName)
	}

	// Networking: the subnet group managed here (or referenced), the VPC
	// default SG when no groups are given (AWS's own default), and
	// AWS-picked AZ placement unless explicitly pinned.
	if createdSubnetGroup != nil {
		args.ClusterSubnetGroupName = createdSubnetGroup.Name
	} else if spec.ClusterSubnetGroupName != "" {
		args.ClusterSubnetGroupName = pulumi.String(spec.ClusterSubnetGroupName)
	}

	if len(spec.SecurityGroupIds) > 0 {
		securityGroupIds := pulumi.StringArray{}
		for _, securityGroupId := range spec.SecurityGroupIds {
			securityGroupIds = append(securityGroupIds, pulumi.String(securityGroupId.GetValue()))
		}
		args.VpcSecurityGroupIds = securityGroupIds
	}

	if spec.AvailabilityZone != "" {
		args.AvailabilityZone = pulumi.String(spec.AvailabilityZone)
	}

	// 0 keeps the AWS default (5439).
	if spec.Port != 0 {
		args.Port = pulumi.Int(int(spec.Port))
	}

	// The static leader-node address only exists on a public cluster;
	// Redshift takes the IP address itself, not an allocation ID.
	if spec.ElasticIp.GetValue() != "" {
		args.ElasticIp = pulumi.String(spec.ElasticIp.GetValue())
	}

	if spec.MasterUsername != "" {
		args.MasterUsername = pulumi.String(spec.MasterUsername)
	}

	// The password contract (CEL enforces exactly one strategy): the
	// AWS-managed Secrets Manager secret (recommended -- no secret in
	// manifest or state) or a directly supplied password.
	// ManageMasterPassword is forwarded ONLY when true: an explicit false
	// conflicts with master_password in the provider's ConflictsWith
	// machinery.
	if spec.ManageMasterPassword {
		args.ManageMasterPassword = pulumi.Bool(true)
		if spec.MasterPasswordSecretKmsKeyId.GetValue() != "" {
			args.MasterPasswordSecretKmsKeyId = pulumi.String(spec.MasterPasswordSecretKmsKeyId.GetValue())
		}
	} else if spec.MasterPassword != "" {
		args.MasterPassword = pulumi.String(spec.MasterPassword)
	}

	// Encryption at rest. AWS defaults new clusters to encrypted and the
	// spec keeps that default; toggling later is an in-place but
	// long-running migration. The provider models encrypted as a nullable
	// bool (a string under the hood -- a Terraform schema quirk the
	// Pulumi SDK inherits), hence the %t formatting.
	if spec.Encrypted != nil {
		args.Encrypted = pulumi.String(fmt.Sprintf("%t", spec.GetEncrypted()))
	}
	if spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
	}

	// IAM roles the warehouse assumes for COPY/UNLOAD/Spectrum. The
	// default role must also be in iam_roles (an AWS requirement the
	// error message makes obvious enough to leave to the API).
	if len(spec.IamRoles) > 0 {
		iamRoles := pulumi.StringArray{}
		for _, iamRole := range spec.IamRoles {
			iamRoles = append(iamRoles, pulumi.String(iamRole.GetValue()))
		}
		args.IamRoles = iamRoles
	}
	if spec.DefaultIamRoleArn.GetValue() != "" {
		args.DefaultIamRoleArn = pulumi.String(spec.DefaultIamRoleArn.GetValue())
	}

	// Snapshots: automated retention bounds recovery; manual retention 0
	// keeps the AWS default (-1, indefinite).
	if spec.AutomatedSnapshotRetentionPeriod != nil {
		args.AutomatedSnapshotRetentionPeriod = pulumi.Int(int(spec.GetAutomatedSnapshotRetentionPeriod()))
	}
	if spec.ManualSnapshotRetentionPeriod != 0 {
		args.ManualSnapshotRetentionPeriod = pulumi.Int(int(spec.ManualSnapshotRetentionPeriod))
	}

	// Maintenance: empty window lets AWS assign one; empty track keeps
	// "current" (the AWS default).
	if spec.PreferredMaintenanceWindow != "" {
		args.PreferredMaintenanceWindow = pulumi.String(spec.PreferredMaintenanceWindow)
	}
	if spec.MaintenanceTrackName != "" {
		args.MaintenanceTrackName = pulumi.String(spec.MaintenanceTrackName)
	}
	if spec.AllowVersionUpgrade != nil {
		args.AllowVersionUpgrade = pulumi.Bool(spec.GetAllowVersionUpgrade())
	}

	if spec.FinalSnapshotIdentifier != "" {
		args.FinalSnapshotIdentifier = pulumi.String(spec.FinalSnapshotIdentifier)
	}

	// Create-time restore sources (mutually exclusive, CEL-enforced): a
	// snapshot by name (with optional source-cluster disambiguation) or
	// by ARN (the cross-account/cross-region sharing shape).
	// owner_account covers snapshots shared by another AWS account. A
	// restored cluster inherits the snapshot's credentials, so
	// master_username stays empty.
	if spec.SnapshotIdentifier != "" {
		args.SnapshotIdentifier = pulumi.String(spec.SnapshotIdentifier)
	}
	if spec.SnapshotArn != "" {
		args.SnapshotArn = pulumi.String(spec.SnapshotArn)
	}
	if spec.SnapshotClusterIdentifier != "" {
		args.SnapshotClusterIdentifier = pulumi.String(spec.SnapshotClusterIdentifier)
	}
	if spec.OwnerAccount != "" {
		args.OwnerAccount = pulumi.String(spec.OwnerAccount)
	}

	// Parameter groups: the managed inline group, an existing referenced
	// group, or the Redshift default.
	if createdParameterGroup != nil {
		args.ClusterParameterGroupName = createdParameterGroup.Name
	} else if spec.ClusterParameterGroupName != "" {
		args.ClusterParameterGroupName = pulumi.String(spec.ClusterParameterGroupName)
	}

	createdCluster, err := redshift.NewCluster(ctx, "cluster", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Redshift cluster")
	}
	return createdCluster, nil
}
