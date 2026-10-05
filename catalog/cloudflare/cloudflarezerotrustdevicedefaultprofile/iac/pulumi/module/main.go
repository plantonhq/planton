package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustdevicedefaultprofilev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustdevicedefaultprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezerotrustdevicedefaultprofilev1alpha1.CloudflareZeroTrustDeviceDefaultProfileIacInput,
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

	// 3. Apply the default device profile (and its folded companions).
	if err := defaultProfile(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to apply default device profile")
	}

	return nil
}
