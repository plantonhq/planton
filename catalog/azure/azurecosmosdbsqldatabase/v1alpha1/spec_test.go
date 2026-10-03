package azurecosmosdbsqldatabasev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureCosmosdbSqlDatabaseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCosmosdbSqlDatabaseSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.DocumentDB/databaseAccounts/planton-cosmos"

// minimal valid spec: a database sharing no throughput (containers bring
// their own, or the account is serverless).
func minimalSpec() *AzureCosmosdbSqlDatabase {
	return &AzureCosmosdbSqlDatabase{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCosmosdbSqlDatabase",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-sql-database",
		},
		Spec: &AzureCosmosdbSqlDatabaseSpec{
			CosmosdbAccountId: literal(accountId),
			DatabaseName:      "app-data",
		},
	}
}

var _ = ginkgo.Describe("AzureCosmosdbSqlDatabaseSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal database", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept fixed throughput in 100 RU/s increments", func() {
			for _, throughput := range []int32{400, 500, 10000} {
				input := minimalSpec()
				input.Spec.Throughput = proto.Int32(throughput)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "throughput %d must be accepted", throughput)
			}
		})

		ginkgo.It("should accept autoscale ceilings in 1000 RU/s increments", func() {
			for _, ceiling := range []int32{1000, 4000, 100000} {
				input := minimalSpec()
				input.Spec.AutoscaleMaxThroughput = proto.Int32(ceiling)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "ceiling %d must be accepted", ceiling)
			}
		})

		ginkgo.It("should accept a 255-character database name", func() {
			input := minimalSpec()
			name := make([]byte, 255)
			for i := range name {
				name[i] = 'a'
			}
			input.Spec.DatabaseName = string(name)
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

		ginkgo.It("should reject a database name over 255 characters", func() {
			input := minimalSpec()
			name := make([]byte, 256)
			for i := range name {
				name[i] = 'a'
			}
			input.Spec.DatabaseName = string(name)
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

		ginkgo.It("should reject an autoscale ceiling below the 1000 RU/s floor", func() {
			input := minimalSpec()
			input.Spec.AutoscaleMaxThroughput = proto.Int32(500)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an autoscale ceiling off the 1000 RU/s increment", func() {
			input := minimalSpec()
			input.Spec.AutoscaleMaxThroughput = proto.Int32(1500)
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
