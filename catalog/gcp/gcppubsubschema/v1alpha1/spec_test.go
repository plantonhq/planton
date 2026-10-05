package gcppubsubschemav1alpha1

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
	ginkgo.RunSpecs(t, "GcpPubSubSchemaSpec Suite")
}

var _ = ginkgo.Describe("GcpPubSubSchemaSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper for StringValueOrRef with a literal value.
	svr := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// Helper to build a minimal valid GcpPubSubSchema (AVRO arm).
	minimal := func() *GcpPubSubSchema {
		return &GcpPubSubSchema{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpPubSubSchema",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-pubsub-schema",
			},
			Spec: &GcpPubSubSchemaSpec{
				SchemaName: "order-events",
				Type:       "AVRO",
				Definition: `{"type":"record","name":"OrderEvent","fields":[{"name":"order_id","type":"string"}]}`,
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid AVRO schema", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a PROTOCOL_BUFFER schema", func() {
		msg := minimal()
		msg.Spec.Type = "PROTOCOL_BUFFER"
		msg.Spec.Definition = "syntax = \"proto3\";\nmessage OrderEvent {\n  string order_id = 1;\n}"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a project_id literal", func() {
		msg := minimal()
		msg.Spec.ProjectId = svr("my-gcp-project")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an omitted project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 3-character schema name (lower boundary)", func() {
		msg := minimal()
		msg.Spec.SchemaName = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 255-character schema name (upper boundary)", func() {
		msg := minimal()
		name := "a"
		for len(name) < 255 {
			name += "b"
		}
		msg.Spec.SchemaName = name
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept hyphens, underscores, periods, tildes, plus and percent in the name", func() {
		msg := minimal()
		msg.Spec.SchemaName = "evt-schema_v1.2~a+b%c"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a name containing goog when not a prefix", func() {
		msg := minimal()
		msg.Spec.SchemaName = "my-google-events"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind", func() {
		msg := minimal()
		msg.Kind = "GcpPubSubTopic"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing schema_name", func() {
		msg := minimal()
		msg.Spec.SchemaName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a 2-character schema name (below minimum)", func() {
		msg := minimal()
		msg.Spec.SchemaName = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a 256-character schema name (above maximum)", func() {
		msg := minimal()
		name := "a"
		for len(name) < 256 {
			name += "b"
		}
		msg.Spec.SchemaName = name
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a schema name starting with a digit", func() {
		msg := minimal()
		msg.Spec.SchemaName = "1events"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a schema name with spaces", func() {
		msg := minimal()
		msg.Spec.SchemaName = "order events"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a schema name with the reserved goog prefix", func() {
		msg := minimal()
		msg.Spec.SchemaName = "goog-events"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing type", func() {
		msg := minimal()
		msg.Spec.Type = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid type", func() {
		msg := minimal()
		msg.Spec.Type = "JSON_SCHEMA"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject the TYPE_UNSPECIFIED sentinel", func() {
		msg := minimal()
		msg.Spec.Type = "TYPE_UNSPECIFIED"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing definition", func() {
		msg := minimal()
		msg.Spec.Definition = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Deletion Policy ────────────────

	ginkgo.It("should accept each valid deletion_policy value", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "deletion_policy %q should be accepted", policy)
		}
	})

	ginkgo.It("should reject an invalid deletion_policy value", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
