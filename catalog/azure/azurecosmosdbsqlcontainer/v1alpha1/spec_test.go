package azurecosmosdbsqlcontainerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureCosmosdbSqlContainerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCosmosdbSqlContainerSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const databaseId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.DocumentDB/databaseAccounts/planton-cosmos/sqlDatabases/app-data"

// minimal valid spec: a single-path Hash-partitioned container sharing
// the database's throughput.
func minimalSpec() *AzureCosmosdbSqlContainer {
	return &AzureCosmosdbSqlContainer{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCosmosdbSqlContainer",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-sql-container",
		},
		Spec: &AzureCosmosdbSqlContainerSpec{
			SqlDatabaseId:     literal(databaseId),
			ContainerName:     "orders",
			PartitionKeyPaths: []string{"/tenantId"},
		},
	}
}

var _ = ginkgo.Describe("AzureCosmosdbSqlContainerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal Hash container", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept an explicit HASH kind with one path", func() {
			input := minimalSpec()
			input.Spec.PartitionKeyKind = AzureCosmosdbSqlContainerPartitionKeyKind_HASH
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a MULTI_HASH hierarchical key with version 2", func() {
			input := minimalSpec()
			input.Spec.PartitionKeyPaths = []string{"/tenantId", "/userId", "/sessionId"}
			input.Spec.PartitionKeyKind = AzureCosmosdbSqlContainerPartitionKeyKind_MULTI_HASH
			input.Spec.PartitionKeyVersion = proto.Int32(2)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a single-path MULTI_HASH key (a one-level hierarchy is legal)", func() {
			input := minimalSpec()
			input.Spec.PartitionKeyKind = AzureCosmosdbSqlContainerPartitionKeyKind_MULTI_HASH
			input.Spec.PartitionKeyVersion = proto.Int32(2)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept fixed and autoscale throughput individually", func() {
			fixed := minimalSpec()
			fixed.Spec.Throughput = proto.Int32(400)
			gomega.Expect(protovalidate.Validate(fixed)).To(gomega.BeNil())

			autoscale := minimalSpec()
			autoscale.Spec.AutoscaleMaxThroughput = proto.Int32(1000)
			gomega.Expect(protovalidate.Validate(autoscale)).To(gomega.BeNil())
		})

		ginkgo.It("should accept TTL semantics (-1 and positive)", func() {
			for _, ttl := range []int32{-1, 3600} {
				input := minimalSpec()
				input.Spec.DefaultTtl = proto.Int32(ttl)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "ttl %d must be accepted", ttl)
			}
		})

		ginkgo.It("should accept an analytical-store TTL", func() {
			input := minimalSpec()
			input.Spec.AnalyticalStorageTtl = proto.Int32(-1)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept unique keys", func() {
			input := minimalSpec()
			input.Spec.UniqueKeys = []*AzureCosmosdbSqlContainerUniqueKey{
				{Paths: []string{"/orderNumber"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full indexing policy", func() {
			input := minimalSpec()
			input.Spec.IndexingPolicy = &AzureCosmosdbSqlContainerIndexingPolicy{
				IndexingMode:  AzureCosmosdbSqlContainerIndexingMode_CONSISTENT,
				IncludedPaths: []*AzureCosmosdbSqlContainerIndexPath{{Path: "/*"}},
				ExcludedPaths: []*AzureCosmosdbSqlContainerIndexPath{{Path: "/payload/*"}},
				CompositeIndexes: []*AzureCosmosdbSqlContainerCompositeIndex{
					{Entries: []*AzureCosmosdbSqlContainerCompositeIndexEntry{
						{Path: "/tenantId", Order: AzureCosmosdbSqlContainerCompositeIndexOrder_ASCENDING},
						{Path: "/createdAt", Order: AzureCosmosdbSqlContainerCompositeIndexOrder_DESCENDING},
					}},
				},
				SpatialIndexes: []*AzureCosmosdbSqlContainerSpatialIndex{{Path: "/location/*"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a NONE indexing mode", func() {
			input := minimalSpec()
			input.Spec.IndexingPolicy = &AzureCosmosdbSqlContainerIndexingPolicy{
				IndexingMode: AzureCosmosdbSqlContainerIndexingMode_NONE,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept last-writer-wins conflict resolution with a path", func() {
			input := minimalSpec()
			input.Spec.ConflictResolutionPolicy = &AzureCosmosdbSqlContainerConflictResolutionPolicy{
				Mode:                   AzureCosmosdbSqlContainerConflictResolutionMode_LAST_WRITER_WINS,
				ConflictResolutionPath: "/_ts",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept custom conflict resolution with a procedure", func() {
			input := minimalSpec()
			input.Spec.ConflictResolutionPolicy = &AzureCosmosdbSqlContainerConflictResolutionPolicy{
				Mode:                        AzureCosmosdbSqlContainerConflictResolutionMode_CUSTOM,
				ConflictResolutionProcedure: "dbs/app-data/colls/orders/sprocs/resolver",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing database reference", func() {
			input := minimalSpec()
			input.Spec.SqlDatabaseId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing container name", func() {
			input := minimalSpec()
			input.Spec.ContainerName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty partition key path list", func() {
			input := minimalSpec()
			input.Spec.PartitionKeyPaths = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a partition key path not starting with '/'", func() {
			input := minimalSpec()
			input.Spec.PartitionKeyPaths = []string{"tenantId"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject more than three partition key paths", func() {
			input := minimalSpec()
			input.Spec.PartitionKeyPaths = []string{"/a", "/b", "/c", "/d"}
			input.Spec.PartitionKeyKind = AzureCosmosdbSqlContainerPartitionKeyKind_MULTI_HASH
			input.Spec.PartitionKeyVersion = proto.Int32(2)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject multiple paths on a HASH key", func() {
			input := minimalSpec()
			input.Spec.PartitionKeyPaths = []string{"/tenantId", "/userId"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject MULTI_HASH without partition key version 2", func() {
			input := minimalSpec()
			input.Spec.PartitionKeyPaths = []string{"/tenantId", "/userId"}
			input.Spec.PartitionKeyKind = AzureCosmosdbSqlContainerPartitionKeyKind_MULTI_HASH
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a partition key version out of range", func() {
			input := minimalSpec()
			input.Spec.PartitionKeyVersion = proto.Int32(3)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject fixed throughput and autoscale together", func() {
			input := minimalSpec()
			input.Spec.Throughput = proto.Int32(400)
			input.Spec.AutoscaleMaxThroughput = proto.Int32(1000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject throughput off the increments", func() {
			fixed := minimalSpec()
			fixed.Spec.Throughput = proto.Int32(450)
			gomega.Expect(protovalidate.Validate(fixed)).NotTo(gomega.BeNil())

			autoscale := minimalSpec()
			autoscale.Spec.AutoscaleMaxThroughput = proto.Int32(1500)
			gomega.Expect(protovalidate.Validate(autoscale)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a TTL below -1", func() {
			input := minimalSpec()
			input.Spec.DefaultTtl = proto.Int32(-2)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a unique key without paths", func() {
			input := minimalSpec()
			input.Spec.UniqueKeys = []*AzureCosmosdbSqlContainerUniqueKey{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a composite index without entries", func() {
			input := minimalSpec()
			input.Spec.IndexingPolicy = &AzureCosmosdbSqlContainerIndexingPolicy{
				CompositeIndexes: []*AzureCosmosdbSqlContainerCompositeIndex{{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a NONE-mode policy that declares paths", func() {
			input := minimalSpec()
			input.Spec.IndexingPolicy = &AzureCosmosdbSqlContainerIndexingPolicy{
				IndexingMode:  AzureCosmosdbSqlContainerIndexingMode_NONE,
				IncludedPaths: []*AzureCosmosdbSqlContainerIndexPath{{Path: "/*"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject declared paths that omit the root path '/*'", func() {
			input := minimalSpec()
			input.Spec.IndexingPolicy = &AzureCosmosdbSqlContainerIndexingPolicy{
				IndexingMode:  AzureCosmosdbSqlContainerIndexingMode_CONSISTENT,
				IncludedPaths: []*AzureCosmosdbSqlContainerIndexPath{{Path: "/tenantId/?"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the root path '/*' in both included and excluded paths", func() {
			input := minimalSpec()
			input.Spec.IndexingPolicy = &AzureCosmosdbSqlContainerIndexingPolicy{
				IndexingMode:  AzureCosmosdbSqlContainerIndexingMode_CONSISTENT,
				IncludedPaths: []*AzureCosmosdbSqlContainerIndexPath{{Path: "/*"}},
				ExcludedPaths: []*AzureCosmosdbSqlContainerIndexPath{{Path: "/*"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept the root path excluded with specific paths included", func() {
			input := minimalSpec()
			input.Spec.IndexingPolicy = &AzureCosmosdbSqlContainerIndexingPolicy{
				IndexingMode:  AzureCosmosdbSqlContainerIndexingMode_CONSISTENT,
				IncludedPaths: []*AzureCosmosdbSqlContainerIndexPath{{Path: "/tenantId/?"}},
				ExcludedPaths: []*AzureCosmosdbSqlContainerIndexPath{{Path: "/*"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject last-writer-wins carrying a procedure", func() {
			input := minimalSpec()
			input.Spec.ConflictResolutionPolicy = &AzureCosmosdbSqlContainerConflictResolutionPolicy{
				Mode:                        AzureCosmosdbSqlContainerConflictResolutionMode_LAST_WRITER_WINS,
				ConflictResolutionProcedure: "dbs/x/colls/y/sprocs/z",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject custom mode carrying a path", func() {
			input := minimalSpec()
			input.Spec.ConflictResolutionPolicy = &AzureCosmosdbSqlContainerConflictResolutionPolicy{
				Mode:                   AzureCosmosdbSqlContainerConflictResolutionMode_CUSTOM,
				ConflictResolutionPath: "/_ts",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a conflict policy without a mode", func() {
			input := minimalSpec()
			input.Spec.ConflictResolutionPolicy = &AzureCosmosdbSqlContainerConflictResolutionPolicy{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
