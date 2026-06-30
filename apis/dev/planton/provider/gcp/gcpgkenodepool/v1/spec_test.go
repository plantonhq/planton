package gcpgkenodepoolv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestGcpGkeNodePoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGkeNodePoolSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpGkeNodePoolSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_gke_node_pool", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &GcpGkeNodePool{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpGkeNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-node-pool",
					},
					Spec: &GcpGkeNodePoolSpec{
						ClusterProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-gke-cluster"},
						},
						ClusterLocation: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "us-central1"},
						},
						NodePoolName: "test-node-pool",
						DiskType:     proto.String("pd-standard"), // Valid disk type
						NodePoolSize: &GcpGkeNodePoolSpec_NodeCount{
							NodeCount: 3, // Required node pool size
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
