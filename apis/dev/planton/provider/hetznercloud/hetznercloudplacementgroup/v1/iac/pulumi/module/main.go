package module

import (
	"github.com/pkg/errors"
	hetznercloudplacementgroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznercloudplacementgroup/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/pulumihcloudprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(
	ctx *pulumi.Context,
	stackInput *hetznercloudplacementgroupv1.HetznerCloudPlacementGroupStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	hcloudProvider, err := pulumihcloudprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup hetzner cloud provider")
	}

	if err := placementGroup(ctx, locals, hcloudProvider); err != nil {
		return errors.Wrap(err, "failed to create placement group")
	}

	return nil
}
