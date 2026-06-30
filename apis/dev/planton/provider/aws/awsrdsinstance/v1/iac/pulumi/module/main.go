package module

import (
	"github.com/pkg/errors"
	awsrdsinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsrdsinstance/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of an AWS RDS DB instance and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsrdsinstancev1.AwsRdsInstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsRdsInstance.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
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
