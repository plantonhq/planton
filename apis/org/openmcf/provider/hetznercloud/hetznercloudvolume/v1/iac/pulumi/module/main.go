package module

import (
	"github.com/pkg/errors"
	hetznercloudvolumev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud/hetznercloudvolume/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/pulumihcloudprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(
	ctx *pulumi.Context,
	stackInput *hetznercloudvolumev1.HetznerCloudVolumeStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	hcloudProvider, err := pulumihcloudprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup hetzner cloud provider")
	}

	if err := volume(ctx, locals, hcloudProvider); err != nil {
		return errors.Wrap(err, "failed to create volume")
	}

	return nil
}
