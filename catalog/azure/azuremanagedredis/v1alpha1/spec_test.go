package azuremanagedredisv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureManagedRedisSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureManagedRedisSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testKeyId      = "https://platform-kv.vault.azure.net/keys/redis-cmk/0123456789abcdef0123456789abcdef"
	testIdentityId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/redis-cmk"
)

// minimal valid spec: the keyless default database on a Balanced_B0.
func minimalSpec() *AzureManagedRedis {
	return &AzureManagedRedis{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureManagedRedis",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-managed-redis",
		},
		Spec: &AzureManagedRedisSpec{
			Region:          "eastus",
			ResourceGroup:   literal("test-rg"),
			ClusterName:     "app-cache",
			SkuName:         AzureManagedRedisSku_BALANCED_B0,
			DefaultDatabase: &AzureManagedRedisDatabase{},
		},
	}
}

var _ = ginkgo.Describe("AzureManagedRedisSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept the minimal keyless spec", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept every SKU family", func() {
			for _, sku := range []AzureManagedRedisSku{
				AzureManagedRedisSku_BALANCED_B1000,
				AzureManagedRedisSku_COMPUTE_OPTIMIZED_X700,
				AzureManagedRedisSku_MEMORY_OPTIMIZED_M2000,
				AzureManagedRedisSku_FLASH_OPTIMIZED_A4500,
			} {
				input := minimalSpec()
				input.Spec.SkuName = sku
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "sku %v must be accepted", sku)
			}
		})

		ginkgo.It("should accept a fully configured database", func() {
			input := minimalSpec()
			input.Spec.SkuName = AzureManagedRedisSku_MEMORY_OPTIMIZED_M10
			input.Spec.DefaultDatabase = &AzureManagedRedisDatabase{
				AccessKeysAuthenticationEnabled: true,
				ClientProtocol:                  AzureManagedRedisClientProtocol_ENCRYPTED,
				ClusteringPolicy:                AzureManagedRedisClusteringPolicy_ENTERPRISE_CLUSTER,
				EvictionPolicy:                  AzureManagedRedisEvictionPolicy_NO_EVICTION,
				Modules: []*AzureManagedRedisModule{
					{Name: "RediSearch"},
					{Name: "RedisJSON"},
					{Name: "RedisBloom", Args: "ERROR_RATE 0.01"},
				},
				PersistenceRedisDatabaseBackupFrequency: stringPtr("6h"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept AOF persistence at its only frequency", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.PersistenceAppendOnlyFileBackupFrequency = stringPtr("1s")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every RDB frequency", func() {
			for _, frequency := range []string{"1h", "6h", "12h"} {
				input := minimalSpec()
				input.Spec.DefaultDatabase.PersistenceRedisDatabaseBackupFrequency = stringPtr(frequency)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "frequency %q must be accepted", frequency)
			}
		})

		ginkgo.It("should accept geo-replication on the B3 sku floor", func() {
			input := minimalSpec()
			input.Spec.SkuName = AzureManagedRedisSku_BALANCED_B3
			input.Spec.DefaultDatabase.GeoReplicationGroupName = "global-cache-group"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a geo-replicated database with the allowed modules", func() {
			input := minimalSpec()
			input.Spec.SkuName = AzureManagedRedisSku_MEMORY_OPTIMIZED_M10
			input.Spec.DefaultDatabase.GeoReplicationGroupName = "global-cache-group"
			input.Spec.DefaultDatabase.ClusteringPolicy = AzureManagedRedisClusteringPolicy_ENTERPRISE_CLUSTER
			input.Spec.DefaultDatabase.EvictionPolicy = AzureManagedRedisEvictionPolicy_NO_EVICTION
			input.Spec.DefaultDatabase.Modules = []*AzureManagedRedisModule{
				{Name: "RediSearch"},
				{Name: "RedisJSON"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept customer-managed-key encryption with a user-assigned identity", func() {
			input := minimalSpec()
			input.Spec.CustomerManagedKey = &AzureManagedRedisCustomerManagedKey{
				KeyVaultKeyId:          literal(testKeyId),
				UserAssignedIdentityId: literal(testIdentityId),
			}
			input.Spec.Identity = &AzureManagedRedisIdentity{
				Type:                    AzureManagedRedisIdentityType_USER_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a system-assigned identity without identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureManagedRedisIdentity{
				Type: AzureManagedRedisIdentityType_SYSTEM_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept disabling high availability and public access", func() {
			input := minimalSpec()
			input.Spec.HighAvailabilityEnabled = boolPtr(false)
			input.Spec.PublicNetworkAccessEnabled = boolPtr(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept user tags", func() {
			input := minimalSpec()
			input.Spec.Tags = map[string]string{"cost-center": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept single-character-interior names at the length bounds", func() {
			input := minimalSpec()
			input.Spec.ClusterName = "a-b"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			input.Spec.ClusterName = "a" + repeat("b", 61) + "c"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing region", func() {
			input := minimalSpec()
			input.Spec.Region = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing resource group", func() {
			input := minimalSpec()
			input.Spec.ResourceGroup = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing cluster name", func() {
			input := minimalSpec()
			input.Spec.ClusterName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject names below the 3-character minimum", func() {
			input := minimalSpec()
			input.Spec.ClusterName = "ab"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject names over 63 characters", func() {
			input := minimalSpec()
			input.Spec.ClusterName = repeat("a", 64)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject names with consecutive hyphens", func() {
			input := minimalSpec()
			input.Spec.ClusterName = "app--cache"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject names starting or ending with a hyphen", func() {
			input := minimalSpec()
			input.Spec.ClusterName = "-app-cache"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

			input.Spec.ClusterName = "app-cache-"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unspecified sku", func() {
			input := minimalSpec()
			input.Spec.SkuName = AzureManagedRedisSku_azure_managed_redis_sku_unspecified
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing default database", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject geo-replication below the B3 sku floor", func() {
			for _, sku := range []AzureManagedRedisSku{
				AzureManagedRedisSku_BALANCED_B0,
				AzureManagedRedisSku_BALANCED_B1,
			} {
				input := minimalSpec()
				input.Spec.SkuName = sku
				input.Spec.DefaultDatabase.GeoReplicationGroupName = "global-cache-group"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "sku %v must reject geo-replication", sku)
			}
		})

		ginkgo.It("should reject a malformed geo-replication group name", func() {
			input := minimalSpec()
			input.Spec.SkuName = AzureManagedRedisSku_BALANCED_B3
			input.Spec.DefaultDatabase.GeoReplicationGroupName = "bad--group"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject enabling both AOF and RDB persistence", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.PersistenceAppendOnlyFileBackupFrequency = stringPtr("1s")
			input.Spec.DefaultDatabase.PersistenceRedisDatabaseBackupFrequency = stringPtr("1h")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject persistence on a geo-replicated database", func() {
			input := minimalSpec()
			input.Spec.SkuName = AzureManagedRedisSku_BALANCED_B3
			input.Spec.DefaultDatabase.GeoReplicationGroupName = "global-cache-group"
			input.Spec.DefaultDatabase.PersistenceRedisDatabaseBackupFrequency = stringPtr("1h")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid AOF frequency", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.PersistenceAppendOnlyFileBackupFrequency = stringPtr("5s")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid RDB frequency", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.PersistenceRedisDatabaseBackupFrequency = stringPtr("2h")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown module name", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.Modules = []*AzureManagedRedisModule{{Name: "RedisGraph"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject more than four modules", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.ClusteringPolicy = AzureManagedRedisClusteringPolicy_ENTERPRISE_CLUSTER
			input.Spec.DefaultDatabase.EvictionPolicy = AzureManagedRedisEvictionPolicy_NO_EVICTION
			input.Spec.DefaultDatabase.Modules = []*AzureManagedRedisModule{
				{Name: "RediSearch"},
				{Name: "RedisJSON"},
				{Name: "RedisBloom"},
				{Name: "RedisTimeSeries"},
				{Name: "RedisJSON"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a module enabled twice", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.Modules = []*AzureManagedRedisModule{
				{Name: "RedisJSON"},
				{Name: "RedisJSON"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject disallowed modules on a geo-replicated database", func() {
			input := minimalSpec()
			input.Spec.SkuName = AzureManagedRedisSku_BALANCED_B3
			input.Spec.DefaultDatabase.GeoReplicationGroupName = "global-cache-group"
			input.Spec.DefaultDatabase.Modules = []*AzureManagedRedisModule{{Name: "RedisBloom"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject RediSearch without NO_EVICTION", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.ClusteringPolicy = AzureManagedRedisClusteringPolicy_ENTERPRISE_CLUSTER
			input.Spec.DefaultDatabase.EvictionPolicy = AzureManagedRedisEvictionPolicy_VOLATILE_LRU
			input.Spec.DefaultDatabase.Modules = []*AzureManagedRedisModule{{Name: "RediSearch"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject RediSearch with the unspecified (OSSCluster-default) clustering policy", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.EvictionPolicy = AzureManagedRedisEvictionPolicy_NO_EVICTION
			input.Spec.DefaultDatabase.Modules = []*AzureManagedRedisModule{{Name: "RediSearch"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject RediSearch with OSS_CLUSTER clustering", func() {
			input := minimalSpec()
			input.Spec.DefaultDatabase.ClusteringPolicy = AzureManagedRedisClusteringPolicy_OSS_CLUSTER
			input.Spec.DefaultDatabase.EvictionPolicy = AzureManagedRedisEvictionPolicy_NO_EVICTION
			input.Spec.DefaultDatabase.Modules = []*AzureManagedRedisModule{{Name: "RediSearch"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject customer-managed key without its key reference", func() {
			input := minimalSpec()
			input.Spec.CustomerManagedKey = &AzureManagedRedisCustomerManagedKey{
				UserAssignedIdentityId: literal(testIdentityId),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject customer-managed key without its identity reference", func() {
			input := minimalSpec()
			input.Spec.CustomerManagedKey = &AzureManagedRedisCustomerManagedKey{
				KeyVaultKeyId: literal(testKeyId),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unspecified identity type", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureManagedRedisIdentity{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject USER_ASSIGNED identity without identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureManagedRedisIdentity{
				Type: AzureManagedRedisIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject SYSTEM_ASSIGNED identity carrying identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureManagedRedisIdentity{
				Type:                    AzureManagedRedisIdentityType_SYSTEM_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})

func stringPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
