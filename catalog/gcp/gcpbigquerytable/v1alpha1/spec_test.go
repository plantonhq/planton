package gcpbigquerytablev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpBigQueryTableSpec Suite")
}

var _ = ginkgo.Describe("GcpBigQueryTableSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	value := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// Helper to build a minimal valid GcpBigQueryTable (a native table).
	minimal := func() *GcpBigQueryTable {
		return &GcpBigQueryTable{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBigQueryTable",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-bq-table",
			},
			Spec: &GcpBigQueryTableSpec{
				DatasetId: value("analytics_prod"),
				TableId:   "events_raw",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal native table (dataset_id + table_id)", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an explicit project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = value("my-gcp-project")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a dataset_id reference (valueFrom)", func() {
		msg := minimal()
		msg.Spec.DatasetId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-dataset"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a schema with metadata, labels, and resource tags", func() {
		msg := minimal()
		msg.Spec.FriendlyName = "Raw Events"
		msg.Spec.Description = "Append-only event stream"
		msg.Spec.Labels = map[string]string{"team": "analytics"}
		msg.Spec.ResourceTags = map[string]string{"123456789012/environment": "production"}
		msg.Spec.Schema = `[{"name":"id","type":"INT64","mode":"REQUIRED"}]`
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept day time-partitioning with clustering and a partition filter", func() {
		msg := minimal()
		msg.Spec.Schema = `[{"name":"event_time","type":"TIMESTAMP"},{"name":"customer_id","type":"INT64"}]`
		msg.Spec.TimePartitioning = &GcpBigQueryTableTimePartitioning{
			Type:         "DAY",
			Field:        "event_time",
			ExpirationMs: 7776000000,
		}
		msg.Spec.Clustering = []string{"customer_id"}
		msg.Spec.RequirePartitionFilter = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every time-partitioning granularity", func() {
		for _, granularity := range []string{"DAY", "HOUR", "MONTH", "YEAR"} {
			msg := minimal()
			msg.Spec.TimePartitioning = &GcpBigQueryTableTimePartitioning{Type: granularity}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should accept integer range partitioning starting at zero", func() {
		msg := minimal()
		msg.Spec.RangePartitioning = &GcpBigQueryTableRangePartitioning{
			Field: "customer_id",
			Range: &GcpBigQueryTableRangePartitioningRange{Start: 0, End: 4000, Interval: 10},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept four clustering columns", func() {
		msg := minimal()
		msg.Spec.Clustering = []string{"a", "b", "c", "d"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a logical view", func() {
		msg := minimal()
		msg.Spec.View = &GcpBigQueryTableView{
			Query: "SELECT customer_id, SUM(amount) AS revenue FROM `p.d.orders` GROUP BY customer_id",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a materialized view with refresh tuning and clustering", func() {
		enableRefresh := true
		msg := minimal()
		msg.Spec.MaterializedView = &GcpBigQueryTableMaterializedView{
			Query:             "SELECT dt, COUNT(*) AS n FROM `p.d.events` GROUP BY dt",
			EnableRefresh:     &enableRefresh,
			RefreshIntervalMs: 900000,
		}
		msg.Spec.Clustering = []string{"dt"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an external GCS parquet table", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect:   true,
			SourceUris:   []string{"gs://my-lake/events/*.parquet"},
			SourceFormat: "PARQUET",
			ParquetOptions: &GcpBigQueryTableParquetOptions{
				EnableListInference: true,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an external CSV table with full csv options", func() {
		quote := `"`
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect:   false,
			SourceUris:   []string{"gs://my-lake/csv/*.csv"},
			SourceFormat: "CSV",
			Schema:       `[{"name":"id","type":"INT64"}]`,
			CsvOptions: &GcpBigQueryTableCsvOptions{
				Quote:           &quote,
				Encoding:        "UTF-8",
				FieldDelimiter:  ",",
				SkipLeadingRows: 1,
			},
			MaxBadRecords:       10,
			IgnoreUnknownValues: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a hive-partitioned external table with a BigLake connection", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect:   true,
			SourceUris:   []string{"gs://my-lake/events/*"},
			SourceFormat: "PARQUET",
			ConnectionId: "projects/p/locations/us/connections/lake",
			HivePartitioningOptions: &GcpBigQueryTableHivePartitioningOptions{
				Mode:            "AUTO",
				SourceUriPrefix: "gs://my-lake/events",
			},
			MetadataCacheMode: "AUTOMATIC",
		}
		msg.Spec.MaxStaleness = "0-0 0 4:0:0"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an object table (object_metadata, no source_format)", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect:     true,
			SourceUris:     []string{"gs://my-images/*"},
			ObjectMetadata: "SIMPLE",
			ConnectionId:   "projects/p/locations/us/connections/lake",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a bigtable-backed external table", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect:   false,
			SourceUris:   []string{"https://googleapis.com/bigtable/projects/p/instances/i/tables/t"},
			SourceFormat: "BIGTABLE",
			BigtableOptions: &GcpBigQueryTableBigtableOptions{
				ReadRowkeyAsString: true,
				ColumnFamilies: []*GcpBigQueryTableBigtableColumnFamily{
					{
						FamilyId: "stats",
						Columns: []*GcpBigQueryTableBigtableColumn{
							{QualifierString: "score", Type: "FLOAT", OnlyReadLatest: true},
						},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept CMEK, expiration, and deletion_protection false", func() {
		deletionProtection := false
		msg := minimal()
		msg.Spec.KmsKeyName = value("projects/p/locations/us/keyRings/r/cryptoKeys/k")
		msg.Spec.ExpirationTime = 1900000000000
		msg.Spec.DeletionProtection = &deletionProtection
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept table constraints with a referenced-table FK ref", func() {
		msg := minimal()
		msg.Spec.TableConstraints = &GcpBigQueryTableConstraints{
			PrimaryKey: &GcpBigQueryTablePrimaryKey{Columns: []string{"id"}},
			ForeignKeys: []*GcpBigQueryTableForeignKey{
				{
					Name: "fk_customer",
					ReferencedTable: &GcpBigQueryTableForeignKeyReferencedTable{
						ProjectId: "my-gcp-project",
						DatasetId: "analytics_prod",
						TableId:   value("customers"),
					},
					ColumnReferences: &GcpBigQueryTableForeignKeyColumnReferences{
						ReferencingColumn: "customer_id",
						ReferencedColumn:  "id",
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept table replication info", func() {
		msg := minimal()
		msg.Spec.TableReplicationInfo = &GcpBigQueryTableReplicationInfo{
			SourceProjectId: "p",
			SourceDatasetId: "d",
			SourceTableId:   "mv",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a BigLake managed-table configuration", func() {
		msg := minimal()
		msg.Spec.BiglakeConfiguration = &GcpBigQueryTableBiglakeConfiguration{
			ConnectionId: "projects/p/locations/us/connections/lake",
			StorageUri:   "gs://my-lake/managed/events",
			FileFormat:   "PARQUET",
			TableFormat:  "ICEBERG",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept external catalog table options", func() {
		msg := minimal()
		msg.Spec.SchemaForeignTypeInfo = &GcpBigQueryTableSchemaForeignTypeInfo{TypeSystem: "HIVE"}
		msg.Spec.ExternalCatalogTableOptions = &GcpBigQueryTableExternalCatalogTableOptions{
			Parameters: map[string]string{"owner": "spark"},
			StorageDescriptor: &GcpBigQueryTableStorageDescriptor{
				LocationUri:  "gs://my-lake/hive/events",
				InputFormat:  "org.apache.hadoop.mapred.TextInputFormat",
				OutputFormat: "org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat",
				SerdeInfo: &GcpBigQueryTableSerDeInfo{
					SerializationLibrary: "org.apache.hadoop.hive.serde2.lazy.LazySimpleSerDe",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject when dataset_id is missing", func() {
		msg := minimal()
		msg.Spec.DatasetId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when table_id is empty", func() {
		msg := minimal()
		msg.Spec.TableId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject table_id with hyphens", func() {
		msg := minimal()
		msg.Spec.TableId = "events-raw"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject table_id exceeding 1024 characters", func() {
		msg := minimal()
		msg.Spec.TableId = strings.Repeat("a", 1025)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject both partitioning methods together", func() {
		msg := minimal()
		msg.Spec.TimePartitioning = &GcpBigQueryTableTimePartitioning{Type: "DAY"}
		msg.Spec.RangePartitioning = &GcpBigQueryTableRangePartitioning{
			Field: "id",
			Range: &GcpBigQueryTableRangePartitioningRange{Start: 0, End: 100, Interval: 10},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid time-partitioning granularity", func() {
		msg := minimal()
		msg.Spec.TimePartitioning = &GcpBigQueryTableTimePartitioning{Type: "WEEK"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject range partitioning with end <= start", func() {
		msg := minimal()
		msg.Spec.RangePartitioning = &GcpBigQueryTableRangePartitioning{
			Field: "id",
			Range: &GcpBigQueryTableRangePartitioningRange{Start: 100, End: 100, Interval: 10},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject range partitioning with a zero interval", func() {
		msg := minimal()
		msg.Spec.RangePartitioning = &GcpBigQueryTableRangePartitioning{
			Field: "id",
			Range: &GcpBigQueryTableRangePartitioningRange{Start: 0, End: 100, Interval: 0},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject five clustering columns", func() {
		msg := minimal()
		msg.Spec.Clustering = []string{"a", "b", "c", "d", "e"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a view combined with a materialized view", func() {
		msg := minimal()
		msg.Spec.View = &GcpBigQueryTableView{Query: "SELECT 1"}
		msg.Spec.MaterializedView = &GcpBigQueryTableMaterializedView{Query: "SELECT 1"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a view combined with external data configuration", func() {
		msg := minimal()
		msg.Spec.View = &GcpBigQueryTableView{Query: "SELECT 1"}
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect: true,
			SourceUris: []string{"gs://b/*"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a logical view carrying a schema", func() {
		msg := minimal()
		msg.Spec.View = &GcpBigQueryTableView{Query: "SELECT 1"}
		msg.Spec.Schema = `[{"name":"x","type":"INT64"}]`
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a logical view carrying partitioning", func() {
		msg := minimal()
		msg.Spec.View = &GcpBigQueryTableView{Query: "SELECT 1"}
		msg.Spec.TimePartitioning = &GcpBigQueryTableTimePartitioning{Type: "DAY"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a materialized view carrying a schema", func() {
		msg := minimal()
		msg.Spec.MaterializedView = &GcpBigQueryTableMaterializedView{Query: "SELECT 1"}
		msg.Spec.Schema = `[{"name":"x","type":"INT64"}]`
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a view without a query", func() {
		msg := minimal()
		msg.Spec.View = &GcpBigQueryTableView{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject external data configuration without source URIs", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect:   true,
			SourceFormat: "PARQUET",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid source format", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect:   true,
			SourceUris:   []string{"gs://b/*"},
			SourceFormat: "XML",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject source_format combined with object_metadata", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect:     true,
			SourceUris:     []string{"gs://b/*"},
			SourceFormat:   "PARQUET",
			ObjectMetadata: "SIMPLE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid hive partitioning mode", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect: true,
			SourceUris: []string{"gs://b/*"},
			HivePartitioningOptions: &GcpBigQueryTableHivePartitioningOptions{
				Mode: "GUESS",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a bigtable column with both qualifier forms", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			Autodetect:   false,
			SourceUris:   []string{"https://googleapis.com/bigtable/projects/p/instances/i/tables/t"},
			SourceFormat: "BIGTABLE",
			BigtableOptions: &GcpBigQueryTableBigtableOptions{
				ColumnFamilies: []*GcpBigQueryTableBigtableColumnFamily{
					{
						FamilyId: "stats",
						Columns: []*GcpBigQueryTableBigtableColumn{
							{QualifierString: "score", QualifierEncoded: "c2NvcmU="},
						},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty primary key", func() {
		msg := minimal()
		msg.Spec.TableConstraints = &GcpBigQueryTableConstraints{
			PrimaryKey: &GcpBigQueryTablePrimaryKey{Columns: []string{}},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a foreign key without column references", func() {
		msg := minimal()
		msg.Spec.TableConstraints = &GcpBigQueryTableConstraints{
			ForeignKeys: []*GcpBigQueryTableForeignKey{
				{
					ReferencedTable: &GcpBigQueryTableForeignKeyReferencedTable{
						ProjectId: "p", DatasetId: "d", TableId: value("t"),
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject replication info missing the source table", func() {
		msg := minimal()
		msg.Spec.TableReplicationInfo = &GcpBigQueryTableReplicationInfo{
			SourceProjectId: "p",
			SourceDatasetId: "d",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a BigLake configuration missing the storage URI", func() {
		msg := minimal()
		msg.Spec.BiglakeConfiguration = &GcpBigQueryTableBiglakeConfiguration{
			ConnectionId: "projects/p/locations/us/connections/lake",
			FileFormat:   "PARQUET",
			TableFormat:  "ICEBERG",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject serde info without a serialization library", func() {
		msg := minimal()
		msg.Spec.ExternalCatalogTableOptions = &GcpBigQueryTableExternalCatalogTableOptions{
			StorageDescriptor: &GcpBigQueryTableStorageDescriptor{
				SerdeInfo: &GcpBigQueryTableSerDeInfo{Name: "no-library"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when metadata is missing", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when spec is missing", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = v
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept the schema-diff toggles", func() {
		msg := minimal()
		msg.Spec.IgnoreAutoGeneratedSchema = true
		msg.Spec.IgnoreSchemaChanges = []string{"dataPolicies"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unsupported ignore_schema_changes entry", func() {
		msg := minimal()
		msg.Spec.IgnoreSchemaChanges = []string{"columnDescriptions"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each table_metadata_view value", func() {
		for _, v := range []string{"TABLE_METADATA_VIEW_UNSPECIFIED", "BASIC", "STORAGE_STATS", "FULL"} {
			msg := minimal()
			msg.Spec.TableMetadataView = v
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should reject an invalid table_metadata_view", func() {
		msg := minimal()
		msg.Spec.TableMetadataView = "EVERYTHING"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept external data with decimal target types and CSV column matching", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			SourceUris:         []string{"gs://my-bucket/data/*.csv"},
			SourceFormat:       "CSV",
			DecimalTargetTypes: []string{"NUMERIC", "BIGNUMERIC", "STRING"},
			CsvOptions: &GcpBigQueryTableCsvOptions{
				SourceColumnMatch: "NAME",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid decimal target type", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			SourceUris:         []string{"gs://my-bucket/data/*.csv"},
			SourceFormat:       "CSV",
			DecimalTargetTypes: []string{"FLOAT64"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid csv source_column_match", func() {
		msg := minimal()
		msg.Spec.ExternalDataConfiguration = &GcpBigQueryTableExternalDataConfiguration{
			SourceUris:   []string{"gs://my-bucket/data/*.csv"},
			SourceFormat: "CSV",
			CsvOptions: &GcpBigQueryTableCsvOptions{
				SourceColumnMatch: "ORDER",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
