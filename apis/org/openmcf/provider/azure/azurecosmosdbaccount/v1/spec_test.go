package azurecosmosdbaccountv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzureCosmosdbAccountSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCosmosdbAccountSpec Validation Tests")
}

// helper to create a minimal valid spec (single-region, no databases)
func minimalSpec() *AzureCosmosdbAccount {
	return &AzureCosmosdbAccount{
		ApiVersion: "azure.openmcf.org/v1",
		Kind:       "AzureCosmosdbAccount",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-cosmos",
		},
		Spec: &AzureCosmosdbAccountSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name: "test-cosmos-account",
			GeoLocations: []*AzureCosmosdbGeoLocation{
				{
					Location:         "eastus",
					FailoverPriority: 0,
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureCosmosdbAccountSpec Validation Tests", func() {

	// -----------------------------------------------------------------------
	// Valid input tests
	// -----------------------------------------------------------------------
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_cosmosdb_account", func() {

			ginkgo.It("should not return a validation error for a minimal account", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with SQL databases and containers", func() {
				input := minimalSpec()
				input.Spec.SqlDatabases = []*AzureCosmosdbSqlDatabase{
					{
						Name:       "mydb",
						Throughput: int32Ptr(400),
						Containers: []*AzureCosmosdbSqlContainer{
							{
								Name:              "mycontainer",
								PartitionKeyPaths: []string{"/tenantId"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with MongoDB databases and collections", func() {
				input := minimalSpec()
				input.Spec.Kind = strPtr("MongoDB")
				input.Spec.MongoServerVersion = strPtr("4.2")
				input.Spec.MongoDatabases = []*AzureCosmosdbMongoDatabase{
					{
						Name: "mydb",
						Collections: []*AzureCosmosdbMongoCollection{
							{
								Name:     "mycollection",
								ShardKey: "tenantId",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multi-region and BoundedStaleness", func() {
				input := minimalSpec()
				input.Spec.ConsistencyPolicy = &AzureCosmosdbConsistencyPolicy{
					ConsistencyLevel:     strPtr("BoundedStaleness"),
					MaxIntervalInSeconds: int32Ptr(300),
					MaxStalenessPrefix:   int32Ptr(100000),
				}
				input.Spec.AutomaticFailoverEnabled = boolPtr(true)
				input.Spec.GeoLocations = []*AzureCosmosdbGeoLocation{
					{Location: "eastus", FailoverPriority: 0, ZoneRedundant: boolPtr(true)},
					{Location: "westus2", FailoverPriority: 1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backup policy Periodic", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureCosmosdbBackupPolicy{
					Type:               "Periodic",
					IntervalInMinutes:  int32Ptr(240),
					RetentionInHours:   int32Ptr(168),
					StorageRedundancy:  strPtr("Geo"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backup policy Continuous", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureCosmosdbBackupPolicy{
					Type: "Continuous",
					Tier: strPtr("Continuous30Days"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with VNet rules and IP filter", func() {
				input := minimalSpec()
				input.Spec.IsVirtualNetworkFilterEnabled = boolPtr(true)
				input.Spec.VirtualNetworkRules = []*AzureCosmosdbVirtualNetworkRule{
					{
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
					},
				}
				input.Spec.IpRangeFilter = []string{"10.0.0.0/24", "203.0.113.1"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with autoscale throughput on container", func() {
				input := minimalSpec()
				input.Spec.SqlDatabases = []*AzureCosmosdbSqlDatabase{
					{
						Name: "mydb",
						Containers: []*AzureCosmosdbSqlContainer{
							{
								Name:                    "autoscale-container",
								PartitionKeyPaths:       []string{"/userId"},
								AutoscaleMaxThroughput:  int32Ptr(4000),
								PartitionKeyKind:        strPtr("Hash"),
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with hierarchical partition key", func() {
				input := minimalSpec()
				input.Spec.SqlDatabases = []*AzureCosmosdbSqlDatabase{
					{
						Name: "mydb",
						Containers: []*AzureCosmosdbSqlContainer{
							{
								Name:              "multi-key-container",
								PartitionKeyPaths: []string{"/tenantId", "/userId"},
								PartitionKeyKind:  strPtr("MultiHash"),
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with free tier and serverless", func() {
				input := minimalSpec()
				input.Spec.FreeTierEnabled = boolPtr(true)
				input.Spec.Capabilities = []string{"EnableServerless"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for all five consistency levels", func() {
				levels := []string{"BoundedStaleness", "ConsistentPrefix", "Eventual", "Session", "Strong"}
				for _, level := range levels {
					input := minimalSpec()
					l := level
					input.Spec.ConsistencyPolicy = &AzureCosmosdbConsistencyPolicy{
						ConsistencyLevel: &l,
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})
	})

	// -----------------------------------------------------------------------
	// Invalid input tests
	// -----------------------------------------------------------------------
	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_cosmosdb_account", func() {

			ginkgo.It("should return a validation error when region is empty", func() {
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

			ginkgo.It("should return a validation error when name is too short", func() {
				input := minimalSpec()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains uppercase", func() {
				input := minimalSpec()
				input.Spec.Name = "Test-Cosmos"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains underscores", func() {
				input := minimalSpec()
				input.Spec.Name = "test_cosmos"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when geo_locations is empty", func() {
				input := minimalSpec()
				input.Spec.GeoLocations = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid kind", func() {
				input := minimalSpec()
				input.Spec.Kind = strPtr("InvalidKind")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid consistency level", func() {
				input := minimalSpec()
				input.Spec.ConsistencyPolicy = &AzureCosmosdbConsistencyPolicy{
					ConsistencyLevel: strPtr("InvalidLevel"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_interval_in_seconds is below minimum", func() {
				input := minimalSpec()
				input.Spec.ConsistencyPolicy = &AzureCosmosdbConsistencyPolicy{
					ConsistencyLevel:     strPtr("BoundedStaleness"),
					MaxIntervalInSeconds: int32Ptr(2),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_staleness_prefix is below minimum", func() {
				input := minimalSpec()
				input.Spec.ConsistencyPolicy = &AzureCosmosdbConsistencyPolicy{
					ConsistencyLevel:   strPtr("BoundedStaleness"),
					MaxStalenessPrefix: int32Ptr(5),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid backup type", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureCosmosdbBackupPolicy{
					Type: "InvalidType",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backup interval is below minimum", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureCosmosdbBackupPolicy{
					Type:              "Periodic",
					IntervalInMinutes: int32Ptr(30),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backup retention is above maximum", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureCosmosdbBackupPolicy{
					Type:             "Periodic",
					RetentionInHours: int32Ptr(800),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid storage redundancy", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureCosmosdbBackupPolicy{
					Type:              "Periodic",
					StorageRedundancy: strPtr("Invalid"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid continuous tier", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureCosmosdbBackupPolicy{
					Type: "Continuous",
					Tier: strPtr("InvalidTier"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid mongo server version", func() {
				input := minimalSpec()
				input.Spec.MongoServerVersion = strPtr("2.4")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid partition key kind", func() {
				input := minimalSpec()
				input.Spec.SqlDatabases = []*AzureCosmosdbSqlDatabase{
					{
						Name: "mydb",
						Containers: []*AzureCosmosdbSqlContainer{
							{
								Name:              "c1",
								PartitionKeyPaths: []string{"/id"},
								PartitionKeyKind:  strPtr("Invalid"),
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when SQL container has no partition key paths", func() {
				input := minimalSpec()
				input.Spec.SqlDatabases = []*AzureCosmosdbSqlDatabase{
					{
						Name: "mydb",
						Containers: []*AzureCosmosdbSqlContainer{
							{
								Name:              "c1",
								PartitionKeyPaths: []string{},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when SQL database throughput is below minimum", func() {
				input := minimalSpec()
				input.Spec.SqlDatabases = []*AzureCosmosdbSqlDatabase{
					{
						Name:       "mydb",
						Throughput: int32Ptr(100),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when autoscale max throughput is below minimum", func() {
				input := minimalSpec()
				input.Spec.SqlDatabases = []*AzureCosmosdbSqlDatabase{
					{
						Name:                   "mydb",
						AutoscaleMaxThroughput: int32Ptr(500),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when geo_location location is empty", func() {
				input := minimalSpec()
				input.Spec.GeoLocations = []*AzureCosmosdbGeoLocation{
					{Location: "", FailoverPriority: 0},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when Mongo collection shard_key is empty", func() {
				input := minimalSpec()
				input.Spec.Kind = strPtr("MongoDB")
				input.Spec.MongoDatabases = []*AzureCosmosdbMongoDatabase{
					{
						Name: "mydb",
						Collections: []*AzureCosmosdbMongoCollection{
							{
								Name:     "coll",
								ShardKey: "",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when Mongo index keys are empty", func() {
				input := minimalSpec()
				input.Spec.Kind = strPtr("MongoDB")
				input.Spec.MongoDatabases = []*AzureCosmosdbMongoDatabase{
					{
						Name: "mydb",
						Collections: []*AzureCosmosdbMongoCollection{
							{
								Name:     "coll",
								ShardKey: "tenantId",
								Indexes: []*AzureCosmosdbMongoIndex{
									{Keys: []string{}},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when VNet rule subnet_id is missing", func() {
				input := minimalSpec()
				input.Spec.VirtualNetworkRules = []*AzureCosmosdbVirtualNetworkRule{
					{SubnetId: nil},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

// Helper functions for optional field pointers
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
