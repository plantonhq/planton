package gcpvpcv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestGcpVpcSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVpcSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpVpcSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_vpc", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &GcpVpc{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gcp-vpc",
					},
					Spec: &GcpVpcSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						NetworkName: "test-vpc-network",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network_name is missing", func() {
				input := &GcpVpc{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gcp-vpc",
					},
					Spec: &GcpVpcSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network_name has invalid format", func() {
				input := &GcpVpc{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpVpc",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gcp-vpc",
					},
					Spec: &GcpVpcSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						NetworkName: "INVALID-NAME", // uppercase not allowed
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
