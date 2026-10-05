package azurecosmosdbaccountv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureCosmosdbAccountSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCosmosdbAccountSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimal valid spec: a single-region SQL-API account with the required
// consistency block.
func minimalSpec() *AzureCosmosdbAccount {
	return &AzureCosmosdbAccount{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCosmosdbAccount",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-cosmos-account",
		},
		Spec: &AzureCosmosdbAccountSpec{
			Region:            "eastus",
			ResourceGroup:     literal("test-rg"),
			AccountName:       "planton-test-cosmos",
			ConsistencyPolicy: &AzureCosmosdbAccountConsistencyPolicy{},
			GeoLocations: []*AzureCosmosdbAccountGeoLocation{
				{Location: "eastus", FailoverPriority: 0},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureCosmosdbAccountSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal single-region account", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept both kinds", func() {
			for _, kind := range []AzureCosmosdbAccountKind{
				AzureCosmosdbAccountKind_GLOBAL_DOCUMENT_DB,
				AzureCosmosdbAccountKind_MONGO_DB,
			} {
				input := minimalSpec()
				input.Spec.Kind = kind
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "kind %v must be accepted", kind)
			}
		})

		ginkgo.It("should accept every consistency level with defaults", func() {
			for _, level := range []AzureCosmosdbAccountConsistencyLevel{
				AzureCosmosdbAccountConsistencyLevel_STRONG,
				AzureCosmosdbAccountConsistencyLevel_BOUNDED_STALENESS,
				AzureCosmosdbAccountConsistencyLevel_SESSION,
				AzureCosmosdbAccountConsistencyLevel_CONSISTENT_PREFIX,
				AzureCosmosdbAccountConsistencyLevel_EVENTUAL,
			} {
				input := minimalSpec()
				input.Spec.ConsistencyPolicy.ConsistencyLevel = level
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "level %v must be accepted", level)
			}
		})

		ginkgo.It("should accept multi-region BoundedStaleness with the explicit floors", func() {
			input := minimalSpec()
			input.Spec.ConsistencyPolicy.ConsistencyLevel = AzureCosmosdbAccountConsistencyLevel_BOUNDED_STALENESS
			input.Spec.ConsistencyPolicy.MaxStalenessPrefix = proto.Int32(100000)
			input.Spec.ConsistencyPolicy.MaxIntervalInSeconds = proto.Int32(300)
			input.Spec.GeoLocations = append(input.Spec.GeoLocations,
				&AzureCosmosdbAccountGeoLocation{Location: "westus2", FailoverPriority: 1})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a serverless SQL account", func() {
			input := minimalSpec()
			input.Spec.Capabilities = []AzureCosmosdbAccountCapability{
				AzureCosmosdbAccountCapability_ENABLE_SERVERLESS,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept Mongo capabilities on a MONGO_DB account", func() {
			input := minimalSpec()
			input.Spec.Kind = AzureCosmosdbAccountKind_MONGO_DB
			input.Spec.MongoServerVersion = AzureCosmosdbAccountMongoServerVersion_MONGO_7_0
			input.Spec.Capabilities = []AzureCosmosdbAccountCapability{
				AzureCosmosdbAccountCapability_ENABLE_MONGO,
				AzureCosmosdbAccountCapability_ENABLE_MONGO_RETRYABLE_WRITES,
				AzureCosmosdbAccountCapability_ENABLE_MONGO_ROLE_BASED_ACCESS_CONTROL,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept MONGO_DB_V34 alongside ENABLE_MONGO", func() {
			input := minimalSpec()
			input.Spec.Kind = AzureCosmosdbAccountKind_MONGO_DB
			input.Spec.Capabilities = []AzureCosmosdbAccountCapability{
				AzureCosmosdbAccountCapability_ENABLE_MONGO,
				AzureCosmosdbAccountCapability_MONGO_DB_V34,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept the extra-API capabilities on a SQL account", func() {
			input := minimalSpec()
			input.Spec.Capabilities = []AzureCosmosdbAccountCapability{
				AzureCosmosdbAccountCapability_ENABLE_CASSANDRA,
				AzureCosmosdbAccountCapability_ENABLE_NO_SQL_VECTOR_SEARCH,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a periodic backup with dials", func() {
			input := minimalSpec()
			input.Spec.Backup = &AzureCosmosdbAccountBackup{
				Type:              AzureCosmosdbAccountBackupType_PERIODIC,
				IntervalInMinutes: proto.Int32(240),
				RetentionInHours:  proto.Int32(24),
				StorageRedundancy: AzureCosmosdbAccountBackupStorageRedundancy_ZONE,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a continuous backup with a tier", func() {
			input := minimalSpec()
			input.Spec.Backup = &AzureCosmosdbAccountBackup{
				Type: AzureCosmosdbAccountBackupType_CONTINUOUS,
				Tier: AzureCosmosdbAccountContinuousTier_CONTINUOUS_7_DAYS,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a system-assigned identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureCosmosdbAccountIdentity{
				Type: AzureCosmosdbAccountIdentityType_SYSTEM_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a user-assigned identity with CMK and a user-assigned default identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureCosmosdbAccountIdentity{
				Type:        AzureCosmosdbAccountIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cmk-identity")},
			}
			input.Spec.DefaultIdentity = &AzureCosmosdbAccountDefaultIdentity{
				Type:                   AzureCosmosdbAccountDefaultIdentityType_USER_ASSIGNED_DEFAULT,
				UserAssignedIdentityId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/cmk-identity"),
			}
			input.Spec.KeyVaultKeyId = literal("https://my-vault.vault.azure.net/keys/cosmos-cmk")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a first-party default identity without an identity reference", func() {
			input := minimalSpec()
			input.Spec.DefaultIdentity = &AzureCosmosdbAccountDefaultIdentity{
				Type: AzureCosmosdbAccountDefaultIdentityType_FIRST_PARTY,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept create_mode RESTORE with continuous backup and a restore block", func() {
			input := minimalSpec()
			input.Spec.Backup = &AzureCosmosdbAccountBackup{Type: AzureCosmosdbAccountBackupType_CONTINUOUS}
			input.Spec.CreateMode = AzureCosmosdbAccountCreateMode_RESTORE
			input.Spec.Restore = &AzureCosmosdbAccountRestore{
				SourceCosmosdbAccountId: "/subscriptions/s/providers/Microsoft.DocumentDB/locations/eastus/restorableDatabaseAccounts/00000000-0000-0000-0000-000000000000",
				RestoreTimestampInUtc:   "2026-07-01T00:00:00Z",
				Databases: []*AzureCosmosdbAccountRestoreDatabase{
					{Name: "app-data", CollectionNames: []string{"orders"}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept create_mode DEFAULT on a continuous-backup account", func() {
			input := minimalSpec()
			input.Spec.Backup = &AzureCosmosdbAccountBackup{Type: AzureCosmosdbAccountBackupType_CONTINUOUS}
			input.Spec.CreateMode = AzureCosmosdbAccountCreateMode_DEFAULT
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a valid CORS rule", func() {
			input := minimalSpec()
			input.Spec.CorsRule = &AzureCosmosdbAccountCorsRule{
				AllowedOrigins:  []string{"https://app.example.com"},
				AllowedMethods:  []string{"GET", "POST"},
				AllowedHeaders:  []string{"*"},
				ExposedHeaders:  []string{"*"},
				MaxAgeInSeconds: proto.Int32(3600),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept IPv4 addresses and CIDR ranges in the IP filter", func() {
			input := minimalSpec()
			input.Spec.IpRangeFilter = []string{"104.42.195.92", "10.0.0.0/16", "0.0.0.0", "255.255.255.255/32"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept TLS 1.2, the only floor Azure still provisions", func() {
			input := minimalSpec()
			input.Spec.MinimalTlsVersion = AzureCosmosdbAccountMinimalTlsVersion_TLS_1_2
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject the retired TLS 1.0/1.1 floors (reserved enum slots)", func() {
			for _, retired := range []AzureCosmosdbAccountMinimalTlsVersion{1, 2} {
				input := minimalSpec()
				input.Spec.MinimalTlsVersion = retired
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "retired TLS floor %d must be rejected", retired)
			}
		})

		ginkgo.It("should accept network posture, capacity, and feature switches", func() {
			input := minimalSpec()
			input.Spec.PublicNetworkAccessEnabled = proto.Bool(false)
			input.Spec.IsVirtualNetworkFilterEnabled = proto.Bool(true)
			input.Spec.VirtualNetworkRules = []*AzureCosmosdbAccountVirtualNetworkRule{
				{SubnetId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/data"), IgnoreMissingVnetServiceEndpoint: proto.Bool(true)},
			}
			input.Spec.NetworkAclBypassForAzureServices = proto.Bool(true)
			input.Spec.NetworkAclBypassIds = []string{"/subscriptions/s/resourceGroups/rg/providers/Microsoft.Synapse/workspaces/syn"}
			input.Spec.Capacity = &AzureCosmosdbAccountCapacity{TotalThroughputLimit: 10000}
			input.Spec.AnalyticalStorageEnabled = proto.Bool(true)
			input.Spec.AnalyticalStorage = &AzureCosmosdbAccountAnalyticalStorage{
				SchemaType: AzureCosmosdbAccountAnalyticalStorageSchemaType_WELL_DEFINED,
			}
			input.Spec.BurstCapacityEnabled = proto.Bool(true)
			input.Spec.PartitionMergeEnabled = proto.Bool(true)
			input.Spec.LocalAuthenticationEnabled = proto.Bool(false)
			input.Spec.AccessKeyMetadataWritesEnabled = proto.Bool(false)
			input.Spec.Tags = map[string]string{"cost-center": "data-platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an unlimited capacity (-1)", func() {
			input := minimalSpec()
			input.Spec.Capacity = &AzureCosmosdbAccountCapacity{TotalThroughputLimit: -1}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing account name", func() {
			input := minimalSpec()
			input.Spec.AccountName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an account name with uppercase letters", func() {
			input := minimalSpec()
			input.Spec.AccountName = "Planton-Cosmos"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing consistency policy", func() {
			input := minimalSpec()
			input.Spec.ConsistencyPolicy = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty geo_locations list", func() {
			input := minimalSpec()
			input.Spec.GeoLocations = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject two write regions (two priority-0 locations)", func() {
			input := minimalSpec()
			input.Spec.GeoLocations = append(input.Spec.GeoLocations,
				&AzureCosmosdbAccountGeoLocation{Location: "westus2", FailoverPriority: 0})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate failover priorities", func() {
			input := minimalSpec()
			input.Spec.GeoLocations = append(input.Spec.GeoLocations,
				&AzureCosmosdbAccountGeoLocation{Location: "westus2", FailoverPriority: 1},
				&AzureCosmosdbAccountGeoLocation{Location: "westeurope", FailoverPriority: 1})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate locations", func() {
			input := minimalSpec()
			input.Spec.GeoLocations = append(input.Spec.GeoLocations,
				&AzureCosmosdbAccountGeoLocation{Location: "eastus", FailoverPriority: 1})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject multi-region BoundedStaleness without the floors", func() {
			input := minimalSpec()
			input.Spec.ConsistencyPolicy.ConsistencyLevel = AzureCosmosdbAccountConsistencyLevel_BOUNDED_STALENESS
			input.Spec.GeoLocations = append(input.Spec.GeoLocations,
				&AzureCosmosdbAccountGeoLocation{Location: "westus2", FailoverPriority: 1})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject multi-region BoundedStaleness with sub-floor values", func() {
			input := minimalSpec()
			input.Spec.ConsistencyPolicy.ConsistencyLevel = AzureCosmosdbAccountConsistencyLevel_BOUNDED_STALENESS
			input.Spec.ConsistencyPolicy.MaxStalenessPrefix = proto.Int32(1000)
			input.Spec.ConsistencyPolicy.MaxIntervalInSeconds = proto.Int32(300)
			input.Spec.GeoLocations = append(input.Spec.GeoLocations,
				&AzureCosmosdbAccountGeoLocation{Location: "westus2", FailoverPriority: 1})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a staleness interval below 5", func() {
			input := minimalSpec()
			input.Spec.ConsistencyPolicy.MaxIntervalInSeconds = proto.Int32(4)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a staleness prefix below 10", func() {
			input := minimalSpec()
			input.Spec.ConsistencyPolicy.MaxStalenessPrefix = proto.Int32(5)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a Mongo-only capability on a SQL account", func() {
			input := minimalSpec()
			input.Spec.Capabilities = []AzureCosmosdbAccountCapability{
				AzureCosmosdbAccountCapability_ENABLE_MONGO,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a SQL-only capability on a MONGO_DB account", func() {
			input := minimalSpec()
			input.Spec.Kind = AzureCosmosdbAccountKind_MONGO_DB
			input.Spec.Capabilities = []AzureCosmosdbAccountCapability{
				AzureCosmosdbAccountCapability_ENABLE_MONGO,
				AzureCosmosdbAccountCapability_ENABLE_CASSANDRA,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject MONGO_DB_V34 without ENABLE_MONGO", func() {
			input := minimalSpec()
			input.Spec.Kind = AzureCosmosdbAccountKind_MONGO_DB
			input.Spec.Capabilities = []AzureCosmosdbAccountCapability{
				AzureCosmosdbAccountCapability_MONGO_DB_V34,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate capabilities", func() {
			input := minimalSpec()
			input.Spec.Capabilities = []AzureCosmosdbAccountCapability{
				AzureCosmosdbAccountCapability_ENABLE_SERVERLESS,
				AzureCosmosdbAccountCapability_ENABLE_SERVERLESS,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a periodic backup carrying a continuous tier", func() {
			input := minimalSpec()
			input.Spec.Backup = &AzureCosmosdbAccountBackup{
				Type: AzureCosmosdbAccountBackupType_PERIODIC,
				Tier: AzureCosmosdbAccountContinuousTier_CONTINUOUS_7_DAYS,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a continuous backup carrying periodic dials", func() {
			input := minimalSpec()
			input.Spec.Backup = &AzureCosmosdbAccountBackup{
				Type:              AzureCosmosdbAccountBackupType_CONTINUOUS,
				IntervalInMinutes: proto.Int32(240),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a backup interval out of range", func() {
			input := minimalSpec()
			input.Spec.Backup = &AzureCosmosdbAccountBackup{
				Type:              AzureCosmosdbAccountBackupType_PERIODIC,
				IntervalInMinutes: proto.Int32(30),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a user-assigned identity without identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureCosmosdbAccountIdentity{
				Type: AzureCosmosdbAccountIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureCosmosdbAccountIdentity{
				Type:        AzureCosmosdbAccountIdentityType_SYSTEM_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/x")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a user-assigned default identity without the identity reference", func() {
			input := minimalSpec()
			input.Spec.DefaultIdentity = &AzureCosmosdbAccountDefaultIdentity{
				Type: AzureCosmosdbAccountDefaultIdentityType_USER_ASSIGNED_DEFAULT,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a first-party default identity carrying an identity reference", func() {
			input := minimalSpec()
			input.Spec.DefaultIdentity = &AzureCosmosdbAccountDefaultIdentity{
				Type:                   AzureCosmosdbAccountDefaultIdentityType_FIRST_PARTY,
				UserAssignedIdentityId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/x"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject create_mode without continuous backup", func() {
			input := minimalSpec()
			input.Spec.CreateMode = AzureCosmosdbAccountCreateMode_DEFAULT
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject create_mode RESTORE without a restore block", func() {
			input := minimalSpec()
			input.Spec.Backup = &AzureCosmosdbAccountBackup{Type: AzureCosmosdbAccountBackupType_CONTINUOUS}
			input.Spec.CreateMode = AzureCosmosdbAccountCreateMode_RESTORE
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a restore block without create_mode RESTORE", func() {
			input := minimalSpec()
			input.Spec.Backup = &AzureCosmosdbAccountBackup{Type: AzureCosmosdbAccountBackupType_CONTINUOUS}
			input.Spec.Restore = &AzureCosmosdbAccountRestore{
				SourceCosmosdbAccountId: "/subscriptions/s/providers/Microsoft.DocumentDB/locations/eastus/restorableDatabaseAccounts/00000000-0000-0000-0000-000000000000",
				RestoreTimestampInUtc:   "2026-07-01T00:00:00Z",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid CORS method", func() {
			input := minimalSpec()
			input.Spec.CorsRule = &AzureCosmosdbAccountCorsRule{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"TRACE"},
				AllowedHeaders: []string{"*"},
				ExposedHeaders: []string{"*"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed IP filter entry", func() {
			input := minimalSpec()
			input.Spec.IpRangeFilter = []string{"not-an-ip"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject IP filter entries with out-of-range octets or prefixes", func() {
			for _, entry := range []string{"999.1.1.1", "10.0.0.256", "10.0.0.0/33", "10.0.0.0/99"} {
				input := minimalSpec()
				input.Spec.IpRangeFilter = []string{entry}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "entry %q must be rejected", entry)
			}
		})

		ginkgo.It("should reject an out-of-vocabulary TLS floor", func() {
			input := minimalSpec()
			input.Spec.MinimalTlsVersion = AzureCosmosdbAccountMinimalTlsVersion(99)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a capacity below -1", func() {
			input := minimalSpec()
			input.Spec.Capacity = &AzureCosmosdbAccountCapacity{TotalThroughputLimit: -2}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
