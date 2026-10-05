package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustlistv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustlist/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezerotrustlistv1alpha1.CloudflareZeroTrustListIacInput,
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

	// 3. Create the Zero Trust list.
	if err := zeroTrustList(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create zero trust list")
	}

	return nil
}
