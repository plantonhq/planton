package azurerediscachev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzureRedisCacheSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureRedisCacheSpec Validation Tests")
}

// helper to create a minimal valid spec (Standard tier, capacity 1)
func minimalSpec() *AzureRedisCache {
	return &AzureRedisCache{
		ApiVersion: "azure.openmcf.org/v1",
		Kind:       "AzureRedisCache",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-redis",
		},
		Spec: &AzureRedisCacheSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name:     "myapp-redis",
			Capacity: 1,
		},
	}
}

var _ = ginkgo.Describe("AzureRedisCacheSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_redis_cache", func() {

			ginkgo.It("should not return a validation error for a minimal Standard cache", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Basic cache with capacity 0", func() {
				sku := "Basic"
				input := minimalSpec()
				input.Spec.SkuName = &sku
				input.Spec.Capacity = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Premium cache", func() {
				sku := "Premium"
				shards := int32(2)
				input := minimalSpec()
				input.Spec.SkuName = &sku
				input.Spec.Capacity = 1
				input.Spec.ShardCount = &shards
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Premium cache with VNet injection", func() {
				sku := "Premium"
				input := minimalSpec()
				input.Spec.SkuName = &sku
				input.Spec.Capacity = 1
				input.Spec.SubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/redis-subnet",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a cache with firewall rules", func() {
				input := minimalSpec()
				input.Spec.FirewallRules = []*AzureRedisFirewallRule{
					{Name: "allow_office", StartIp: "203.0.113.0", EndIp: "203.0.113.255"},
					{Name: "allow_azure", StartIp: "0.0.0.0", EndIp: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a cache with patch schedules", func() {
				startHour := int32(3)
				input := minimalSpec()
				input.Spec.PatchSchedules = []*AzureRedisPatchSchedule{
					{DayOfWeek: "Saturday", StartHourUtc: &startHour},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a cache with zones", func() {
				input := minimalSpec()
				input.Spec.Zones = []string{"1", "2"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for each valid SKU name", func() {
				skus := []string{"Basic", "Standard", "Premium"}
				for _, s := range skus {
					sku := s
					input := minimalSpec()
					input.Spec.SkuName = &sku
					if sku == "Premium" {
						input.Spec.Capacity = 1
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for each valid redis version", func() {
				versions := []string{"4", "6"}
				for _, v := range versions {
					ver := v
					input := minimalSpec()
					input.Spec.RedisVersion = &ver
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for each valid maxmemory policy", func() {
				policies := []string{
					"allkeys-lfu", "allkeys-lru", "allkeys-random",
					"noeviction", "volatile-lfu", "volatile-lru",
					"volatile-random", "volatile-ttl",
				}
				for _, p := range policies {
					policy := p
					input := minimalSpec()
					input.Spec.MaxmemoryPolicy = &policy
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for each valid TLS version", func() {
				versions := []string{"1.0", "1.1", "1.2"}
				for _, v := range versions {
					ver := v
					input := minimalSpec()
					input.Spec.MinimumTlsVersion = &ver
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for each valid patch schedule day", func() {
				days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday", "Everyday", "Weekend"}
				for _, d := range days {
					day := d
					input := minimalSpec()
					input.Spec.PatchSchedules = []*AzureRedisPatchSchedule{
						{DayOfWeek: day},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				sku := "Premium"
				version := "6"
				shards := int32(3)
				nonSsl := true
				tls := "1.2"
				publicAccess := false
				policy := "allkeys-lru"
				startHour := int32(2)
				window := "PT3H"
				input := minimalSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.SkuName = &sku
				input.Spec.Capacity = 2
				input.Spec.RedisVersion = &version
				input.Spec.SubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/redis",
					},
				}
				input.Spec.Zones = []string{"1", "2", "3"}
				input.Spec.ShardCount = &shards
				input.Spec.NonSslPortEnabled = &nonSsl
				input.Spec.MinimumTlsVersion = &tls
				input.Spec.PublicNetworkAccessEnabled = &publicAccess
				input.Spec.MaxmemoryPolicy = &policy
				input.Spec.PatchSchedules = []*AzureRedisPatchSchedule{
					{DayOfWeek: "Saturday", StartHourUtc: &startHour, MaintenanceWindow: &window},
					{DayOfWeek: "Sunday", StartHourUtc: &startHour},
				}
				input.Spec.FirewallRules = []*AzureRedisFirewallRule{
					{Name: "allow_vpn", StartIp: "10.0.0.1", EndIp: "10.0.0.1"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for resource_group", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureResourceGroup,
							Name:      "shared-rg",
							FieldPath: "status.outputs.resource_group_name",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for subnet_id", func() {
				sku := "Premium"
				input := minimalSpec()
				input.Spec.SkuName = &sku
				input.Spec.Capacity = 1
				input.Spec.SubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureSubnet,
							Name:      "redis-subnet",
							FieldPath: "status.outputs.subnet_id",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_redis_cache", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := minimalSpec()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := minimalSpec()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 63 characters", func() {
				tooLong := "a"
				for len(tooLong) < 64 {
					tooLong += "b"
				}
				input := minimalSpec()
				input.Spec.Name = tooLong
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a number", func() {
				input := minimalSpec()
				input.Spec.Name = "1-invalid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains uppercase", func() {
				input := minimalSpec()
				input.Spec.Name = "Invalid-Name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sku_name is invalid", func() {
				invalidSku := "Enterprise"
				input := minimalSpec()
				input.Spec.SkuName = &invalidSku
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when capacity exceeds maximum", func() {
				input := minimalSpec()
				input.Spec.Capacity = 7
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when capacity is negative", func() {
				input := minimalSpec()
				input.Spec.Capacity = -1
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when redis_version is invalid", func() {
				invalidVersion := "5"
				input := minimalSpec()
				input.Spec.RedisVersion = &invalidVersion
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when maxmemory_policy is invalid", func() {
				invalidPolicy := "random-eviction"
				input := minimalSpec()
				input.Spec.MaxmemoryPolicy = &invalidPolicy
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when minimum_tls_version is invalid", func() {
				invalidTls := "1.3"
				input := minimalSpec()
				input.Spec.MinimumTlsVersion = &invalidTls
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shard_count exceeds maximum", func() {
				shards := int32(11)
				input := minimalSpec()
				input.Spec.ShardCount = &shards
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shard_count is zero", func() {
				shards := int32(0)
				input := minimalSpec()
				input.Spec.ShardCount = &shards
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when patch schedule day is invalid", func() {
				input := minimalSpec()
				input.Spec.PatchSchedules = []*AzureRedisPatchSchedule{
					{DayOfWeek: "Funday"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when patch schedule start_hour exceeds 23", func() {
				startHour := int32(25)
				input := minimalSpec()
				input.Spec.PatchSchedules = []*AzureRedisPatchSchedule{
					{DayOfWeek: "Monday", StartHourUtc: &startHour},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule name contains hyphens", func() {
				input := minimalSpec()
				input.Spec.FirewallRules = []*AzureRedisFirewallRule{
					{Name: "allow-office", StartIp: "0.0.0.0", EndIp: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule name is empty", func() {
				input := minimalSpec()
				input.Spec.FirewallRules = []*AzureRedisFirewallRule{
					{Name: "", StartIp: "0.0.0.0", EndIp: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule start_ip is empty", func() {
				input := minimalSpec()
				input.Spec.FirewallRules = []*AzureRedisFirewallRule{
					{Name: "bad_rule", StartIp: "", EndIp: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule end_ip is empty", func() {
				input := minimalSpec()
				input.Spec.FirewallRules = []*AzureRedisFirewallRule{
					{Name: "bad_rule", StartIp: "0.0.0.0", EndIp: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalSpec()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureRedisCache{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureRedisCache",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-redis",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := minimalSpec()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := minimalSpec()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
