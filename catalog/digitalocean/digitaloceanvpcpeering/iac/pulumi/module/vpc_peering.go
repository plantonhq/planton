package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// vpcPeering provisions the peering connection and exports its outputs.
// Only the name updates in place; changing either VPC replaces the
// peering. The provider waits for ACTIVE on create and retries the delete
// while DigitalOcean returns 403 during peering settlement.
func vpcPeering(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.VpcPeering, error) {
	spec := locals.DigitalOceanVpcPeering.Spec

	createdPeering, err := digitalocean.NewVpcPeering(
		ctx,
		"peering",
		&digitalocean.VpcPeeringArgs{
			Name: pulumi.String(spec.PeeringName),
			// References resolve to literal VPC UUIDs before the module
			// runs; the provider treats the pair as an unordered set.
			VpcIds: pulumi.StringArray{
				pulumi.String(spec.Vpc_1.GetValue()),
				pulumi.String(spec.Vpc_2.GetValue()),
			},
		},
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean vpc peering")
	}

	// The lifecycle status is deliberately not exported: the provider waits
	// for ACTIVE before this apply succeeds, so a stored status could only
	// ever read "ACTIVE" and would go stale the moment DigitalOcean moved
	// the peering. Live status belongs to whoever reads the API.
	ctx.Export(OpPeeringId, createdPeering.ID())

	return createdPeering, nil
}
