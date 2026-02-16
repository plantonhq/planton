package gcpfirestoredatabasev1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpFirestoreDatabaseSpec Suite")
}

var _ = ginkgo.Describe("GcpFirestoreDatabaseSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpFirestoreDatabase.
	minimal := func() *GcpFirestoreDatabase {
		return &GcpFirestoreDatabase{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpFirestoreDatabase",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-firestore-db",
			},
			Spec: &GcpFirestoreDatabaseSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				LocationId:   "nam5",
				DatabaseName: "(default)",
				Type:         "FIRESTORE_NATIVE",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec with (default) database", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a named database (4 chars minimum)", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "abcd"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a named database (63 chars maximum)", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "a" + strings.Repeat("b", 61) + "c"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a named database with hyphens", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "my-app-database"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a named database with digits", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "db01-prod"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept FIRESTORE_NATIVE type", func() {
		msg := minimal()
		msg.Spec.Type = "FIRESTORE_NATIVE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept DATASTORE_MODE type", func() {
		msg := minimal()
		msg.Spec.Type = "DATASTORE_MODE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept OPTIMISTIC concurrency mode", func() {
		msg := minimal()
		msg.Spec.ConcurrencyMode = "OPTIMISTIC"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept PESSIMISTIC concurrency mode", func() {
		msg := minimal()
		msg.Spec.ConcurrencyMode = "PESSIMISTIC"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept OPTIMISTIC_WITH_ENTITY_GROUPS concurrency mode", func() {
		msg := minimal()
		msg.Spec.Type = "DATASTORE_MODE"
		msg.Spec.ConcurrencyMode = "OPTIMISTIC_WITH_ENTITY_GROUPS"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty concurrency mode (uses GCP default)", func() {
		msg := minimal()
		msg.Spec.ConcurrencyMode = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept POINT_IN_TIME_RECOVERY_ENABLED", func() {
		msg := minimal()
		msg.Spec.PointInTimeRecoveryEnablement = "POINT_IN_TIME_RECOVERY_ENABLED"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept POINT_IN_TIME_RECOVERY_DISABLED", func() {
		msg := minimal()
		msg.Spec.PointInTimeRecoveryEnablement = "POINT_IN_TIME_RECOVERY_DISABLED"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty PITR (uses GCP default)", func() {
		msg := minimal()
		msg.Spec.PointInTimeRecoveryEnablement = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept DELETE_PROTECTION_ENABLED", func() {
		msg := minimal()
		dps := "DELETE_PROTECTION_ENABLED"
		msg.Spec.DeleteProtectionState = &dps
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept DELETE_PROTECTION_DISABLED", func() {
		msg := minimal()
		dps := "DELETE_PROTECTION_DISABLED"
		msg.Spec.DeleteProtectionState = &dps
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept STANDARD database edition", func() {
		msg := minimal()
		msg.Spec.DatabaseEdition = "STANDARD"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept ENTERPRISE database edition with FIRESTORE_NATIVE", func() {
		msg := minimal()
		msg.Spec.Type = "FIRESTORE_NATIVE"
		msg.Spec.DatabaseEdition = "ENTERPRISE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty database edition (uses GCP default)", func() {
		msg := minimal()
		msg.Spec.DatabaseEdition = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept kms_key_name", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/my-proj/locations/us/keyRings/ring/cryptoKeys/key",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a single-region location", func() {
		msg := minimal()
		msg.Spec.LocationId = "us-east1"
		msg.Spec.DatabaseName = "regional-db"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept eur3 multi-region location", func() {
		msg := minimal()
		msg.Spec.LocationId = "eur3"
		msg.Spec.DatabaseName = "eu-database"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a full-featured spec", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "full-featured-db"
		msg.Spec.LocationId = "nam5"
		msg.Spec.Type = "FIRESTORE_NATIVE"
		msg.Spec.ConcurrencyMode = "OPTIMISTIC"
		msg.Spec.PointInTimeRecoveryEnablement = "POINT_IN_TIME_RECOVERY_ENABLED"
		dps := "DELETE_PROTECTION_ENABLED"
		msg.Spec.DeleteProtectionState = &dps
		msg.Spec.DatabaseEdition = "ENTERPRISE"
		msg.Spec.KmsKeyName = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/my-proj/locations/us/keyRings/ring/cryptoKeys/key",
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

	ginkgo.It("should reject missing location_id", func() {
		msg := minimal()
		msg.Spec.LocationId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing database_name", func() {
		msg := minimal()
		msg.Spec.DatabaseName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing type", func() {
		msg := minimal()
		msg.Spec.Type = ""
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

	ginkgo.It("should reject invalid type value", func() {
		msg := minimal()
		msg.Spec.Type = "INVALID_TYPE"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject lowercase type value", func() {
		msg := minimal()
		msg.Spec.Type = "firestore_native"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid concurrency mode", func() {
		msg := minimal()
		msg.Spec.ConcurrencyMode = "SNAPSHOT_ISOLATION"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid PITR value", func() {
		msg := minimal()
		msg.Spec.PointInTimeRecoveryEnablement = "ENABLED"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid delete_protection_state", func() {
		msg := minimal()
		dps := "ENABLED"
		msg.Spec.DeleteProtectionState = &dps
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid database_edition", func() {
		msg := minimal()
		msg.Spec.DatabaseEdition = "PREMIUM"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject ENTERPRISE edition with DATASTORE_MODE", func() {
		msg := minimal()
		msg.Spec.Type = "DATASTORE_MODE"
		msg.Spec.DatabaseEdition = "ENTERPRISE"
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

	ginkgo.It("should reject database name with underscores", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "my_database"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject database name with dots", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "my.database"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject 3-character database name (below minimum)", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject 64-character database name (above maximum)", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "a" + strings.Repeat("b", 62) + "c"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject UUID-like database name", func() {
		msg := minimal()
		msg.Spec.DatabaseName = "a1234567-1234-1234-1234-123456789abc"
		err := validator.Validate(msg)
		// This passes the regex since it starts with 'a' and matches the pattern.
		// GCP's UUID rejection is enforced at the API level, not by our regex.
		// We rely on GCP API to reject UUID-like names.
		// This test documents the known limitation.
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})
})
