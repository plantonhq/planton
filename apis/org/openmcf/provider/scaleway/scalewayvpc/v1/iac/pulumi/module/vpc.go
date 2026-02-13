package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/network"
)

// vpc provisions the Scaleway VPC and exports its ID.
func vpc(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) (*network.Vpc, error) {

	// 1. Build the resource arguments from the proto fields.
	vpcArgs := &network.VpcArgs{
		Name:   pulumi.String(locals.ScalewayVpc.Metadata.Name),
		Region: pulumi.String(locals.ScalewayVpc.Spec.Region),
		Tags:   pulumi.ToStringArray(locals.ScalewayTags),
	}

	// 2. Set routing flags when enabled.
	// These are one-way toggles: once enabled, they cannot be disabled.
	if locals.ScalewayVpc.Spec.EnableRouting {
		vpcArgs.EnableRouting = pulumi.Bool(true)
	}
	if locals.ScalewayVpc.Spec.EnableCustomRoutesPropagation {
		vpcArgs.EnableCustomRoutesPropagation = pulumi.Bool(true)
	}

	// 3. Create the VPC.
	createdVpc, err := network.NewVpc(
		ctx,
		"vpc",
		vpcArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scaleway vpc")
	}

	// 4. Export stack output.
	ctx.Export(OpVpcId, createdVpc.ID())

	return createdVpc, nil
}
