package azureeventhubclusterv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventHubClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventHubClusterSpec Validation Tests")
}

// helper to create a minimal valid cluster
func minimalCluster() *AzureEventHubCluster {
	return &AzureEventHubCluster{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventHubCluster",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-eh-cluster",
		},
		Spec: &AzureEventHubClusterSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			ClusterName: "myapp-eventhub-cluster",
		},
	}
}

var _ = ginkgo.Describe("AzureEventHubClusterSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_event_hub_cluster", func() {

			ginkgo.It("should accept a minimal cluster", func() {
				err := protovalidate.Validate(minimalCluster())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept explicit capacity units and tags", func() {
				capacityUnits := int32(2)
				input := minimalCluster()
				input.Spec.CapacityUnits = &capacityUnits
				input.Spec.Tags = map[string]string{"team": "streaming", "cost-center": "cc-42"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_event_hub_cluster", func() {

			ginkgo.It("should reject a missing region", func() {
				input := minimalCluster()
				input.Spec.Region = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource_group", func() {
				input := minimalCluster()
				input.Spec.ResourceGroup = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing cluster_name", func() {
				input := minimalCluster()
				input.Spec.ClusterName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a cluster name longer than 50 characters", func() {
				input := minimalCluster()
				input.Spec.ClusterName = "a" + strings.Repeat("b", 50)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a cluster name with invalid characters", func() {
				input := minimalCluster()
				input.Spec.ClusterName = "bad!cluster*name"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a cluster name ending with a hyphen", func() {
				input := minimalCluster()
				input.Spec.ClusterName = "invalid-name-"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject zero capacity units", func() {
				capacityUnits := int32(0)
				input := minimalCluster()
				input.Spec.CapacityUnits = &capacityUnits
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
