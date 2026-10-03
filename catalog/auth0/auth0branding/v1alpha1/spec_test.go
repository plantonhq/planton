package auth0brandingv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestAuth0Branding(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0Branding Suite")
}

func branding(spec *Auth0BrandingSpec) *Auth0Branding {
	return &Auth0Branding{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0Branding",
		Metadata:   &shared.CatalogObjectMetadata{Name: "branding"},
		Spec:       spec,
	}
}

// themed wraps a theme in an otherwise empty spec.
func themed(theme *Auth0BrandingTheme) *Auth0Branding {
	return branding(&Auth0BrandingSpec{Theme: theme})
}

const pageTemplate = `<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>`

func identifiers(loginDisplay, formatting, masking string) *Auth0BrandingThemeIdentifiers {
	return &Auth0BrandingThemeIdentifiers{
		LoginDisplay:    loginDisplay,
		OtpAutocomplete: true,
		PhoneDisplay: &Auth0BrandingThemePhoneDisplay{
			Formatting: formatting,
			Masking:    masking,
		},
	}
}

var _ = ginkgo.Describe("Auth0Branding Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a logo and colors", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{
				LogoUrl:    "https://assets.planton.ai/brand-kit/logo/p-mark-email.png",
				FaviconUrl: "https://assets.example.com/favicon.png",
				Colors: &Auth0BrandingColors{
					Primary:        "#0059d6",
					PageBackground: `{"type":"linear-gradient","start":"#000000","end":"#333333","angle_deg":35}`,
				},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a three-digit primary color", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{
				Colors: &Auth0BrandingColors{Primary: "#000"},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a font alone", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{FontUrl: "https://assets.example.com/brand.woff2"}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a page template carrying both tags", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{UniversalLoginTemplate: pageTemplate}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an empty theme alone, applied at Auth0's defaults", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a theme setting every block", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				DisplayName: "Planton",
				Borders: &Auth0BrandingThemeBorders{
					ButtonsStyle:       proto.String("pill"),
					ButtonBorderRadius: proto.Float64(10),
					ButtonBorderWeight: proto.Float64(0),
					InputsStyle:        proto.String("sharp"),
					InputBorderRadius:  proto.Float64(0),
					InputBorderWeight:  proto.Float64(3),
					ShowWidgetShadow:   proto.Bool(false),
					WidgetCornerRadius: proto.Float64(50),
					WidgetBorderWeight: proto.Float64(10),
				},
				Colors: &Auth0BrandingThemeColors{
					PrimaryButton:      proto.String("#000000"),
					CaptchaWidgetTheme: proto.String("dark"),
				},
				Fonts: &Auth0BrandingThemeFonts{
					FontUrl:           "https://assets.example.com/brand.woff2",
					LinksStyle:        proto.String("underlined"),
					ReferenceTextSize: proto.Float64(24),
					BodyText:          &Auth0BrandingThemeTextStyle{Size: proto.Float64(0)},
					Links:             &Auth0BrandingThemeTextStyle{Bold: proto.Bool(false), Size: proto.Float64(150)},
					Title:             &Auth0BrandingThemeTextStyle{Bold: proto.Bool(true), Size: proto.Float64(75)},
				},
				PageBackground: &Auth0BrandingThemePageBackground{
					BackgroundColor:    proto.String("#0b1120"),
					BackgroundImageUrl: "https://assets.example.com/background.jpg",
					PageLayout:         proto.String("left"),
				},
				Widget: &Auth0BrandingThemeWidget{
					HeaderTextAlignment: proto.String("right"),
					LogoHeight:          proto.Float64(100),
					LogoPosition:        proto.String("none"),
					LogoUrl:             "https://assets.example.com/logo.png",
					SocialButtonsLayout: proto.String("top"),
				},
				Identifiers: identifiers("separate", "regional", "hide_country_code"),
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts the title at its largest size", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Fonts: &Auth0BrandingThemeFonts{Title: &Auth0BrandingThemeTextStyle{Size: proto.Float64(150)}},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		for _, style := range []string{"pill", "rounded", "sharp"} {
			ginkgo.It("accepts the "+style+" button and input style", func() {
				err := protovalidate.Validate(themed(&Auth0BrandingTheme{
					Borders: &Auth0BrandingThemeBorders{ButtonsStyle: proto.String(style), InputsStyle: proto.String(style)},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		}

		for _, captcha := range []string{"auto", "dark", "light"} {
			ginkgo.It("accepts the "+captcha+" captcha theme", func() {
				err := protovalidate.Validate(themed(&Auth0BrandingTheme{
					Colors: &Auth0BrandingThemeColors{CaptchaWidgetTheme: proto.String(captcha)},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		}

		for _, linksStyle := range []string{"normal", "underlined"} {
			ginkgo.It("accepts the "+linksStyle+" links style", func() {
				err := protovalidate.Validate(themed(&Auth0BrandingTheme{
					Fonts: &Auth0BrandingThemeFonts{LinksStyle: proto.String(linksStyle)},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		}

		for _, placement := range []string{"center", "left", "right"} {
			ginkgo.It("accepts the "+placement+" page layout and header alignment", func() {
				err := protovalidate.Validate(themed(&Auth0BrandingTheme{
					PageBackground: &Auth0BrandingThemePageBackground{PageLayout: proto.String(placement)},
					Widget:         &Auth0BrandingThemeWidget{HeaderTextAlignment: proto.String(placement)},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		}

		for _, position := range []string{"center", "left", "right", "none"} {
			ginkgo.It("accepts the "+position+" logo position", func() {
				err := protovalidate.Validate(themed(&Auth0BrandingTheme{
					Widget: &Auth0BrandingThemeWidget{LogoPosition: proto.String(position)},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		}

		for _, layout := range []string{"bottom", "top"} {
			ginkgo.It("accepts the "+layout+" social buttons layout", func() {
				err := protovalidate.Validate(themed(&Auth0BrandingTheme{
					Widget: &Auth0BrandingThemeWidget{SocialButtonsLayout: proto.String(layout)},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		}

		for _, loginDisplay := range []string{"unified", "separate"} {
			for _, formatting := range []string{"international", "regional"} {
				for _, masking := range []string{"mask_digits", "hide_country_code", "show_all"} {
					ginkgo.It("accepts "+loginDisplay+" identifiers with "+formatting+" "+masking+" phone numbers", func() {
						err := protovalidate.Validate(themed(&Auth0BrandingTheme{
							Identifiers: identifiers(loginDisplay, formatting, masking),
						}))
						gomega.Expect(err).To(gomega.BeNil())
					})
				}
			}
		}
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a spec that manages nothing", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("configure at least one branding setting or a theme")))
		})

		ginkgo.It("refuses a logo that is not a URL", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{LogoUrl: "logo.png"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a favicon that is not a URL", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{FaviconUrl: "favicon.png"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a font that is not a URL", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{FontUrl: "brand.woff2"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a primary color that is not hex", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{
				Colors: &Auth0BrandingColors{Primary: "blue"},
			}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("colors.primary is a hex color")))
		})

		ginkgo.It("refuses a primary color without its hash", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{
				Colors: &Auth0BrandingColors{Primary: "0059d6"},
			}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("colors.primary is a hex color")))
		})

		ginkgo.It("refuses a page template without the head tag", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{
				UniversalLoginTemplate: `<!DOCTYPE html><html><head></head><body>{%- auth0:widget -%}</body></html>`,
			}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("universal_login_template must contain both")))
		})

		ginkgo.It("refuses a page template without the widget tag", func() {
			err := protovalidate.Validate(branding(&Auth0BrandingSpec{
				UniversalLoginTemplate: `<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body></body></html>`,
			}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("universal_login_template must contain both")))
		})

		ginkgo.It("refuses a title smaller than 75", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Fonts: &Auth0BrandingThemeFonts{Title: &Auth0BrandingThemeTextStyle{Size: proto.Float64(74)}},
			}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("fonts.title.size must be between 75 and 150")))
		})

		ginkgo.It("refuses a title larger than 150", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Fonts: &Auth0BrandingThemeFonts{Title: &Auth0BrandingThemeTextStyle{Size: proto.Float64(151)}},
			}))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("fonts.title.size must be between 75 and 150")))
		})

		ginkgo.It("refuses a text style larger than 150", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Fonts: &Auth0BrandingThemeFonts{BodyText: &Auth0BrandingThemeTextStyle{Size: proto.Float64(151)}},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown button style", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Borders: &Auth0BrandingThemeBorders{ButtonsStyle: proto.String("square")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown input style", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Borders: &Auth0BrandingThemeBorders{InputsStyle: proto.String("round")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a button radius below 1", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Borders: &Auth0BrandingThemeBorders{ButtonBorderRadius: proto.Float64(0)},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an input border weight above 3", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Borders: &Auth0BrandingThemeBorders{InputBorderWeight: proto.Float64(4)},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a widget corner radius above 50", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Borders: &Auth0BrandingThemeBorders{WidgetCornerRadius: proto.Float64(51)},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown captcha theme", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Colors: &Auth0BrandingThemeColors{CaptchaWidgetTheme: proto.String("system")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown links style", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Fonts: &Auth0BrandingThemeFonts{LinksStyle: proto.String("bold")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a reference text size below 12", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Fonts: &Auth0BrandingThemeFonts{ReferenceTextSize: proto.Float64(11)},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a theme font that is not a URL", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Fonts: &Auth0BrandingThemeFonts{FontUrl: "brand.woff2"},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown page layout", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				PageBackground: &Auth0BrandingThemePageBackground{PageLayout: proto.String("top")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a background image that is not a URL", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				PageBackground: &Auth0BrandingThemePageBackground{BackgroundImageUrl: "background.jpg"},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown header alignment", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Widget: &Auth0BrandingThemeWidget{HeaderTextAlignment: proto.String("justify")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown logo position", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Widget: &Auth0BrandingThemeWidget{LogoPosition: proto.String("top")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a logo height below 1", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Widget: &Auth0BrandingThemeWidget{LogoHeight: proto.Float64(0)},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a widget logo that is not a URL", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Widget: &Auth0BrandingThemeWidget{LogoUrl: "logo.png"},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown social buttons layout", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Widget: &Auth0BrandingThemeWidget{SocialButtonsLayout: proto.String("left")},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses identifiers without a login display", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Identifiers: identifiers("", "international", "mask_digits"),
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown login display", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Identifiers: identifiers("combined", "international", "mask_digits"),
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses identifiers without a phone display", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Identifiers: &Auth0BrandingThemeIdentifiers{LoginDisplay: "unified"},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown phone formatting", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Identifiers: identifiers("unified", "national", "mask_digits"),
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an unknown phone masking", func() {
			err := protovalidate.Validate(themed(&Auth0BrandingTheme{
				Identifiers: identifiers("unified", "international", "hide_all"),
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
