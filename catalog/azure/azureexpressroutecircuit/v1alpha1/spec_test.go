package azureexpressroutecircuitv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureExpressRouteCircuitSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureExpressRouteCircuitSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// validResource returns a minimal valid service-provider-mode circuit
// that individual cases mutate into the shape under test.
func validResource() *AzureExpressRouteCircuit {
	return &AzureExpressRouteCircuit{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureExpressRouteCircuit",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-erc",
		},
		Spec: &AzureExpressRouteCircuitSpec{
			Region:              "eastus",
			ResourceGroup:       literal("test-rg"),
			Name:                "hq-circuit",
			SkuTier:             AzureExpressRouteCircuitSkuTier_STANDARD,
			SkuFamily:           AzureExpressRouteCircuitSkuFamily_METERED_DATA,
			ServiceProviderName: "Equinix",
			PeeringLocation:     "Washington DC",
			BandwidthInMbps:     50,
		},
	}
}

var _ = ginkgo.Describe("AzureExpressRouteCircuitSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_express_route_circuit", func() {

			ginkgo.It("should not return a validation error for a service-provider circuit", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ExpressRoute Direct circuit", func() {
				input := validResource()
				input.Spec.ServiceProviderName = ""
				input.Spec.PeeringLocation = ""
				input.Spec.BandwidthInMbps = 0
				input.Spec.ExpressRoutePortId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/expressRoutePorts/port1")
				input.Spec.BandwidthInGbps = 5
				input.Spec.RateLimitingEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a LOCAL unlimited circuit with authorizations", func() {
				input := validResource()
				input.Spec.SkuTier = AzureExpressRouteCircuitSkuTier_LOCAL
				input.Spec.SkuFamily = AzureExpressRouteCircuitSkuFamily_UNLIMITED_DATA
				input.Spec.Authorizations = []*AzureExpressRouteCircuitAuthorization{
					{Name: "partner-team"},
					{Name: "dr-site"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a redeemed authorization key reference", func() {
				input := validResource()
				input.Spec.AuthorizationKey = literal("8158df85-3d6b-4d9f-8a3c-247b63cab0a8")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_express_route_circuit", func() {

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

			ginkgo.It("should return a validation error when the SKU tier is unspecified", func() {
				input := validResource()
				input.Spec.SkuTier = AzureExpressRouteCircuitSkuTier_azure_express_route_circuit_sku_tier_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the SKU family is unspecified", func() {
				input := validResource()
				input.Spec.SkuFamily = AzureExpressRouteCircuitSkuFamily_azure_express_route_circuit_sku_family_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both provisioning modes are set", func() {
				input := validResource()
				input.Spec.ExpressRoutePortId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/expressRoutePorts/port1")
				input.Spec.BandwidthInGbps = 5
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither provisioning mode is set", func() {
				input := validResource()
				input.Spec.ServiceProviderName = ""
				input.Spec.PeeringLocation = ""
				input.Spec.BandwidthInMbps = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the provider trio is missing its location", func() {
				input := validResource()
				input.Spec.PeeringLocation = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the provider trio is missing its bandwidth", func() {
				input := validResource()
				input.Spec.BandwidthInMbps = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for gbps bandwidth on a provider circuit", func() {
				input := validResource()
				input.Spec.BandwidthInGbps = 5
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a Direct circuit without gbps bandwidth", func() {
				input := validResource()
				input.Spec.ServiceProviderName = ""
				input.Spec.PeeringLocation = ""
				input.Spec.BandwidthInMbps = 0
				input.Spec.ExpressRoutePortId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/expressRoutePorts/port1")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for provider fields on a Direct circuit", func() {
				input := validResource()
				input.Spec.ServiceProviderName = ""
				input.Spec.ExpressRoutePortId = literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/expressRoutePorts/port1")
				input.Spec.BandwidthInGbps = 5
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate authorization names", func() {
				input := validResource()
				input.Spec.Authorizations = []*AzureExpressRouteCircuitAuthorization{
					{Name: "partner-team"},
					{Name: "partner-team"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an authorization without a name", func() {
				input := validResource()
				input.Spec.Authorizations = []*AzureExpressRouteCircuitAuthorization{
					{Name: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
