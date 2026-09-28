package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applyBranding manages the branding of the EXISTING tenant the provider's
// credential belongs to -- the logo, favicon, colors and font every Universal
// Login page shares, and the page template the login box renders inside -- and
// declares nothing when the spec manages none of those settings. Each setting
// is sent only when the spec sets it. Auth0 has no delete for the tenant's
// branding: the resource's delete removes the page template (on a tenant with a
// custom domain) and leaves the last-applied logo, favicon, colors and font in
// place. The provider checks for a custom domain on every read, update and
// delete, and refuses a page template on a tenant without one. The Terraform
// module's auth0_branding (iac/tf/main.tf) is its twin.
func applyBranding(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.Branding, error) {
	if !locals.ManageBranding {
		return nil, nil
	}

	args := &auth0.BrandingArgs{
		LogoUrl:    pulumi.StringPtrFromPtr(locals.LogoUrl),
		FaviconUrl: pulumi.StringPtrFromPtr(locals.FaviconUrl),
	}
	if locals.Colors != nil {
		args.Colors = &auth0.BrandingColorsArgs{
			Primary:        pulumi.StringPtrFromPtr(locals.Colors.Primary),
			PageBackground: pulumi.StringPtrFromPtr(locals.Colors.PageBackground),
		}
	}
	// Removing font_url after it was applied returns the pages to Auth0's font.
	if locals.FontUrl != nil {
		args.Font = &auth0.BrandingFontArgs{
			Url: pulumi.String(*locals.FontUrl),
		}
	}
	// Removing the template returns the pages to Auth0's default page.
	if locals.UniversalLoginTemplate != nil {
		args.UniversalLogin = &auth0.BrandingUniversalLoginArgs{
			Body: pulumi.String(*locals.UniversalLoginTemplate),
		}
	}

	branding, err := auth0.NewBranding(ctx, locals.ResourceName, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to apply the branding of the Auth0 tenant for %s", locals.ResourceName)
	}
	return branding, nil
}
