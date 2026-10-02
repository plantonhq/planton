package gcpvectorsearchcollectionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVectorSearchCollectionSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpVectorSearchCollectionSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	denseField := func() *GcpVectorSearchCollectionVectorSchema {
		return &GcpVectorSearchCollectionVectorSchema{
			FieldName: "text_embedding",
			DenseVector: &GcpVectorSearchCollectionDenseVector{
				Dimensions: proto.Int32(768),
				VertexEmbeddingConfig: &GcpVectorSearchCollectionVertexEmbeddingConfig{
					ModelId:      "text-embedding-005",
					TaskType:     "RETRIEVAL_DOCUMENT",
					TextTemplate: "Title: {title} ---- Body: {body}",
				},
			},
		}
	}

	// One dense field with Vertex-computed embeddings and one index over it.
	minimal := func() *GcpVectorSearchCollection {
		return &GcpVectorSearchCollection{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVectorSearchCollection",
			Metadata:   &shared.CloudResourceMetadata{Name: "product-docs"},
			Spec: &GcpVectorSearchCollectionSpec{
				Location:      "us-central1",
				DataSchema:    `{"type":"object","properties":{"title":{"type":"string"},"body":{"type":"string"}}}`,
				VectorSchemas: []*GcpVectorSearchCollectionVectorSchema{denseField()},
				Indexes: []*GcpVectorSearchCollectionIndex{{
					IndexId:    "docs-ann",
					IndexField: "text_embedding",
				}},
			},
		}
	}

	ginkgo.It("should accept a dense collection with one index", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a bare collection with no vector fields and no indexes", func() {
		msg := minimal()
		msg.Spec.VectorSchemas = nil
		msg.Spec.Indexes = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every index lever, a sparse field, CMEK, and an explicit id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nameRef("ai-project")
		msg.Spec.CollectionId = "product-docs-v2"
		msg.Spec.DisplayName = "Product docs"
		msg.Spec.Description = "Search over the product documentation"
		msg.Spec.Labels = map[string]string{"team": "search"}
		msg.Spec.KmsKeyName = litRef("projects/ai-project/locations/us-central1/keyRings/ai/cryptoKeys/vs")
		msg.Spec.VectorSchemas = append(msg.Spec.VectorSchemas, &GcpVectorSearchCollectionVectorSchema{
			FieldName:    "lexical",
			SparseVector: true,
		})
		msg.Spec.Indexes[0].DisplayName = "Docs ANN"
		msg.Spec.Indexes[0].Description = "Cosine over normalized embeddings"
		msg.Spec.Indexes[0].Labels = map[string]string{"tier": "prod"}
		msg.Spec.Indexes[0].DistanceMetric = "COSINE_DISTANCE"
		msg.Spec.Indexes[0].FeatureNormType = "UNIT_L2_NORM"
		msg.Spec.Indexes[0].FilterFields = []string{"title"}
		msg.Spec.Indexes[0].StoreFields = []string{"title", "body"}
		msg.Spec.Indexes[0].DedicatedInfrastructure = &GcpVectorSearchCollectionDedicatedInfrastructure{
			Mode: "STORAGE_OPTIMIZED",
			AutoscalingSpec: &GcpVectorSearchCollectionAutoscalingSpec{
				MinReplicaCount: proto.Int32(2),
				MaxReplicaCount: proto.Int32(4),
			},
		}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a collection or index id that is not RFC 1035", func() {
		msg := minimal()
		msg.Spec.CollectionId = "Product_Docs"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Indexes[0].IndexId = "-docs"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a vector field to be exactly one of dense or sparse", func() {
		msg := minimal()
		msg.Spec.VectorSchemas[0].SparseVector = true
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.VectorSchemas[0].DenseVector = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require every embedding config field", func() {
		msg := minimal()
		msg.Spec.VectorSchemas[0].DenseVector.VertexEmbeddingConfig.TaskType = "EMBED"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.VectorSchemas[0].DenseVector.VertexEmbeddingConfig.TextTemplate = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require index_id and index_field on every index", func() {
		msg := minimal()
		msg.Spec.Indexes[0].IndexField = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Indexes[0].IndexId = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject unknown metric, norm, and mode values and out-of-range replicas", func() {
		msg := minimal()
		msg.Spec.Indexes[0].DistanceMetric = "EUCLIDEAN"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Indexes[0].FeatureNormType = "L1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Indexes[0].DedicatedInfrastructure = &GcpVectorSearchCollectionDedicatedInfrastructure{Mode: "FAST"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Indexes[0].DedicatedInfrastructure = &GcpVectorSearchCollectionDedicatedInfrastructure{
			AutoscalingSpec: &GcpVectorSearchCollectionAutoscalingSpec{MaxReplicaCount: proto.Int32(1001)},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
