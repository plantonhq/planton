package azuredatafactorydatasetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureDataFactoryDatasetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataFactoryDatasetSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testFactoryId       = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DataFactory/factories/app-df"
	testLinkedServiceId = testFactoryId + "/linkedservices/sqldb-conn"
)

// blobLocation builds a minimal valid blob storage location.
func blobLocation() *AzureDataFactoryDatasetBlobStorageLocation {
	return &AzureDataFactoryDatasetBlobStorageLocation{
		Container: "landing",
		Path:      "raw/orders",
		Filename:  "orders.csv",
	}
}

// httpLocation builds an HTTP server location with all fields (the
// strictest formats require path and filename).
func httpLocation() *AzureDataFactoryDatasetHttpServerLocation {
	return &AzureDataFactoryDatasetHttpServerLocation{
		RelativeUrl: "exports/daily",
		Path:        "2026/08",
		Filename:    "orders.csv",
	}
}

// validResource returns a valid delimited text dataset (the flagship
// variant) that individual cases mutate into the shape under test.
func validResource() *AzureDataFactoryDataset {
	return &AzureDataFactoryDataset{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataFactoryDataset",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-adf-dataset",
		},
		Spec: &AzureDataFactoryDatasetSpec{
			DataFactoryId:     literal(testFactoryId),
			Name:              "orders-csv",
			LinkedServiceName: literal("blob-conn"),
			Variant: &AzureDataFactoryDatasetSpec_DelimitedText{
				DelimitedText: &AzureDataFactoryDatasetDelimitedText{
					AzureBlobStorageLocation: blobLocation(),
					FirstRowAsHeader:         proto.Bool(true),
				},
			},
		},
	}
}

// withoutVariant clears the fixture's delimited_text variant so a
// case can install a different one.
func withoutVariant() *AzureDataFactoryDataset {
	input := validResource()
	input.Spec.Variant = nil
	return input
}

