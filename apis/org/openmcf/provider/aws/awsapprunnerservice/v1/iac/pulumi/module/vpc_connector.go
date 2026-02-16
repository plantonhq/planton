package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/internal/valuefrom"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apprunner"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// vpcConnector creates an App Runner VPC Connector so the service can reach
// resources inside a VPC (databases, caches, internal APIs). It is only called
// when subnet_ids are provided in the spec (and no existing vpc_connector_arn
// is referenced).
func vpcConnector(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*apprunner.VpcConnector, error) {
	spec := locals.AwsAppRunnerService.Spec
	resourceName := locals.AwsAppRunnerService.Metadata.Name

	// Resolve repeated StringValueOrRef fields to plain string slices.
	subnetIds := valuefrom.ToStringArray(spec.GetSubnetIds())
	securityGroupIds := valuefrom.ToStringArray(spec.GetSecurityGroupIds())

	connector, err := apprunner.NewVpcConnector(ctx, resourceName+"-vpc-connector", &apprunner.VpcConnectorArgs{
		VpcConnectorName: pulumi.String(resourceName),
		Subnets:          pulumi.ToStringArray(subnetIds),
		SecurityGroups:   pulumi.ToStringArray(securityGroupIds),
		Tags:             pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create App Runner VPC connector")
	}

	// Export the VPC Connector ARN for cross-resource references.
	ctx.Export(OpVpcConnectorArn, connector.Arn)

	return connector, nil
}
