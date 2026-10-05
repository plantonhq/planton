package azureprivatelinkservicev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePrivateLinkServiceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateLinkServiceSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// validResource returns a minimal valid service in the classic
// load-balancer-frontend shape that individual cases mutate into the
// shape under test.
func validResource() *AzurePrivateLinkService {
	return &AzurePrivateLinkService{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePrivateLinkService",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-pls",
		},
		Spec: &AzurePrivateLinkServiceSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "orders-api",
			NatIpConfigurations: []*AzurePrivateLinkServiceNatIpConfiguration{
				{
					Name:     "nat-1",
					SubnetId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/pls"),
					Primary:  true,
				},
			},
			LoadBalancerFrontendIpConfigurationIds: []*foreignkeyv1.StringValueOrRef{
				literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/lb/frontendIPConfigurations/internal"),
			},
		},
	}
}

var _ = ginkgo.Describe("AzurePrivateLinkServiceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_link_service", func() {

			ginkgo.It("should not return a validation error for the load-balancer shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the destination-ip shape instead of a load balancer", func() {
				input := validResource()
				input.Spec.LoadBalancerFrontendIpConfigurationIds = nil
				input.Spec.DestinationIpAddress = "10.0.5.10"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple NAT configurations with a single primary", func() {
				input := validResource()
				input.Spec.NatIpConfigurations = append(input.Spec.NatIpConfigurations,
					&AzurePrivateLinkServiceNatIpConfiguration{
						Name:     "nat-2",
						SubnetId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/pls"),
					})
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a static NAT address with proxy protocol and fqdns", func() {
				input := validResource()
				input.Spec.NatIpConfigurations[0].PrivateIpAddress = "10.0.5.7"
				input.Spec.ProxyProtocolEnabled = true
				input.Spec.Fqdns = []string{"orders.internal.example.com"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept subscription visibility and auto-approval lists", func() {
				input := validResource()
				input.Spec.VisibilitySubscriptionIds = []string{"8158df85-3d6b-4d9f-8a3c-247b63cab0a8"}
				input.Spec.AutoApprovalSubscriptionIds = []string{"8158df85-3d6b-4d9f-8a3c-247b63cab0a8"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the public-visibility wildcard", func() {
				input := validResource()
				input.Spec.VisibilitySubscriptionIds = []string{"*"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_private_link_service", func() {

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

			ginkgo.It("should return a validation error for a name starting with a hyphen", func() {
				input := validResource()
				input.Spec.Name = "-orders"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a name ending with a period", func() {
				input := validResource()
				input.Spec.Name = "orders."
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both destinations are set", func() {
				input := validResource()
				input.Spec.DestinationIpAddress = "10.0.5.10"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither destination is set", func() {
				input := validResource()
				input.Spec.LoadBalancerFrontendIpConfigurationIds = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a non-IP destination address", func() {
				input := validResource()
				input.Spec.LoadBalancerFrontendIpConfigurationIds = nil
				input.Spec.DestinationIpAddress = "not-an-ip"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error with zero NAT configurations", func() {
				input := validResource()
				input.Spec.NatIpConfigurations = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error with nine NAT configurations", func() {
				input := validResource()
				for i := 0; i < 8; i++ {
					input.Spec.NatIpConfigurations = append(input.Spec.NatIpConfigurations,
						&AzurePrivateLinkServiceNatIpConfiguration{
							Name:     "extra-" + string(rune('a'+i)),
							SubnetId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/pls"),
						})
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when no configuration is primary", func() {
				input := validResource()
				input.Spec.NatIpConfigurations[0].Primary = false
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when two configurations are primary", func() {
				input := validResource()
				input.Spec.NatIpConfigurations = append(input.Spec.NatIpConfigurations,
					&AzurePrivateLinkServiceNatIpConfiguration{
						Name:     "nat-2",
						SubnetId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/pls"),
						Primary:  true,
					})
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate NAT configuration names", func() {
				input := validResource()
				input.Spec.NatIpConfigurations = append(input.Spec.NatIpConfigurations,
					&AzurePrivateLinkServiceNatIpConfiguration{
						Name:     "nat-1",
						SubnetId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/pls"),
					})
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a NAT configuration has no subnet", func() {
				input := validResource()
				input.Spec.NatIpConfigurations[0].SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a non-IP static NAT address", func() {
				input := validResource()
				input.Spec.NatIpConfigurations[0].PrivateIpAddress = "not-an-ip"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an IPv6 address version", func() {
				input := validResource()
				version := "IPv6"
				input.Spec.NatIpConfigurations[0].PrivateIpAddressVersion = &version
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a non-UUID auto-approval entry", func() {
				input := validResource()
				input.Spec.AutoApprovalSubscriptionIds = []string{"not-a-uuid"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a non-UUID visibility entry", func() {
				input := validResource()
				input.Spec.VisibilitySubscriptionIds = []string{"everyone"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
