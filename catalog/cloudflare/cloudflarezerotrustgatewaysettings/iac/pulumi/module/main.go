package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustgatewaysettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustgatewaysettings/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezerotrustgatewaysettingsv1alpha1.CloudflareZeroTrustGatewaySettingsIacInput,
) error {
	// 1. Prepare locals (metadata, credentials).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Apply the Gateway configuration surfaces (settings singleton,
	// logging singleton, PAC-file collection).
	if err := gatewaySettings(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to apply gateway settings")
	}

	return nil
}
