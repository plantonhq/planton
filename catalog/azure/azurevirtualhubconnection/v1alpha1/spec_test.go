package azurevirtualhubconnectionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVirtualHubConnectionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualHubConnectionSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// criteriaPtr returns a pointer to the given override-criteria enum
// value (the field is optional, so the generated Go type is a pointer).
func criteriaPtr(value AzureVirtualHubConnectionStaticVnetLocalRouteOverrideCriteria) *AzureVirtualHubConnectionStaticVnetLocalRouteOverrideCriteria {
	return &value
}

// boolPtr returns a pointer to the given bool (for optional fields).
func boolPtr(value bool) *bool {
	return &value
}

const (
	testHubId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus"
	testVnetId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/spoke-app"
)

// validResource returns a minimal valid spoke attachment that
// individual cases mutate into the shape under test.
func validResource() *AzureVirtualHubConnection {
	return &AzureVirtualHubConnection{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVirtualHubConnection",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-vhub-connection",
		},
		Spec: &AzureVirtualHubConnectionSpec{
			Name:                   "spoke-app",
			VirtualHubId:           literal(testHubId),
			RemoteVirtualNetworkId: literal(testVnetId),
		},
	}
}

var _ = ginkgo.Describe("AzureVirtualHubConnectionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_virtual_hub_connection", func() {

			ginkgo.It("should not return a validation error for a minimal attachment", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept internet security", func() {
				input := validResource()
				input.Spec.InternetSecurityEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept routing with an associated route table", func() {
				input := validResource()
				input.Spec.Routing = &AzureVirtualHubConnectionRouting{
					AssociatedRouteTableId: literal(testHubId + "/hubRouteTables/prod-isolated"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept routing that only propagates to labels", func() {
				input := validResource()
				input.Spec.Routing = &AzureVirtualHubConnectionRouting{
					PropagatedRouteTable: &AzureVirtualHubConnectionPropagatedRouteTable{
						Labels: []string{"prod", "shared"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a full routing block with static routes", func() {
				input := validResource()
				input.Spec.Routing = &AzureVirtualHubConnectionRouting{
					AssociatedRouteTableId: literal(testHubId + "/hubRouteTables/defaultRouteTable"),
					InboundRouteMapId:      literal(testHubId + "/routeMaps/ingress"),
					OutboundRouteMapId:     literal(testHubId + "/routeMaps/egress"),
					PropagatedRouteTable: &AzureVirtualHubConnectionPropagatedRouteTable{
						RouteTableIds: []*foreignkeyv1.StringValueOrRef{
							literal(testHubId + "/hubRouteTables/defaultRouteTable"),
						},
					},
					StaticVnetRoutes: []*AzureVirtualHubConnectionStaticVnetRoute{
						{
							Name:             "to-nva",
							AddressPrefixes:  []string{"10.50.0.0/16"},
							NextHopIpAddress: "10.20.1.4",
						},
					},
					StaticVnetLocalRouteOverrideCriteria:   criteriaPtr(AzureVirtualHubConnectionStaticVnetLocalRouteOverrideCriteria_EQUAL),
					StaticVnetPropagateStaticRoutesEnabled: boolPtr(false),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_virtual_hub_connection", func() {

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a single-character name (the provider's regex requires 2-80)", func() {
				input := validResource()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a name starting with a hyphen", func() {
				input := validResource()
				input.Spec.Name = "-spoke"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a name ending in a period", func() {
				input := validResource()
				input.Spec.Name = "spoke."
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when virtual_hub_id is missing", func() {
				input := validResource()
				input.Spec.VirtualHubId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when remote_virtual_network_id is missing", func() {
				input := validResource()
				input.Spec.RemoteVirtualNetworkId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a routing block that configures nothing", func() {
				input := validResource()
				input.Spec.Routing = &AzureVirtualHubConnectionRouting{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for propagation with no targets", func() {
				input := validResource()
				input.Spec.Routing = &AzureVirtualHubConnectionRouting{
					PropagatedRouteTable: &AzureVirtualHubConnectionPropagatedRouteTable{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a static route without a name", func() {
				input := validResource()
				input.Spec.Routing = &AzureVirtualHubConnectionRouting{
					StaticVnetRoutes: []*AzureVirtualHubConnectionStaticVnetRoute{
						{AddressPrefixes: []string{"10.50.0.0/16"}, NextHopIpAddress: "10.20.1.4"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a static route with a prefix missing its mask", func() {
				input := validResource()
				input.Spec.Routing = &AzureVirtualHubConnectionRouting{
					StaticVnetRoutes: []*AzureVirtualHubConnectionStaticVnetRoute{
						{Name: "to-nva", AddressPrefixes: []string{"10.50.0.0"}, NextHopIpAddress: "10.20.1.4"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a static route with a malformed next hop", func() {
				input := validResource()
				input.Spec.Routing = &AzureVirtualHubConnectionRouting{
					StaticVnetRoutes: []*AzureVirtualHubConnectionStaticVnetRoute{
						{Name: "to-nva", AddressPrefixes: []string{"10.50.0.0/16"}, NextHopIpAddress: "nva.local"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an undefined override criteria", func() {
				input := validResource()
				input.Spec.Routing = &AzureVirtualHubConnectionRouting{
					AssociatedRouteTableId:               literal(testHubId + "/hubRouteTables/defaultRouteTable"),
					StaticVnetLocalRouteOverrideCriteria: criteriaPtr(AzureVirtualHubConnectionStaticVnetLocalRouteOverrideCriteria(99)),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
