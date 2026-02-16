package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/redshift"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func redshiftCluster(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdSg *ec2.SecurityGroup,
	createdSubnetGroup *redshift.SubnetGroup,
	createdParamGroup *redshift.ParameterGroup,
) (*redshift.Cluster, error) {
	spec := locals.AwsRedshiftCluster.Spec

	args := &redshift.ClusterArgs{
		ClusterIdentifier:                pulumi.String(locals.AwsRedshiftCluster.Metadata.Id),
		NodeType:                         pulumi.String(spec.NodeType),
		NumberOfNodes:                    pulumi.Int(int(spec.GetNumberOfNodes())),
		DatabaseName:                     pulumi.String(spec.GetDatabaseName()),
		Port:                             pulumi.Int(int(spec.GetPort())),
		AutomatedSnapshotRetentionPeriod: pulumi.Int(int(spec.GetAutomatedSnapshotRetentionPeriod())),
		SkipFinalSnapshot:                pulumi.Bool(spec.SkipFinalSnapshot),
		FinalSnapshotIdentifier:          pulumi.String(spec.FinalSnapshotIdentifier),
		AllowVersionUpgrade:              pulumi.Bool(spec.GetAllowVersionUpgrade()),
		ApplyImmediately:                 pulumi.Bool(spec.ApplyImmediately),
		PubliclyAccessible:               pulumi.Bool(spec.PubliclyAccessible),
		EnhancedVpcRouting:               pulumi.Bool(spec.EnhancedVpcRouting),
		MultiAz:                          pulumi.Bool(spec.MultiAz),
		Tags:                             pulumi.ToStringMap(locals.Labels),
	}

	// Encrypted uses StringPtrInput in the Pulumi SDK (Terraform nullable bool quirk)
	if spec.Encrypted != nil {
		args.Encrypted = pulumi.String(fmt.Sprintf("%t", spec.GetEncrypted()))
	}

	if spec.PreferredMaintenanceWindow != "" {
		args.PreferredMaintenanceWindow = pulumi.String(spec.PreferredMaintenanceWindow)
	}

	if spec.MaintenanceTrackName != "" {
		args.MaintenanceTrackName = pulumi.String(spec.MaintenanceTrackName)
	}

	// KMS key for encryption
	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
	}

	// Master user credentials handling
	if spec.ManageMasterPassword {
		args.ManageMasterPassword = pulumi.Bool(true)
		if spec.MasterPasswordSecretKmsKeyId != nil && spec.MasterPasswordSecretKmsKeyId.GetValue() != "" {
			args.MasterPasswordSecretKmsKeyId = pulumi.String(spec.MasterPasswordSecretKmsKeyId.GetValue())
		}
		if spec.GetMasterUsername() != "" {
			args.MasterUsername = pulumi.String(spec.GetMasterUsername())
		}
	} else {
		if spec.GetMasterUsername() != "" {
			args.MasterUsername = pulumi.String(spec.GetMasterUsername())
		}
		if spec.MasterPassword != "" {
			args.MasterPassword = pulumi.String(spec.MasterPassword)
		}
	}

	// Subnet group selection
	if createdSubnetGroup != nil {
		args.ClusterSubnetGroupName = createdSubnetGroup.Name
	} else if spec.ClusterSubnetGroupName != nil && spec.ClusterSubnetGroupName.GetValue() != "" {
		args.ClusterSubnetGroupName = pulumi.String(spec.ClusterSubnetGroupName.GetValue())
	}

	// Parameter group
	if createdParamGroup != nil {
		args.ClusterParameterGroupName = createdParamGroup.Name
	} else if spec.ClusterParameterGroupName != "" {
		args.ClusterParameterGroupName = pulumi.String(spec.ClusterParameterGroupName)
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

	// IAM roles
	var iamRoles pulumi.StringArray
	for _, role := range spec.IamRoles {
		if role.GetValue() != "" {
			iamRoles = append(iamRoles, pulumi.String(role.GetValue()))
		}
	}
	if len(iamRoles) > 0 {
		args.IamRoles = iamRoles
	}

	// Default IAM role
	if spec.DefaultIamRoleArn != nil && spec.DefaultIamRoleArn.GetValue() != "" {
		args.DefaultIamRoleArn = pulumi.String(spec.DefaultIamRoleArn.GetValue())
	}

	cluster, err := redshift.NewCluster(ctx, "redshift-cluster", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create redshift cluster")
	}
	return cluster, nil
}
