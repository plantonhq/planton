package azurefirewallv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureFirewallSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFirewallSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// validResource returns a minimal valid VNet-deployed AzureFirewall that
// individual cases then mutate into the shape under test.
func validResource() *AzureFirewall {
	return &AzureFirewall{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFirewall",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-firewall",
		},
		Spec: &AzureFirewallSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "hub-egress-fw",
			IpConfigurations: []*AzureFirewallIpConfiguration{
				{
					Name:              "primary",
					SubnetId:          ref("firewall-subnet"),
					PublicIpAddressId: ref("firewall-pip"),
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureFirewallSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_firewall", func() {

			ginkgo.It("should not return a validation error for a minimal VNet firewall", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit AZFW_VNET sku with STANDARD tier", func() {
				input := validResource()
				input.Spec.SkuName = AzureFirewallSkuName_AZFW_VNET
				input.Spec.SkuTier = AzureFirewallSkuTier_STANDARD
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a policy attachment and zones", func() {
				input := validResource()
				input.Spec.FirewallPolicyId = ref("egress-baseline")
				input.Spec.Zones = []string{"1", "2", "3"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a hub firewall with virtual_hub and no ip configurations", func() {
				input := validResource()
				input.Spec.SkuName = AzureFirewallSkuName_AZFW_HUB
				input.Spec.IpConfigurations = nil
				input.Spec.VirtualHub = &AzureFirewallVirtualHub{
					VirtualHubId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualHubs/hub"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a private-only data path with a management configuration", func() {
				input := validResource()
				input.Spec.IpConfigurations[0].PublicIpAddressId = nil
				input.Spec.ManagementIpConfiguration = &AzureFirewallManagementIpConfiguration{
					Name:              "management",
					SubnetId:          ref("firewall-mgmt-subnet"),
					PublicIpAddressId: ref("firewall-mgmt-pip"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept extra public-ip-only configurations alongside the subnet-bearing one", func() {
				input := validResource()
				input.Spec.IpConfigurations = append(input.Spec.IpConfigurations, &AzureFirewallIpConfiguration{
					Name:              "extra-pip",
					PublicIpAddressId: ref("firewall-pip-2"),
				})
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept dns servers with proxy and SNAT ranges on a policy-less firewall", func() {
				input := validResource()
				input.Spec.DnsServers = []string{"10.0.0.4"}
				input.Spec.DnsProxyEnabled = true
				input.Spec.PrivateIpRanges = []string{"IANAPrivateRanges", "100.64.0.0/10"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_firewall", func() {

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

			ginkgo.It("should return a validation error when name ends with a hyphen", func() {
				input := validResource()
				input.Spec.Name = "bad-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a VNet firewall without ip configurations", func() {
				input := validResource()
				input.Spec.IpConfigurations = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an AZFW_HUB firewall without virtual_hub", func() {
				input := validResource()
				input.Spec.SkuName = AzureFirewallSkuName_AZFW_HUB
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a virtual_hub on a VNet firewall", func() {
				input := validResource()
				input.Spec.VirtualHub = &AzureFirewallVirtualHub{
					VirtualHubId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualHubs/hub"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a hub firewall that keeps ip configurations", func() {
				input := validResource()
				input.Spec.SkuName = AzureFirewallSkuName_AZFW_HUB
				input.Spec.VirtualHub = &AzureFirewallVirtualHub{
					VirtualHubId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualHubs/hub"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when no configuration carries a subnet", func() {
				input := validResource()
				input.Spec.IpConfigurations[0].SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when two configurations carry subnets", func() {
				input := validResource()
				input.Spec.IpConfigurations = append(input.Spec.IpConfigurations, &AzureFirewallIpConfiguration{
					Name:     "second-subnet",
					SubnetId: ref("another-subnet"),
				})
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a private-only data path without a management configuration", func() {
				input := validResource()
				input.Spec.IpConfigurations[0].PublicIpAddressId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the management configuration reuses a data-path name", func() {
				input := validResource()
				input.Spec.ManagementIpConfiguration = &AzureFirewallManagementIpConfiguration{
					Name:              "primary",
					SubnetId:          ref("firewall-mgmt-subnet"),
					PublicIpAddressId: ref("firewall-mgmt-pip"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the management configuration omits its public IP", func() {
				input := validResource()
				input.Spec.ManagementIpConfiguration = &AzureFirewallManagementIpConfiguration{
					Name:     "management",
					SubnetId: ref("firewall-mgmt-subnet"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a virtual_hub without its hub id", func() {
				input := validResource()
				input.Spec.SkuName = AzureFirewallSkuName_AZFW_HUB
				input.Spec.IpConfigurations = nil
				input.Spec.VirtualHub = &AzureFirewallVirtualHub{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for dns_proxy_enabled on a policy-attached firewall", func() {
				input := validResource()
				input.Spec.FirewallPolicyId = ref("egress-baseline")
				input.Spec.DnsProxyEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for dns_servers on a policy-attached firewall", func() {
				input := validResource()
				input.Spec.FirewallPolicyId = ref("egress-baseline")
				input.Spec.DnsServers = []string{"10.0.0.4"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a zero public_ip_count", func() {
				input := validResource()
				input.Spec.SkuName = AzureFirewallSkuName_AZFW_HUB
				input.Spec.IpConfigurations = nil
				count := int32(0)
				input.Spec.VirtualHub = &AzureFirewallVirtualHub{
					VirtualHubId:  literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualHubs/hub"),
					PublicIpCount: &count,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
