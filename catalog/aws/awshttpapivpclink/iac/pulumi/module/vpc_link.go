package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// vpcLink creates the API Gateway v2 VPC link -- a set of AWS-managed ENIs
// provisioned into the referenced subnets that HTTP API private integrations
// route through to reach internal ALBs, NLBs, or Cloud Map services. AWS has
// no update API for the network attachment: changing subnets or security
// groups replaces the link (only the name mutates in place).
func vpcLink(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.AwsHttpApiVpcLink.Spec

	// Referenced subnet/SG IDs arrive pre-resolved by the orchestrator;
	// GetValue() unwraps the literal arm.
	subnetIds := make([]string, 0, len(spec.SubnetIds))
	for _, s := range spec.SubnetIds {
		subnetIds = append(subnetIds, s.GetValue())
	}

	// Security groups may be empty: AWS then applies no filtering on the
	// link side and reachability is governed solely by the target's
	// security groups.
	securityGroupIds := make([]string, 0, len(spec.SecurityGroupIds))
	for _, sg := range spec.SecurityGroupIds {
		securityGroupIds = append(securityGroupIds, sg.GetValue())
	}

	createdVpcLink, err := apigatewayv2.NewVpcLink(ctx, locals.AwsHttpApiVpcLink.Metadata.Name, &apigatewayv2.VpcLinkArgs{
		// The cloud name is metadata.name -- the same basis the Terraform
		// module uses, so both engines create the same physical identity.
		Name:             pulumi.String(locals.AwsHttpApiVpcLink.Metadata.Name),
		SubnetIds:        pulumi.ToStringArray(subnetIds),
		SecurityGroupIds: pulumi.ToStringArray(securityGroupIds),
		Tags:             pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create VPC link")
	}

	// Export outputs matching AwsHttpApiVpcLinkOutputs.
	ctx.Export(OpVpcLinkId, createdVpcLink.ID())
	ctx.Export(OpVpcLinkArn, createdVpcLink.Arn)

	return nil
}
