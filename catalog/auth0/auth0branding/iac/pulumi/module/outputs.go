package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the branding as applied: the theme's id when the spec
// declares a theme, and the logo the tenant's pages show when the spec manages
// a branding setting. Each is empty when its resource is not declared. The
// Terraform module's outputs.tf is its twin.
func exportOutputs(ctx *pulumi.Context, branding *auth0.Branding, theme *auth0.BrandingTheme) error {
	if theme != nil {
		ctx.Export("theme_id", theme.ID())
	} else {
		ctx.Export("theme_id", pulumi.String(""))
	}

	if branding != nil {
		ctx.Export("logo_url", branding.LogoUrl)
	} else {
		ctx.Export("logo_url", pulumi.String(""))
	}

	return nil
}
