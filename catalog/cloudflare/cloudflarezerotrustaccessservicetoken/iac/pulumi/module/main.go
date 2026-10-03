package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustaccessservicetokenv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustaccessservicetoken/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezerotrustaccessservicetokenv1alpha1.CloudflareZeroTrustAccessServiceTokenIacInput,
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

	// 3. Create the Access service token.
	if err := serviceToken(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create access service token")
	}

	return nil
}
