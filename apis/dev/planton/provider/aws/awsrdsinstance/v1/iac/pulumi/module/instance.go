package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dbInstance creates an RDS DB Instance based on the provided spec and optional subnet group.
func dbInstance(ctx *pulumi.Context, locals *Locals, provider *aws.Provider, createdSubnetGroup *rds.SubnetGroup) (*rds.Instance, error) {
	spec := locals.AwsRdsInstance.Spec

	// Use Metadata.Id if available, otherwise fall back to Metadata.Name (consistent with Terraform version)
	instanceIdentifier := locals.AwsRdsInstance.Metadata.Id
	if instanceIdentifier == "" {
		instanceIdentifier = locals.AwsRdsInstance.Metadata.Name
	}

	args := &rds.InstanceArgs{
		Identifier:         pulumi.String(instanceIdentifier),
		Engine:             pulumi.String(spec.Engine),
		EngineVersion:      pulumi.String(spec.EngineVersion),
		InstanceClass:      pulumi.String(spec.InstanceClass),
		AllocatedStorage:   pulumi.Int(int(spec.AllocatedStorageGb)),
		StorageEncrypted:   pulumi.Bool(spec.StorageEncrypted),
		Port:               pulumi.Int(int(spec.Port)),
		PubliclyAccessible: pulumi.Bool(spec.PubliclyAccessible),
		MultiAz:            pulumi.Bool(spec.MultiAz),
		Tags:               pulumi.ToStringMap(locals.Labels),
	}

	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
	}
	if createdSubnetGroup != nil {
		args.DbSubnetGroupName = createdSubnetGroup.Name
	} else if spec.DbSubnetGroupName != nil && spec.DbSubnetGroupName.GetValue() != "" {
		args.DbSubnetGroupName = pulumi.String(spec.DbSubnetGroupName.GetValue())
	}
	if len(spec.SecurityGroupIds) > 0 {
		var sgIds pulumi.StringArray
		for _, sg := range spec.SecurityGroupIds {
			if sg.GetValue() != "" {
				sgIds = append(sgIds, pulumi.String(sg.GetValue()))
			}
		}
		if len(sgIds) > 0 {
			args.VpcSecurityGroupIds = sgIds
		}
	}
	if spec.ParameterGroupName != "" {
		args.ParameterGroupName = pulumi.String(spec.ParameterGroupName)
	}
	if spec.OptionGroupName != "" {
		args.OptionGroupName = pulumi.String(spec.OptionGroupName)
	}
	if spec.Username != "" {
		args.Username = pulumi.String(spec.Username)
	}
	if spec.Password != "" {
		args.Password = pulumi.String(spec.Password)
	}

	inst, err := rds.NewInstance(ctx, "rds-instance", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create rds instance")
	}
	return inst, nil
}
