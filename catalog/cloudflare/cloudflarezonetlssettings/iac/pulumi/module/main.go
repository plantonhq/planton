package module

import (
	"github.com/pkg/errors"
	cloudflarezonetlssettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezonetlssettings/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezonetlssettingsv1alpha1.CloudflareZoneTlsSettingsIacInput,
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

	// 3. Emit the managed TLS settings. Unset fields are never sent; the
	// singleton toggles (Universal SSL, Total TLS, auto origin key exchange, CA
	// associations) have NO delete at Cloudflare -- destroy abandons the
	// last-applied values (see the spec's contract).
	if err := tlsSettings(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create zone tls settings")
	}

	// 4. Export the zone id -- the singleton's identity.
	ctx.Export(OpZoneId, pulumi.String(locals.CloudflareZoneTlsSettings.Spec.ZoneId.GetValue()))

	return nil
}
