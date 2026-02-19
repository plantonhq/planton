package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func placementGroup(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	createdPlacementGroup, err := hcloud.NewPlacementGroup(
		ctx,
		"placement-group",
		&hcloud.PlacementGroupArgs{
			Name:   pulumi.String(locals.HetznerCloudPlacementGroup.Metadata.Name),
			Type:   pulumi.String(locals.HetznerCloudPlacementGroup.Spec.GetType().String()),
			Labels: pulumi.ToStringMap(locals.Labels),
		},
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud placement group")
	}

	ctx.Export(OpPlacementGroupId, createdPlacementGroup.ID())

	return nil
}
