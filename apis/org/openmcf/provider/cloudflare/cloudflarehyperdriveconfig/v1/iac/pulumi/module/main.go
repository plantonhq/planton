package module

import (
	"github.com/pkg/errors"
	cloudflarehyperdriveconfigv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarehyperdriveconfig/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module's main.tf.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflarehyperdriveconfigv1.CloudflareHyperdriveConfigStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if _, err := hyperdriveConfig(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare hyperdrive config")
	}

	return nil
}
