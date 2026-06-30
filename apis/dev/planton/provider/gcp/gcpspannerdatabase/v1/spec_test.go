package gcpspannerdatabasev1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpSpannerDatabaseSpec Suite")
}

var _ = ginkgo.Describe("GcpSpannerDatabaseSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpSpannerDatabase.
	minimal := func() *GcpSpannerDatabase {
		return &GcpSpannerDatabase{
			ApiVersion: "gcp.planton.dev/v1",
			Kind:       "GcpSpannerDatabase",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-spanner-db",
			},
			Spec: &GcpSpannerDatabaseSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-spanner-instance",
					},
				},
				DatabaseName: "my-database",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept GOOGLE_STANDARD_SQL dialect", func() {
		msg := minimal()
		msg.Spec.DatabaseDialect = "GOOGLE_STANDARD_SQL"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept POSTGRESQL dialect", func() {
		msg := minimal()
		msg.Spec.DatabaseDialect = "POSTGRESQL"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty dialect (defaults to GOOGLE_STANDARD_SQL)", func() {
		msg := minimal()
		msg.Spec.DatabaseDialect = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a database name with hyphens", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "my-app-db"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a database name with underscores", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "my_app_db"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 2-character database name (minimum)", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 30-character database name (maximum)", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "a" + strings.Repeat("b", 28) + "c"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a database name with mixed chars", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "app-db_v2"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept version_retention_period", func() {
		msg := minimal()
		msg.Spec.VersionRetentionPeriod = "7d"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept DDL statements", func() {
		msg := minimal()
		msg.Spec.Ddl = []string{
			"CREATE TABLE Users (UserId STRING(36) NOT NULL) PRIMARY KEY (UserId)",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept enable_drop_protection true", func() {
		msg := minimal()
		msg.Spec.EnableDropProtection = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept kms_key_name", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/my-proj/locations/us-central1/keyRings/ring/cryptoKeys/key",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept default_time_zone", func() {
		msg := minimal()
		msg.Spec.DefaultTimeZone = "UTC"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a full-featured spec", func() {
		msg := minimal()
		msg.Spec.DatabaseDialect = "GOOGLE_STANDARD_SQL"
		msg.Spec.VersionRetentionPeriod = "3d"
		msg.Spec.DefaultTimeZone = "America/New_York"
		msg.Spec.EnableDropProtection = true
		msg.Spec.Ddl = []string{
			"CREATE TABLE Accounts (Id STRING(36) NOT NULL) PRIMARY KEY (Id)",
			"CREATE INDEX AccountsById ON Accounts(Id)",
		}
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/my-proj/locations/us-central1/keyRings/ring/cryptoKeys/key",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing instance", func() {
		msg := minimal()
		msg.Spec.Instance = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing database_name", func() {
		msg := minimal()
		msg.Spec.DatabaseName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject database name starting with a digit", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "1database"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject database name starting with a hyphen", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "-database"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject database name ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "database-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject database name ending with an underscore", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "database_"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject database name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "MyDatabase"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject database name with spaces", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "my database"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject database name with dots", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "my.database"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject 1-character database name (below minimum)", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "a"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject 31-character database name (above maximum)", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "a" + strings.Repeat("b", 29) + "c"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid database_dialect", func() {
		msg := minimal()
		msg.Spec.DatabaseDialect = "MYSQL"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject another invalid database_dialect", func() {
		msg := minimal()
		msg.Spec.DatabaseDialect = "google_standard_sql"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
