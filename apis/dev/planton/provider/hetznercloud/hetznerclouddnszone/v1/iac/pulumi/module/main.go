package module

import (
	"github.com/pkg/errors"
	hetznerclouddnszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznerclouddnszone/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/pulumihcloudprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(
	ctx *pulumi.Context,
	stackInput *hetznerclouddnszonev1.HetznerCloudDnsZoneStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	hcloudProvider, err := pulumihcloudprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup hetzner cloud provider")
	}

	if err := zone(ctx, locals, hcloudProvider); err != nil {
		return errors.Wrap(err, "failed to create dns zone")
	}

	return nil
}
