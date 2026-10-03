package gcpvertexaideployedindexv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiDeployedIndexSpec Suite")
}

var _ = ginkgo.Describe("GcpVertexAiDeployedIndexSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	strRef := func(val string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: val,
			},
		}
	}

	minimal := func() *GcpVertexAiDeployedIndex {
		return &GcpVertexAiDeployedIndex{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiDeployedIndex",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-deployed-index",
			},
			Spec: &GcpVertexAiDeployedIndexSpec{
				Location:        "us-central1",
				DeployedIndexId: "products_v1",
				Index:           strRef("projects/my-project/locations/us-central1/indexes/1234567890"),
				IndexEndpoint:   strRef("projects/my-project/locations/us-central1/indexEndpoints/9876543210"),
			},
		}
	}

	intPtr := func(v int32) *int32 { return &v }
	strPtr := func(v string) *string { return &v }

	// Suppress unused variable warnings.
	_ = intPtr
	_ = strPtr

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with display_name", func() {
		msg := minimal()
		msg.Spec.DisplayName = "Products v1 Deployment"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept deployed_index_id with underscores and digits", func() {
		msg := minimal()
		msg.Spec.DeployedIndexId = "a_1_b_2"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a single-letter deployed_index_id", func() {
		msg := minimal()
		msg.Spec.DeployedIndexId = "a"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept deployed_index_id at max length (128 chars)", func() {
		msg := minimal()
		msg.Spec.DeployedIndexId = "a" + strings.Repeat("b", 127)
		gomega.Expect(len(msg.Spec.DeployedIndexId)).To(gomega.Equal(128))
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept automatic resources with bounds", func() {
		msg := minimal()
		msg.Spec.AutomaticResources = &GcpVertexAiDeployedIndexAutomaticResources{
			MinReplicaCount: intPtr(2),
			MaxReplicaCount: 10,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty automatic resources (GCP defaults)", func() {
		msg := minimal()
		msg.Spec.AutomaticResources = &GcpVertexAiDeployedIndexAutomaticResources{}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept dedicated resources with machine type", func() {
		msg := minimal()
		msg.Spec.DedicatedResources = &GcpVertexAiDeployedIndexDedicatedResources{
			MachineType:     "e2-standard-16",
			MinReplicaCount: 2,
			MaxReplicaCount: 5,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept dedicated resources without machine type (API default)", func() {
		msg := minimal()
		msg.Spec.DedicatedResources = &GcpVertexAiDeployedIndexDedicatedResources{
			MinReplicaCount: 1,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept max_replica_count at the 1000 boundary", func() {
		msg := minimal()
		msg.Spec.DedicatedResources = &GcpVertexAiDeployedIndexDedicatedResources{
			MinReplicaCount: 2,
			MaxReplicaCount: 1000,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a deployment_group", func() {
		msg := minimal()
		msg.Spec.DeploymentGroup = strPtr("prod")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept enable_access_logging", func() {
		msg := minimal()
		msg.Spec.EnableAccessLogging = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept reserved_ip_ranges", func() {
		msg := minimal()
		msg.Spec.ReservedIpRanges = []*foreignkeyv1.StringValueOrRef{
			strRef("vertex-ai-range-a"),
			strRef("vertex-ai-range-b"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept auth config with issuers and audiences", func() {
		msg := minimal()
		msg.Spec.AuthConfig = &GcpVertexAiDeployedIndexAuthConfig{
			AllowedIssuers: []*foreignkeyv1.StringValueOrRef{
				strRef("query-sa@my-project.iam.gserviceaccount.com"),
			},
			Audiences: []string{"vector-search-clients"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept full-featured dedicated spec", func() {
		msg := minimal()
		msg.Spec.DisplayName = "Production Deployment"
		msg.Spec.DedicatedResources = &GcpVertexAiDeployedIndexDedicatedResources{
			MachineType:     "e2-highmem-16",
			MinReplicaCount: 3,
			MaxReplicaCount: 12,
		}
		msg.Spec.DeploymentGroup = strPtr("prod")
		msg.Spec.EnableAccessLogging = true
		msg.Spec.ReservedIpRanges = []*foreignkeyv1.StringValueOrRef{
			strRef("vertex-ai-range-a"),
		}
		msg.Spec.AuthConfig = &GcpVertexAiDeployedIndexAuthConfig{
			AllowedIssuers: []*foreignkeyv1.StringValueOrRef{
				strRef("query-sa@my-project.iam.gserviceaccount.com"),
			},
			Audiences: []string{"vector-search-clients"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject spec with missing location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with missing deployed_index_id", func() {
		msg := minimal()
		msg.Spec.DeployedIndexId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject deployed_index_id starting with a digit", func() {
		msg := minimal()
		msg.Spec.DeployedIndexId = "1products"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject deployed_index_id starting with an underscore", func() {
		msg := minimal()
		msg.Spec.DeployedIndexId = "_products"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject deployed_index_id with hyphens", func() {
		msg := minimal()
		msg.Spec.DeployedIndexId = "products-v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject deployed_index_id exceeding 128 chars", func() {
		msg := minimal()
		msg.Spec.DeployedIndexId = "a" + strings.Repeat("b", 128)
		gomega.Expect(len(msg.Spec.DeployedIndexId)).To(gomega.Equal(129))
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with missing index", func() {
		msg := minimal()
		msg.Spec.Index = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with missing index_endpoint", func() {
		msg := minimal()
		msg.Spec.IndexEndpoint = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject both sizing arms set together", func() {
		msg := minimal()
		msg.Spec.AutomaticResources = &GcpVertexAiDeployedIndexAutomaticResources{}
		msg.Spec.DedicatedResources = &GcpVertexAiDeployedIndexDedicatedResources{
			MinReplicaCount: 2,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject dedicated resources without min_replica_count", func() {
		msg := minimal()
		msg.Spec.DedicatedResources = &GcpVertexAiDeployedIndexDedicatedResources{
			MachineType: "e2-standard-16",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject automatic min_replica_count of 0", func() {
		msg := minimal()
		msg.Spec.AutomaticResources = &GcpVertexAiDeployedIndexAutomaticResources{
			MinReplicaCount: intPtr(0),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject max_replica_count above 1000", func() {
		msg := minimal()
		msg.Spec.AutomaticResources = &GcpVertexAiDeployedIndexAutomaticResources{
			MaxReplicaCount: 1001,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject dedicated max_replica_count above 1000", func() {
		msg := minimal()
		msg.Spec.DedicatedResources = &GcpVertexAiDeployedIndexDedicatedResources{
			MinReplicaCount: 2,
			MaxReplicaCount: 1001,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject display_name exceeding 128 chars", func() {
		msg := minimal()
		msg.Spec.DisplayName = strings.Repeat("d", 129)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject deployment_group exceeding 64 chars", func() {
		msg := minimal()
		msg.Spec.DeploymentGroup = strPtr(strings.Repeat("g", 65))
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with missing spec", func() {
		msg := &GcpVertexAiDeployedIndex{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiDeployedIndex",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-deployed-index",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Deletion policy ────────────────

	ginkgo.It("should accept all deletion_policy values", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("deletion_policy"))
	})
})
