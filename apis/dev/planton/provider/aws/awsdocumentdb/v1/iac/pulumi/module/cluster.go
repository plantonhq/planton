package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/docdb"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func docdbCluster(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSg *ec2.SecurityGroup,
	createdSubnetGroup *docdb.SubnetGroup,
	createdParamGroup *docdb.ClusterParameterGroup,
) (*docdb.Cluster, error) {
	spec := locals.AwsDocumentDb.Spec

	args := &docdb.ClusterArgs{
		ClusterIdentifier:            pulumi.String(locals.AwsDocumentDb.Metadata.Id),
		Engine:                       pulumi.String("docdb"),
		EngineVersion:                pulumi.String(spec.GetEngineVersion()),
		MasterUsername:               pulumi.String(spec.GetMasterUsername()),
		MasterPassword:               pulumi.String(spec.MasterPassword),
		Port:                         pulumi.Int(getEffectivePort(spec)),
		DeletionProtection:           pulumi.Bool(spec.DeletionProtection),
		StorageEncrypted:             pulumi.Bool(spec.GetStorageEncrypted()),
		BackupRetentionPeriod:        pulumi.Int(spec.GetBackupRetentionPeriod()),
		SkipFinalSnapshot:            pulumi.Bool(spec.GetSkipFinalSnapshot()),
		ApplyImmediately:             pulumi.Bool(spec.ApplyImmediately),
		Tags:                         pulumi.ToStringMap(locals.Labels),
		EnabledCloudwatchLogsExports: pulumi.ToStringArray(spec.EnabledCloudwatchLogsExports),
	}

	// Preferred backup window
	if spec.PreferredBackupWindow != "" {
		args.PreferredBackupWindow = pulumi.String(spec.PreferredBackupWindow)
	}

	// Preferred maintenance window
	if spec.PreferredMaintenanceWindow != "" {
		args.PreferredMaintenanceWindow = pulumi.String(spec.PreferredMaintenanceWindow)
	}

	// Final snapshot identifier (required when skip_final_snapshot is false)
	if spec.FinalSnapshotIdentifier != "" {
		args.FinalSnapshotIdentifier = pulumi.String(spec.FinalSnapshotIdentifier)
	}

	// KMS key for storage encryption
	if spec.KmsKey != nil && spec.KmsKey.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKey.GetValue())
	}

	// Subnet group selection
	if createdSubnetGroup != nil {
		args.DbSubnetGroupName = createdSubnetGroup.Name
	} else if spec.DbSubnetGroup != nil && spec.DbSubnetGroup.GetValue() != "" {
		args.DbSubnetGroupName = pulumi.String(spec.DbSubnetGroup.GetValue())
	}

	// Parameter group
	if createdParamGroup != nil {
		args.DbClusterParameterGroupName = createdParamGroup.Name
	} else if spec.ClusterParameterGroupName != "" {
		args.DbClusterParameterGroupName = pulumi.String(spec.ClusterParameterGroupName)
	}

	// Security groups
	var vpcSecurityGroupIds pulumi.StringArray
	for _, sg := range spec.SecurityGroups {
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

	cluster, err := docdb.NewCluster(ctx, "docdb-cluster", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create documentdb cluster")
	}
	return cluster, nil
}