var _ = ginkgo.Describe("AzureDataFactoryDatasetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("variant selection -- one minimal document per variant", func() {

			ginkgo.It("should accept a delimited text dataset on blob storage", func() {
				gomega.Expect(protovalidate.Validate(validResource())).To(gomega.BeNil())
			})

			ginkgo.It("should accept an azure blob dataset", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_AzureBlob{AzureBlob: &AzureDataFactoryDatasetAzureBlob{
					Path:     "raw/orders",
					Filename: "orders.csv",
					SchemaColumn: []*AzureDataFactoryDatasetSchemaColumn{
						{Name: "order_id", Type: "Int64"},
						{Name: "placed_at", Type: "DateTime", Description: "Order timestamp"},
					},
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an azure sql table dataset referencing its linked service by ID", func() {
				input := withoutVariant()
				input.Spec.LinkedServiceName = nil
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_AzureSqlTable{AzureSqlTable: &AzureDataFactoryDatasetAzureSqlTable{
					LinkedServiceId: literal(testLinkedServiceId),
					Schema:          "dbo",
					Table:           "orders",
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a binary dataset on SFTP with compression", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Binary{Binary: &AzureDataFactoryDatasetBinary{
					SftpServerLocation: &AzureDataFactoryDatasetSftpServerLocation{
						Path:     "outbound",
						Filename: "archive.tar.gz",
					},
					Compression: &AzureDataFactoryDatasetBinaryCompression{
						Type:  "TarGZip",
						Level: "Optimal",
					},
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a cosmosdb SQL API dataset", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_CosmosdbSqlapi{CosmosdbSqlapi: &AzureDataFactoryDatasetCosmosdbSqlapi{
					CollectionName: "orders",
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom dataset carrying its own linked service reference", func() {
				input := withoutVariant()
				input.Spec.LinkedServiceName = nil
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Custom{Custom: &AzureDataFactoryDatasetCustom{
					LinkedService: &AzureDataFactoryDatasetCustomLinkedService{
						Name:       literal("blob-conn"),
						Parameters: map[string]string{"container": "landing"},
					},
					Type:               "Excel",
					TypePropertiesJson: `{"location":{"type":"AzureBlobStorageLocation","container":"landing"},"sheetName":"Sheet1"}`,
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an http dataset", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Http{Http: &AzureDataFactoryDatasetHttp{
					RelativeUrl:   "exports/daily.csv",
					RequestMethod: "GET",
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a json dataset on blob storage with path and filename", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Json{Json: &AzureDataFactoryDatasetJson{
					AzureBlobStorageLocation: blobLocation(),
					Encoding:                 "UTF-8",
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a mysql dataset", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Mysql{Mysql: &AzureDataFactoryDatasetMysql{
					TableName: "orders",
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a parquet dataset on Data Lake Gen2", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Parquet{Parquet: &AzureDataFactoryDatasetParquet{
					AzureBlobFsLocation: &AzureDataFactoryDatasetBlobFsLocation{
						FileSystem: "lake",
						Path:       "curated/orders",
					},
					CompressionCodec: "snappy",
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a postgresql dataset", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Postgresql{Postgresql: &AzureDataFactoryDatasetPostgresql{
					TableName: "public.orders",
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a snowflake dataset with snowflake-typed columns", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Snowflake{Snowflake: &AzureDataFactoryDatasetSnowflake{
					TableName:  "ORDERS",
					SchemaName: "PUBLIC",
					SchemaColumn: []*AzureDataFactoryDatasetSnowflakeSchemaColumn{
						{Name: "ORDER_ID", Type: "NUMBER", Precision: 38},
						{Name: "TOTAL", Type: "DECIMAL", Precision: 12, Scale: 2},
					},
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a sql server table dataset", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_SqlServerTable{SqlServerTable: &AzureDataFactoryDatasetSqlServerTable{
					TableName: "dbo.orders",
				}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("shared root fields", func() {

			ginkgo.It("should accept description, annotations, parameters, additional properties, and folder", func() {
				input := validResource()
				input.Spec.Description = "Daily order exports"
				input.Spec.Annotations = []string{"orders", "daily"}
				input.Spec.Parameters = map[string]string{"window": "2026-08-13"}
				input.Spec.AdditionalProperties = map[string]string{"customKey": "customValue"}
				input.Spec.Folder = "ingest/orders"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name mixing special characters with regular ones", func() {
				input := validResource()
				input.Spec.Name = "orders-2026.csv"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept delimited text parse settings", func() {
				input := validResource()
				input.Spec.GetDelimitedText().ColumnDelimiter = ";"
				input.Spec.GetDelimitedText().RowDelimiter = "\n"
				input.Spec.GetDelimitedText().QuoteCharacter = "'"
				input.Spec.GetDelimitedText().EscapeCharacter = "/"
				input.Spec.GetDelimitedText().Encoding = "UTF-8"
				input.Spec.GetDelimitedText().NullValue = "NULL"
				input.Spec.GetDelimitedText().CompressionCodec = "gzip"
				input.Spec.GetDelimitedText().CompressionLevel = "Fastest"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept dynamic expression flags on locations", func() {
				input := validResource()
				input.Spec.Parameters = map[string]string{"runDate": ""}
				input.Spec.GetDelimitedText().AzureBlobStorageLocation.Path = "raw/@{dataset().runDate}"
				input.Spec.GetDelimitedText().AzureBlobStorageLocation.DynamicPathEnabled = true
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("root contracts", func() {

			ginkgo.It("should reject a spec with no variant block", func() {
				input := withoutVariant()
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("carries exactly one variant by construction (setting a second arm replaces the first)", func() {
				input := validResource()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Mysql{Mysql: &AzureDataFactoryDatasetMysql{TableName: "orders"}}
				gomega.Expect(input.Spec.GetDelimitedText()).To(gomega.BeNil())
				gomega.Expect(input.Spec.GetMysql()).NotTo(gomega.BeNil())
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("accepts a variant whose selection is its only content (an HTTP file with every setting left to Data Factory)", func() {
				input := validResource()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Http{Http: &AzureDataFactoryDatasetHttp{}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should reject a name-wired variant without linked_service_name", func() {
				input := validResource()
				input.Spec.LinkedServiceName = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject linked_service_name alongside azure_sql_table", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_AzureSqlTable{AzureSqlTable: &AzureDataFactoryDatasetAzureSqlTable{
					LinkedServiceId: literal(testLinkedServiceId),
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject linked_service_name alongside custom", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Custom{Custom: &AzureDataFactoryDatasetCustom{
					LinkedService:      &AzureDataFactoryDatasetCustomLinkedService{Name: literal("blob-conn")},
					Type:               "Excel",
					TypePropertiesJson: `{}`,
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing data_factory_id", func() {
				input := validResource()
				input.Spec.DataFactoryId = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name consisting entirely of special characters", func() {
				input := validResource()
				input.Spec.Name = "-.+?/<>*%&:\\"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty annotation string", func() {
				input := validResource()
				input.Spec.Annotations = []string{""}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("location contracts", func() {

			ginkgo.It("should reject a delimited text dataset with no location", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_DelimitedText{DelimitedText: &AzureDataFactoryDatasetDelimitedText{}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a delimited text dataset with two locations", func() {
				input := validResource()
				input.Spec.GetDelimitedText().HttpServerLocation = httpLocation()
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a delimited text HTTP location missing path", func() {
				input := withoutVariant()
				loc := httpLocation()
				loc.Path = ""
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_DelimitedText{DelimitedText: &AzureDataFactoryDatasetDelimitedText{HttpServerLocation: loc}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a binary dataset with no location", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Binary{Binary: &AzureDataFactoryDatasetBinary{}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a binary HTTP location missing filename", func() {
				input := withoutVariant()
				loc := httpLocation()
				loc.Filename = ""
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Binary{Binary: &AzureDataFactoryDatasetBinary{HttpServerLocation: loc}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a json dataset with no location", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Json{Json: &AzureDataFactoryDatasetJson{}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a json blob location missing filename", func() {
				input := withoutVariant()
				loc := blobLocation()
				loc.Filename = ""
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Json{Json: &AzureDataFactoryDatasetJson{AzureBlobStorageLocation: loc}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a parquet HTTP location missing filename", func() {
				input := withoutVariant()
				loc := httpLocation()
				loc.Filename = ""
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Parquet{Parquet: &AzureDataFactoryDatasetParquet{HttpServerLocation: loc}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should accept a parquet HTTP location without a path (the one format where it is optional)", func() {
				input := withoutVariant()
				loc := httpLocation()
				loc.Path = ""
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Parquet{Parquet: &AzureDataFactoryDatasetParquet{HttpServerLocation: loc}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should reject an HTTP location missing relative_url", func() {
				input := withoutVariant()
				loc := httpLocation()
				loc.RelativeUrl = ""
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_DelimitedText{DelimitedText: &AzureDataFactoryDatasetDelimitedText{HttpServerLocation: loc}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a blob storage location missing container", func() {
				input := validResource()
				input.Spec.GetDelimitedText().AzureBlobStorageLocation.Container = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an SFTP location missing path", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Binary{Binary: &AzureDataFactoryDatasetBinary{
					SftpServerLocation: &AzureDataFactoryDatasetSftpServerLocation{Filename: "archive.bin"},
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("variant field contracts", func() {

			ginkgo.It("should reject a binary compression without a type", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Binary{Binary: &AzureDataFactoryDatasetBinary{
					AzureBlobStorageLocation: blobLocation(),
					Compression:              &AzureDataFactoryDatasetBinaryCompression{Level: "Optimal"},
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a binary compression with an unknown codec", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Binary{Binary: &AzureDataFactoryDatasetBinary{
					AzureBlobStorageLocation: blobLocation(),
					Compression:              &AzureDataFactoryDatasetBinaryCompression{Type: "snappy"},
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a delimited text compression codec outside the vocabulary", func() {
				input := validResource()
				input.Spec.GetDelimitedText().CompressionCodec = "zstd"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a parquet compression codec of None (delimited-text-only token)", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Parquet{Parquet: &AzureDataFactoryDatasetParquet{
					AzureBlobStorageLocation: blobLocation(),
					CompressionCodec:         "None",
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a compression level outside the vocabulary", func() {
				input := validResource()
				input.Spec.GetDelimitedText().CompressionLevel = "Maximum"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a schema column without a name", func() {
				input := validResource()
				input.Spec.GetDelimitedText().SchemaColumn = []*AzureDataFactoryDatasetSchemaColumn{{Type: "String"}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a schema column type outside the interim vocabulary", func() {
				input := validResource()
				input.Spec.GetDelimitedText().SchemaColumn = []*AzureDataFactoryDatasetSchemaColumn{{Name: "order_id", Type: "BIGINT"}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a snowflake column type outside snowflake's vocabulary", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Snowflake{Snowflake: &AzureDataFactoryDatasetSnowflake{
					SchemaColumn: []*AzureDataFactoryDatasetSnowflakeSchemaColumn{{Name: "ORDER_ID", Type: "Int64"}},
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a negative snowflake column precision", func() {
				input := withoutVariant()
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Snowflake{Snowflake: &AzureDataFactoryDatasetSnowflake{
					SchemaColumn: []*AzureDataFactoryDatasetSnowflakeSchemaColumn{{Name: "ORDER_ID", Type: "NUMBER", Precision: -1}},
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an azure sql table dataset without linked_service_id", func() {
				input := withoutVariant()
				input.Spec.LinkedServiceName = nil
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_AzureSqlTable{AzureSqlTable: &AzureDataFactoryDatasetAzureSqlTable{Table: "orders"}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a custom dataset without a type", func() {
				input := withoutVariant()
				input.Spec.LinkedServiceName = nil
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Custom{Custom: &AzureDataFactoryDatasetCustom{
					LinkedService:      &AzureDataFactoryDatasetCustomLinkedService{Name: literal("blob-conn")},
					TypePropertiesJson: `{}`,
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a custom dataset without type properties", func() {
				input := withoutVariant()
				input.Spec.LinkedServiceName = nil
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Custom{Custom: &AzureDataFactoryDatasetCustom{
					LinkedService: &AzureDataFactoryDatasetCustomLinkedService{Name: literal("blob-conn")},
					Type:          "Excel",
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a custom dataset without a linked service block", func() {
				input := withoutVariant()
				input.Spec.LinkedServiceName = nil
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Custom{Custom: &AzureDataFactoryDatasetCustom{
					Type:               "Excel",
					TypePropertiesJson: `{}`,
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a custom linked service block without a name", func() {
				input := withoutVariant()
				input.Spec.LinkedServiceName = nil
				input.Spec.Variant = &AzureDataFactoryDatasetSpec_Custom{Custom: &AzureDataFactoryDatasetCustom{
					LinkedService:      &AzureDataFactoryDatasetCustomLinkedService{},
					Type:               "Excel",
					TypePropertiesJson: `{}`,
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
