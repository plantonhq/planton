package azurevirtualnetworkv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVirtualNetworkSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualNetworkSpec Validation Tests")
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

// validResource returns a minimal valid AzureVirtualNetwork that individual
// cases then mutate into the shape under test.
func validResource() *AzureVirtualNetwork {
	return &AzureVirtualNetwork{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVirtualNetwork",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-network",
		},
		Spec: &AzureVirtualNetworkSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "test-vnet",
			AddressSpaces: []string{"10.0.0.0/16"},
		},
	}
}

var _ = ginkgo.Describe("AzureVirtualNetworkSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_virtual_network", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the resource group as a reference", func() {
				input := validResource()
				input.Spec.ResourceGroup = ref("platform-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple address spaces", func() {
				input := validResource()
				input.Spec.AddressSpaces = []string{"10.0.0.0/16", "10.1.0.0/16", "fd00:db8::/48"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept IPAM pool allocation instead of address spaces", func() {
				input := validResource()
				input.Spec.AddressSpaces = nil
				input.Spec.IpAddressPools = []*AzureVirtualNetworkIpAddressPool{
					{
						Id:                  "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/networkManagers/nm/ipamPools/pool",
						NumberOfIpAddresses: "256",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept custom DNS servers", func() {
				input := validResource()
				input.Spec.DnsServers = []string{"10.0.0.4", "10.0.0.5"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a well-formed BGP community", func() {
				input := validResource()
				input.Spec.BgpCommunity = "12076:20010"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a DDoS protection plan attachment", func() {
				input := validResource()
				input.Spec.DdosProtectionPlan = &AzureVirtualNetworkDdosProtectionPlan{
					Id:     "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/ddosProtectionPlans/plan",
					Enable: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept encryption with allow-unencrypted enforcement", func() {
				input := validResource()
				input.Spec.Encryption = AzureVirtualNetworkEncryptionEnforcement_ALLOW_UNENCRYPTED
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a flow timeout inside the 4-30 range", func() {
				input := validResource()
				timeout := int32(15)
				input.Spec.FlowTimeoutInMinutes = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{
					"cost-center": "platform",
					"owner":       "network-team",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name at the 2-character minimum", func() {
				input := validResource()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name ending with an underscore", func() {
				input := validResource()
				input.Spec.Name = "vnet_"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_virtual_network", func() {

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

			ginkgo.It("should return a validation error when name has invalid characters", func() {
				input := validResource()
				input.Spec.Name = "bad name!"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name ends with a period", func() {
				input := validResource()
				input.Spec.Name = "vnet."
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither address source is set", func() {
				input := validResource()
				input.Spec.AddressSpaces = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both address sources are set", func() {
				input := validResource()
				input.Spec.IpAddressPools = []*AzureVirtualNetworkIpAddressPool{
					{
						Id:                  "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/networkManagers/nm/ipamPools/pool",
						NumberOfIpAddresses: "256",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when more than two IPAM pools are set", func() {
				input := validResource()
				input.Spec.AddressSpaces = nil
				pool := func(name string) *AzureVirtualNetworkIpAddressPool {
					return &AzureVirtualNetworkIpAddressPool{
						Id:                  "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/networkManagers/nm/ipamPools/" + name,
						NumberOfIpAddresses: "256",
					}
				}
				input.Spec.IpAddressPools = []*AzureVirtualNetworkIpAddressPool{pool("a"), pool("b"), pool("c")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an IPAM pool count is not a positive number", func() {
				input := validResource()
				input.Spec.AddressSpaces = nil
				input.Spec.IpAddressPools = []*AzureVirtualNetworkIpAddressPool{
					{
						Id:                  "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/networkManagers/nm/ipamPools/pool",
						NumberOfIpAddresses: "0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed BGP community", func() {
				input := validResource()
				input.Spec.BgpCommunity = "12076"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a BGP community segment out of range", func() {
				input := validResource()
				input.Spec.BgpCommunity = "12076:0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a BGP community segment above 65534", func() {
				input := validResource()
				input.Spec.BgpCommunity = "12076:65535"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a DDoS plan attachment omits the plan id", func() {
				input := validResource()
				input.Spec.DdosProtectionPlan = &AzureVirtualNetworkDdosProtectionPlan{Enable: true}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when flow timeout is below 4 minutes", func() {
				input := validResource()
				timeout := int32(3)
				input.Spec.FlowTimeoutInMinutes = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when flow timeout exceeds 30 minutes", func() {
				input := validResource()
				timeout := int32(31)
				input.Spec.FlowTimeoutInMinutes = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := validResource()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := validResource()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := validResource()
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
