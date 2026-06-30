package azurevpcv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

func TestAzureVpcSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVpcSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureVpcSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_vpc", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureVpc{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-vpc",
					},
					Spec: &AzureVpcSpec{
						Region:           "eastus",
						ResourceGroup:    stringRef("test-rg"),
						AddressSpaceCidr: "10.0.0.0/16",
						NodesSubnetCidr:  "10.0.0.0/18",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
