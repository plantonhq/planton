package module

import (
	"github.com/pkg/errors"
	cloudflarezonesettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezonesettings/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflarezonesettingsv1alpha1.CloudflareZoneSettingsIacInput,
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

	// 3. Emit one cloudflare_zone_setting per managed setting field. Unset
	// fields are never sent: whatever the zone already carries stays untouched
	// (Cloudflare has no delete for zone settings -- see the spec's contract).
	if err := zoneSettings(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create zone settings")
	}

	// 4. Emit the companion zone-configuration resources (managed transforms,
	// URL normalization, origin cloud regions, waiting-room crawler bypass).
	if err := companions(ctx, locals, cloudflareProvider); err != nil {
		return errors.Wrap(err, "failed to create companion zone configuration")
	}

	// 5. Export the zone id -- the singleton's identity.
	ctx.Export(OpZoneId, pulumi.String(locals.CloudflareZoneSettings.Spec.ZoneId.GetValue()))

	return nil
}
