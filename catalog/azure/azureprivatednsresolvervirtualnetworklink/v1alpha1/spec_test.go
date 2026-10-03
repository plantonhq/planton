package azureprivatednsresolvervirtualnetworklinkv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePrivateDnsResolverVirtualNetworkLinkSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateDnsResolverVirtualNetworkLinkSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testRulesetId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/dnsForwardingRulesets/platform-ruleset"
	testVnetId    = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/spoke-rg/providers/Microsoft.Network/virtualNetworks/spoke-vnet"
)

// validResource returns a minimal valid link that individual cases
// mutate into the shape under test.
func validResource() *AzurePrivateDnsResolverVirtualNetworkLink {
	return &AzurePrivateDnsResolverVirtualNetworkLink{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePrivateDnsResolverVirtualNetworkLink",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-link",
		},
		Spec: &AzurePrivateDnsResolverVirtualNetworkLinkSpec{
			Name:                   "spoke-vnet",
			DnsForwardingRulesetId: literal(testRulesetId),
			VirtualNetworkId:       literal(testVnetId),
		},
	}
}

var _ = ginkgo.Describe("AzurePrivateDnsResolverVirtualNetworkLinkSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_dns_resolver_virtual_network_link", func() {

			ginkgo.It("should not return a validation error for the minimal link", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a link carrying metadata annotations", func() {
				input := validResource()
				input.Spec.Metadata = map[string]string{"owner": "payments-team", "env": "prod"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_private_dns_resolver_virtual_network_link", func() {

			ginkgo.It("should reject a missing link name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing ruleset reference", func() {
				input := validResource()
				input.Spec.DnsForwardingRulesetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing virtual network reference", func() {
				input := validResource()
				input.Spec.VirtualNetworkId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong api version", func() {
				input := validResource()
				input.ApiVersion = "azure.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong kind", func() {
				input := validResource()
				input.Kind = "AzurePrivateDnsZoneVirtualNetworkLink"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
