package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustorganizationv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustorganization/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezerotrustorganizationv1alpha1.CloudflareZeroTrustOrganizationIacInput,
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

	// 3. Apply the organization configuration (and the folded service-key
	// rotation cadence when declared).
	if err := organization(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to apply zero trust organization")
	}

	return nil
}
