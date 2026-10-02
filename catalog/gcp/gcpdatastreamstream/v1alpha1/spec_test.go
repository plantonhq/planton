package gcpdatastreamstreamv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDatastreamStreamSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

const (
	sourceProfile      = "projects/p/locations/us-central1/connectionProfiles/orders-postgres"
	destinationProfile = "projects/p/locations/us-central1/connectionProfiles/bigquery"
)

var _ = ginkgo.Describe("GcpDatastreamStreamSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	postgresRdbms := func() *GcpDatastreamStreamPostgresqlRdbms {
		return &GcpDatastreamStreamPostgresqlRdbms{
			PostgresqlSchemas: []*GcpDatastreamStreamPostgresqlSchema{{
				Schema: "public",
				PostgresqlTables: []*GcpDatastreamStreamPostgresqlTable{{
					Table:             "orders",
					PostgresqlColumns: []*GcpDatastreamStreamPostgresqlColumn{{Column: "id", PrimaryKey: true}},
				}},
			}},
		}
	}
	bigquery := func() *GcpDatastreamStreamDestinationConfig {
		return &GcpDatastreamStreamDestinationConfig{
			DestinationConnectionProfile: litRef(destinationProfile),
			BigqueryDestinationConfig: &GcpDatastreamStreamBigqueryDestinationConfig{
				SourceHierarchyDatasets: &GcpDatastreamStreamSourceHierarchyDatasets{
					DatasetTemplate: &GcpDatastreamStreamDatasetTemplate{Location: "US", DatasetIdPrefix: "orders"},
				},
				DataFreshness: "900s",
				WriteMode:     "MERGE",
			},
		}
	}
	base := func() *GcpDatastreamStream {
		return &GcpDatastreamStream{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDatastreamStream",
			Metadata:   &shared.CloudResourceMetadata{Name: "orders-cdc"},
			Spec: &GcpDatastreamStreamSpec{
				Location: "us-central1",
				SourceConfig: &GcpDatastreamStreamSourceConfig{
					SourceConnectionProfile: litRef(sourceProfile),
					PostgresqlSourceConfig: &GcpDatastreamStreamPostgresqlSourceConfig{
						IncludeObjects:  postgresRdbms(),
						ReplicationSlot: "datastream_slot",
						Publication:     "datastream_pub",
					},
				},
				DestinationConfig: bigquery(),
				BackfillNone:      true,
			},
		}
	}

	ginkgo.It("should accept a complete PostgreSQL-to-BigQuery stream", func() {
		r := base()
		r.Spec.ProjectId = litRef("p")
		r.Spec.StreamId = "orders-cdc"
		r.Spec.DisplayName = "Orders CDC"
		r.Spec.Labels = map[string]string{"team": "data"}
		r.Spec.BackfillNone = false
		r.Spec.BackfillAll = &GcpDatastreamStreamBackfillAll{PostgresqlExcludedObjects: postgresRdbms()}
		r.Spec.CustomerManagedEncryptionKey = litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k")
		r.Spec.DesiredState = "RUNNING"
		r.Spec.CreateWithoutValidation = true
		r.Spec.DeletionPolicy = "PREVENT"
		r.Spec.RuleSets = []*GcpDatastreamStreamRuleSet{{
			ObjectFilter: &GcpDatastreamStreamObjectFilter{SourceObjectIdentifier: &GcpDatastreamStreamSourceObjectIdentifier{
				PostgresqlIdentifier: &GcpDatastreamStreamSchemaTableIdentifier{Schema: "public", Table: "orders"},
			}},
			CustomizationRules: []*GcpDatastreamStreamCustomizationRule{
				{BigqueryPartitioning: &GcpDatastreamStreamBigqueryPartitioning{
					TimeUnitPartition:      &GcpDatastreamStreamTimeUnitPartition{Column: "created_at", PartitioningTimeGranularity: "PARTITIONING_TIME_GRANULARITY_DAY"},
					RequirePartitionFilter: true,
				}},
				{BigqueryClustering: &GcpDatastreamStreamBigqueryClustering{Columns: []string{"customer_id"}}},
			},
		}}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each source arm on its own", func() {
		arms := []func(*GcpDatastreamStreamSourceConfig){
			func(s *GcpDatastreamStreamSourceConfig) {
				s.MysqlSourceConfig = &GcpDatastreamStreamMysqlSourceConfig{
					IncludeObjects: &GcpDatastreamStreamMysqlRdbms{MysqlDatabases: []*GcpDatastreamStreamMysqlDatabase{{
						Database: "shop", MysqlTables: []*GcpDatastreamStreamMysqlTable{{Table: "orders"}},
					}}},
					MaxConcurrentBackfillTasks: 8, MaxConcurrentCdcTasks: 4, CdcMethod: "GTID",
				}
			},
			func(s *GcpDatastreamStreamSourceConfig) {
				s.OracleSourceConfig = &GcpDatastreamStreamOracleSourceConfig{
					IncludeObjects:       &GcpDatastreamStreamOracleRdbms{OracleSchemas: []*GcpDatastreamStreamOracleSchema{{Schema: "HR"}}},
					LargeObjectsHandling: "STREAM",
				}
			},
			func(s *GcpDatastreamStreamSourceConfig) {
				s.SqlServerSourceConfig = &GcpDatastreamStreamSqlServerSourceConfig{
					IncludeObjects: &GcpDatastreamStreamSqlServerRdbms{Schemas: []*GcpDatastreamStreamSqlServerSchema{{Schema: "dbo"}}},
					CdcMethod:      "CHANGE_TABLES",
				}
			},
			func(s *GcpDatastreamStreamSourceConfig) {
				s.MongodbSourceConfig = &GcpDatastreamStreamMongodbSourceConfig{
					IncludeObjects:             &GcpDatastreamStreamMongodbCluster{Databases: []*GcpDatastreamStreamMongodbDatabase{{Database: "app"}}},
					MaxConcurrentBackfillTasks: 50,
				}
			},
			func(s *GcpDatastreamStreamSourceConfig) {
				s.SalesforceSourceConfig = &GcpDatastreamStreamSalesforceSourceConfig{
					IncludeObjects:  &GcpDatastreamStreamSalesforceOrg{Objects: []*GcpDatastreamStreamSalesforceObject{{ObjectName: "Account"}}},
					PollingInterval: "300s",
				}
			},
			func(s *GcpDatastreamStreamSourceConfig) {
				s.SpannerSourceConfig = &GcpDatastreamStreamSpannerSourceConfig{
					ChangeStreamName: "orders_changes", BackfillDataBoostEnabled: true, SpannerRpcPriority: "LOW", FgacRole: "analyst",
				}
			},
		}
		for i, arm := range arms {
			r := base()
			r.Spec.SourceConfig.PostgresqlSourceConfig = nil
			arm(r.Spec.SourceConfig)
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "arm %d", i)
		}
	})

	ginkgo.It("should reject no source arm and two source arms", func() {
		r := base()
		r.Spec.SourceConfig.PostgresqlSourceConfig = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
		r = base()
		r.Spec.SourceConfig.SpannerSourceConfig = &GcpDatastreamStreamSpannerSourceConfig{ChangeStreamName: "c"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a missing source profile, destination, or source config", func() {
		r := base()
		r.Spec.SourceConfig.SourceConnectionProfile = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "source profile")
		r = base()
		r.Spec.DestinationConfig = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "destination")
		r = base()
		r.Spec.SourceConfig = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "source")
	})

	ginkgo.It("should require exactly one backfill mode", func() {
		r := base()
		r.Spec.BackfillNone = false
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "neither")
		r.Spec.BackfillNone = true
		r.Spec.BackfillAll = &GcpDatastreamStreamBackfillAll{}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "both")
		r.Spec.BackfillNone = false
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "backfill everything")
	})

	ginkgo.It("should require names in MongoDB and Spanner backfill exclusions only", func() {
		r := base()
		r.Spec.BackfillNone = false
		r.Spec.BackfillAll = &GcpDatastreamStreamBackfillAll{MongodbExcludedObjects: &GcpDatastreamStreamMongodbCluster{}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "no databases")
		r.Spec.BackfillAll.MongodbExcludedObjects.Databases = []*GcpDatastreamStreamMongodbDatabase{{
			Database: "app", Collections: []*GcpDatastreamStreamMongodbCollection{{}},
		}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "unnamed collection")
		r.Spec.BackfillAll.MongodbExcludedObjects.Databases[0].Collections[0].Collection = "events"
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "named")

		r.Spec.BackfillAll = &GcpDatastreamStreamBackfillAll{SpannerExcludedObjects: &GcpDatastreamStreamSpannerDatabase{
			Schemas: []*GcpDatastreamStreamSpannerSchema{{Schema: "s", Tables: []*GcpDatastreamStreamSpannerTable{{
				Table: "t", Columns: []*GcpDatastreamStreamSpannerColumn{{}},
			}}}},
		}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "unnamed Spanner column")

		// The same unnamed levels are fine in a source's include list.
		r = base()
		r.Spec.SourceConfig.PostgresqlSourceConfig = nil
		r.Spec.SourceConfig.MongodbSourceConfig = &GcpDatastreamStreamMongodbSourceConfig{
			IncludeObjects: &GcpDatastreamStreamMongodbCluster{Databases: []*GcpDatastreamStreamMongodbDatabase{{
				Collections: []*GcpDatastreamStreamMongodbCollection{{}},
			}}},
		}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "include list")
	})

	ginkgo.It("should reject empty object sets where Google requires a first level", func() {
		r := base()
		r.Spec.SourceConfig.PostgresqlSourceConfig.IncludeObjects = &GcpDatastreamStreamPostgresqlRdbms{}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "no schemas")
		r = base()
		r.Spec.SourceConfig.PostgresqlSourceConfig.IncludeObjects.PostgresqlSchemas[0].Schema = ""
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "unnamed schema")
	})

	ginkgo.It("should reject a missing publication or replication slot", func() {
		r := base()
		r.Spec.SourceConfig.PostgresqlSourceConfig.Publication = ""
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
		r = base()
		r.Spec.SourceConfig.PostgresqlSourceConfig.ReplicationSlot = ""
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject values outside Google's lists and ranges", func() {
		cases := map[string]func(*GcpDatastreamStreamSpec){
			"mysql cdc method": func(s *GcpDatastreamStreamSpec) {
				s.SourceConfig.PostgresqlSourceConfig = nil
				s.SourceConfig.MysqlSourceConfig = &GcpDatastreamStreamMysqlSourceConfig{CdcMethod: "BINLOG"}
			},
			"sql server cdc method": func(s *GcpDatastreamStreamSpec) {
				s.SourceConfig.PostgresqlSourceConfig = nil
				s.SourceConfig.SqlServerSourceConfig = &GcpDatastreamStreamSqlServerSourceConfig{CdcMethod: "GTID"}
			},
			"oracle large objects": func(s *GcpDatastreamStreamSpec) {
				s.SourceConfig.PostgresqlSourceConfig = nil
				s.SourceConfig.OracleSourceConfig = &GcpDatastreamStreamOracleSourceConfig{LargeObjectsHandling: "KEEP"}
			},
			"mongodb concurrency": func(s *GcpDatastreamStreamSpec) {
				s.SourceConfig.PostgresqlSourceConfig = nil
				s.SourceConfig.MongodbSourceConfig = &GcpDatastreamStreamMongodbSourceConfig{MaxConcurrentBackfillTasks: 51}
			},
			"spanner priority": func(s *GcpDatastreamStreamSpec) {
				s.SourceConfig.PostgresqlSourceConfig = nil
				s.SourceConfig.SpannerSourceConfig = &GcpDatastreamStreamSpannerSourceConfig{ChangeStreamName: "c", SpannerRpcPriority: "URGENT"}
			},
			"spanner change stream": func(s *GcpDatastreamStreamSpec) {
				s.SourceConfig.PostgresqlSourceConfig = nil
				s.SourceConfig.SpannerSourceConfig = &GcpDatastreamStreamSpannerSourceConfig{}
			},
			"salesforce polling interval": func(s *GcpDatastreamStreamSpec) {
				s.SourceConfig.PostgresqlSourceConfig = nil
				s.SourceConfig.SalesforceSourceConfig = &GcpDatastreamStreamSalesforceSourceConfig{PollingInterval: "5m"}
			},
			"negative concurrency": func(s *GcpDatastreamStreamSpec) {
				s.SourceConfig.PostgresqlSourceConfig.MaxConcurrentBackfillTasks = -1
			},
			"write mode": func(s *GcpDatastreamStreamSpec) {
				s.DestinationConfig.BigqueryDestinationConfig.WriteMode = "UPSERT"
			},
			"data freshness": func(s *GcpDatastreamStreamSpec) {
				s.DestinationConfig.BigqueryDestinationConfig.DataFreshness = "15 minutes"
			},
			"desired state": func(s *GcpDatastreamStreamSpec) { s.DesiredState = "STOPPED" },
			"deletion policy": func(s *GcpDatastreamStreamSpec) { s.DeletionPolicy = "FORCE" },
			"location": func(s *GcpDatastreamStreamSpec) { s.Location = "US" },
		}
		for name, mutate := range cases {
			r := base()
			mutate(r.Spec)
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), name)
		}
	})

	ginkgo.It("should require exactly one destination arm and one BigQuery dataset config", func() {
		r := base()
		r.Spec.DestinationConfig.GcsDestinationConfig = &GcpDatastreamStreamGcsDestinationConfig{AvroFileFormat: true}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "two destinations")

		r = base()
		r.Spec.DestinationConfig.BigqueryDestinationConfig = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "no destination arm")

		r = base()
		r.Spec.DestinationConfig.BigqueryDestinationConfig.SingleTargetDataset = &GcpDatastreamStreamSingleTargetDataset{
			DatasetId: litRef("projects/p/datasets/orders"),
		}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "two dataset configs")

		r.Spec.DestinationConfig.BigqueryDestinationConfig.SourceHierarchyDatasets = nil
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "single dataset")

		r.Spec.DestinationConfig.BigqueryDestinationConfig.SingleTargetDataset = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "no dataset config")
	})

	ginkgo.It("should accept BigLake managed tables and reject values Google does not offer", func() {
		r := base()
		r.Spec.DestinationConfig.BigqueryDestinationConfig.BlmtConfig = &GcpDatastreamStreamBlmtConfig{
			Bucket: litRef("lake"), RootPath: "/iceberg", ConnectionName: litRef("p.us-central1.lake"),
			FileFormat: "PARQUET", TableFormat: "ICEBERG",
		}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		r.Spec.DestinationConfig.BigqueryDestinationConfig.BlmtConfig.TableFormat = "DELTA"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should enforce the Cloud Storage file rules", func() {
		gcs := func(g *GcpDatastreamStreamGcsDestinationConfig) *GcpDatastreamStream {
			r := base()
			r.Spec.DestinationConfig.BigqueryDestinationConfig = nil
			r.Spec.DestinationConfig.GcsDestinationConfig = g
			return r
		}
		gomega.Expect(validator.Validate(gcs(&GcpDatastreamStreamGcsDestinationConfig{
			Path: "/orders", FileRotationInterval: "60s", FileRotationMb: 50, AvroFileFormat: true,
		}))).To(gomega.Succeed(), "avro")
		gomega.Expect(validator.Validate(gcs(&GcpDatastreamStreamGcsDestinationConfig{
			JsonFileFormat: &GcpDatastreamStreamGcsJsonFileFormat{Compression: "GZIP", SchemaFileFormat: "AVRO_SCHEMA_FILE"},
		}))).To(gomega.Succeed(), "json")
		gomega.Expect(validator.Validate(gcs(&GcpDatastreamStreamGcsDestinationConfig{}))).ToNot(gomega.Succeed(), "no format")
		gomega.Expect(validator.Validate(gcs(&GcpDatastreamStreamGcsDestinationConfig{
			AvroFileFormat: true, JsonFileFormat: &GcpDatastreamStreamGcsJsonFileFormat{},
		}))).ToNot(gomega.Succeed(), "two formats")
		for _, interval := range []string{"14s", "61s", "1m"} {
			gomega.Expect(validator.Validate(gcs(&GcpDatastreamStreamGcsDestinationConfig{
				AvroFileFormat: true, FileRotationInterval: interval,
			}))).ToNot(gomega.Succeed(), interval)
		}
		gomega.Expect(validator.Validate(gcs(&GcpDatastreamStreamGcsDestinationConfig{
			JsonFileFormat: &GcpDatastreamStreamGcsJsonFileFormat{Compression: "ZSTD"},
		}))).ToNot(gomega.Succeed(), "compression")
	})

	ginkgo.It("should enforce the rule-set unions", func() {
		identifier := &GcpDatastreamStreamSourceObjectIdentifier{
			PostgresqlIdentifier: &GcpDatastreamStreamSchemaTableIdentifier{Schema: "public", Table: "orders"},
		}
		clustering := &GcpDatastreamStreamCustomizationRule{BigqueryClustering: &GcpDatastreamStreamBigqueryClustering{Columns: []string{"id"}}}
		withRules := func(rs *GcpDatastreamStreamRuleSet) *GcpDatastreamStream {
			r := base()
			r.Spec.RuleSets = []*GcpDatastreamStreamRuleSet{rs}
			return r
		}

		gomega.Expect(validator.Validate(withRules(&GcpDatastreamStreamRuleSet{
			ObjectFilter: &GcpDatastreamStreamObjectFilter{SourceObjectIdentifier: identifier},
		}))).ToNot(gomega.Succeed(), "no rules")

		gomega.Expect(validator.Validate(withRules(&GcpDatastreamStreamRuleSet{
			CustomizationRules: []*GcpDatastreamStreamCustomizationRule{clustering},
		}))).ToNot(gomega.Succeed(), "no object filter")

		gomega.Expect(validator.Validate(withRules(&GcpDatastreamStreamRuleSet{
			ObjectFilter: &GcpDatastreamStreamObjectFilter{SourceObjectIdentifier: &GcpDatastreamStreamSourceObjectIdentifier{
				PostgresqlIdentifier: identifier.PostgresqlIdentifier,
				MysqlIdentifier:      &GcpDatastreamStreamMysqlIdentifier{Database: "d", Table: "t"},
			}},
			CustomizationRules: []*GcpDatastreamStreamCustomizationRule{clustering},
		}))).ToNot(gomega.Succeed(), "two identifiers")

		gomega.Expect(validator.Validate(withRules(&GcpDatastreamStreamRuleSet{
			ObjectFilter: &GcpDatastreamStreamObjectFilter{SourceObjectIdentifier: identifier},
			CustomizationRules: []*GcpDatastreamStreamCustomizationRule{{
				BigqueryClustering:   clustering.BigqueryClustering,
				BigqueryPartitioning: &GcpDatastreamStreamBigqueryPartitioning{IngestionTimePartition: &GcpDatastreamStreamIngestionTimePartition{}},
			}},
		}))).ToNot(gomega.Succeed(), "two rule kinds")

		gomega.Expect(validator.Validate(withRules(&GcpDatastreamStreamRuleSet{
			ObjectFilter: &GcpDatastreamStreamObjectFilter{SourceObjectIdentifier: identifier},
			CustomizationRules: []*GcpDatastreamStreamCustomizationRule{{
				BigqueryPartitioning: &GcpDatastreamStreamBigqueryPartitioning{
					IngestionTimePartition: &GcpDatastreamStreamIngestionTimePartition{},
					IntegerRangePartition:  &GcpDatastreamStreamIntegerRangePartition{Column: "id", Start: 0, End: 1000, Interval: 10},
				},
			}},
		}))).ToNot(gomega.Succeed(), "two partitioning kinds")

		gomega.Expect(validator.Validate(withRules(&GcpDatastreamStreamRuleSet{
			ObjectFilter: &GcpDatastreamStreamObjectFilter{SourceObjectIdentifier: identifier},
			CustomizationRules: []*GcpDatastreamStreamCustomizationRule{{
				BigqueryPartitioning: &GcpDatastreamStreamBigqueryPartitioning{
					IntegerRangePartition: &GcpDatastreamStreamIntegerRangePartition{Column: "id", Start: 0, End: 1000, Interval: 10},
				},
			}},
		}))).To(gomega.Succeed(), "integer ranges")
	})
})
