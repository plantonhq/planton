package module

import (
	"reflect"
	"testing"

	auth0brandingv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0branding/v1alpha1"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func iacInput(spec *auth0brandingv1alpha1.Auth0BrandingSpec) *auth0brandingv1alpha1.Auth0BrandingIacInput {
	return &auth0brandingv1alpha1.Auth0BrandingIacInput{
		Target: &auth0brandingv1alpha1.Auth0Branding{
			Metadata: &shared.CatalogObjectMetadata{Name: "branding"},
			Spec:     spec,
		},
	}
}

func TestResolveTheme(t *testing.T) {
	t.Run("no theme declares no theme resource", func(t *testing.T) {
		if got := resolveTheme(nil); got != nil {
			t.Errorf("got %+v, want nil", got)
		}
	})

	t.Run("an empty theme is sent whole at Auth0's defaults, without identifiers", func(t *testing.T) {
		got := resolveTheme(&auth0brandingv1alpha1.Auth0BrandingTheme{})
		want := themeDefaults
		if !reflect.DeepEqual(*got, want) {
			t.Errorf("got %+v, want %+v", *got, want)
		}
	})

	t.Run("each text style falls back to its own default", func(t *testing.T) {
		got := resolveTheme(&auth0brandingv1alpha1.Auth0BrandingTheme{
			Fonts: &auth0brandingv1alpha1.Auth0BrandingThemeFonts{
				Title: &auth0brandingv1alpha1.Auth0BrandingThemeTextStyle{Bold: proto.Bool(true)},
				Links: &auth0brandingv1alpha1.Auth0BrandingThemeTextStyle{Size: proto.Float64(100)},
			},
		}).Fonts
		want := themeFonts{
			FontUrl:           "",
			LinksStyle:        "normal",
			ReferenceTextSize: 16,
			BodyText:          themeTextStyle{Bold: false, Size: 87.5},
			ButtonsText:       themeTextStyle{Bold: false, Size: 100},
			InputLabels:       themeTextStyle{Bold: false, Size: 100},
			Links:             themeTextStyle{Bold: true, Size: 100},
			Subtitle:          themeTextStyle{Bold: false, Size: 87.5},
			Title:             themeTextStyle{Bold: true, Size: 150},
		}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("a set field wins over the default, an explicit zero included", func(t *testing.T) {
		got := resolveTheme(&auth0brandingv1alpha1.Auth0BrandingTheme{
			DisplayName: "Planton",
			Borders: &auth0brandingv1alpha1.Auth0BrandingThemeBorders{
				ButtonsStyle:     proto.String("pill"),
				ShowWidgetShadow: proto.Bool(false),
			},
			Colors: &auth0brandingv1alpha1.Auth0BrandingThemeColors{
				PrimaryButton: proto.String("#000000"),
			},
			Widget: &auth0brandingv1alpha1.Auth0BrandingThemeWidget{
				LogoUrl: "https://assets.planton.ai/brand-kit/logo/p-mark-email.png",
			},
		})
		if got.DisplayName == nil || *got.DisplayName != "Planton" {
			t.Errorf("display name: got %v, want Planton", got.DisplayName)
		}
		wantBorders := themeDefaults.Borders
		wantBorders.ButtonsStyle = "pill"
		wantBorders.ShowWidgetShadow = false
		if got.Borders != wantBorders {
			t.Errorf("borders: got %+v, want %+v", got.Borders, wantBorders)
		}
		wantColors := themeDefaults.Colors
		wantColors.PrimaryButton = "#000000"
		if got.Colors != wantColors {
			t.Errorf("colors: got %+v, want %+v", got.Colors, wantColors)
		}
		wantWidget := themeDefaults.Widget
		wantWidget.LogoUrl = "https://assets.planton.ai/brand-kit/logo/p-mark-email.png"
		if got.Widget != wantWidget {
			t.Errorf("widget: got %+v, want %+v", got.Widget, wantWidget)
		}
		if got.PageBackground != themeDefaults.PageBackground {
			t.Errorf("page background: got %+v, want %+v", got.PageBackground, themeDefaults.PageBackground)
		}
	})

	t.Run("identifiers are sent when declared", func(t *testing.T) {
		got := resolveTheme(&auth0brandingv1alpha1.Auth0BrandingTheme{
			Identifiers: &auth0brandingv1alpha1.Auth0BrandingThemeIdentifiers{
				LoginDisplay: "unified",
				PhoneDisplay: &auth0brandingv1alpha1.Auth0BrandingThemePhoneDisplay{
					Formatting: "international",
					Masking:    "mask_digits",
				},
			},
		}).Identifiers
		want := &themeIdentifiers{
			LoginDisplay:           "unified",
			OtpAutocomplete:        false,
			PhoneDisplayFormatting: "international",
			PhoneDisplayMasking:    "mask_digits",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
}

func TestManageBranding(t *testing.T) {
	cases := []struct {
		name string
		spec *auth0brandingv1alpha1.Auth0BrandingSpec
		want bool
	}{
		{
			name: "a theme alone leaves the branding undeclared",
			spec: &auth0brandingv1alpha1.Auth0BrandingSpec{Theme: &auth0brandingv1alpha1.Auth0BrandingTheme{}},
			want: false,
		},
		{
			name: "a logo declares the branding",
			spec: &auth0brandingv1alpha1.Auth0BrandingSpec{LogoUrl: "https://assets.example.com/logo.png"},
			want: true,
		},
		{
			name: "declared colors declare the branding",
			spec: &auth0brandingv1alpha1.Auth0BrandingSpec{Colors: &auth0brandingv1alpha1.Auth0BrandingColors{PageBackground: "#000000"}},
			want: true,
		},
		{
			name: "a page template declares the branding",
			spec: &auth0brandingv1alpha1.Auth0BrandingSpec{UniversalLoginTemplate: "<html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := initializeLocals(iacInput(tc.spec)).ManageBranding; got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
