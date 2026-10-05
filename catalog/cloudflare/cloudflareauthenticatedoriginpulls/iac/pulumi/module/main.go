package module

import (
	"github.com/pkg/errors"
	cloudflareauthenticatedoriginpullsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareauthenticatedoriginpulls/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflareauthenticatedoriginpullsv1alpha1.CloudflareAuthenticatedOriginPullsIacInput,
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

	// 3. Manage the zone's Authenticated Origin Pulls surface.
	if err := authenticatedOriginPulls(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to manage authenticated origin pulls")
	}

	return nil
}
