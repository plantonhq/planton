package azurevirtualnetworkpeeringv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVirtualNetworkPeeringSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualNetworkPeeringSpec Validation Tests")
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

const (
	testLocalNetworkId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/hub-rg/providers/Microsoft.Network/virtualNetworks/hub-vnet"
	testRemoteNetworkId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/spoke-rg/providers/Microsoft.Network/virtualNetworks/spoke-vnet"
)

// validResource returns a minimal valid AzureVirtualNetworkPeering that
// individual cases then mutate into the shape under test.
func validResource() *AzureVirtualNetworkPeering {
	return &AzureVirtualNetworkPeering{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVirtualNetworkPeering",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-peering",
		},
		Spec: &AzureVirtualNetworkPeeringSpec{
			Name:                   "hub-to-spoke",
			VirtualNetworkId:       literal(testLocalNetworkId),
			RemoteVirtualNetworkId: literal(testRemoteNetworkId),
		},
	}
}

var _ = ginkgo.Describe("AzureVirtualNetworkPeeringSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_virtual_network_peering", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept both networks as references", func() {
				input := validResource()
				input.Spec.VirtualNetworkId = ref("hub-vnet")
				input.Spec.RemoteVirtualNetworkId = ref("spoke-vnet")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the hub-side gateway transit shape", func() {
				input := validResource()
				forwarded := true
				transit := true
				input.Spec.AllowForwardedTraffic = &forwarded
				input.Spec.AllowGatewayTransit = &transit
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the spoke-side remote gateway shape", func() {
				input := validResource()
				useRemote := true
				input.Spec.UseRemoteGateways = &useRemote
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept subnet-scoped peering with subnet names", func() {
				input := validResource()
				complete := false
				input.Spec.PeerCompleteVirtualNetworksEnabled = &complete
				input.Spec.LocalSubnetNames = []string{"app"}
				input.Spec.RemoteSubnetNames = []string{"data"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept IPv6-only peering", func() {
				input := validResource()
				ipv6 := true
				input.Spec.OnlyIpv6PeeringEnabled = &ipv6
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_virtual_network_peering", func() {

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a non-alphanumeric character", func() {
				input := validResource()
				input.Spec.Name = "-bad"
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

			ginkgo.It("should return a validation error when the local network is missing", func() {
				input := validResource()
				input.Spec.VirtualNetworkId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the remote network is missing", func() {
				input := validResource()
				input.Spec.RemoteVirtualNetworkId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when local subnet names are set without subnet-scoped peering", func() {
				input := validResource()
				input.Spec.LocalSubnetNames = []string{"app"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet names are set with complete peering explicitly on", func() {
				input := validResource()
				complete := true
				input.Spec.PeerCompleteVirtualNetworksEnabled = &complete
				input.Spec.RemoteSubnetNames = []string{"data"}
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
