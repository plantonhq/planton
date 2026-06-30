package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// vpc provisions the VPC itself and exports its ID.
func vpc(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.Vpc, error) {

	// 1. Build the resource arguments straight from the proto fields.
	vpcArgs := &digitalocean.VpcArgs{
		Name:   pulumi.String(locals.DigitalOceanVpc.Metadata.Name),
		Region: pulumi.String(locals.DigitalOceanVpc.Spec.Region.String()),
	}

	// 2. Add optional description if provided
	if locals.DigitalOceanVpc.Spec.Description != "" {
		vpcArgs.Description = pulumi.String(locals.DigitalOceanVpc.Spec.Description)
	}

	// 3. Add IP range if explicitly specified (80/20: optional for auto-generation)
	// When omitted, DigitalOcean auto-generates a non-conflicting /20 CIDR block
	if locals.DigitalOceanVpc.Spec.IpRangeCidr != "" {
		vpcArgs.IpRange = pulumi.String(locals.DigitalOceanVpc.Spec.IpRangeCidr)
	}

	// 4. Create the VPC.
	createdVpc, err := digitalocean.NewVpc(
		ctx,
		"vpc",
		vpcArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean vpc")
	}

	// 5. Export stack output.
	ctx.Export(OpVpcId, createdVpc.ID())

	return createdVpc, nil
}
