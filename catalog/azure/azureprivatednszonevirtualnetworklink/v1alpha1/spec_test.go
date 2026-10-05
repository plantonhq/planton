package azureprivatednszonevirtualnetworklinkv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePrivateDnsZoneVirtualNetworkLinkSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateDnsZoneVirtualNetworkLinkSpec Validation Tests")
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

// validResource returns a minimal valid AzurePrivateDnsZoneVirtualNetworkLink
// that individual cases then mutate into the shape under test.
func validResource() *AzurePrivateDnsZoneVirtualNetworkLink {
	return &AzurePrivateDnsZoneVirtualNetworkLink{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePrivateDnsZoneVirtualNetworkLink",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-link",
		},
		Spec: &AzurePrivateDnsZoneVirtualNetworkLinkSpec{
			Name:             "hub-vnet",
			PrivateDnsZoneId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/privateDnsZones/corp.internal"),
			VirtualNetworkId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/hub"),
		},
	}
}

var _ = ginkgo.Describe("AzurePrivateDnsZoneVirtualNetworkLinkSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_dns_zone_virtual_network_link", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the zone and network as references", func() {
				input := validResource()
				input.Spec.PrivateDnsZoneId = ref("postgres-privatelink-zone")
				input.Spec.VirtualNetworkId = ref("hub-network")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept enabling VM auto-registration", func() {
				input := validResource()
				enabled := true
				input.Spec.RegistrationEnabled = &enabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the NX_DOMAIN_REDIRECT resolution policy", func() {
				input := validResource()
				input.Spec.ResolutionPolicy = AzurePrivateDnsZoneVirtualNetworkLinkResolutionPolicy_NX_DOMAIN_REDIRECT
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

			ginkgo.It("should accept a single-character name", func() {
				input := validResource()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_private_dns_zone_virtual_network_link", func() {

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

			ginkgo.It("should return a validation error when name ends with a hyphen", func() {
				input := validResource()
				input.Spec.Name = "link-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 80 characters", func() {
				input := validResource()
				long := make([]byte, 81)
				for i := range long {
					long[i] = 'a'
				}
				input.Spec.Name = string(long)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when private_dns_zone_id is missing", func() {
				input := validResource()
				input.Spec.PrivateDnsZoneId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when virtual_network_id is missing", func() {
				input := validResource()
				input.Spec.VirtualNetworkId = nil
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
