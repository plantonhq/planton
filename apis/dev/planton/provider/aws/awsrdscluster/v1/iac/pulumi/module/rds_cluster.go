package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func rdsCluster(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSg *ec2.SecurityGroup,
	createdSubnetGroup *rds.SubnetGroup,
	createdParamGroup *rds.ClusterParameterGroup,
) (*rds.Cluster, error) {
	spec := locals.AwsRdsCluster.Spec

	args := &rds.ClusterArgs{
		ClusterIdentifier:                pulumi.String(locals.AwsRdsCluster.Metadata.Id),
		Engine:                           pulumi.String(spec.Engine),
		EngineVersion:                    pulumi.String(spec.EngineVersion),
		DatabaseName:                     pulumi.String(spec.DatabaseName),
		Port:                             pulumi.Int(spec.Port),
		DeletionProtection:               pulumi.Bool(spec.DeletionProtection),
		PreferredMaintenanceWindow:       pulumi.String(spec.PreferredMaintenanceWindow),
		BackupRetentionPeriod:            pulumi.Int(spec.BackupRetentionPeriod),
		PreferredBackupWindow:            pulumi.String(spec.PreferredBackupWindow),
		CopyTagsToSnapshot:               pulumi.Bool(spec.CopyTagsToSnapshot),
		SkipFinalSnapshot:                pulumi.Bool(spec.SkipFinalSnapshot),
		FinalSnapshotIdentifier:          pulumi.String(spec.FinalSnapshotIdentifier),
		IamDatabaseAuthenticationEnabled: pulumi.Bool(spec.IamDatabaseAuthenticationEnabled),
		EnableHttpEndpoint:               pulumi.Bool(spec.EnableHttpEndpoint),
		Tags:                             pulumi.ToStringMap(locals.Labels),
		EnabledCloudwatchLogsExports:     pulumi.ToStringArray(spec.EnabledCloudwatchLogsExports),
		StorageEncrypted:                 pulumi.Bool(spec.StorageEncrypted),
		EngineMode:                       pulumi.String(spec.EngineMode),
	}

	// Storage type (aurora or aurora-iopt1)
	if spec.StorageType != "" {
		args.StorageType = pulumi.String(spec.StorageType)
	}

	// Serverless v2 scaling configuration
	if spec.ServerlessV2Scaling != nil {
		args.Serverlessv2ScalingConfiguration = &rds.ClusterServerlessv2ScalingConfigurationArgs{
			MinCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MinCapacity),
			MaxCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MaxCapacity),
		}
	}

	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
	}
	if spec.ReplicationSourceIdentifier != "" {
		args.ReplicationSourceIdentifier = pulumi.String(spec.ReplicationSourceIdentifier)
	}
	if spec.SnapshotIdentifier != "" {
		args.SnapshotIdentifier = pulumi.String(spec.SnapshotIdentifier)
	}

	// Master user credentials handling
	if spec.ManageMasterUserPassword {
		args.ManageMasterUserPassword = pulumi.Bool(true)
		if spec.MasterUserSecretKmsKeyId != nil && spec.MasterUserSecretKmsKeyId.GetValue() != "" {
			args.MasterUserSecretKmsKeyId = pulumi.String(spec.MasterUserSecretKmsKeyId.GetValue())
		}
		if spec.GetUsername() != "" {
			args.MasterUsername = pulumi.String(spec.GetUsername())
		}
	} else {
		if spec.GetUsername() != "" {
			args.MasterUsername = pulumi.String(spec.GetUsername())
		}
		if spec.Password != "" {
			args.MasterPassword = pulumi.String(spec.Password)
		}
	}

	// Subnet group selection
	if createdSubnetGroup != nil {
		args.DbSubnetGroupName = createdSubnetGroup.Name
	} else if spec.DbSubnetGroupName != nil && spec.DbSubnetGroupName.GetValue() != "" {
		args.DbSubnetGroupName = pulumi.String(spec.DbSubnetGroupName.GetValue())
	}

	// Parameter group
	if createdParamGroup != nil {
		args.DbClusterParameterGroupName = createdParamGroup.Name
	} else if spec.DbClusterParameterGroupName != "" {
		args.DbClusterParameterGroupName = pulumi.String(spec.DbClusterParameterGroupName)
	}

	// Security groups (associate existing + created if present)
	var vpcSecurityGroupIds pulumi.StringArray
	for _, sg := range spec.AssociateSecurityGroupIds {
		vpcSecurityGroupIds = append(vpcSecurityGroupIds, pulumi.String(sg.GetValue()))
	}
	if createdSg != nil {
		vpcSecurityGroupIds = append(vpcSecurityGroupIds, createdSg.ID())
	}
	if len(vpcSecurityGroupIds) > 0 {
		args.VpcSecurityGroupIds = vpcSecurityGroupIds
	}

	cluster, err := rds.NewCluster(ctx, "rds-cluster", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create rds cluster")
	}
	return cluster, nil
}
