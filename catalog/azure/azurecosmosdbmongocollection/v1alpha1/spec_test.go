package azurecosmosdbmongocollectionv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureCosmosdbMongoCollectionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCosmosdbMongoCollectionSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const databaseId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.DocumentDB/databaseAccounts/planton-cosmos-mongo/mongodbDatabases/app-data"

// minimal valid spec: an unsharded collection sharing the database's
// throughput, carrying the _id index Azure requires on every Mongo
// collection.
func minimalSpec() *AzureCosmosdbMongoCollection {
	return &AzureCosmosdbMongoCollection{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCosmosdbMongoCollection",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-mongo-collection",
		},
		Spec: &AzureCosmosdbMongoCollectionSpec{
			MongoDatabaseId: literal(databaseId),
			CollectionName:  "events",
			Indexes: []*AzureCosmosdbMongoCollectionIndex{
				{Keys: []string{"_id"}, Unique: proto.Bool(true)},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureCosmosdbMongoCollectionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal unsharded collection", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a sharded collection with compound indexes", func() {
			input := minimalSpec()
			input.Spec.ShardKey = "tenantId"
			input.Spec.Indexes = []*AzureCosmosdbMongoCollectionIndex{
				{Keys: []string{"_id"}, Unique: proto.Bool(true)},
				{Keys: []string{"tenantId", "createdAt"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept the _id index case-insensitively, mirroring Azure's check", func() {
			input := minimalSpec()
			input.Spec.Indexes = []*AzureCosmosdbMongoCollectionIndex{
				{Keys: []string{"_ID"}, Unique: proto.Bool(true)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a collection name at the 255-character ceiling", func() {
			input := minimalSpec()
			input.Spec.CollectionName = strings.Repeat("a", 255)
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
			for _, ttl := range []int32{-1, 86400} {
				input := minimalSpec()
				input.Spec.DefaultTtlSeconds = proto.Int32(ttl)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "ttl %d must be accepted", ttl)
			}
		})

		ginkgo.It("should accept an analytical-store TTL", func() {
			input := minimalSpec()
			input.Spec.AnalyticalStorageTtl = proto.Int32(-1)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing database reference", func() {
			input := minimalSpec()
			input.Spec.MongoDatabaseId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing collection name", func() {
			input := minimalSpec()
			input.Spec.CollectionName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a collection name over 255 characters", func() {
			input := minimalSpec()
			input.Spec.CollectionName = strings.Repeat("a", 256)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty index list -- Azure requires the _id index", func() {
			input := minimalSpec()
			input.Spec.Indexes = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject indexes that lack the _id key", func() {
			input := minimalSpec()
			input.Spec.Indexes = []*AzureCosmosdbMongoCollectionIndex{
				{Keys: []string{"tenantId", "createdAt"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a TTL of 0", func() {
			input := minimalSpec()
			input.Spec.DefaultTtlSeconds = proto.Int32(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a TTL below -1", func() {
			input := minimalSpec()
			input.Spec.DefaultTtlSeconds = proto.Int32(-2)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject throughput off the increments", func() {
			fixed := minimalSpec()
			fixed.Spec.Throughput = proto.Int32(450)
			gomega.Expect(protovalidate.Validate(fixed)).NotTo(gomega.BeNil())

			autoscale := minimalSpec()
			autoscale.Spec.AutoscaleMaxThroughput = proto.Int32(2500)
			gomega.Expect(protovalidate.Validate(autoscale)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject throughput below the floors", func() {
			// 300 satisfies the 100-increment rule but sits below the
			// 400 RU/s floor; 0 satisfies the 1000-increment rule but
			// sits below the 1000 RU/s autoscale floor.
			fixed := minimalSpec()
			fixed.Spec.Throughput = proto.Int32(300)
			gomega.Expect(protovalidate.Validate(fixed)).NotTo(gomega.BeNil())

			autoscale := minimalSpec()
			autoscale.Spec.AutoscaleMaxThroughput = proto.Int32(0)
			gomega.Expect(protovalidate.Validate(autoscale)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an analytical-store TTL below -1", func() {
			input := minimalSpec()
			input.Spec.AnalyticalStorageTtl = proto.Int32(-2)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject fixed throughput and autoscale together", func() {
			input := minimalSpec()
			input.Spec.Throughput = proto.Int32(400)
			input.Spec.AutoscaleMaxThroughput = proto.Int32(1000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an index without keys", func() {
			input := minimalSpec()
			input.Spec.Indexes = []*AzureCosmosdbMongoCollectionIndex{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
