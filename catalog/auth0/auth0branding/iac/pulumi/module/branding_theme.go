package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applyTheme applies the no-code look of the login box when the spec declares a
// theme, and declares nothing otherwise. The theme is sent whole: every block
// the provider requires is always sent, each field the spec leaves unset at
// Auth0's default (resolved in locals.go); identifiers is sent only when
// declared. A tenant has one theme: the provider adopts the tenant's existing
// theme on create, and destroy deletes it, returning the login box to Auth0's
// look. The Terraform module's auth0_branding_theme (iac/tf/main.tf) is its
// twin.
func applyTheme(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.BrandingTheme, error) {
	t := locals.Theme
	if t == nil {
		return nil, nil
	}

	args := &auth0.BrandingThemeArgs{
		DisplayName: pulumi.StringPtrFromPtr(t.DisplayName),
		Borders: auth0.BrandingThemeBordersArgs{
			ButtonsStyle:       pulumi.String(t.Borders.ButtonsStyle),
			ButtonBorderRadius: pulumi.Float64(t.Borders.ButtonBorderRadius),
			ButtonBorderWeight: pulumi.Float64(t.Borders.ButtonBorderWeight),
			InputsStyle:        pulumi.String(t.Borders.InputsStyle),
			InputBorderRadius:  pulumi.Float64(t.Borders.InputBorderRadius),
			InputBorderWeight:  pulumi.Float64(t.Borders.InputBorderWeight),
			ShowWidgetShadow:   pulumi.Bool(t.Borders.ShowWidgetShadow),
			WidgetCornerRadius: pulumi.Float64(t.Borders.WidgetCornerRadius),
			WidgetBorderWeight: pulumi.Float64(t.Borders.WidgetBorderWeight),
		},
		Colors: auth0.BrandingThemeColorsArgs{
			BaseFocusColor:          pulumi.String(t.Colors.BaseFocusColor),
			BaseHoverColor:          pulumi.String(t.Colors.BaseHoverColor),
			BodyText:                pulumi.String(t.Colors.BodyText),
			CaptchaWidgetTheme:      pulumi.String(t.Colors.CaptchaWidgetTheme),
			Error:                   pulumi.String(t.Colors.Error),
			Header:                  pulumi.String(t.Colors.Header),
			Icons:                   pulumi.String(t.Colors.Icons),
			InputBackground:         pulumi.String(t.Colors.InputBackground),
			InputBorder:             pulumi.String(t.Colors.InputBorder),
			InputFilledText:         pulumi.String(t.Colors.InputFilledText),
			InputLabelsPlaceholders: pulumi.String(t.Colors.InputLabelsPlaceholders),
			LinksFocusedComponents:  pulumi.String(t.Colors.LinksFocusedComponents),
			PrimaryButton:           pulumi.String(t.Colors.PrimaryButton),
			PrimaryButtonLabel:      pulumi.String(t.Colors.PrimaryButtonLabel),
			SecondaryButtonBorder:   pulumi.String(t.Colors.SecondaryButtonBorder),
			SecondaryButtonLabel:    pulumi.String(t.Colors.SecondaryButtonLabel),
			Success:                 pulumi.String(t.Colors.Success),
			WidgetBackground:        pulumi.String(t.Colors.WidgetBackground),
			WidgetBorder:            pulumi.String(t.Colors.WidgetBorder),
		},
		Fonts: auth0.BrandingThemeFontsArgs{
			FontUrl:           pulumi.String(t.Fonts.FontUrl),
			LinksStyle:        pulumi.String(t.Fonts.LinksStyle),
			ReferenceTextSize: pulumi.Float64(t.Fonts.ReferenceTextSize),
			BodyText: auth0.BrandingThemeFontsBodyTextArgs{
				Bold: pulumi.Bool(t.Fonts.BodyText.Bold),
				Size: pulumi.Float64(t.Fonts.BodyText.Size),
			},
			ButtonsText: auth0.BrandingThemeFontsButtonsTextArgs{
				Bold: pulumi.Bool(t.Fonts.ButtonsText.Bold),
				Size: pulumi.Float64(t.Fonts.ButtonsText.Size),
			},
			InputLabels: auth0.BrandingThemeFontsInputLabelsArgs{
				Bold: pulumi.Bool(t.Fonts.InputLabels.Bold),
				Size: pulumi.Float64(t.Fonts.InputLabels.Size),
			},
			Links: auth0.BrandingThemeFontsLinksArgs{
				Bold: pulumi.Bool(t.Fonts.Links.Bold),
				Size: pulumi.Float64(t.Fonts.Links.Size),
			},
			Subtitle: auth0.BrandingThemeFontsSubtitleArgs{
				Bold: pulumi.Bool(t.Fonts.Subtitle.Bold),
				Size: pulumi.Float64(t.Fonts.Subtitle.Size),
			},
			Title: auth0.BrandingThemeFontsTitleArgs{
				Bold: pulumi.Bool(t.Fonts.Title.Bold),
				Size: pulumi.Float64(t.Fonts.Title.Size),
			},
		},
		PageBackground: auth0.BrandingThemePageBackgroundArgs{
			BackgroundColor:    pulumi.String(t.PageBackground.BackgroundColor),
			BackgroundImageUrl: pulumi.String(t.PageBackground.BackgroundImageUrl),
			PageLayout:         pulumi.String(t.PageBackground.PageLayout),
		},
		Widget: auth0.BrandingThemeWidgetArgs{
			HeaderTextAlignment: pulumi.String(t.Widget.HeaderTextAlignment),
			LogoHeight:          pulumi.Float64(t.Widget.LogoHeight),
			LogoPosition:        pulumi.String(t.Widget.LogoPosition),
			LogoUrl:             pulumi.String(t.Widget.LogoUrl),
			SocialButtonsLayout: pulumi.String(t.Widget.SocialButtonsLayout),
		},
	}

	// Sent only when declared. Once applied, Auth0 keeps it: removing the block
	// leaves the last-applied values in place.
	if t.Identifiers != nil {
		args.Identifiers = &auth0.BrandingThemeIdentifiersArgs{
			LoginDisplay:    pulumi.String(t.Identifiers.LoginDisplay),
			OtpAutocomplete: pulumi.Bool(t.Identifiers.OtpAutocomplete),
			PhoneDisplay: auth0.BrandingThemeIdentifiersPhoneDisplayArgs{
				Formatting: pulumi.String(t.Identifiers.PhoneDisplayFormatting),
				Masking:    pulumi.String(t.Identifiers.PhoneDisplayMasking),
			},
		}
	}

	brandingTheme, err := auth0.NewBrandingTheme(ctx, locals.ResourceName+"-theme", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to apply the branding theme of the Auth0 tenant for %s", locals.ResourceName)
	}
	return brandingTheme, nil
}
