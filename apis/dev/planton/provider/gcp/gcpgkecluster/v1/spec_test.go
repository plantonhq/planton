package gcpgkeclusterv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestGcpGkeClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGkeClusterSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpGkeClusterSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_gke_cluster", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &GcpGkeCluster{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpGkeCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gke-cluster",
					},
					Spec: &GcpGkeClusterSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						NetworkSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Location: "us-central1",
						SubnetworkSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/regions/us-central1/subnetworks/test-subnet"},
						},
						ClusterSecondaryRangeName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "pods-range"},
						},
						ServicesSecondaryRangeName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "services-range"},
						},
						MasterIpv4CidrBlock: "10.0.0.0/28",
						ClusterName:         "test-gke-cluster",
						RouterNatName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-nat-gateway"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
