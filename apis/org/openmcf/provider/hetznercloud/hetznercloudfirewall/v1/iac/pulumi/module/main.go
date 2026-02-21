package module

import (
	"github.com/pkg/errors"
	hetznercloudfirewallv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud/hetznercloudfirewall/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/pulumihcloudprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	stackInput *hetznercloudfirewallv1.HetznerCloudFirewallStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	hcloudProvider, err := pulumihcloudprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup hetzner cloud provider")
	}

	if err := firewall(ctx, locals, hcloudProvider); err != nil {
		return errors.Wrap(err, "failed to create firewall")
	}

	return nil
}
