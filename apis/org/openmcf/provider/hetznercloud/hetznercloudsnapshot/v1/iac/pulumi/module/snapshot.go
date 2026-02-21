package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func snapshot(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	spec := locals.HetznerCloudSnapshot.Spec

	serverIdInt, err := strconv.Atoi(spec.ServerId.GetValue())
	if err != nil {
		return errors.Wrapf(err, "failed to parse server_id %q as integer",
			spec.ServerId.GetValue())
	}

	snapshotArgs := &hcloud.SnapshotArgs{
		ServerId: pulumi.Int(serverIdInt),
		Labels:   pulumi.ToStringMap(locals.Labels),
	}

	if spec.Description != "" {
		snapshotArgs.Description = pulumi.StringPtr(spec.Description)
	}

	createdSnapshot, err := hcloud.NewSnapshot(
		ctx,
		"snapshot",
		snapshotArgs,
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud snapshot")
	}

	ctx.Export(OpSnapshotId, createdSnapshot.ID())

	return nil
}
