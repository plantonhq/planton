package module

import (
	"github.com/pkg/errors"
	awsrdsinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsrdsinstance/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of an AWS RDS DB instance and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsrdsinstancev1.AwsRdsInstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsRdsInstance.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsRdsInstance.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// Subnet group (only when subnetIds provided and no name supplied)
	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "subnet group")
	}

	// Create the RDS Instance
	instance, err := dbInstance(ctx, locals, provider, createdSubnetGroup)
	if err != nil {
		return errors.Wrap(err, "rds instance")
	}

	// Export outputs as defined in AwsRdsInstanceStackOutputs
	ctx.Export(OpRdsInstanceId, instance.ID())
	ctx.Export(OpRdsInstanceArn, instance.Arn)
	ctx.Export(OpRdsInstanceEndpoint, instance.Address)
	ctx.Export(OpRdsInstancePort, instance.Port)
	if createdSubnetGroup != nil {
		ctx.Export(OpRdsSubnetGroup, createdSubnetGroup.Name)
	} else if locals.AwsRdsInstance.Spec != nil && locals.AwsRdsInstance.Spec.DbSubnetGroupName != nil && locals.AwsRdsInstance.Spec.DbSubnetGroupName.GetValue() != "" {
		ctx.Export(OpRdsSubnetGroup, pulumi.String(locals.AwsRdsInstance.Spec.DbSubnetGroupName.GetValue()))
	}
	if locals.AwsRdsInstance.Spec != nil && locals.AwsRdsInstance.Spec.ParameterGroupName != "" {
		ctx.Export(OpRdsParameterGroup, pulumi.String(locals.AwsRdsInstance.Spec.ParameterGroupName))
	}
	if locals.AwsRdsInstance.Spec != nil && len(locals.AwsRdsInstance.Spec.SecurityGroupIds) > 0 {
		first := locals.AwsRdsInstance.Spec.SecurityGroupIds[0].GetValue()
		if first != "" {
			ctx.Export(OpRdsSecurityGroup, pulumi.String(first))
		}
	}
	return nil
}
