package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func egressOnlyInternetGateway(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) (*ec2.EgressOnlyInternetGateway, error) {
	spec := locals.AwsEgressOnlyInternetGateway.Spec
	name := locals.AwsEgressOnlyInternetGateway.Metadata.Name

	createdEgressOnlyInternetGateway, err := ec2.NewEgressOnlyInternetGateway(ctx, name, &ec2.EgressOnlyInternetGatewayArgs{
		VpcId: pulumi.String(spec.VpcId.GetValue()),
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
			stringmaps.AddEntry(locals.AwsTags, "Name", name)),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create egress-only internet gateway")
	}

	ctx.Export(OpEgressOnlyInternetGatewayId, createdEgressOnlyInternetGateway.ID())
	ctx.Export(OpVpcId, pulumi.String(spec.VpcId.GetValue()))
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return createdEgressOnlyInternetGateway, nil
}
