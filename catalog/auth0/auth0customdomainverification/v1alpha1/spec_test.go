package auth0customdomainverificationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAuth0CustomDomainVerification(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0CustomDomainVerification Suite")
}

func verification(spec *Auth0CustomDomainVerificationSpec) *Auth0CustomDomainVerification {
	return &Auth0CustomDomainVerification{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0CustomDomainVerification",
		Metadata:   &shared.CatalogObjectMetadata{Name: "sign-in-domain-verification"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("Auth0CustomDomainVerification Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a custom domain id", func() {
			err := protovalidate.Validate(verification(&Auth0CustomDomainVerificationSpec{
				CustomDomainId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "cd_0123456789abcdef"},
				},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a reference to the Auth0CustomDomain", func() {
			err := protovalidate.Validate(verification(&Auth0CustomDomainVerificationSpec{
				CustomDomainId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_Auth0CustomDomain,
						Name:      "sign-in-domain",
						FieldPath: "status.outputs.id",
					}},
				},
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing custom domain", func() {
			err := protovalidate.Validate(verification(&Auth0CustomDomainVerificationSpec{}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty custom domain id", func() {
			err := protovalidate.Validate(verification(&Auth0CustomDomainVerificationSpec{
				CustomDomainId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: ""},
				},
			}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
