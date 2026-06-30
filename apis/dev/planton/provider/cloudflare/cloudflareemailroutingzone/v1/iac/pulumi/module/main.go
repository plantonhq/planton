package module

import (
	"github.com/pkg/errors"
	cloudflareemailroutingzonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflareemailroutingzone/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module's main.tf.
func Resources(
	ctx *pulumi.Context,
	stackInput *cloudflareemailroutingzonev1.CloudflareEmailRoutingZoneStackInput,
) error {
	locals := initializeLocals(ctx, stackInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if err := emailRoutingZone(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to enable cloudflare email routing")
	}

	return nil
}
