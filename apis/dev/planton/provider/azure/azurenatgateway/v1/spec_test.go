package azurenatgatewayv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAzureNatGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureNatGatewaySpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureNatGatewaySpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_nat_gateway", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureNatGateway{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureNatGateway",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nat-gateway",
					},
					Spec: &AzureNatGatewaySpec{
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet"},
						},
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}
