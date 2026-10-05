package gcpfirestoreindexv1alpha1

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
	ginkgo.RunSpecs(t, "GcpFirestoreIndexSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpFirestoreIndexSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpFirestoreIndex {
		return &GcpFirestoreIndex{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpFirestoreIndex",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "orders-by-customer",
			},
			Spec: &GcpFirestoreIndexSpec{
				Collection: "orders",
				Fields: []*GcpFirestoreIndexField{
					{FieldPath: "customerId", Order: "ASCENDING"},
					{FieldPath: "createdAt", Order: "DESCENDING"},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal composite index", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an omitted project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an omitted database (defaults to project default)", func() {
		msg := minimal()
		msg.Spec.Database = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("my-gcp-project")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a database literal", func() {
		msg := minimal()
		msg.Spec.Database = litRef("analytics-db")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept COLLECTION_GROUP query_scope", func() {
		msg := minimal()
		msg.Spec.QueryScope = proto.String("COLLECTION_GROUP")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept COLLECTION_RECURSIVE query_scope", func() {
		msg := minimal()
		msg.Spec.QueryScope = proto.String("COLLECTION_RECURSIVE")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept DATASTORE_MODE_API api_scope", func() {
		msg := minimal()
		msg.Spec.ApiScope = proto.String("DATASTORE_MODE_API")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each valid density value", func() {
		for _, density := range []string{"SPARSE_ALL", "SPARSE_ANY", "DENSE"} {
			msg := minimal()
			msg.Spec.Density = density
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept an array-contains field", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{FieldPath: "tags", ArrayConfig: "CONTAINS"},
			{FieldPath: "createdAt", Order: "DESCENDING"},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a vector field last", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{FieldPath: "category", Order: "ASCENDING"},
			{
				FieldPath: "embedding",
				VectorConfig: &GcpFirestoreIndexVectorConfig{
					Dimension: 768,
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	// The MongoDB-compatible scope requires an explicit COLLECTION_GROUP
	// query scope (API-enforced; the COLLECTION default is rejected).
	ginkgo.It("should accept MONGODB_COMPATIBLE_API api_scope with COLLECTION_GROUP", func() {
		msg := minimal()
		msg.Spec.ApiScope = proto.String("MONGODB_COMPATIBLE_API")
		msg.Spec.QueryScope = proto.String("COLLECTION_GROUP")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a multikey index under the MongoDB-compatible scope", func() {
		msg := minimal()
		msg.Spec.ApiScope = proto.String("MONGODB_COMPATIBLE_API")
		msg.Spec.QueryScope = proto.String("COLLECTION_GROUP")
		msg.Spec.Multikey = true
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a unique index and skip_wait", func() {
		msg := minimal()
		msg.Spec.Unique = true
		msg.Spec.SkipWait = true
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each valid deletion_policy value", func() {
		for _, policy := range []string{"DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept a text search field", func() {
		msg := minimal()
		msg.Spec.ApiScope = proto.String("MONGODB_COMPATIBLE_API")
		msg.Spec.QueryScope = proto.String("COLLECTION_GROUP")
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{
				FieldPath: "description",
				SearchConfig: &GcpFirestoreIndexSearchConfig{
					TextSpec: &GcpFirestoreIndexTextSpec{
						IndexSpecs: []*GcpFirestoreIndexTextIndexSpec{
							{IndexType: "TOKENIZED", MatchType: "MATCH_GLOBALLY"},
						},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a geo search field under the default scope", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{
				FieldPath: "location",
				SearchConfig: &GcpFirestoreIndexSearchConfig{
					GeoSpec: &GcpFirestoreIndexGeoSpec{GeoJsonIndexingDisabled: true},
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject multikey outside the MongoDB-compatible scope", func() {
		msg := minimal()
		msg.Spec.Multikey = true
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("MONGODB_COMPATIBLE_API"))
	})

	ginkgo.It("should reject the MongoDB-compatible scope without COLLECTION_GROUP", func() {
		msg := minimal()
		msg.Spec.ApiScope = proto.String("MONGODB_COMPATIBLE_API")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("COLLECTION_GROUP"))
	})

	// The API rejects indexes mixing search and non-search fields.
	ginkgo.It("should reject mixed search and non-search fields", func() {
		msg := minimal()
		msg.Spec.ApiScope = proto.String("MONGODB_COMPATIBLE_API")
		msg.Spec.QueryScope = proto.String("COLLECTION_GROUP")
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{FieldPath: "category", Order: "ASCENDING"},
			{
				FieldPath: "title",
				SearchConfig: &GcpFirestoreIndexSearchConfig{
					TextSpec: &GcpFirestoreIndexTextSpec{
						IndexSpecs: []*GcpFirestoreIndexTextIndexSpec{
							{IndexType: "TOKENIZED", MatchType: "MATCH_GLOBALLY"},
						},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("search"))
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty search_config (neither text nor geo)", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{FieldPath: "description", SearchConfig: &GcpFirestoreIndexSearchConfig{}},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a text_spec without index_specs", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{
				FieldPath: "description",
				SearchConfig: &GcpFirestoreIndexSearchConfig{
					TextSpec: &GcpFirestoreIndexTextSpec{},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject search_config combined with order (two roles)", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{
				FieldPath: "description",
				Order:     "ASCENDING",
				SearchConfig: &GcpFirestoreIndexSearchConfig{
					GeoSpec: &GcpFirestoreIndexGeoSpec{},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one of order"))
	})

	ginkgo.It("should reject a missing collection", func() {
		msg := minimal()
		msg.Spec.Collection = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty fields list", func() {
		msg := minimal()
		msg.Spec.Fields = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a field with no role", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{FieldPath: "customerId"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one of order"))
	})

	ginkgo.It("should reject a field with multiple roles", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{FieldPath: "customerId", Order: "ASCENDING", ArrayConfig: "CONTAINS"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one of order"))
	})

	ginkgo.It("should reject an invalid order value", func() {
		msg := minimal()
		msg.Spec.Fields[0].Order = "RANDOM"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid array_config value", func() {
		msg := minimal()
		msg.Spec.Fields[0].ArrayConfig = "SORTED"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a vector_config without dimension", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{FieldPath: "embedding", VectorConfig: &GcpFirestoreIndexVectorConfig{}},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a zero vector dimension", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{FieldPath: "embedding", VectorConfig: &GcpFirestoreIndexVectorConfig{Dimension: 0}},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid query_scope", func() {
		msg := minimal()
		msg.Spec.QueryScope = proto.String("GLOBAL")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid api_scope", func() {
		msg := minimal()
		msg.Spec.ApiScope = proto.String("NATIVE_ONLY")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid density", func() {
		msg := minimal()
		msg.Spec.Density = "COMPACT"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing field_path", func() {
		msg := minimal()
		msg.Spec.Fields = []*GcpFirestoreIndexField{
			{Order: "ASCENDING"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		msg := minimal()
		msg.Kind = "GcpFirestoreIndexes"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
