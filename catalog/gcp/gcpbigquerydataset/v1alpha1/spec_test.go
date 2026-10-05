package gcpbigquerydatasetv1alpha1

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
	ginkgo.RunSpecs(t, "GcpBigQueryDatasetSpec Suite")
}

var _ = ginkgo.Describe("GcpBigQueryDatasetSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpBigQueryDataset.
	minimal := func() *GcpBigQueryDataset {
		return &GcpBigQueryDataset{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBigQueryDataset",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-bq-dataset",
			},
			Spec: &GcpBigQueryDatasetSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				DatasetId: "analytics_prod",
				Location:  "US",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec (project_id + dataset_id + location)", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept dataset_id with underscores and numbers", func() {
		msg := minimal()
		msg.Spec.DatasetId = "raw_events_2024_Q1"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept dataset_id with uppercase letters", func() {
		msg := minimal()
		msg.Spec.DatasetId = "MyDataset"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all optional string fields", func() {
		msg := minimal()
		msg.Spec.FriendlyName = "Production Analytics"
		msg.Spec.Description = "Main analytics dataset for production workloads"
		msg.Spec.DefaultCollation = "und:ci"
		msg.Spec.StorageBillingModel = "PHYSICAL"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept storage_billing_model LOGICAL", func() {
		msg := minimal()
		msg.Spec.StorageBillingModel = "LOGICAL"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept storage_billing_model PHYSICAL", func() {
		msg := minimal()
		msg.Spec.StorageBillingModel = "PHYSICAL"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept default_table_expiration_ms at minimum (3600000)", func() {
		msg := minimal()
		msg.Spec.DefaultTableExpirationMs = 3600000
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept default_table_expiration_ms above minimum", func() {
		msg := minimal()
		msg.Spec.DefaultTableExpirationMs = 86400000 // 1 day
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept max_time_travel_hours at minimum (48)", func() {
		msg := minimal()
		msg.Spec.MaxTimeTravelHours = 48
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept max_time_travel_hours at maximum (168)", func() {
		msg := minimal()
		msg.Spec.MaxTimeTravelHours = 168
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept max_time_travel_hours at midpoint (96)", func() {
		msg := minimal()
		msg.Spec.MaxTimeTravelHours = 96
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept is_case_insensitive set to true", func() {
		msg := minimal()
		msg.Spec.IsCaseInsensitive = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept delete_contents_on_destroy set to true", func() {
		msg := minimal()
		msg.Spec.DeleteContentsOnDestroy = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with CMEK encryption", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with access entries (user by email)", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{Role: "OWNER", UserByEmail: "admin@example.com"},
			{Role: "READER", GroupByEmail: "analysts@example.com"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with access entry for special group", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{Role: "OWNER", SpecialGroup: "projectOwners"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with view-based access (no role)", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{
				View: &GcpBigQueryDatasetAccessView{
					ProjectId: "my-gcp-project",
					DatasetId: "shared_views",
					TableId:   "revenue_summary",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept regional location", func() {
		msg := minimal()
		msg.Spec.Location = "us-central1"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an omitted project_id (ambient project contract)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept user labels and resource tags", func() {
		msg := minimal()
		msg.Spec.Labels = map[string]string{"team": "analytics", "cost-center": "cc-1234"}
		msg.Spec.ResourceTags = map[string]string{"123456789012/environment": "production"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an iam_member access entry", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{Role: "READER", IamMember: "serviceAccount:etl@my-gcp-project.iam.gserviceaccount.com"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a domain access entry", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{Role: "READER", Domain: "example.com"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept routine-based access (no role)", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{
				Routine: &GcpBigQueryDatasetAccessRoutine{
					ProjectId: "my-gcp-project",
					DatasetId: "shared_udfs",
					RoutineId: "anonymize",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept authorized-dataset access (no role)", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{
				Dataset: &GcpBigQueryDatasetAccessDataset{
					ProjectId:   "my-gcp-project",
					DatasetId:   "reporting_views",
					TargetTypes: []string{"VIEWS"},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a principal grant gated by an IAM condition", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{
				Role:        "READER",
				UserByEmail: "contractor@example.com",
				Condition: &GcpBigQueryDatasetAccessCondition{
					Expression: `request.time < timestamp("2030-01-01T00:00:00Z")`,
					Title:      "expires-2030",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an external dataset reference", func() {
		msg := minimal()
		msg.Spec.ExternalDatasetReference = &GcpBigQueryDatasetExternalDatasetReference{
			ExternalSource: "aws-glue://arn:aws:glue:us-east-1:123456789012:database/sales",
			Connection:     "projects/my-gcp-project/locations/aws-us-east-1/connections/glue-conn",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept external catalog options", func() {
		msg := minimal()
		msg.Spec.ExternalCatalogOptions = &GcpBigQueryDatasetExternalCatalogOptions{
			DefaultStorageLocationUri: "gs://my-lakehouse/warehouse",
			Parameters:                map[string]string{"hive.metastore.database.owner": "spark"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject when dataset_id is empty", func() {
		msg := minimal()
		msg.Spec.DatasetId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when location is empty", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject dataset_id with hyphens", func() {
		msg := minimal()
		msg.Spec.DatasetId = "invalid-dataset-name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject dataset_id with spaces", func() {
		msg := minimal()
		msg.Spec.DatasetId = "invalid dataset"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject dataset_id with dots", func() {
		msg := minimal()
		msg.Spec.DatasetId = "invalid.dataset"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject dataset_id exceeding 1024 characters", func() {
		msg := minimal()
		msg.Spec.DatasetId = strings.Repeat("a", 1025)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject default_table_expiration_ms below minimum (non-zero)", func() {
		msg := minimal()
		msg.Spec.DefaultTableExpirationMs = 1000 // 1 second, below 1 hour minimum
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject max_time_travel_hours below range", func() {
		msg := minimal()
		msg.Spec.MaxTimeTravelHours = 47
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject max_time_travel_hours above range", func() {
		msg := minimal()
		msg.Spec.MaxTimeTravelHours = 169
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid storage_billing_model", func() {
		msg := minimal()
		msg.Spec.StorageBillingModel = "INVALID"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an access entry with no identity", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{Role: "READER"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an access entry with two identities", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{Role: "READER", UserByEmail: "a@example.com", GroupByEmail: "g@example.com"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a principal grant without a role", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{UserByEmail: "a@example.com"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a view authorization that also sets a role", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{
				Role: "READER",
				View: &GcpBigQueryDatasetAccessView{
					ProjectId: "p", DatasetId: "d", TableId: "t",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a routine authorization that also sets a role", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{
				Role: "READER",
				Routine: &GcpBigQueryDatasetAccessRoutine{
					ProjectId: "p", DatasetId: "d", RoutineId: "r",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an authorized dataset with empty target_types", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{
				Dataset: &GcpBigQueryDatasetAccessDataset{
					ProjectId: "p", DatasetId: "d", TargetTypes: []string{},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an authorized dataset with an invalid target type", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{
				Dataset: &GcpBigQueryDatasetAccessDataset{
					ProjectId: "p", DatasetId: "d", TargetTypes: []string{"TABLES"},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an external dataset reference missing the connection", func() {
		msg := minimal()
		msg.Spec.ExternalDatasetReference = &GcpBigQueryDatasetExternalDatasetReference{
			ExternalSource: "aws-glue://arn:aws:glue:us-east-1:123456789012:database/sales",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an IAM condition without an expression", func() {
		msg := minimal()
		msg.Spec.Access = []*GcpBigQueryDatasetAccessEntry{
			{
				Role:        "READER",
				UserByEmail: "a@example.com",
				Condition:   &GcpBigQueryDatasetAccessCondition{Title: "no-expression"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an external catalog storage URI over 1024 characters", func() {
		msg := minimal()
		msg.Spec.ExternalCatalogOptions = &GcpBigQueryDatasetExternalCatalogOptions{
			DefaultStorageLocationUri: "gs://" + strings.Repeat("a", 1024),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when metadata is missing", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when spec is missing", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = v
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
