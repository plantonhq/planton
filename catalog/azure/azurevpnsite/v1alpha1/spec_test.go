package azurevpnsitev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureVpnSiteSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVpnSiteSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testWanId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualWans/test-wan"

// validResource returns a minimal valid VPN site that individual cases
// mutate into the shape under test.
func validResource() *AzureVpnSite {
	return &AzureVpnSite{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVpnSite",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-vpn-site",
		},
		Spec: &AzureVpnSiteSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "branch-london",
			VirtualWanId:  literal(testWanId),
			AddressCidrs:  []string{"192.168.10.0/24"},
		},
	}
}

// validLink returns a connectable single-endpoint link.
func validLink(name string) *AzureVpnSiteLink {
	return &AzureVpnSiteLink{
		Name:      name,
		IpAddress: "203.0.113.10",
	}
}

var _ = ginkgo.Describe("AzureVpnSiteSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_vpn_site", func() {

			ginkgo.It("should not return a validation error for a minimal site", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a link with a static IP endpoint", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{validLink("primary-isp")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a link with an FQDN endpoint", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{
					{Name: "primary-isp", Fqdn: "vpn.branch.example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept two links with distinct names (the active-active shape)", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{
					validLink("primary-isp"),
					{Name: "backup-isp", IpAddress: "198.51.100.7", ProviderName: "Airtel", SpeedInMbps: 100},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a BGP-speaking link and an empty address space", func() {
				input := validResource()
				input.Spec.AddressCidrs = nil
				input.Spec.Links = []*AzureVpnSiteLink{
					{
						Name:      "primary-isp",
						IpAddress: "203.0.113.10",
						Bgp:       &AzureVpnSiteLinkBgp{Asn: 65010, PeeringAddress: "169.254.21.5"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an O365 policy that states breakout on with every category off", func() {
				input := validResource()
				input.Spec.O365Policy = &AzureVpnSiteO365Policy{
					TrafficCategory: &AzureVpnSiteO365TrafficCategory{
						AllowEndpointEnabled:    proto.Bool(false),
						DefaultEndpointEnabled:  proto.Bool(false),
						OptimizeEndpointEnabled: proto.Bool(false),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept device metadata and an O365 breakout policy", func() {
				input := validResource()
				input.Spec.DeviceVendor = "Cisco"
				input.Spec.DeviceModel = "ISR4331"
				input.Spec.O365Policy = &AzureVpnSiteO365Policy{
					TrafficCategory: &AzureVpnSiteO365TrafficCategory{
						OptimizeEndpointEnabled: proto.Bool(true),
						AllowEndpointEnabled:    proto.Bool(true),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an IPv6 link endpoint", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{
					{Name: "primary-isp", IpAddress: "2001:db8::7"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_vpn_site", func() {

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

			ginkgo.It("should return a validation error for a name carrying a provider-forbidden character", func() {
				input := validResource()
				input.Spec.Name = "branch/london"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when virtual_wan_id is missing", func() {
				input := validResource()
				input.Spec.VirtualWanId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a link with no endpoint", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{{Name: "primary-isp"}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a link with a malformed IP", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{
					{Name: "primary-isp", IpAddress: "not-an-ip"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate link names", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{
					validLink("primary-isp"),
					{Name: "primary-isp", IpAddress: "198.51.100.7"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a link without a name", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{{IpAddress: "203.0.113.10"}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a BGP block with a malformed peering address", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{
					{
						Name:      "primary-isp",
						IpAddress: "203.0.113.10",
						Bgp:       &AzureVpnSiteLinkBgp{Asn: 65010, PeeringAddress: "peer.local"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a BGP block with ASN 0", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{
					{
						Name:      "primary-isp",
						IpAddress: "203.0.113.10",
						Bgp:       &AzureVpnSiteLinkBgp{Asn: 0, PeeringAddress: "169.254.21.5"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a negative link speed", func() {
				input := validResource()
				input.Spec.Links = []*AzureVpnSiteLink{
					{Name: "primary-isp", IpAddress: "203.0.113.10", SpeedInMbps: -1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
