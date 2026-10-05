package kubernetestemporalv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesTemporal(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTemporal Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func strPtr(s string) *string { return &s }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

func testPostgres() *KubernetesTemporalPostgres {
	return &KubernetesTemporalPostgres{
		Host:     literal("temporal-pg-rw"),
		Username: "temporal",
		PasswordSecret: &KubernetesTemporalPasswordSecret{
			SecretName: literal("temporal-pg-app"),
		},
	}
}

var _ = ginkgo.Describe("KubernetesTemporal Validation Tests", func() {
	var input *KubernetesTemporal

	ginkgo.BeforeEach(func() {
		input = &KubernetesTemporal{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesTemporal",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "workflows",
			},
			Spec: &KubernetesTemporalSpec{
				Namespace: literal("temporal"),
				Database: &KubernetesTemporalDatabase{
					Backend: &KubernetesTemporalDatabase_Postgres{Postgres: testPostgres()},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (postgres arm, every optional block omitted) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "temporal", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("postgres host and password composed from a KubernetesPostgres should be valid", func() {
			pg := testPostgres()
			pg.Host = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "temporal-pg", "status.outputs.rw_service")
			pg.PasswordSecret.SecretName = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "temporal-pg", "status.outputs.password_secret.name")
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("postgres connection tuning should be valid", func() {
			pg := testPostgres()
			pg.MaxConns = int32Ptr(50)
			pg.MaxIdleConns = int32Ptr(10)
			pg.MaxConnLifetime = strPtr("30m")
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the mysql arm should be valid", func() {
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Mysql{
				Mysql: &KubernetesTemporalMysql{
					Host:     literal("temporal-mysql-primary"),
					Username: "root",
					PasswordSecret: &KubernetesTemporalMysqlPasswordSecret{
						SecretName: literal("temporal-mysql-secrets"),
						SecretKey:  "root",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the cassandra arm WITH a sql visibility block should be valid", func() {
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Cassandra{
				Cassandra: &KubernetesTemporalCassandra{
					Hosts:    []string{"cassandra-0.example.internal", "cassandra-1.example.internal"},
					Username: "temporal",
					PasswordSecret: &KubernetesTemporalPasswordSecret{
						SecretName: literal("cassandra-credentials"),
					},
					Datacenter: "dc1",
				},
			}
			input.Spec.Database.Visibility = &KubernetesTemporalVisibility{
				Backend: &KubernetesTemporalVisibility_Postgres{Postgres: testPostgres()},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a separate sql visibility store alongside the postgres default store should be valid", func() {
			input.Spec.Database.Visibility = &KubernetesTemporalVisibility{
				Backend:      &KubernetesTemporalVisibility_Postgres{Postgres: testPostgres()},
				DatabaseName: strPtr("temporal_vis"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("database tls with host verification should be valid", func() {
			pg := testPostgres()
			pg.Tls = &KubernetesTemporalDatabaseTls{Enabled: true, HostVerification: true, ServerName: "db.example.com"}
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("per-service sizing should be valid", func() {
			input.Spec.Services = &KubernetesTemporalServices{
				Frontend: &KubernetesTemporalServiceConfig{Replicas: int32Ptr(2)},
				History:  &KubernetesTemporalServiceConfig{Replicas: int32Ptr(3)},
				Matching: &KubernetesTemporalServiceConfig{Replicas: int32Ptr(2)},
				Worker:   &KubernetesTemporalServiceConfig{Replicas: int32Ptr(1)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("num_history_shards inside the bound should be valid", func() {
			input.Spec.NumHistoryShards = int32Ptr(4096)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("declarative temporal namespaces should be valid", func() {
			input.Spec.TemporalNamespaces = []*KubernetesTemporalNamespace{
				{Name: "default"},
				{Name: "payments", Retention: strPtr("30d")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("dynamic config limits should be valid", func() {
			input.Spec.DynamicConfig = &KubernetesTemporalDynamicConfig{
				HistorySizeLimitError: int64Ptr(104857600),
				BlobSizeLimitError:    int64Ptr(10485760),
				BlobSizeLimitWarn:     int64Ptr(5242880),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("s3 archival with s3:// URIs should be valid", func() {
			input.Spec.Archival = &KubernetesTemporalArchival{
				Provider:      &KubernetesTemporalArchival_S3{S3: &KubernetesTemporalArchivalS3{Region: "us-east-1"}},
				HistoryUri:    "s3://temporal-archival/history",
				VisibilityUri: "s3://temporal-archival/visibility",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gcs archival with gs:// URIs should be valid", func() {
			input.Spec.Archival = &KubernetesTemporalArchival{
				Provider:      &KubernetesTemporalArchival_Gcs{Gcs: &KubernetesTemporalArchivalGcs{}},
				HistoryUri:    "gs://temporal-archival/history",
				VisibilityUri: "gs://temporal-archival/visibility",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("filestore archival with file:// URIs should be valid", func() {
			input.Spec.Archival = &KubernetesTemporalArchival{
				Provider:      &KubernetesTemporalArchival_Filestore{Filestore: &KubernetesTemporalArchivalFilestore{}},
				HistoryUri:    "file:///var/temporal/archival/history",
				VisibilityUri: "file:///var/temporal/archival/visibility",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("web ui, admin tools, internal frontend and log level knobs should be valid", func() {
			enabled := true
			input.Spec.WebUi = &KubernetesTemporalWebUi{Enabled: &enabled, Replicas: int32Ptr(2)}
			input.Spec.AdminToolsEnabled = &enabled
			input.Spec.InternalFrontendEnabled = true
			input.Spec.LogLevel = strPtr("debug,info")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("create_databases and skip_schema_setup toggles should be valid", func() {
			input.Spec.Database.CreateDatabases = true
			input.Spec.Database.SkipSchemaSetup = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a spec without a database should fail", func() {
			input.Spec.Database = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a database without a backend arm should fail (oneof required)", func() {
			input.Spec.Database = &KubernetesTemporalDatabase{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("the cassandra arm WITHOUT a visibility block should fail (visibility removed upstream in v1.21)", func() {
			input.Spec.Database = &KubernetesTemporalDatabase{
				Backend: &KubernetesTemporalDatabase_Cassandra{
					Cassandra: &KubernetesTemporalCassandra{
						Hosts:    []string{"cassandra.example.internal"},
						Username: "temporal",
						PasswordSecret: &KubernetesTemporalPasswordSecret{
							SecretName: literal("cassandra-credentials"),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a cassandra arm with no hosts should fail", func() {
			input.Spec.Database = &KubernetesTemporalDatabase{
				Backend: &KubernetesTemporalDatabase_Cassandra{
					Cassandra: &KubernetesTemporalCassandra{
						Hosts:    []string{},
						Username: "temporal",
						PasswordSecret: &KubernetesTemporalPasswordSecret{
							SecretName: literal("cassandra-credentials"),
						},
					},
				},
				Visibility: &KubernetesTemporalVisibility{
					Backend: &KubernetesTemporalVisibility_Postgres{Postgres: testPostgres()},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres arm without a password secret should fail", func() {
			pg := testPostgres()
			pg.PasswordSecret = nil
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a mysql password secret without a key should fail", func() {
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Mysql{
				Mysql: &KubernetesTemporalMysql{
					Host:     literal("temporal-mysql-primary"),
					Username: "root",
					PasswordSecret: &KubernetesTemporalMysqlPasswordSecret{
						SecretName: literal("temporal-mysql-secrets"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("num_history_shards of 0 should fail", func() {
			input.Spec.NumHistoryShards = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("num_history_shards above 16384 should fail", func() {
			input.Spec.NumHistoryShards = int32Ptr(20000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a database name with dashes should fail (SQL identifier pattern)", func() {
			input.Spec.Database.DatabaseName = strPtr("temporal-db")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a temporal namespace with a bad retention form should fail", func() {
			input.Spec.TemporalNamespaces = []*KubernetesTemporalNamespace{
				{Name: "default", Retention: strPtr("three-days")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a single-character temporal namespace name should fail", func() {
			input.Spec.TemporalNamespaces = []*KubernetesTemporalNamespace{{Name: "a"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("s3 archival with gs:// URIs should fail (scheme mismatch)", func() {
			input.Spec.Archival = &KubernetesTemporalArchival{
				Provider:      &KubernetesTemporalArchival_S3{S3: &KubernetesTemporalArchivalS3{Region: "us-east-1"}},
				HistoryUri:    "gs://temporal-archival/history",
				VisibilityUri: "gs://temporal-archival/visibility",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("archival without a provider should fail (oneof required)", func() {
			input.Spec.Archival = &KubernetesTemporalArchival{
				HistoryUri:    "s3://temporal-archival/history",
				VisibilityUri: "s3://temporal-archival/visibility",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("archival without URIs should fail", func() {
			input.Spec.Archival = &KubernetesTemporalArchival{
				Provider: &KubernetesTemporalArchival_S3{S3: &KubernetesTemporalArchivalS3{Region: "us-east-1"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("dynamic config below the floor should fail", func() {
			input.Spec.DynamicConfig = &KubernetesTemporalDynamicConfig{
				BlobSizeLimitError: int64Ptr(1024),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("tls host_verification without enabled should fail", func() {
			pg := testPostgres()
			pg.Tls = &KubernetesTemporalDatabaseTls{HostVerification: true}
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad log level should fail", func() {
			input.Spec.LogLevel = strPtr("verbose")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres port out of range should fail", func() {
			pg := testPostgres()
			pg.Port = int32Ptr(70000)
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad max_conn_lifetime should fail", func() {
			pg := testPostgres()
			pg.MaxConnLifetime = strPtr("1 hour")
			input.Spec.Database.Backend = &KubernetesTemporalDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("web ui replicas above the bound should fail", func() {
			input.Spec.WebUi = &KubernetesTemporalWebUi{Replicas: int32Ptr(50)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("service replicas above the bound should fail", func() {
			input.Spec.Services = &KubernetesTemporalServices{
				History: &KubernetesTemporalServiceConfig{Replicas: int32Ptr(500)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
