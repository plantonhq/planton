package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apprunner"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// vpcConnector creates the App Runner VPC connector -- a set of AWS-managed
// ENIs provisioned into the referenced subnets that App Runner services route
// their OUTBOUND traffic through to reach private VPC resources. AWS has no
// update API for connectors: changing subnets or security groups replaces
// the connector (registered as a new revision under the same name).
func vpcConnector(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.AwsAppRunnerVpcConnector.Spec

	// Referenced subnet/SG IDs arrive pre-resolved by the orchestrator;
	// GetValue() unwraps the literal arm.
	subnetIds := make([]string, 0, len(spec.SubnetIds))
	for _, s := range spec.SubnetIds {
		subnetIds = append(subnetIds, s.GetValue())
	}

	// AWS requires at least one security group on a connector (spec-
	// enforced) -- the groups govern what the connected services can reach.
	securityGroupIds := make([]string, 0, len(spec.SecurityGroupIds))
	for _, sg := range spec.SecurityGroupIds {
		securityGroupIds = append(securityGroupIds, sg.GetValue())
	}

	createdConnector, err := apprunner.NewVpcConnector(ctx, locals.AwsAppRunnerVpcConnector.Metadata.Name, &apprunner.VpcConnectorArgs{
		// The cloud name is metadata.name -- the same basis the Terraform
		// module uses, so both engines create the same physical identity.
		VpcConnectorName: pulumi.String(locals.AwsAppRunnerVpcConnector.Metadata.Name),
		Subnets:          pulumi.ToStringArray(subnetIds),
		SecurityGroups:   pulumi.ToStringArray(securityGroupIds),
		Tags:             pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create VPC connector")
	}

	// Export outputs matching AwsAppRunnerVpcConnectorOutputs.
	ctx.Export(OpVpcConnectorArn, createdConnector.Arn)
	ctx.Export(OpVpcConnectorRevision, createdConnector.VpcConnectorRevision)
	ctx.Export(OpStatus, createdConnector.Status)

	return nil
}
