package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/neptune"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func neptuneCluster(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSg *ec2.SecurityGroup,
	createdSubnetGroup *neptune.SubnetGroup,
	createdParamGroup *neptune.ClusterParameterGroup,
) (*neptune.Cluster, error) {
	spec := locals.AwsNeptuneCluster.Spec

	args := &neptune.ClusterArgs{
		ClusterIdentifier:                pulumi.String(locals.AwsNeptuneCluster.Metadata.Id),
		Engine:                           pulumi.String("neptune"),
		EngineVersion:                    pulumi.String(spec.GetEngineVersion()),
		Port:                             pulumi.Int(getEffectivePort(spec)),
		StorageEncrypted:                 pulumi.Bool(spec.GetStorageEncrypted()),
		DeletionProtection:               pulumi.Bool(spec.DeletionProtection),
		BackupRetentionPeriod:            pulumi.Int(spec.GetBackupRetentionPeriod()),
		SkipFinalSnapshot:                pulumi.Bool(spec.GetSkipFinalSnapshot()),
		ApplyImmediately:                 pulumi.Bool(spec.ApplyImmediately),
		CopyTagsToSnapshot:              pulumi.Bool(spec.CopyTagsToSnapshot),
		AllowMajorVersionUpgrade:         pulumi.Bool(spec.AllowMajorVersionUpgrade),
		IamDatabaseAuthenticationEnabled: pulumi.Bool(spec.IamDatabaseAuthenticationEnabled),
		Tags:                             pulumi.ToStringMap(locals.Labels),
		EnableCloudwatchLogsExports:      pulumi.ToStringArray(spec.EnabledCloudwatchLogsExports),
	}

	// Storage type
	if spec.StorageType != "" {
		args.StorageType = pulumi.String(spec.StorageType)
	}

	// Serverless v2 scaling configuration
	if spec.ServerlessV2Scaling != nil {
		args.ServerlessV2ScalingConfiguration = &neptune.ClusterServerlessV2ScalingConfigurationArgs{
			MinCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MinCapacity),
			MaxCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MaxCapacity),
		}
	}

	// Preferred backup window
	if spec.PreferredBackupWindow != "" {
		args.PreferredBackupWindow = pulumi.String(spec.PreferredBackupWindow)
	}

	// Preferred maintenance window
	if spec.PreferredMaintenanceWindow != "" {
		args.PreferredMaintenanceWindow = pulumi.String(spec.PreferredMaintenanceWindow)
	}

	// Final snapshot identifier
	if spec.FinalSnapshotIdentifier != "" {
		args.FinalSnapshotIdentifier = pulumi.String(spec.FinalSnapshotIdentifier)
	}

	// KMS key for storage encryption
	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyArn = pulumi.String(spec.KmsKeyId.GetValue())
	}

	// IAM roles for S3 data loading and other service integrations
	if len(spec.IamRoles) > 0 {
		var roleArns pulumi.StringArray
		for _, role := range spec.IamRoles {
			if role.GetValue() != "" {
				roleArns = append(roleArns, pulumi.String(role.GetValue()))
			}
		}
		if len(roleArns) > 0 {
			args.IamRoles = roleArns
		}
	}

	// Subnet group selection
	if createdSubnetGroup != nil {
		args.NeptuneSubnetGroupName = createdSubnetGroup.Name
	} else if spec.NeptuneSubnetGroupName != nil && spec.NeptuneSubnetGroupName.GetValue() != "" {
		args.NeptuneSubnetGroupName = pulumi.String(spec.NeptuneSubnetGroupName.GetValue())
	}

	// Parameter group
	if createdParamGroup != nil {
		args.NeptuneClusterParameterGroupName = createdParamGroup.Name
	} else if spec.ClusterParameterGroupName != "" {
		args.NeptuneClusterParameterGroupName = pulumi.String(spec.ClusterParameterGroupName)
	}

	// Security groups
	var vpcSecurityGroupIds pulumi.StringArray
	for _, sg := range spec.SecurityGroupIds {
		if sg.GetValue() != "" {
			vpcSecurityGroupIds = append(vpcSecurityGroupIds, pulumi.String(sg.GetValue()))
		}
	}
	if createdSg != nil {
		vpcSecurityGroupIds = append(vpcSecurityGroupIds, createdSg.ID())
	}
	if len(vpcSecurityGroupIds) > 0 {
		args.VpcSecurityGroupIds = vpcSecurityGroupIds
	}

	cluster, err := neptune.NewCluster(ctx, "neptune-cluster", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create neptune cluster")
	}
	return cluster, nil
}
