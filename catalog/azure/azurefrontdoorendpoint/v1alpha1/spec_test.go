package azurefrontdoorendpointv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureFrontDoorEndpointSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorEndpointSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const profileId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd"

// minimal valid spec: an enabled endpoint on a referenced profile.
func minimalSpec() *AzureFrontDoorEndpoint {
	return &AzureFrontDoorEndpoint{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorEndpoint",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-front-door-endpoint",
		},
		Spec: &AzureFrontDoorEndpointSpec{
			ProfileId:    literal(profileId),
			EndpointName: "test-endpoint",
		},
	}
}

var _ = ginkgo.Describe("AzureFrontDoorEndpointSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal endpoint", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a profile reference via valueFrom", func() {
			input := minimalSpec()
			input.Spec.ProfileId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-profile"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept endpoint name boundaries (2 and 46 characters)", func() {
			input := minimalSpec()
			input.Spec.EndpointName = "ab"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.EndpointName = "a" + strings.Repeat("b", 44) + "c"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a disabled endpoint", func() {
			input := minimalSpec()
			input.Spec.Enabled = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept user tags", func() {
			input := minimalSpec()
			input.Spec.Tags = map[string]string{"cost-center": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing profile reference", func() {
			input := minimalSpec()
			input.Spec.ProfileId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing endpoint name", func() {
			input := minimalSpec()
			input.Spec.EndpointName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a single-character endpoint name", func() {
			input := minimalSpec()
			input.Spec.EndpointName = "a"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an endpoint name over 46 characters", func() {
			input := minimalSpec()
			input.Spec.EndpointName = "a" + strings.Repeat("b", 45) + "c"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an endpoint name with a leading hyphen", func() {
			input := minimalSpec()
			input.Spec.EndpointName = "-endpoint"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an endpoint name with a trailing hyphen", func() {
			input := minimalSpec()
			input.Spec.EndpointName = "endpoint-"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an endpoint name with invalid characters", func() {
			input := minimalSpec()
			input.Spec.EndpointName = "my.endpoint"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong api version", func() {
			input := minimalSpec()
			input.ApiVersion = "azure.planton.dev/v2"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong kind", func() {
			input := minimalSpec()
			input.Kind = "AzureFrontDoorEndpoints"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing metadata", func() {
			input := minimalSpec()
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
