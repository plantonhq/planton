package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// vpc provisions the VPC and exports its outputs.
//
// The IP range is immutable: DigitalOcean assigns one when ip_range is unset
// (the assigned range is reported through the ip_range output), and a change
// to a set range REPLACES the VPC.
//
// Destroy caveat the module cannot fix: in a region with no VPC yet, the
// first VPC created becomes the region's DEFAULT, and the API refuses to
// delete default VPCs (403). The `default` attribute is computed and cannot
// be steered from here; the kind's GUIDE carries the account-level recovery.
func vpc(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.Vpc, error) {
	spec := locals.DigitalOceanVpc.Spec

	// The VPC's name is its Planton identity -- it comes from metadata.name,
	// never from separate spec surface.
	vpcArgs := &digitalocean.VpcArgs{
		Name:   pulumi.String(locals.DigitalOceanVpc.Metadata.Name),
		Region: pulumi.String(spec.Region.String()),
	}

	if spec.Description != "" {
		vpcArgs.Description = pulumi.String(spec.Description)
	}

	if spec.IpRangeCidr != "" {
		vpcArgs.IpRange = pulumi.String(spec.IpRangeCidr)
	}

	createdVpc, err := digitalocean.NewVpc(
		ctx,
		"vpc",
		vpcArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean vpc")
	}

	// Stack outputs -- exactly the DigitalOceanVpcStackOutputs contract,
	// from the SDK's real field names (the urn output is VpcUrn).
	ctx.Export(OpVpcId, createdVpc.ID())
	ctx.Export(OpIpRange, createdVpc.IpRange)
	ctx.Export(OpUrn, createdVpc.VpcUrn)

	return createdVpc, nil
}
