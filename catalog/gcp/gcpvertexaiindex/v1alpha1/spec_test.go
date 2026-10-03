package gcpvertexaiindexv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiIndexSpec Suite")
}

var _ = ginkgo.Describe("GcpVertexAiIndexSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpVertexAiIndex {
		return &GcpVertexAiIndex{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiIndex",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-index",
			},
			Spec: &GcpVertexAiIndexSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Location:    "us-central1",
				DisplayName: "My Vector Index",
				Config: &GcpVertexAiIndexConfig{
					Dimensions: 768,
				},
			},
		}
	}

	intPtr := func(v int32) *int32 { return &v }
	strPtr := func(v string) *string { return &v }

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with description", func() {
		msg := minimal()
		msg.Spec.Description = "Embeddings index for product search"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec without project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept BATCH_UPDATE index_update_method", func() {
		msg := minimal()
		msg.Spec.IndexUpdateMethod = strPtr("BATCH_UPDATE")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept STREAM_UPDATE index_update_method", func() {
		msg := minimal()
		msg.Spec.IndexUpdateMethod = strPtr("STREAM_UPDATE")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty index_update_method (GCP default)", func() {
		msg := minimal()
		msg.Spec.IndexUpdateMethod = strPtr("")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a gs:// contents_delta_uri", func() {
		msg := minimal()
		msg.Spec.ContentsDeltaUri = "gs://my-bucket/embeddings/"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept contents_delta_uri with is_complete_overwrite", func() {
		msg := minimal()
		msg.Spec.ContentsDeltaUri = "gs://my-bucket/embeddings/"
		msg.Spec.IsCompleteOverwrite = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept tree-AH algorithm with approximate_neighbors_count", func() {
		msg := minimal()
		msg.Spec.Config.ApproximateNeighborsCount = 150
		msg.Spec.Config.TreeAhConfig = &GcpVertexAiIndexTreeAhConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept tree-AH with explicit leaf tuning", func() {
		msg := minimal()
		msg.Spec.Config.ApproximateNeighborsCount = 150
		msg.Spec.Config.TreeAhConfig = &GcpVertexAiIndexTreeAhConfig{
			LeafNodeEmbeddingCount:   intPtr(500),
			LeafNodesToSearchPercent: intPtr(25),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept leaf_nodes_to_search_percent at the 1 boundary", func() {
		msg := minimal()
		msg.Spec.Config.ApproximateNeighborsCount = 100
		msg.Spec.Config.TreeAhConfig = &GcpVertexAiIndexTreeAhConfig{
			LeafNodesToSearchPercent: intPtr(1),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept leaf_nodes_to_search_percent at the 100 boundary", func() {
		msg := minimal()
		msg.Spec.Config.ApproximateNeighborsCount = 100
		msg.Spec.Config.TreeAhConfig = &GcpVertexAiIndexTreeAhConfig{
			LeafNodesToSearchPercent: intPtr(100),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept brute-force algorithm without approximate_neighbors_count", func() {
		msg := minimal()
		msg.Spec.Config.BruteForceConfig = &GcpVertexAiIndexBruteForceConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every valid shard_size", func() {
		for _, size := range []string{"SHARD_SIZE_SMALL", "SHARD_SIZE_MEDIUM", "SHARD_SIZE_LARGE"} {
			msg := minimal()
			msg.Spec.Config.ShardSize = size
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "shard_size %s", size)
		}
	})

	ginkgo.It("should accept every valid distance_measure_type", func() {
		for _, dm := range []string{"SQUARED_L2_DISTANCE", "L1_DISTANCE", "COSINE_DISTANCE", "DOT_PRODUCT_DISTANCE"} {
			msg := minimal()
			msg.Spec.Config.DistanceMeasureType = strPtr(dm)
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "distance_measure_type %s", dm)
		}
	})

	ginkgo.It("should accept every valid feature_norm_type", func() {
		for _, fn := range []string{"UNIT_L2_NORM", "NONE"} {
			msg := minimal()
			msg.Spec.Config.FeatureNormType = strPtr(fn)
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "feature_norm_type %s", fn)
		}
	})

	ginkgo.It("should accept dimensions of 1 (boundary)", func() {
		msg := minimal()
		msg.Spec.Config.Dimensions = 1
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with user labels", func() {
		msg := minimal()
		msg.Spec.Labels = map[string]string{"team": "ml-platform", "cost-center": "research"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with display_name at max length (128 chars)", func() {
		msg := minimal()
		msg.Spec.DisplayName = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"abcdefghijklmnopqrstuvwxyz" +
			"0123456789012345678901234567890123456789" +
			"abcdefghijklmnopqrstuvwxyz0123456789"
		gomega.Expect(len(msg.Spec.DisplayName)).To(gomega.Equal(128))
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept full-featured streaming spec", func() {
		msg := minimal()
		msg.Spec.Description = "Production streaming vector index"
		msg.Spec.IndexUpdateMethod = strPtr("STREAM_UPDATE")
		msg.Spec.ContentsDeltaUri = "gs://prod-embeddings/products/"
		msg.Spec.Config = &GcpVertexAiIndexConfig{
			Dimensions:                1536,
			ApproximateNeighborsCount: 150,
			ShardSize:                 "SHARD_SIZE_MEDIUM",
			DistanceMeasureType:       strPtr("COSINE_DISTANCE"),
			FeatureNormType:           strPtr("UNIT_L2_NORM"),
			TreeAhConfig: &GcpVertexAiIndexTreeAhConfig{
				LeafNodeEmbeddingCount:   intPtr(1000),
				LeafNodesToSearchPercent: intPtr(10),
			},
		}
		msg.Spec.Labels = map[string]string{"env": "prod"}
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

	ginkgo.It("should reject spec with missing display_name", func() {
		msg := minimal()
		msg.Spec.DisplayName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with display_name exceeding 128 chars", func() {
		msg := minimal()
		msg.Spec.DisplayName = "A very long display name that exceeds the maximum allowed length of one hundred and twenty-eight characters and should be rejected by validation"
		gomega.Expect(len(msg.Spec.DisplayName)).To(gomega.BeNumerically(">", 128))
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject spec with missing config", func() {
		msg := minimal()
		msg.Spec.Config = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject config with missing dimensions", func() {
		msg := minimal()
		msg.Spec.Config.Dimensions = 0
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject config with negative dimensions", func() {
		msg := minimal()
		msg.Spec.Config.Dimensions = -5
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid index_update_method", func() {
		msg := minimal()
		msg.Spec.IndexUpdateMethod = strPtr("REALTIME_UPDATE")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject lowercase index_update_method", func() {
		msg := minimal()
		msg.Spec.IndexUpdateMethod = strPtr("batch_update")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject contents_delta_uri without gs:// scheme", func() {
		msg := minimal()
		msg.Spec.ContentsDeltaUri = "s3://my-bucket/embeddings/"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject contents_delta_uri that is a bare path", func() {
		msg := minimal()
		msg.Spec.ContentsDeltaUri = "my-bucket/embeddings/"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid shard_size", func() {
		msg := minimal()
		msg.Spec.Config.ShardSize = "SHARD_SIZE_XL"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid distance_measure_type", func() {
		msg := minimal()
		msg.Spec.Config.DistanceMeasureType = strPtr("HAMMING_DISTANCE")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid feature_norm_type", func() {
		msg := minimal()
		msg.Spec.Config.FeatureNormType = strPtr("L1_NORM")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject both algorithm arms set together", func() {
		msg := minimal()
		msg.Spec.Config.ApproximateNeighborsCount = 100
		msg.Spec.Config.TreeAhConfig = &GcpVertexAiIndexTreeAhConfig{}
		msg.Spec.Config.BruteForceConfig = &GcpVertexAiIndexBruteForceConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject tree-AH without approximate_neighbors_count", func() {
		msg := minimal()
		msg.Spec.Config.TreeAhConfig = &GcpVertexAiIndexTreeAhConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject negative approximate_neighbors_count", func() {
		msg := minimal()
		msg.Spec.Config.ApproximateNeighborsCount = -1
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject zero leaf_node_embedding_count", func() {
		msg := minimal()
		msg.Spec.Config.ApproximateNeighborsCount = 100
		msg.Spec.Config.TreeAhConfig = &GcpVertexAiIndexTreeAhConfig{
			LeafNodeEmbeddingCount: intPtr(0),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject leaf_nodes_to_search_percent above 100", func() {
		msg := minimal()
		msg.Spec.Config.ApproximateNeighborsCount = 100
		msg.Spec.Config.TreeAhConfig = &GcpVertexAiIndexTreeAhConfig{
			LeafNodesToSearchPercent: intPtr(101),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject leaf_nodes_to_search_percent of 0", func() {
		msg := minimal()
		msg.Spec.Config.ApproximateNeighborsCount = 100
		msg.Spec.Config.TreeAhConfig = &GcpVertexAiIndexTreeAhConfig{
			LeafNodesToSearchPercent: intPtr(0),
		}
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
		msg := &GcpVertexAiIndex{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiIndex",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-index",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── CMEK and deletion policy ────────────────

	ginkgo.It("should accept a CMEK key", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/p/locations/us-central1/keyRings/kr/cryptoKeys/k",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

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
