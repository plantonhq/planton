package module

import (
	"github.com/pkg/errors"
	hetznercloudnetworkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud/hetznercloudnetwork/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/pulumihcloudprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(
	ctx *pulumi.Context,
	stackInput *hetznercloudnetworkv1.HetznerCloudNetworkStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	hcloudProvider, err := pulumihcloudprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup hetzner cloud provider")
	}

	if err := network(ctx, locals, hcloudProvider); err != nil {
		return errors.Wrap(err, "failed to create network")
	}

	return nil
}
