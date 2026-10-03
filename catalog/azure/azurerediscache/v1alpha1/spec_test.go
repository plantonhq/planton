package azurerediscachev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureRedisCacheSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureRedisCacheSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

func boolPtr(v bool) *bool { return &v }

func stringPtr(v string) *string { return &v }

const subnetId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/redis"

// minimal valid spec: a Standard-tier cache (sku unspecified = STANDARD).
func minimalSpec() *AzureRedisCache {
	return &AzureRedisCache{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureRedisCache",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-cache",
		},
		Spec: &AzureRedisCacheSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			CacheName:     "planton-test-cache",
		},
	}
}

// premiumSpec returns a valid PREMIUM-tier baseline (capacity P1).
func premiumSpec() *AzureRedisCache {
	input := minimalSpec()
	input.Spec.SkuName = AzureRedisCacheSku_PREMIUM
	input.Spec.Capacity = 1
	return input
}

var _ = ginkgo.Describe("AzureRedisCacheSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal Standard cache", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept every explicit tier at a legal capacity", func() {
			for _, sku := range []AzureRedisCacheSku{AzureRedisCacheSku_BASIC, AzureRedisCacheSku_STANDARD} {
				input := minimalSpec()
				input.Spec.SkuName = sku
				input.Spec.Capacity = 6
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "sku %v must be accepted", sku)
			}
			gomega.Expect(protovalidate.Validate(premiumSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a resource group reference", func() {
			input := minimalSpec()
			input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureResourceGroup,
						Name:      "app-rg",
						FieldPath: "status.outputs.resource_group_name",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept single-label and hyphenated cache names", func() {
			for _, name := range []string{"c", "cache1", "my-app-cache", "0cache"} {
				input := minimalSpec()
				input.Spec.CacheName = name
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "name %q must be accepted", name)
			}
		})

		ginkgo.It("should accept a VNet-injected Premium cache with a static IP", func() {
			input := premiumSpec()
			input.Spec.SubnetId = literal(subnetId)
			input.Spec.PrivateStaticIpAddress = stringPtr("10.0.1.10")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept clustering on Premium", func() {
			input := premiumSpec()
			input.Spec.ShardCount = int32Ptr(3)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept extra replicas on Premium", func() {
			input := premiumSpec()
			input.Spec.ReplicasPerPrimary = int32Ptr(2)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept zones and tags", func() {
			input := minimalSpec()
			input.Spec.Zones = []string{"1", "2"}
			input.Spec.Tags = map[string]string{"cost-center": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept disabling access keys once Entra auth is on", func() {
			input := minimalSpec()
			input.Spec.AccessKeysAuthenticationEnabled = boolPtr(false)
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				ActiveDirectoryAuthenticationEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept disabling Redis authentication on a VNet-injected cache", func() {
			input := premiumSpec()
			input.Spec.SubnetId = literal(subnetId)
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				AuthenticationEnabled: boolPtr(false),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept memory dials on Standard", func() {
			input := minimalSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				MaxmemoryReserved:              int32Ptr(50),
				MaxmemoryDelta:                 int32Ptr(50),
				MaxfragmentationmemoryReserved: int32Ptr(50),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept RDB persistence on Premium with a connection string", func() {
			input := premiumSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				RdbBackupEnabled:           true,
				RdbBackupFrequency:         int32Ptr(60),
				RdbBackupMaxSnapshotCount:  int32Ptr(1),
				RdbStorageConnectionString: "DefaultEndpointsProtocol=https;AccountName=x;AccountKey=y",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept RDB persistence via managed identity without a connection string", func() {
			input := premiumSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				RdbBackupEnabled:                    true,
				DataPersistenceAuthenticationMethod: AzureRedisCachePersistenceAuthMethod_MANAGED_IDENTITY,
			}
			input.Spec.Identity = &AzureRedisCacheIdentity{
				Type: AzureRedisCacheIdentityType_SYSTEM_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept AOF persistence on Premium", func() {
			input := premiumSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				AofBackupEnabled:             true,
				AofStorageConnectionString_0: "DefaultEndpointsProtocol=https;AccountName=x;AccountKey=y",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a user-assigned identity with entries", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureRedisCacheIdentity{
				Type:                    AzureRedisCacheIdentityType_USER_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept patch schedules with explicit windows", func() {
			input := minimalSpec()
			input.Spec.PatchSchedules = []*AzureRedisCachePatchSchedule{
				{
					DayOfWeek:         AzureRedisCachePatchScheduleDay_SUNDAY,
					StartHourUtc:      int32Ptr(2),
					MaintenanceWindow: stringPtr("PT6H30M"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept firewall rules with valid names and IPv4 ranges", func() {
			input := minimalSpec()
			input.Spec.FirewallRules = []*AzureRedisCacheFirewallRule{
				{Name: "office_range", StartIp: "203.0.113.0", EndIp: "203.0.113.255"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every documented eviction policy", func() {
			for _, policy := range []string{"allkeys-lfu", "allkeys-lru", "allkeys-random", "noeviction", "volatile-lfu", "volatile-lru", "volatile-random", "volatile-ttl"} {
				input := minimalSpec()
				input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
					MaxmemoryPolicy: stringPtr(policy),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "policy %q must be accepted", policy)
			}
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

		ginkgo.It("should reject names that break the DNS-label rules", func() {
			for _, name := range []string{"", "-cache", "cache-", "my--cache", "my_cache", "MY CACHE"} {
				input := minimalSpec()
				input.Spec.CacheName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a name longer than 63 characters", func() {
			input := minimalSpec()
			input.Spec.CacheName = "a123456789b123456789c123456789d123456789e123456789f123456789g123"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined sku enum value", func() {
			input := minimalSpec()
			input.Spec.SkuName = AzureRedisCacheSku(99)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject Premium capacity outside P1-P5", func() {
			for _, capacity := range []int32{0, 6} {
				input := premiumSpec()
				input.Spec.Capacity = capacity
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "capacity %d must be rejected on Premium", capacity)
			}
		})

		ginkgo.It("should reject capacity above the C-family range", func() {
			input := minimalSpec()
			input.Spec.Capacity = 7
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unsupported redis version", func() {
			input := minimalSpec()
			input.Spec.RedisVersion = stringPtr("5")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject VNet injection outside Premium", func() {
			input := minimalSpec()
			input.Spec.SubnetId = literal(subnetId)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a static IP without VNet injection", func() {
			input := premiumSpec()
			input.Spec.PrivateStaticIpAddress = stringPtr("10.0.1.10")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed static IP", func() {
			input := premiumSpec()
			input.Spec.SubnetId = literal(subnetId)
			input.Spec.PrivateStaticIpAddress = stringPtr("not-an-ip")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject clustering outside Premium", func() {
			input := minimalSpec()
			input.Spec.ShardCount = int32Ptr(2)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject extra replicas outside Premium", func() {
			input := minimalSpec()
			input.Spec.ReplicasPerPrimary = int32Ptr(2)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject combining clustering with extra replicas", func() {
			input := premiumSpec()
			input.Spec.ShardCount = int32Ptr(2)
			input.Spec.ReplicasPerPrimary = int32Ptr(2)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a zone outside 1-3", func() {
			input := minimalSpec()
			input.Spec.Zones = []string{"4"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject disabling access keys without Entra auth", func() {
			input := minimalSpec()
			input.Spec.AccessKeysAuthenticationEnabled = boolPtr(false)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject disabling Redis authentication without VNet injection", func() {
			input := minimalSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				AuthenticationEnabled: boolPtr(false),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject memory dials on the Basic tier", func() {
			input := minimalSpec()
			input.Spec.SkuName = AzureRedisCacheSku_BASIC
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				MaxmemoryReserved: int32Ptr(50),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid eviction policy", func() {
			input := minimalSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				MaxmemoryPolicy: stringPtr("evict-everything"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject RDB persistence outside Premium", func() {
			input := minimalSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				RdbBackupEnabled:           true,
				RdbStorageConnectionString: "conn",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject RDB persistence without a connection string under SAS auth", func() {
			input := premiumSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				RdbBackupEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid RDB backup frequency", func() {
			input := premiumSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				RdbBackupEnabled:           true,
				RdbBackupFrequency:         int32Ptr(45),
				RdbStorageConnectionString: "conn",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject AOF persistence outside Premium", func() {
			input := minimalSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				AofBackupEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject MANAGED_IDENTITY persistence auth without an identity block", func() {
			input := premiumSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				DataPersistenceAuthenticationMethod: AzureRedisCachePersistenceAuthMethod_MANAGED_IDENTITY,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed storage account subscription id", func() {
			input := premiumSpec()
			input.Spec.RedisConfiguration = &AzureRedisCacheConfiguration{
				StorageAccountSubscriptionId: stringPtr("not-a-uuid"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a system-assigned identity carrying user-assigned entries", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureRedisCacheIdentity{
				Type:                    AzureRedisCacheIdentityType_SYSTEM_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/id")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a user-assigned identity without entries", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureRedisCacheIdentity{
				Type: AzureRedisCacheIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unspecified patch-schedule day", func() {
			input := minimalSpec()
			input.Spec.PatchSchedules = []*AzureRedisCachePatchSchedule{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a start hour outside 0-23", func() {
			input := minimalSpec()
			input.Spec.PatchSchedules = []*AzureRedisCachePatchSchedule{
				{DayOfWeek: AzureRedisCachePatchScheduleDay_MONDAY, StartHourUtc: int32Ptr(24)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject malformed maintenance windows", func() {
			for _, window := range []string{"PT", "5H", "P1D", "five hours"} {
				input := minimalSpec()
				input.Spec.PatchSchedules = []*AzureRedisCachePatchSchedule{
					{DayOfWeek: AzureRedisCachePatchScheduleDay_MONDAY, MaintenanceWindow: stringPtr(window)},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "window %q must be rejected", window)
			}
		})

		ginkgo.It("should reject firewall rule names with hyphens", func() {
			input := minimalSpec()
			input.Spec.FirewallRules = []*AzureRedisCacheFirewallRule{
				{Name: "office-range", StartIp: "203.0.113.0", EndIp: "203.0.113.255"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject firewall rules with malformed addresses", func() {
			input := minimalSpec()
			input.Spec.FirewallRules = []*AzureRedisCacheFirewallRule{
				{Name: "bad_rule", StartIp: "203.0.113.0", EndIp: "not-an-ip"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
