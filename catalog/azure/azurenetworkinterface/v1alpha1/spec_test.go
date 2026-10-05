package azurenetworkinterfacev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

const testSubnetId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/net-rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/app"

// validSpec returns a minimal valid single-configuration NIC the failure
// cases mutate one field at a time.
func validSpec() *AzureNetworkInterfaceSpec {
	return &AzureNetworkInterfaceSpec{
		Region:        "eastus",
		ResourceGroup: stringRef("test-rg"),
		Name:          "app-nic",
		IpConfigurations: []*AzureNetworkInterfaceIpConfiguration{
			{
				Name:     "primary",
				SubnetId: stringRef(testSubnetId),
			},
		},
	}
}

func validInput(spec *AzureNetworkInterfaceSpec) *AzureNetworkInterface {
	return &AzureNetworkInterface{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureNetworkInterface",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-network-interface",
		},
		Spec: spec,
	}
}

func TestAzureNetworkInterfaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureNetworkInterfaceSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureNetworkInterfaceSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal single-configuration NIC", func() {
			err := protovalidate.Validate(validInput(validSpec()))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a static private address", func() {
			spec := validSpec()
			spec.IpConfigurations[0].PrivateIpAllocation = AzureNetworkInterfacePrivateIpAllocation_STATIC
			spec.IpConfigurations[0].PrivateIpAddress = "10.0.1.10"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a public-IP-fronted configuration", func() {
			spec := validSpec()
			spec.IpConfigurations[0].PublicIpAddressId = stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/pip")
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a dual-stack NIC with the first configuration primary", func() {
			spec := validSpec()
			spec.IpConfigurations[0].Primary = true
			spec.IpConfigurations = append(spec.IpConfigurations, &AzureNetworkInterfaceIpConfiguration{
				Name:             "ipv6",
				PrivateIpVersion: AzureNetworkInterfacePrivateIpVersion_IPV6,
			})
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an appliance NIC with forwarding, acceleration, and an NSG", func() {
			spec := validSpec()
			spec.IpForwardingEnabled = true
			spec.AcceleratedNetworkingEnabled = true
			spec.NetworkSecurityGroupId = stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/networkSecurityGroups/nsg")
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a paired auxiliary mode and SKU", func() {
			spec := validSpec()
			spec.AuxiliaryMode = AzureNetworkInterfaceAuxiliaryMode_ACCELERATED_CONNECTIONS
			spec.AuxiliarySku = AzureNetworkInterfaceAuxiliarySku_A2
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept ASG memberships, DNS overrides, and tags", func() {
			spec := validSpec()
			spec.ApplicationSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/applicationSecurityGroups/web"),
			}
			spec.DnsServers = []string{"10.0.0.4"}
			spec.InternalDnsNameLabel = "app-1"
			spec.Tags = map[string]string{"cost-center": "platform"}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept load-balancer pool, NAT rule, and App Gateway pool memberships", func() {
			spec := validSpec()
			spec.IpConfigurations[0].LoadBalancerBackendAddressPoolIds = []*foreignkeyv1.StringValueOrRef{
				stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/lb/backendAddressPools/web"),
			}
			spec.IpConfigurations[0].LoadBalancerInboundNatRuleIds = []*foreignkeyv1.StringValueOrRef{
				stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/lb/inboundNatRules/ssh-admin"),
			}
			spec.IpConfigurations[0].ApplicationGatewayBackendAddressPoolIds = []*foreignkeyv1.StringValueOrRef{
				stringRef("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/applicationGateways/agw/backendAddressPools/web"),
			}
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing region", func() {
			spec := validSpec()
			spec.Region = ""
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			spec := validSpec()
			spec.Name = ""
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a name with invalid characters", func() {
			spec := validSpec()
			spec.Name = "-bad-nic-"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a NIC without ip_configurations", func() {
			spec := validSpec()
			spec.IpConfigurations = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an ip_configuration without a name", func() {
			spec := validSpec()
			spec.IpConfigurations[0].Name = ""
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an IPv4 configuration without a subnet", func() {
			spec := validSpec()
			spec.IpConfigurations[0].SubnetId = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject STATIC allocation without an address", func() {
			spec := validSpec()
			spec.IpConfigurations[0].PrivateIpAllocation = AzureNetworkInterfacePrivateIpAllocation_STATIC
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a pinned address under DYNAMIC allocation", func() {
			spec := validSpec()
			spec.IpConfigurations[0].PrivateIpAddress = "10.0.1.10"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject multiple configurations when the first is not primary", func() {
			spec := validSpec()
			spec.IpConfigurations = append(spec.IpConfigurations, &AzureNetworkInterfaceIpConfiguration{
				Name:     "secondary",
				SubnetId: stringRef(testSubnetId),
				Primary:  true,
			})
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject more than one primary configuration", func() {
			spec := validSpec()
			spec.IpConfigurations[0].Primary = true
			spec.IpConfigurations = append(spec.IpConfigurations, &AzureNetworkInterfaceIpConfiguration{
				Name:     "secondary",
				SubnetId: stringRef(testSubnetId),
				Primary:  true,
			})
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an auxiliary mode without a SKU", func() {
			spec := validSpec()
			spec.AuxiliaryMode = AzureNetworkInterfaceAuxiliaryMode_MAX_CONNECTIONS
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an auxiliary SKU without a mode", func() {
			spec := validSpec()
			spec.AuxiliarySku = AzureNetworkInterfaceAuxiliarySku_A1
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
