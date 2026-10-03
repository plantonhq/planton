package azureprivatednsresolverv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePrivateDnsResolverSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateDnsResolverSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testVnetId      = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet"
	testInSubnetId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet/subnets/dns-inbound"
	testOutSubnetId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet/subnets/dns-outbound"
)

// validResource returns a minimal valid resolver (no endpoints) that
// individual cases mutate into the shape under test.
func validResource() *AzurePrivateDnsResolver {
	return &AzurePrivateDnsResolver{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePrivateDnsResolver",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-resolver",
		},
		Spec: &AzurePrivateDnsResolverSpec{
			Region:           "eastus",
			ResourceGroup:    literal("platform-rg"),
			Name:             "platform-dns-resolver",
			VirtualNetworkId: literal(testVnetId),
		},
	}
}

var _ = ginkgo.Describe("AzurePrivateDnsResolverSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_dns_resolver", func() {

			ginkgo.It("should not return a validation error for the minimal resolver (no endpoints)", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept one dynamic inbound endpoint (allocation method unspecified)", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{Name: "inbound", SubnetId: literal(testInSubnetId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit DYNAMIC inbound endpoint without an address", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{
						Name:                      "inbound",
						SubnetId:                  literal(testInSubnetId),
						PrivateIpAllocationMethod: AzurePrivateDnsResolverIpAllocationMethod_DYNAMIC,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a STATIC inbound endpoint pinning an address", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{
						Name:                      "inbound",
						SubnetId:                  literal(testInSubnetId),
						PrivateIpAllocationMethod: AzurePrivateDnsResolverIpAllocationMethod_STATIC,
						PrivateIpAddress:          "10.10.0.4",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full hybrid shape: inbound plus outbound endpoints", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{Name: "inbound", SubnetId: literal(testInSubnetId)},
				}
				input.Spec.OutboundEndpoints = []*AzurePrivateDnsResolverOutboundEndpoint{
					{Name: "outbound", SubnetId: literal(testOutSubnetId)},
				}
				input.Spec.Tags = map[string]string{"team": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple endpoints with unique names", func() {
				input := validResource()
				input.Spec.OutboundEndpoints = []*AzurePrivateDnsResolverOutboundEndpoint{
					{Name: "outbound-1", SubnetId: literal(testOutSubnetId)},
					{Name: "outbound-2", SubnetId: literal(testInSubnetId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_private_dns_resolver", func() {

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resolver name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing virtual network", func() {
				input := validResource()
				input.Spec.VirtualNetworkId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate inbound endpoint names", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{Name: "inbound", SubnetId: literal(testInSubnetId)},
					{Name: "inbound", SubnetId: literal(testOutSubnetId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "unique")).To(gomega.BeTrue())
			})

			ginkgo.It("should reject duplicate outbound endpoint names", func() {
				input := validResource()
				input.Spec.OutboundEndpoints = []*AzurePrivateDnsResolverOutboundEndpoint{
					{Name: "outbound", SubnetId: literal(testOutSubnetId)},
					{Name: "outbound", SubnetId: literal(testInSubnetId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "unique")).To(gomega.BeTrue())
			})

			ginkgo.It("should reject an inbound endpoint without a name", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{SubnetId: literal(testInSubnetId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an inbound endpoint without a subnet", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{Name: "inbound"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an outbound endpoint without a subnet", func() {
				input := validResource()
				input.Spec.OutboundEndpoints = []*AzurePrivateDnsResolverOutboundEndpoint{
					{Name: "outbound"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject STATIC allocation without an address (the provider's own contract)", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{
						Name:                      "inbound",
						SubnetId:                  literal(testInSubnetId),
						PrivateIpAllocationMethod: AzurePrivateDnsResolverIpAllocationMethod_STATIC,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "STATIC")).To(gomega.BeTrue())
			})

			ginkgo.It("should reject an address with explicit DYNAMIC allocation (the provider's own contract)", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{
						Name:                      "inbound",
						SubnetId:                  literal(testInSubnetId),
						PrivateIpAllocationMethod: AzurePrivateDnsResolverIpAllocationMethod_DYNAMIC,
						PrivateIpAddress:          "10.10.0.4",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an address when the allocation method is unspecified (defaults to DYNAMIC)", func() {
				input := validResource()
				input.Spec.InboundEndpoints = []*AzurePrivateDnsResolverInboundEndpoint{
					{
						Name:             "inbound",
						SubnetId:         literal(testInSubnetId),
						PrivateIpAddress: "10.10.0.4",
					},
				}
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
				input.Kind = "AzureDnsResolver"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
