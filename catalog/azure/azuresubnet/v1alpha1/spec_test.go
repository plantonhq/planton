package azuresubnetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureSubnetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureSubnetSpec Validation Tests")
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

const testVirtualNetworkId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet"

// validResource returns a minimal valid AzureSubnet that individual cases
// then mutate into the shape under test.
func validResource() *AzureSubnet {
	return &AzureSubnet{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureSubnet",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-subnet",
		},
		Spec: &AzureSubnetSpec{
			VirtualNetworkId: literal(testVirtualNetworkId),
			Name:             "my-subnet",
			AddressPrefixes:  []string{"10.0.1.0/24"},
		},
	}
}

var _ = ginkgo.Describe("AzureSubnetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_subnet", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the virtual network as a reference", func() {
				input := validResource()
				input.Spec.VirtualNetworkId = ref("prod-vnet")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple address prefixes (dual-stack)", func() {
				input := validResource()
				input.Spec.AddressPrefixes = []string{"10.0.1.0/24", "fd00:db8:deca:1::/64"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an IPAM pool instead of address prefixes", func() {
				input := validResource()
				input.Spec.AddressPrefixes = nil
				input.Spec.IpAddressPool = &AzureSubnetIpAddressPool{
					Id:                  "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/networkManagers/nm/ipamPools/pool1",
					NumberOfIpAddresses: "256",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept service endpoints with policy ids", func() {
				input := validResource()
				input.Spec.ServiceEndpoints = []string{"Microsoft.Storage", "Microsoft.Sql"}
				input.Spec.ServiceEndpointPolicyIds = []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/serviceEndpointPolicies/storage-only",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a delegation without actions", func() {
				input := validResource()
				input.Spec.Delegations = []*AzureSubnetDelegation{
					{Name: "container-apps", ServiceName: "Microsoft.App/environments"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple delegations with actions", func() {
				input := validResource()
				input.Spec.Delegations = []*AzureSubnetDelegation{
					{
						Name:        "postgresql",
						ServiceName: "Microsoft.DBforPostgreSQL/flexibleServers",
						Actions:     []string{"Microsoft.Network/virtualNetworks/subnets/join/action"},
					},
					{Name: "netapp", ServiceName: "Microsoft.Netapp/volumes"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every private endpoint network policy mode", func() {
				for _, mode := range []AzureSubnetPrivateEndpointNetworkPolicies{
					AzureSubnetPrivateEndpointNetworkPolicies_ENABLED,
					AzureSubnetPrivateEndpointNetworkPolicies_NETWORK_SECURITY_GROUP_ENABLED,
					AzureSubnetPrivateEndpointNetworkPolicies_ROUTE_TABLE_ENABLED,
				} {
					input := validResource()
					input.Spec.PrivateEndpointNetworkPolicies = mode
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept private link service network policies disabled", func() {
				input := validResource()
				disabled := false
				input.Spec.PrivateLinkServiceNetworkPoliciesEnabled = &disabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept disabling default outbound access", func() {
				input := validResource()
				disabled := false
				input.Spec.DefaultOutboundAccessEnabled = &disabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept tenant sharing scope when default outbound access is explicitly false", func() {
				input := validResource()
				disabled := false
				input.Spec.DefaultOutboundAccessEnabled = &disabled
				input.Spec.SharingScope = AzureSubnetSharingScope_TENANT
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the route table attachment as a reference", func() {
				input := validResource()
				input.Spec.RouteTableId = ref("egress-firewall-routes")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept all three attachments together", func() {
				input := validResource()
				input.Spec.RouteTableId = literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/routeTables/rt")
				input.Spec.NetworkSecurityGroupId = literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/networkSecurityGroups/nsg")
				input.Spec.NatGatewayId = literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/natGateways/nat")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_subnet", func() {

			ginkgo.It("should return a validation error when virtual_network_id is missing", func() {
				input := validResource()
				input.Spec.VirtualNetworkId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 80 characters", func() {
				input := validResource()
				tooLongName := ""
				for len(tooLongName) < 81 {
					tooLongName += "a"
				}
				input.Spec.Name = tooLongName
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a non-alphanumeric character", func() {
				input := validResource()
				input.Spec.Name = "-bad-name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name ends with a period", func() {
				input := validResource()
				input.Spec.Name = "bad-name."
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when no address source is set", func() {
				input := validResource()
				input.Spec.AddressPrefixes = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both address sources are set", func() {
				input := validResource()
				input.Spec.IpAddressPool = &AzureSubnetIpAddressPool{
					Id:                  "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/networkManagers/nm/ipamPools/pool1",
					NumberOfIpAddresses: "256",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the IPAM pool count is not a positive number", func() {
				input := validResource()
				input.Spec.AddressPrefixes = nil
				input.Spec.IpAddressPool = &AzureSubnetIpAddressPool{
					Id:                  "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/networkManagers/nm/ipamPools/pool1",
					NumberOfIpAddresses: "0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a delegation is missing its name", func() {
				input := validResource()
				input.Spec.Delegations = []*AzureSubnetDelegation{
					{ServiceName: "Microsoft.DBforPostgreSQL/flexibleServers"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a delegation is missing its service name", func() {
				input := validResource()
				input.Spec.Delegations = []*AzureSubnetDelegation{
					{Name: "postgresql"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sharing scope is set without disabling default outbound access", func() {
				input := validResource()
				input.Spec.SharingScope = AzureSubnetSharingScope_TENANT
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sharing scope is set with default outbound access explicitly true", func() {
				input := validResource()
				enabled := true
				input.Spec.DefaultOutboundAccessEnabled = &enabled
				input.Spec.SharingScope = AzureSubnetSharingScope_TENANT
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
