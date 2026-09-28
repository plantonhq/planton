package auth0tenantsettingsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAuth0TenantSettings(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0TenantSettings Suite")
}

func settings(spec *Auth0TenantSettingsSpec) *Auth0TenantSettings {
	return &Auth0TenantSettings{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0TenantSettings",
		Metadata:   &shared.CloudResourceMetadata{Name: "tenant-settings"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("Auth0TenantSettings Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts every setting", func() {
			err := protovalidate.Validate(settings(&Auth0TenantSettingsSpec{
				FriendlyName: "Acme",
				PictureUrl:   "https://assets.acme.com/logo.png",
				SupportEmail: "support@acme.com",
				SupportUrl:   "https://acme.com/support",
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts the default domain alone, read from a verified custom domain", func() {
			err := protovalidate.Validate(settings(&Auth0TenantSettingsSpec{
				DefaultCustomDomain: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      cloudresourcekind.CloudResourceKind_Auth0CustomDomainVerification,
						Name:      "sign-in-domain-verification",
						FieldPath: "status.outputs.domain",
					}},
				},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts the canonical domain as the default", func() {
			err := protovalidate.Validate(settings(&Auth0TenantSettingsSpec{
				DefaultCustomDomain: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "acme.eu.auth0.com"},
				},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts one setting, leaving the others unmanaged", func() {
			err := protovalidate.Validate(settings(&Auth0TenantSettingsSpec{FriendlyName: "Acme"}))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a spec that manages nothing", func() {
			err := protovalidate.Validate(settings(&Auth0TenantSettingsSpec{}))
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("configure at least one tenant setting"))
		})

		ginkgo.It("refuses a support email that is not an address", func() {
			err := protovalidate.Validate(settings(&Auth0TenantSettingsSpec{SupportEmail: "support"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a logo that is not a URL", func() {
			err := protovalidate.Validate(settings(&Auth0TenantSettingsSpec{PictureUrl: "logo.png"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses a support page that is not a URL", func() {
			err := protovalidate.Validate(settings(&Auth0TenantSettingsSpec{SupportUrl: "acme support"}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("refuses an empty default domain", func() {
			err := protovalidate.Validate(settings(&Auth0TenantSettingsSpec{
				DefaultCustomDomain: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: ""},
				},
			}))
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
