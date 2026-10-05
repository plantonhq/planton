package module

import (
	"github.com/pkg/errors"
	cloudflarecustomhostnamefallbackoriginv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarecustomhostnamefallbackorigin/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—kept small to mirror a Terraform module's main.tf.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarecustomhostnamefallbackoriginv1alpha1.CloudflareCustomHostnameFallbackOriginIacInput,
) error {
	locals := initializeLocals(ctx, iacInput)

	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	if err := fallbackOrigin(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cloudflare custom hostname fallback origin")
	}

	return nil
}
