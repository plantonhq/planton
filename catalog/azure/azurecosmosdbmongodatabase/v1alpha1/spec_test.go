package azurecosmosdbmongodatabasev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureCosmosdbMongoDatabaseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCosmosdbMongoDatabaseSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.DocumentDB/databaseAccounts/planton-cosmos-mongo"

// minimal valid spec: a database sharing no throughput (collections
// bring their own, or the account is serverless).
func minimalSpec() *AzureCosmosdbMongoDatabase {
	return &AzureCosmosdbMongoDatabase{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCosmosdbMongoDatabase",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-mongo-database",
		},
		Spec: &AzureCosmosdbMongoDatabaseSpec{
			CosmosdbAccountId: literal(accountId),
			DatabaseName:      "app-data",
		},
	}
}

var _ = ginkgo.Describe("AzureCosmosdbMongoDatabaseSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal database", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept fixed throughput in 100 RU/s increments", func() {
			input := minimalSpec()
			input.Spec.Throughput = proto.Int32(400)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept autoscale ceilings in 1000 RU/s increments", func() {
			input := minimalSpec()
			input.Spec.AutoscaleMaxThroughput = proto.Int32(4000)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing account reference", func() {
			input := minimalSpec()
			input.Spec.CosmosdbAccountId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing database name", func() {
			input := minimalSpec()
			input.Spec.DatabaseName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject throughput below the 400 RU/s floor", func() {
			input := minimalSpec()
			input.Spec.Throughput = proto.Int32(300)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject throughput off the 100 RU/s increment", func() {
			input := minimalSpec()
			input.Spec.Throughput = proto.Int32(450)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an autoscale ceiling off the 1000 RU/s increment", func() {
			input := minimalSpec()
			input.Spec.AutoscaleMaxThroughput = proto.Int32(2500)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject fixed throughput and autoscale together", func() {
			input := minimalSpec()
			input.Spec.Throughput = proto.Int32(400)
			input.Spec.AutoscaleMaxThroughput = proto.Int32(1000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
