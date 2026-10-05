package module

import (
	"github.com/pkg/errors"
	cloudflarecachesettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarecachesettings/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarecachesettingsv1alpha1.CloudflareCacheSettingsIacInput,
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

	// 3. Emit the managed cache settings. Unset fields are never sent -- and
	// most of these have NO delete at Cloudflare, so anything sent once is
	// owned until reverted explicitly (see the spec's contract; Argo Smart
	// Routing even keeps billing after destroy).
	if err := cacheSettings(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create cache settings")
	}

	// 4. Export the zone id -- the singleton's identity.
	ctx.Export(OpZoneId, pulumi.String(locals.CloudflareCacheSettings.Spec.ZoneId.GetValue()))

	return nil
}
