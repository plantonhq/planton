package auth0customdomainv1alpha1

import (
	"fmt"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAuth0CustomDomain(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0CustomDomain Suite")
}

func customDomain(spec *Auth0CustomDomainSpec) *Auth0CustomDomain {
	return &Auth0CustomDomain{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0CustomDomain",
		Metadata:   &shared.CatalogObjectMetadata{Name: "sign-in-domain"},
		Spec:       spec,
	}
}

// pairs builds n distinct domain_metadata entries.
func pairs(n int) map[string]string {
	m := make(map[string]string, n)
	for i := 0; i < n; i++ {
		m[fmt.Sprintf("key%d", i)] = "value"
	}
	return m
}

var _ = ginkgo.Describe("Auth0CustomDomain Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts an Auth0-managed domain", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain: "id.example.com",
				Type:   "auth0_managed_certs",
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an Auth0-managed domain with the recommended TLS policy", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:    "id.example.com",
				Type:      "auth0_managed_certs",
				TlsPolicy: "recommended",
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a self-managed domain behind a proxy that names the client's address", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:               "login.example.com",
				Type:                 "self_managed_certs",
				CustomClientIpHeader: "cf-connecting-ip",
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a parent domain as the passkeys' relying party", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:                 "id.example.com",
				Type:                   "auth0_managed_certs",
				RelyingPartyIdentifier: "example.com",
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts ten metadata pairs", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:         "id.example.com",
				Type:           "auth0_managed_certs",
				DomainMetadata: pairs(10),
			}))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing domain", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{Type: "auth0_managed_certs"}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a domain that is not a host name", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain: "https://id.example.com",
				Type:   "auth0_managed_certs",
			}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing type", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{Domain: "id.example.com"}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown type", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain: "id.example.com",
				Type:   "managed",
			}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects the retired compatible TLS policy", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:    "id.example.com",
				Type:      "auth0_managed_certs",
				TlsPolicy: "compatible",
			}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a TLS policy on a self-managed domain", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:    "login.example.com",
				Type:      "self_managed_certs",
				TlsPolicy: "recommended",
			}))
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("tls_policy applies only to an Auth0-managed domain"))
		})

		ginkgo.It("rejects an unknown client IP header", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:               "id.example.com",
				Type:                 "auth0_managed_certs",
				CustomClientIpHeader: "x-real-ip",
			}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects more than ten metadata pairs", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:         "id.example.com",
				Type:           "auth0_managed_certs",
				DomainMetadata: pairs(11),
			}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a metadata value longer than 255 characters", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:         "id.example.com",
				Type:           "auth0_managed_certs",
				DomainMetadata: map[string]string{"brand": strings.Repeat("a", 256)},
			}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a relying party that is not a host name", func() {
			err := protovalidate.Validate(customDomain(&Auth0CustomDomainSpec{
				Domain:                 "id.example.com",
				Type:                   "auth0_managed_certs",
				RelyingPartyIdentifier: "https://example.com",
			}))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
