package module

import (
	auth0brandingv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0branding/v1alpha1"
)

// Locals holds the values the module computes from the stack input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	// The tenant's branding settings. A nil pointer is NOT MANAGED: the provider
	// never sends it, and the tenant keeps whatever it already carries. Empty
	// strings are the proto's zero value for "unset", so they map to nil.
	LogoUrl                *string
	FaviconUrl             *string
	FontUrl                *string
	UniversalLoginTemplate *string

	// Colors is nil when the spec declares no colors; each color in it is sent
	// only when set.
	Colors *brandingColors

	// ManageBranding is true when the spec manages one of the branding's
	// settings; otherwise the branding resource is not declared, and a spec that
	// declares only a theme leaves the branding untouched.
	ManageBranding bool

	// Theme is the theme to send whole, every unset field at Auth0's default, or
	// nil when the spec declares no theme (the theme resource is then not
	// declared).
	Theme *theme
}

// brandingColors are the colors every Universal Login page shares.
type brandingColors struct {
	Primary        *string
	PageBackground *string
}

// theme is a declared theme as sent: every field resolved.
type theme struct {
	// DisplayName is nil when the spec leaves it unset (never sent).
	DisplayName    *string
	Borders        themeBorders
	Colors         themeColors
	Fonts          themeFonts
	PageBackground themePageBackground
	Widget         themeWidget

	// Identifiers is nil when the spec leaves it out (never sent).
	Identifiers *themeIdentifiers
}

type themeBorders struct {
	ButtonsStyle       string
	ButtonBorderRadius float64
	ButtonBorderWeight float64
	InputsStyle        string
	InputBorderRadius  float64
	InputBorderWeight  float64
	ShowWidgetShadow   bool
	WidgetCornerRadius float64
	WidgetBorderWeight float64
}

type themeColors struct {
	BaseFocusColor          string
	BaseHoverColor          string
	BodyText                string
	CaptchaWidgetTheme      string
	Error                   string
	Header                  string
	Icons                   string
	InputBackground         string
	InputBorder             string
	InputFilledText         string
	InputLabelsPlaceholders string
	LinksFocusedComponents  string
	PrimaryButton           string
	PrimaryButtonLabel      string
	SecondaryButtonBorder   string
	SecondaryButtonLabel    string
	Success                 string
	WidgetBackground        string
	WidgetBorder            string
}

type themeFonts struct {
	FontUrl           string
	LinksStyle        string
	ReferenceTextSize float64
	BodyText          themeTextStyle
	ButtonsText       themeTextStyle
	InputLabels       themeTextStyle
	Links             themeTextStyle
	Subtitle          themeTextStyle
	Title             themeTextStyle
}

// themeTextStyle is one text style: whether it is bold, and its size as a
// percentage of the reference text size.
type themeTextStyle struct {
	Bold bool
	Size float64
}

type themePageBackground struct {
	BackgroundColor    string
	BackgroundImageUrl string
	PageLayout         string
}

type themeWidget struct {
	HeaderTextAlignment string
	LogoHeight          float64
	LogoPosition        string
	LogoUrl             string
	SocialButtonsLayout string
}

type themeIdentifiers struct {
	LoginDisplay           string
	OtpAutocomplete        bool
	PhoneDisplayFormatting string
	PhoneDisplayMasking    string
}

// themeDefaults is the theme Auth0 applies to every field left unset: the
// Default of each argument in the auth0/auth0 provider's
// internal/auth0/branding/resource_theme.go -- the one place this module states
// them. The Terraform module's theme_defaults (iac/tf/locals.tf) carries the
// same values -- keep them in lockstep.
var themeDefaults = theme{
	Borders: themeBorders{
		ButtonsStyle:       "rounded",
		ButtonBorderRadius: 3,
		ButtonBorderWeight: 1,
		InputsStyle:        "rounded",
		InputBorderRadius:  3,
		InputBorderWeight:  1,
		ShowWidgetShadow:   true,
		WidgetCornerRadius: 5,
		WidgetBorderWeight: 0,
	},
	Colors: themeColors{
		BaseFocusColor:          "#635dff",
		BaseHoverColor:          "#000000",
		BodyText:                "#1e212a",
		CaptchaWidgetTheme:      "auto",
		Error:                   "#d03c38",
		Header:                  "#1e212a",
		Icons:                   "#65676e",
		InputBackground:         "#ffffff",
		InputBorder:             "#c9cace",
		InputFilledText:         "#000000",
		InputLabelsPlaceholders: "#65676e",
		LinksFocusedComponents:  "#635dff",
		PrimaryButton:           "#635dff",
		PrimaryButtonLabel:      "#ffffff",
		SecondaryButtonBorder:   "#c9cace",
		SecondaryButtonLabel:    "#1e212a",
		Success:                 "#13a688",
		WidgetBackground:        "#ffffff",
		WidgetBorder:            "#c9cace",
	},
	Fonts: themeFonts{
		FontUrl:           "",
		LinksStyle:        "normal",
		ReferenceTextSize: 16,
		// Each text style's own default.
		BodyText:    themeTextStyle{Bold: false, Size: 87.5},
		ButtonsText: themeTextStyle{Bold: false, Size: 100},
		InputLabels: themeTextStyle{Bold: false, Size: 100},
		Links:       themeTextStyle{Bold: true, Size: 87.5},
		Subtitle:    themeTextStyle{Bold: false, Size: 87.5},
		Title:       themeTextStyle{Bold: false, Size: 150},
	},
	PageBackground: themePageBackground{
		BackgroundColor:    "#000000",
		BackgroundImageUrl: "",
		PageLayout:         "center",
	},
	Widget: themeWidget{
		HeaderTextAlignment: "center",
		LogoHeight:          52,
		LogoPosition:        "center",
		LogoUrl:             "",
		SocialButtonsLayout: "bottom",
	},
}

