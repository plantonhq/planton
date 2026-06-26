package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// fallbackOrigin sets the default origin all of a SaaS zone's custom hostnames route
// to (one per zone) and exports its status.
func fallbackOrigin(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareCustomHostnameFallbackOrigin.Spec

	created, err := cloudflare.NewCustomHostnameFallbackOrigin(
		ctx,
		"custom-hostname-fallback-origin",
		&cloudflare.CustomHostnameFallbackOriginArgs{
			ZoneId: pulumi.String(spec.ZoneId.GetValue()),
			Origin: pulumi.String(spec.Origin.GetValue()),
		},
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare custom hostname fallback origin")
	}

	ctx.Export(OpStatus, created.Status)
	ctx.Export(OpCreatedAt, created.CreatedAt)
	ctx.Export(OpUpdatedAt, created.UpdatedAt)
	ctx.Export(OpErrors, created.Errors)

	return nil
}
