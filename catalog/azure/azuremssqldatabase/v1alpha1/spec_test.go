package azuremssqldatabasev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMssqlDatabaseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMssqlDatabaseSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	serverId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Sql/servers/orders-sql"
	poolId   = serverId + "/elasticPools/tenant-pool"
	sourceId = serverId + "/databases/orders"
	uaiId    = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/db-uai"
	cmkKeyId = "https://vault.vault.azure.net/keys/db-tde/0123456789abcdef0123456789abcdef"
)

// minimal valid spec: a fresh general-purpose database.
func minimalSpec() *AzureMssqlDatabase {
	return &AzureMssqlDatabase{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMssqlDatabase",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-db",
		},
		Spec: &AzureMssqlDatabaseSpec{
			ServerId:     literal(serverId),
			DatabaseName: "orders",
			SkuName:      "GP_Gen5_2",
		},
	}
}

var _ = ginkgo.Describe("AzureMssqlDatabaseSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal vCore database", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept representative SKUs across every family", func() {
			for _, sku := range []string{"Basic", "S0", "S12", "P15", "GP_Gen5_2", "GP_S_Gen5_1", "BC_Gen5_4", "HS_Gen5_2", "HS_S_Gen5_2", "DW100c", "DS100", "Free"} {
				input := minimalSpec()
				input.Spec.SkuName = sku
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "sku %q must be accepted", sku)
			}
		})

		ginkgo.It("should accept an unset sku (Azure's serverless default)", func() {
			input := minimalSpec()
			input.Spec.SkuName = ""
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a pooled database with the ElasticPool sku", func() {
			input := minimalSpec()
			input.Spec.SkuName = "ElasticPool"
			input.Spec.ElasticPoolId = literal(poolId)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a fractional max size", func() {
			size := 0.5
			input := minimalSpec()
			input.Spec.SkuName = "Basic"
			input.Spec.MaxSizeGb = &size
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept serverless dials on a serverless sku", func() {
			pause := int32(60)
			minCap := 0.5
			input := minimalSpec()
			input.Spec.SkuName = "GP_S_Gen5_2"
			input.Spec.AutoPauseDelayInMinutes = &pause
			input.Spec.MinCapacity = &minCap
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept disabling auto-pause with -1", func() {
			pause := int32(-1)
			input := minimalSpec()
			input.Spec.SkuName = "GP_S_Gen5_2"
			input.Spec.AutoPauseDelayInMinutes = &pause
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept Hyperscale read replicas", func() {
			replicas := int32(2)
			input := minimalSpec()
			input.Spec.SkuName = "HS_Gen5_2"
			input.Spec.ReadReplicaCount = &replicas
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept availability and integrity dials", func() {
			input := minimalSpec()
			input.Spec.SkuName = "BC_Gen5_2"
			input.Spec.ReadScale = true
			input.Spec.ZoneRedundant = true
			input.Spec.LedgerEnabled = true
			input.Spec.EnclaveType = AzureMssqlDatabaseEnclaveType_VBS
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a copy of a source database", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMssqlDatabaseCreateMode_COPY
			input.Spec.CreationSourceDatabaseId = literal(sourceId)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a geo secondary with a secondary type", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMssqlDatabaseCreateMode_SECONDARY
			input.Spec.CreationSourceDatabaseId = literal(sourceId)
			input.Spec.SecondaryType = AzureMssqlDatabaseSecondaryType_GEO
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a point-in-time restore with source and timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMssqlDatabaseCreateMode_POINT_IN_TIME_RESTORE
			input.Spec.CreationSourceDatabaseId = literal(sourceId)
			input.Spec.RestorePointInTime = "2026-07-01T08:30:00Z"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a dropped-database restore", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMssqlDatabaseCreateMode_RESTORE
			input.Spec.RestoreDroppedDatabaseId = sourceId
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an LTR-backup restore", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMssqlDatabaseCreateMode_RESTORE_LONG_TERM_RETENTION_BACKUP
			input.Spec.RestoreLongTermRetentionBackupId = "/subscriptions/s/providers/Microsoft.Sql/locations/eastus/longTermRetentionServers/orders-sql/longTermRetentionDatabases/orders/longTermRetentionBackups/b1"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept database-scoped TDE CMK with an identity and rotation", func() {
			input := minimalSpec()
			input.Spec.UserAssignedIdentityIds = []*foreignkeyv1.StringValueOrRef{literal(uaiId)}
			input.Spec.TransparentDataEncryptionKeyVaultKeyId = literal(cmkKeyId)
			input.Spec.TransparentDataEncryptionKeyAutomaticRotationEnabled = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a bacpac import on a fresh database", func() {
			input := minimalSpec()
			input.Spec.Import = &AzureMssqlDatabaseImport{
				StorageUri:                 "https://account.blob.core.windows.net/bacpacs/app.bacpac",
				StorageKey:                 literal("key=="),
				StorageKeyType:             AzureMssqlDatabaseImportStorageKeyType_STORAGE_ACCESS_KEY,
				AdministratorLogin:         "sqladmin",
				AdministratorLoginPassword: literal("P@ssw0rd1234!"),
				AuthenticationType:         AzureMssqlDatabaseImportAuthenticationType_SQL,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept retention policies", func() {
			interval := int32(24)
			week := int32(26)
			input := minimalSpec()
			input.Spec.ShortTermRetentionPolicy = &AzureMssqlDatabaseShortTermRetentionPolicy{
				RetentionDays:         14,
				BackupIntervalInHours: &interval,
			}
			input.Spec.LongTermRetentionPolicy = &AzureMssqlDatabaseLongTermRetentionPolicy{
				WeeklyRetention:  "P5W",
				MonthlyRetention: "P12M",
				YearlyRetention:  "P5Y",
				WeekOfYear:       &week,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a threat-detection policy with storage export", func() {
			input := minimalSpec()
			input.Spec.ThreatDetectionPolicy = &AzureMssqlDatabaseThreatDetectionPolicy{
				State:                   AzureMssqlDatabaseThreatDetectionState_ENABLED,
				DisabledAlerts:          []string{"Sql_Injection"},
				EmailAddresses:          []string{"secops@contoso.com"},
				StorageEndpoint:         "https://export.blob.core.windows.net",
				StorageAccountAccessKey: literal("key=="),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept backup storage redundancy and DW geo backup", func() {
			off := false
			input := minimalSpec()
			input.Spec.SkuName = "DW100c"
			input.Spec.StorageAccountType = AzureMssqlDatabaseBackupStorageAccountType_ZONE_REDUNDANT
			input.Spec.GeoBackupEnabled = &off
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept valueFrom references for the server and pool", func() {
			input := minimalSpec()
			input.Spec.ServerId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureMssqlServer,
						Name:      "orders-sql",
						FieldPath: "status.outputs.server_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing server", func() {
			input := minimalSpec()
			input.Spec.ServerId = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a database name with illegal characters or endings", func() {
			for _, name := range []string{"bad:name", "bad/name", "bad.", "bad ", "bad*"} {
				input := minimalSpec()
				input.Spec.DatabaseName = name
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a malformed sku", func() {
			for _, sku := range []string{"GP_Gen5", "gp_gen5_2", "Standard", "S", "HS_X"} {
				input := minimalSpec()
				input.Spec.SkuName = sku
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil(), "sku %q must be rejected", sku)
			}
		})

		ginkgo.It("should reject a pool reference without the ElasticPool sku", func() {
			input := minimalSpec()
			input.Spec.ElasticPoolId = literal(poolId)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject the ElasticPool sku without a pool reference", func() {
			input := minimalSpec()
			input.Spec.SkuName = "ElasticPool"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a maintenance window on a pooled database", func() {
			input := minimalSpec()
			input.Spec.SkuName = "ElasticPool"
			input.Spec.ElasticPoolId = literal(poolId)
			input.Spec.MaintenanceConfigurationName = "SQL_EastUS_DB_1"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject serverless dials on a provisioned sku", func() {
			pause := int32(60)
			input := minimalSpec()
			input.Spec.AutoPauseDelayInMinutes = &pause
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an out-of-range auto-pause delay", func() {
			pause := int32(30)
			input := minimalSpec()
			input.Spec.SkuName = "GP_S_Gen5_2"
			input.Spec.AutoPauseDelayInMinutes = &pause
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject read replicas outside Hyperscale", func() {
			replicas := int32(2)
			input := minimalSpec()
			input.Spec.ReadReplicaCount = &replicas
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a copy without a source", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMssqlDatabaseCreateMode_COPY
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a source on a fresh database", func() {
			input := minimalSpec()
			input.Spec.CreationSourceDatabaseId = literal(sourceId)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a point-in-time restore without a timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMssqlDatabaseCreateMode_POINT_IN_TIME_RESTORE
			input.Spec.CreationSourceDatabaseId = literal(sourceId)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed restore timestamp", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMssqlDatabaseCreateMode_POINT_IN_TIME_RESTORE
			input.Spec.CreationSourceDatabaseId = literal(sourceId)
			input.Spec.RestorePointInTime = "yesterday"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a secondary type outside the secondary modes", func() {
			input := minimalSpec()
			input.Spec.SecondaryType = AzureMssqlDatabaseSecondaryType_GEO
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject recovery sources outside their create modes", func() {
			input := minimalSpec()
			input.Spec.RecoverDatabaseId = sourceId
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())

			input = minimalSpec()
			input.Spec.RestoreDroppedDatabaseId = sourceId
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())

			input = minimalSpec()
			input.Spec.RestoreLongTermRetentionBackupId = sourceId
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject TDE CMK without an attached identity", func() {
			input := minimalSpec()
			input.Spec.TransparentDataEncryptionKeyVaultKeyId = literal(cmkKeyId)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject CMK auto-rotation without a key", func() {
			input := minimalSpec()
			input.Spec.TransparentDataEncryptionKeyAutomaticRotationEnabled = true
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an import on a non-default create mode", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureMssqlDatabaseCreateMode_COPY
			input.Spec.CreationSourceDatabaseId = literal(sourceId)
			input.Spec.Import = &AzureMssqlDatabaseImport{
				StorageUri:                 "https://account.blob.core.windows.net/bacpacs/app.bacpac",
				StorageKey:                 literal("key=="),
				StorageKeyType:             AzureMssqlDatabaseImportStorageKeyType_STORAGE_ACCESS_KEY,
				AdministratorLogin:         "sqladmin",
				AdministratorLoginPassword: literal("P@ssw0rd1234!"),
				AuthenticationType:         AzureMssqlDatabaseImportAuthenticationType_SQL,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an import without a credential type", func() {
			input := minimalSpec()
			input.Spec.Import = &AzureMssqlDatabaseImport{
				StorageUri:                 "https://account.blob.core.windows.net/bacpacs/app.bacpac",
				StorageKey:                 literal("key=="),
				AdministratorLogin:         "sqladmin",
				AdministratorLoginPassword: literal("P@ssw0rd1234!"),
				AuthenticationType:         AzureMssqlDatabaseImportAuthenticationType_SQL,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject short-term retention outside 1-35 days", func() {
			input := minimalSpec()
			input.Spec.ShortTermRetentionPolicy = &AzureMssqlDatabaseShortTermRetentionPolicy{
				RetentionDays: 60,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a differential-backup interval other than 12 or 24", func() {
			interval := int32(6)
			input := minimalSpec()
			input.Spec.ShortTermRetentionPolicy = &AzureMssqlDatabaseShortTermRetentionPolicy{
				RetentionDays:         7,
				BackupIntervalInHours: &interval,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an empty long-term retention policy", func() {
			input := minimalSpec()
			input.Spec.LongTermRetentionPolicy = &AzureMssqlDatabaseLongTermRetentionPolicy{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed LTR duration", func() {
			input := minimalSpec()
			input.Spec.LongTermRetentionPolicy = &AzureMssqlDatabaseLongTermRetentionPolicy{
				WeeklyRetention: "5 weeks",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a threat-detection storage endpoint without its key", func() {
			input := minimalSpec()
			input.Spec.ThreatDetectionPolicy = &AzureMssqlDatabaseThreatDetectionPolicy{
				State:           AzureMssqlDatabaseThreatDetectionState_ENABLED,
				StorageEndpoint: "https://export.blob.core.windows.net",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a max size beyond 4096 GB", func() {
			size := 5000.0
			input := minimalSpec()
			input.Spec.MaxSizeGb = &size
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid sample name", func() {
			input := minimalSpec()
			input.Spec.SampleName = "Northwind"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
