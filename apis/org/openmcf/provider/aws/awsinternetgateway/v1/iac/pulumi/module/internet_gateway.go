package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func internetGateway(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) (*ec2.InternetGateway, error) {
	spec := locals.AwsInternetGateway.Spec
	name := locals.AwsInternetGateway.Metadata.Name

	createdInternetGateway, err := ec2.NewInternetGateway(ctx, name, &ec2.InternetGatewayArgs{
		VpcId: pulumi.String(spec.VpcId.GetValue()),
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
			stringmaps.AddEntry(locals.AwsTags, "Name", name)),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create internet gateway")
	}

	ctx.Export(OpInternetGatewayId, createdInternetGateway.ID())
	ctx.Export(OpInternetGatewayArn, createdInternetGateway.Arn)
	ctx.Export(OpVpcId, pulumi.String(spec.VpcId.GetValue()))
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return createdInternetGateway, nil
}
