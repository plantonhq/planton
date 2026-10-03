package gcpbigtabletablev1alpha1

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
	ginkgo.RunSpecs(t, "GcpBigtableTableSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpBigtableTableSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpBigtableTable {
		return &GcpBigtableTable{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBigtableTable",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "events",
			},
			Spec: &GcpBigtableTableSpec{
				Instance: litRef("my-bt-instance"),
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal spec (instance only, name from metadata)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit table_name with dots and underscores", func() {
		target := minimal()
		target.Spec.TableName = "events_raw.v2"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a column family without a GC policy", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an aggregate column family type", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "counters", Type: "intsum"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a max_age-only GC policy", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1", GcPolicy: &GcpBigtableTableGcPolicy{MaxAge: "720h"}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a max_versions-only GC policy", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1", GcPolicy: &GcpBigtableTableGcPolicy{MaxVersions: 3}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a UNION policy combining age and versions", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1", GcPolicy: &GcpBigtableTableGcPolicy{
				Mode:        "UNION",
				MaxAge:      "8760h",
				MaxVersions: 5,
			}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a raw gc_rules JSON tree", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1", GcPolicy: &GcpBigtableTableGcPolicy{
				GcRules: `{"mode":"union","rules":[{"max_age":"720h"},{"max_version":2}]}`,
			}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept split keys", func() {
		target := minimal()
		target.Spec.SplitKeys = []string{"user1", "user5", "user9"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a change stream retention", func() {
		target := minimal()
		target.Spec.ChangeStreamRetention = "24h0m0s"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept 0 to disable change streams", func() {
		target := minimal()
		target.Spec.ChangeStreamRetention = "0"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an automated backup policy", func() {
		target := minimal()
		target.Spec.AutomatedBackupPolicy = &GcpBigtableTableAutomatedBackupPolicy{
			RetentionPeriod: "72h",
			Frequency:       "24h",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept explicit UNPROTECTED deletion protection", func() {
		target := minimal()
		unprotected := "UNPROTECTED"
		target.Spec.DeletionProtection = &unprotected
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a row key schema", func() {
		target := minimal()
		target.Spec.RowKeySchema = `{"structType":{"fields":[{"fieldName":"user_id","type":{"stringType":{"encoding":{"utf8Bytes":{}}}}}],"encoding":{"delimitedBytes":{"delimiter":"Iw=="}}}}`
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a reference-shaped instance", func() {
		target := minimal()
		target.Spec.Instance = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-bt-instance"},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing instance", func() {
		target := minimal()
		target.Spec.Instance = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a table_name over 50 characters", func() {
		target := minimal()
		target.Spec.TableName = "a123456789a123456789a123456789a123456789a1234567891"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a table_name with invalid characters", func() {
		target := minimal()
		target.Spec.TableName = "events table"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a column family without a name", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{{}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject gc_rules combined with typed GC fields", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1", GcPolicy: &GcpBigtableTableGcPolicy{
				GcRules: `{"rules":[{"max_age":"720h"}]}`,
				MaxAge:  "720h",
			}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("mutually exclusive"))
	})

	ginkgo.It("should reject mode without both conditions", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1", GcPolicy: &GcpBigtableTableGcPolicy{
				Mode:   "UNION",
				MaxAge: "720h",
			}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject both conditions without a mode", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1", GcPolicy: &GcpBigtableTableGcPolicy{
				MaxAge:      "720h",
				MaxVersions: 3,
			}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid GC mode", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1", GcPolicy: &GcpBigtableTableGcPolicy{
				Mode:        "EITHER",
				MaxAge:      "720h",
				MaxVersions: 3,
			}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a malformed max_age duration", func() {
		target := minimal()
		target.Spec.ColumnFamilies = []*GcpBigtableTableColumnFamily{
			{Family: "cf1", GcPolicy: &GcpBigtableTableGcPolicy{MaxAge: "30 days"}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a malformed change_stream_retention", func() {
		target := minimal()
		target.Spec.ChangeStreamRetention = "1 day"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an automated backup policy without retention", func() {
		target := minimal()
		target.Spec.AutomatedBackupPolicy = &GcpBigtableTableAutomatedBackupPolicy{
			Frequency: "24h",
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid deletion_protection value", func() {
		target := minimal()
		invalid := "DISABLED"
		target.Spec.DeletionProtection = &invalid
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing spec", func() {
		target := minimal()
		target.Spec = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing metadata", func() {
		target := minimal()
		target.Metadata = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		target := minimal()
		target.Kind = "GcpBigTableTable"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong api_version literal", func() {
		target := minimal()
		target.ApiVersion = "gcp.planton.dev/v2"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept automated backup locations", func() {
		target := minimal()
		target.Spec.AutomatedBackupPolicy = &GcpBigtableTableAutomatedBackupPolicy{
			RetentionPeriod: "72h",
			Frequency:       "24h",
			Locations:       []string{"projects/my-project/locations/us-central1-a"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept deletion_policy DELETE, PREVENT, ABANDON, and empty", func() {
		for _, policy := range []string{"DELETE", "PREVENT", "ABANDON", ""} {
			target := minimal()
			target.Spec.DeletionPolicy = policy
			err := validator.Validate(target)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "policy %q", policy)
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "RETAIN"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
