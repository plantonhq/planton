package azureexpressroutegatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureExpressRouteGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureExpressRouteGatewaySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testHubId     = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus"
	testPeeringId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/expressRouteCircuits/dc-circuit/peerings/AzurePrivatePeering"
)

// validResource returns a minimal valid gateway that individual cases
// mutate into the shape under test.
func validResource() *AzureExpressRouteGateway {
	return &AzureExpressRouteGateway{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureExpressRouteGateway",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ergw",
		},
		Spec: &AzureExpressRouteGatewaySpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "hub-er-gateway",
			VirtualHubId:  literal(testHubId),
			ScaleUnits:    1,
		},
	}
}

// validConnection returns a connection joining the test peering.
func validConnection(name string) *AzureExpressRouteGatewayConnection {
	return &AzureExpressRouteGatewayConnection{
		Name:                         name,
		ExpressRouteCircuitPeeringId: literal(testPeeringId),
	}
}

var _ = ginkgo.Describe("AzureExpressRouteGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_express_route_gateway", func() {

			ginkgo.It("should not return a validation error for a minimal gateway", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the maximum scale units", func() {
				input := validResource()
				input.Spec.ScaleUnits = 10
				input.Spec.AllowNonVirtualWanTraffic = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a connection to a circuit peering", func() {
				input := validResource()
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{validConnection("dc-primary")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a cross-subscription connection with an authorization key", func() {
				input := validResource()
				connection := validConnection("partner-dc")
				connection.AuthorizationKey = "12345678-1234-1234-1234-123456789abc"
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{connection}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a managed-secret reference in the authorization key", func() {
				// Sensitive fields carry secret references on consuming
				// platforms -- the UUID shape is taught in the field comment,
				// never enforced on the stored value.
				input := validResource()
				connection := validConnection("partner-dc")
				connection.AuthorizationKey = "${secrets-group/partner-circuit/authorization-key}"
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{connection}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fully-tuned connection", func() {
				input := validResource()
				connection := validConnection("dc-primary")
				connection.InternetSecurityEnabled = true
				connection.ExpressRouteGatewayBypassEnabled = true
				connection.RoutingWeight = 32000
				connection.Routing = &AzureExpressRouteGatewayConnectionRouting{
					AssociatedRouteTableId: literal(testHubId + "/hubRouteTables/on-prem"),
					PropagatedRouteTable: &AzureExpressRouteGatewayConnectionPropagatedRouteTable{
						Labels: []string{"default"},
					},
				}
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{connection}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_express_route_gateway", func() {

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

			ginkgo.It("should return a validation error when virtual_hub_id is missing", func() {
				input := validResource()
				input.Spec.VirtualHubId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when scale_units is missing", func() {
				input := validResource()
				input.Spec.ScaleUnits = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for scale_units above ARM's cap of 10", func() {
				input := validResource()
				input.Spec.ScaleUnits = 11
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate connection names", func() {
				input := validResource()
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{
					validConnection("dc"),
					validConnection("dc"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a connection name starting with a period", func() {
				input := validResource()
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{validConnection(".dc")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a connection without a circuit peering", func() {
				input := validResource()
				connection := validConnection("dc")
				connection.ExpressRouteCircuitPeeringId = nil
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{connection}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a routing weight above 32000", func() {
				input := validResource()
				connection := validConnection("dc")
				connection.RoutingWeight = 32001
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{connection}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a routing block that configures nothing", func() {
				input := validResource()
				connection := validConnection("dc")
				connection.Routing = &AzureExpressRouteGatewayConnectionRouting{}
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{connection}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for propagation with no targets", func() {
				input := validResource()
				connection := validConnection("dc")
				connection.Routing = &AzureExpressRouteGatewayConnectionRouting{
					PropagatedRouteTable: &AzureExpressRouteGatewayConnectionPropagatedRouteTable{},
				}
				input.Spec.Connections = []*AzureExpressRouteGatewayConnection{connection}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