func initializeLocals(stackInput *auth0brandingv1alpha1.Auth0BrandingStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	locals := &Locals{
		ResourceName:           target.Metadata.Name,
		LogoUrl:                managed(spec.LogoUrl),
		FaviconUrl:             managed(spec.FaviconUrl),
		FontUrl:                managed(spec.FontUrl),
		UniversalLoginTemplate: managed(spec.UniversalLoginTemplate),
		Theme:                  resolveTheme(spec.Theme),
	}
	if spec.Colors != nil {
		locals.Colors = &brandingColors{
			Primary:        managed(spec.Colors.Primary),
			PageBackground: managed(spec.Colors.PageBackground),
		}
	}
	locals.ManageBranding = locals.LogoUrl != nil ||
		locals.FaviconUrl != nil ||
		locals.Colors != nil ||
		locals.FontUrl != nil ||
		locals.UniversalLoginTemplate != nil
	return locals
}

// resolveTheme overlays the fields a declared theme sets on themeDefaults, so
// the theme is sent whole; a block the spec leaves out is its defaults alone. It
// returns nil when the spec declares no theme.
func resolveTheme(spec *auth0brandingv1alpha1.Auth0BrandingTheme) *theme {
	if spec == nil {
		return nil
	}
	d := themeDefaults

	borders := spec.GetBorders()
	if borders == nil {
		borders = &auth0brandingv1alpha1.Auth0BrandingThemeBorders{}
	}
	colors := spec.GetColors()
	if colors == nil {
		colors = &auth0brandingv1alpha1.Auth0BrandingThemeColors{}
	}
	fonts := spec.GetFonts()
	if fonts == nil {
		fonts = &auth0brandingv1alpha1.Auth0BrandingThemeFonts{}
	}
	pageBackground := spec.GetPageBackground()
	if pageBackground == nil {
		pageBackground = &auth0brandingv1alpha1.Auth0BrandingThemePageBackground{}
	}
	widget := spec.GetWidget()
	if widget == nil {
		widget = &auth0brandingv1alpha1.Auth0BrandingThemeWidget{}
	}

	resolved := &theme{
		DisplayName: managed(spec.DisplayName),
		Borders: themeBorders{
			ButtonsStyle:       orDefault(borders.ButtonsStyle, d.Borders.ButtonsStyle),
			ButtonBorderRadius: orDefault(borders.ButtonBorderRadius, d.Borders.ButtonBorderRadius),
			ButtonBorderWeight: orDefault(borders.ButtonBorderWeight, d.Borders.ButtonBorderWeight),
			InputsStyle:        orDefault(borders.InputsStyle, d.Borders.InputsStyle),
			InputBorderRadius:  orDefault(borders.InputBorderRadius, d.Borders.InputBorderRadius),
			InputBorderWeight:  orDefault(borders.InputBorderWeight, d.Borders.InputBorderWeight),
			ShowWidgetShadow:   orDefault(borders.ShowWidgetShadow, d.Borders.ShowWidgetShadow),
			WidgetCornerRadius: orDefault(borders.WidgetCornerRadius, d.Borders.WidgetCornerRadius),
			WidgetBorderWeight: orDefault(borders.WidgetBorderWeight, d.Borders.WidgetBorderWeight),
		},
		Colors: themeColors{
			BaseFocusColor:          orDefault(colors.BaseFocusColor, d.Colors.BaseFocusColor),
			BaseHoverColor:          orDefault(colors.BaseHoverColor, d.Colors.BaseHoverColor),
			BodyText:                orDefault(colors.BodyText, d.Colors.BodyText),
			CaptchaWidgetTheme:      orDefault(colors.CaptchaWidgetTheme, d.Colors.CaptchaWidgetTheme),
			Error:                   orDefault(colors.Error, d.Colors.Error),
			Header:                  orDefault(colors.Header, d.Colors.Header),
			Icons:                   orDefault(colors.Icons, d.Colors.Icons),
			InputBackground:         orDefault(colors.InputBackground, d.Colors.InputBackground),
			InputBorder:             orDefault(colors.InputBorder, d.Colors.InputBorder),
			InputFilledText:         orDefault(colors.InputFilledText, d.Colors.InputFilledText),
			InputLabelsPlaceholders: orDefault(colors.InputLabelsPlaceholders, d.Colors.InputLabelsPlaceholders),
			LinksFocusedComponents:  orDefault(colors.LinksFocusedComponents, d.Colors.LinksFocusedComponents),
			PrimaryButton:           orDefault(colors.PrimaryButton, d.Colors.PrimaryButton),
			PrimaryButtonLabel:      orDefault(colors.PrimaryButtonLabel, d.Colors.PrimaryButtonLabel),
			SecondaryButtonBorder:   orDefault(colors.SecondaryButtonBorder, d.Colors.SecondaryButtonBorder),
			SecondaryButtonLabel:    orDefault(colors.SecondaryButtonLabel, d.Colors.SecondaryButtonLabel),
			Success:                 orDefault(colors.Success, d.Colors.Success),
			WidgetBackground:        orDefault(colors.WidgetBackground, d.Colors.WidgetBackground),
			WidgetBorder:            orDefault(colors.WidgetBorder, d.Colors.WidgetBorder),
		},
		Fonts: themeFonts{
			FontUrl:           fonts.FontUrl,
			LinksStyle:        orDefault(fonts.LinksStyle, d.Fonts.LinksStyle),
			ReferenceTextSize: orDefault(fonts.ReferenceTextSize, d.Fonts.ReferenceTextSize),
			BodyText:          resolveTextStyle(fonts.BodyText, d.Fonts.BodyText),
			ButtonsText:       resolveTextStyle(fonts.ButtonsText, d.Fonts.ButtonsText),
			InputLabels:       resolveTextStyle(fonts.InputLabels, d.Fonts.InputLabels),
			Links:             resolveTextStyle(fonts.Links, d.Fonts.Links),
			Subtitle:          resolveTextStyle(fonts.Subtitle, d.Fonts.Subtitle),
			Title:             resolveTextStyle(fonts.Title, d.Fonts.Title),
		},
		PageBackground: themePageBackground{
			BackgroundColor:    orDefault(pageBackground.BackgroundColor, d.PageBackground.BackgroundColor),
			BackgroundImageUrl: pageBackground.BackgroundImageUrl,
			PageLayout:         orDefault(pageBackground.PageLayout, d.PageBackground.PageLayout),
		},
		Widget: themeWidget{
			HeaderTextAlignment: orDefault(widget.HeaderTextAlignment, d.Widget.HeaderTextAlignment),
			LogoHeight:          orDefault(widget.LogoHeight, d.Widget.LogoHeight),
			LogoPosition:        orDefault(widget.LogoPosition, d.Widget.LogoPosition),
			LogoUrl:             widget.LogoUrl,
			SocialButtonsLayout: orDefault(widget.SocialButtonsLayout, d.Widget.SocialButtonsLayout),
		},
	}

	// identifiers is the one theme block sent only when declared: Auth0 offers
	// it only on tenants with the feature enabled.
	if identifiers := spec.GetIdentifiers(); identifiers != nil {
		resolved.Identifiers = &themeIdentifiers{
			LoginDisplay:           identifiers.LoginDisplay,
			OtpAutocomplete:        identifiers.OtpAutocomplete,
			PhoneDisplayFormatting: identifiers.GetPhoneDisplay().GetFormatting(),
			PhoneDisplayMasking:    identifiers.GetPhoneDisplay().GetMasking(),
		}
	}
	return resolved
}

// resolveTextStyle overlays the fields a text style sets on that style's own
// default; a style the spec leaves out is its default alone.
func resolveTextStyle(spec *auth0brandingv1alpha1.Auth0BrandingThemeTextStyle, fallback themeTextStyle) themeTextStyle {
	if spec == nil {
		return fallback
	}
	return themeTextStyle{
		Bold: orDefault(spec.Bold, fallback.Bold),
		Size: orDefault(spec.Size, fallback.Size),
	}
}

// orDefault is a proto3 optional field's value when the spec sets it (an
// explicit zero included), and Auth0's default otherwise.
func orDefault[T any](value *T, fallback T) T {
	if value == nil {
		return fallback
	}
	return *value
}

// managed maps the proto's "unset" (the empty string) to nil, so the setting is
// never sent.
func managed(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
