package azurelocalnetworkgatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureLocalNetworkGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureLocalNetworkGatewaySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// validResource returns a minimal valid site description (static public
// IP, one reachable prefix) that individual cases mutate into the shape
// under test.
func validResource() *AzureLocalNetworkGateway {
	return &AzureLocalNetworkGateway{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureLocalNetworkGateway",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-lngw",
		},
		Spec: &AzureLocalNetworkGatewaySpec{
			Region:         "eastus",
			ResourceGroup:  literal("test-rg"),
			Name:           "hq-datacenter",
			GatewayAddress: "203.0.113.10",
			AddressSpaces:  []string{"192.168.100.0/24"},
		},
	}
}

var _ = ginkgo.Describe("AzureLocalNetworkGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_local_network_gateway", func() {

			ginkgo.It("should not return a validation error for a minimal static site", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an FQDN endpoint instead of an address", func() {
				input := validResource()
				input.Spec.GatewayAddress = ""
				input.Spec.GatewayFqdn = "vpn.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a BGP-only site with no static address spaces", func() {
				input := validResource()
				input.Spec.AddressSpaces = nil
				input.Spec.BgpSettings = &AzureLocalNetworkGatewayBgpSettings{
					Asn:               65010,
					BgpPeeringAddress: "10.255.255.1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept BGP alongside static address spaces with a peer weight", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureLocalNetworkGatewayBgpSettings{
					Asn:               65010,
					BgpPeeringAddress: "10.255.255.1",
					PeerWeight:        50,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_local_network_gateway", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both endpoint forms are set", func() {
				input := validResource()
				input.Spec.GatewayFqdn = "vpn.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither endpoint form is set", func() {
				input := validResource()
				input.Spec.GatewayAddress = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a non-IP gateway_address", func() {
				input := validResource()
				input.Spec.GatewayAddress = "not-an-ip"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when no routing source exists", func() {
				input := validResource()
				input.Spec.AddressSpaces = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for BGP without a peering address", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureLocalNetworkGatewayBgpSettings{
					Asn: 65010,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a zero BGP ASN", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureLocalNetworkGatewayBgpSettings{
					Asn:               0,
					BgpPeeringAddress: "10.255.255.1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range peer weight", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureLocalNetworkGatewayBgpSettings{
					Asn:               65010,
					BgpPeeringAddress: "10.255.255.1",
					PeerWeight:        101,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
