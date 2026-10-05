package kubernetesairflowv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesAirflow(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAirflow Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

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

func testPostgres() *KubernetesAirflowPostgres {
	return &KubernetesAirflowPostgres{
		Host: literal("airflow-pg-rw"),
		PasswordSecret: &KubernetesAirflowPasswordSecret{
			SecretName: literal("airflow-pg-app"),
		},
	}
}

func testMysql() *KubernetesAirflowMysql {
	return &KubernetesAirflowMysql{
		Host: literal("airflow-mysql"),
		PasswordSecret: &KubernetesAirflowMysqlPasswordSecret{
			SecretName: literal("airflow-mysql-auth"),
		},
	}
}

var _ = ginkgo.Describe("KubernetesAirflow Validation Tests", func() {
	var input *KubernetesAirflow

	ginkgo.BeforeEach(func() {
		input = &KubernetesAirflow{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAirflow",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "pipelines",
			},
			Spec: &KubernetesAirflowSpec{
				Namespace: literal("airflow"),
				Database: &KubernetesAirflowDatabase{
					Backend: &KubernetesAirflowDatabase_Postgres{Postgres: testPostgres()},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (postgres arm, every optional block omitted) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "airflow", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("postgres host and password composed from a KubernetesPostgres should be valid", func() {
			pg := testPostgres()
			pg.Host = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "airflow-pg", "status.outputs.rw_service")
			pg.PasswordSecret.SecretName = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "airflow-pg", "status.outputs.password_secret.name")
			input.Spec.Database.Backend = &KubernetesAirflowDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("postgres tuning (port, database, user, ssl mode) should be valid", func() {
			pg := testPostgres()
			pg.Port = int32Ptr(5433)
			pg.DatabaseName = strPtr("airflow_prod")
			pg.Username = strPtr("airflow_svc")
			pg.SslMode = strPtr("verify-full")
			input.Spec.Database.Backend = &KubernetesAirflowDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("mysql arm should be valid", func() {
			input.Spec.Database.Backend = &KubernetesAirflowDatabase_Mysql{Mysql: testMysql()}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("explicit KubernetesExecutor without a broker should be valid", func() {
			input.Spec.Executor = strPtr("KubernetesExecutor")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("CeleryExecutor with the bundled redis broker should be valid", func() {
			input.Spec.Executor = strPtr("CeleryExecutor")
			input.Spec.Broker = &KubernetesAirflowBroker{
				Backend: &KubernetesAirflowBroker_BundledRedis{BundledRedis: &KubernetesAirflowBundledRedis{}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("CeleryExecutor with a composed Valkey broker should be valid", func() {
			input.Spec.Executor = strPtr("CeleryExecutor")
			input.Spec.Broker = &KubernetesAirflowBroker{
				Backend: &KubernetesAirflowBroker_Valkey{Valkey: &KubernetesAirflowValkeyBroker{
					Host:     valueFrom(catalogkind.CatalogKind_KubernetesValkey, "airflow-broker", "status.outputs.service"),
					Username: "airflow",
					PasswordSecret: &KubernetesAirflowBrokerPasswordSecret{
						SecretName: valueFrom(catalogkind.CatalogKind_KubernetesValkey, "airflow-broker", "status.outputs.password_secret.name"),
						SecretKey:  strPtr("airflow"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("CeleryExecutor with an existing broker-url Secret should be valid", func() {
			input.Spec.Executor = strPtr("CeleryExecutor")
			input.Spec.Broker = &KubernetesAirflowBroker{
				Backend: &KubernetesAirflowBroker_ExistingBrokerUrlSecret{ExistingBrokerUrlSecret: &KubernetesAirflowBrokerUrlSecret{
					SecretName: "my-broker-url",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multi-executor list with a Celery member and a broker should be valid", func() {
			input.Spec.Executor = strPtr("CeleryExecutor,KubernetesExecutor")
			input.Spec.Broker = &KubernetesAirflowBroker{
				Backend: &KubernetesAirflowBroker_BundledRedis{BundledRedis: &KubernetesAirflowBundledRedis{}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a custom executor class path without a broker should be valid", func() {
			input.Spec.Executor = strPtr("airflow.providers.amazon.aws.executors.batch.AwsBatchExecutor")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("git-sync from a public https repository should be valid", func() {
			input.Spec.Dags = &KubernetesAirflowDags{
				Source: &KubernetesAirflowDags_GitSync{GitSync: &KubernetesAirflowGitSync{
					Repo:    "https://github.com/example/dags.git",
					Ref:     "main",
					SubPath: "dags",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("git-sync over ssh with a key Secret and pinned known_hosts should be valid", func() {
			input.Spec.Dags = &KubernetesAirflowDags{
				Source: &KubernetesAirflowDags_GitSync{GitSync: &KubernetesAirflowGitSync{
					Repo:         "git@github.com:example/dags.git",
					SshKeySecret: "dags-ssh-key",
					KnownHosts:   "github.com ssh-ed25519 AAAA...",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("git-sync over https with a credentials Secret should be valid", func() {
			input.Spec.Dags = &KubernetesAirflowDags{
				Source: &KubernetesAirflowDags_GitSync{GitSync: &KubernetesAirflowGitSync{
					Repo:              "https://github.com/example/private-dags.git",
					CredentialsSecret: "dags-git-token",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("dags on a shared persistent volume should be valid", func() {
			input.Spec.Dags = &KubernetesAirflowDags{
				Source: &KubernetesAirflowDags_Persistence{Persistence: &KubernetesAirflowDagsPersistence{
					Size:         strPtr("2Gi"),
					StorageClass: "efs-rwx",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full component sizing with KEDA autoscaling should be valid", func() {
			input.Spec.Executor = strPtr("CeleryExecutor")
			input.Spec.Broker = &KubernetesAirflowBroker{
				Backend: &KubernetesAirflowBroker_BundledRedis{BundledRedis: &KubernetesAirflowBundledRedis{}},
			}
			input.Spec.Components = &KubernetesAirflowComponents{
				ApiServer: &KubernetesAirflowApiServer{Replicas: int32Ptr(2)},
				Scheduler: &KubernetesAirflowScheduler{Replicas: int32Ptr(2)},
				Triggerer: &KubernetesAirflowTriggerer{PersistenceSize: strPtr("1Gi")},
				Workers: &KubernetesAirflowWorkers{
					PersistenceSize: strPtr("1Gi"),
					Keda: &KubernetesAirflowWorkersKeda{
						Enabled:     true,
						MinReplicas: int32Ptr(0),
						MaxReplicas: int32Ptr(20),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("pgbouncer on the postgres arm should be valid", func() {
			input.Spec.Pgbouncer = &KubernetesAirflowPgBouncer{
				Enabled:          true,
				MetadataPoolSize: int32Ptr(20),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("opensearch log read path composed from a KubernetesOpenSearch should be valid", func() {
			input.Spec.Logging = &KubernetesAirflowLogging{
				RemoteRead: &KubernetesAirflowLogging_Opensearch{Opensearch: &KubernetesAirflowLogSearchBackend{
					Host:     valueFrom(catalogkind.CatalogKind_KubernetesOpenSearch, "logs", "status.outputs.service_name"),
					Username: "airflow",
					PasswordSecret: &KubernetesAirflowLogBackendPasswordSecret{
						SecretName: valueFrom(catalogkind.CatalogKind_KubernetesOpenSearch, "logs", "status.outputs.admin_credentials_secret_name"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("opensearch log read path with an explicit credential Secret should be valid", func() {
			input.Spec.Logging = &KubernetesAirflowLogging{
				RemoteRead: &KubernetesAirflowLogging_Opensearch{Opensearch: &KubernetesAirflowLogSearchBackend{
					Host:     literal("logs-opensearch"),
					Username: "airflow",
					PasswordSecret: &KubernetesAirflowLogBackendPasswordSecret{
						SecretName: literal("logs-airflow-auth"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("unauthenticated elasticsearch log read path should be valid", func() {
			input.Spec.Logging = &KubernetesAirflowLogging{
				Persistence: &KubernetesAirflowLogsPersistence{Enabled: true, Size: strPtr("10Gi")},
				RemoteRead: &KubernetesAirflowLogging_Elasticsearch{Elasticsearch: &KubernetesAirflowLogSearchBackend{
					Host: literal("es-http"),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("admin user with an existing password Secret should be valid", func() {
			input.Spec.AdminUser = &KubernetesAirflowAdminUser{
				Username: strPtr("platform-admin"),
				Email:    strPtr("platform@example.com"),
				PasswordSecret: &KubernetesAirflowExistingSecretRef{
					SecretName: "airflow-admin-cred",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("bring-your-own security Secrets should be valid", func() {
			input.Spec.Security = &KubernetesAirflowSecurity{
				FernetKeySecretName:    "shared-fernet",
				ApiSecretKeySecretName: "shared-api-key",
				JwtSecretName:          "shared-jwt",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("image overrides, scheduling, statsd and helm values should be valid", func() {
			input.Spec.StatsdEnabled = boolPtr(false)
			input.Spec.LoadExamples = true
			input.Spec.Images = &KubernetesAirflowImages{
				AirflowRepository: "mirror.example.com/airflow",
				AirflowTag:        "3.2.2-custom",
			}
			input.Spec.Scheduling = &KubernetesAirflowScheduling{
				NodeSelector: map[string]string{"workload": "data"},
			}
			input.Spec.HelmValues = "workers:\n  terminationGracePeriodSeconds: 300\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing database should fail", func() {
			input.Spec.Database = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("database with no backend arm should fail", func() {
			input.Spec.Database = &KubernetesAirflowDatabase{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("postgres without a password Secret should fail", func() {
			pg := testPostgres()
			pg.PasswordSecret = nil
			input.Spec.Database.Backend = &KubernetesAirflowDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("CeleryExecutor without a broker should fail the pairing rule", func() {
			input.Spec.Executor = strPtr("CeleryExecutor")
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("message broker"))
		})

		ginkgo.It("CeleryKubernetesExecutor without a broker should fail the pairing rule", func() {
			input.Spec.Executor = strPtr("CeleryKubernetesExecutor")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a broker with the KubernetesExecutor should fail", func() {
			input.Spec.Executor = strPtr("KubernetesExecutor")
			input.Spec.Broker = &KubernetesAirflowBroker{
				Backend: &KubernetesAirflowBroker_BundledRedis{BundledRedis: &KubernetesAirflowBundledRedis{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a broker with executor unset should fail (the effective default is KubernetesExecutor)", func() {
			input.Spec.Broker = &KubernetesAirflowBroker{
				Backend: &KubernetesAirflowBroker_BundledRedis{BundledRedis: &KubernetesAirflowBundledRedis{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a lowercase executor name should fail the pattern", func() {
			input.Spec.Executor = strPtr("celeryexecutor")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a non-executor class path should fail the pattern", func() {
			input.Spec.Executor = strPtr("CeleryExecutor,NotAnExec")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an Airflow 2 version should fail the v3-line rule", func() {
			input.Spec.AirflowVersion = strPtr("2.10.5")
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("Airflow 3"))
		})

		ginkgo.It("pgbouncer on the mysql arm should fail", func() {
			input.Spec.Database.Backend = &KubernetesAirflowDatabase_Mysql{Mysql: testMysql()}
			input.Spec.Pgbouncer = &KubernetesAirflowPgBouncer{Enabled: true}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("PgBouncer"))
		})

		ginkgo.It("an invalid postgres ssl mode should fail", func() {
			pg := testPostgres()
			pg.SslMode = strPtr("mandatory")
			input.Spec.Database.Backend = &KubernetesAirflowDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres port outside the valid range should fail", func() {
			pg := testPostgres()
			pg.Port = int32Ptr(70000)
			input.Spec.Database.Backend = &KubernetesAirflowDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a hyphenated database name should fail the identifier pattern", func() {
			pg := testPostgres()
			pg.DatabaseName = strPtr("air-flow")
			input.Spec.Database.Backend = &KubernetesAirflowDatabase_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a dags block with no source arm should fail", func() {
			input.Spec.Dags = &KubernetesAirflowDags{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("git-sync with an ssh key against an https repo should fail", func() {
			input.Spec.Dags = &KubernetesAirflowDags{
				Source: &KubernetesAirflowDags_GitSync{GitSync: &KubernetesAirflowGitSync{
					Repo:         "https://github.com/example/dags.git",
					SshKeySecret: "dags-ssh-key",
				}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("SSH"))
		})

		ginkgo.It("git-sync with https credentials against an ssh repo should fail", func() {
			input.Spec.Dags = &KubernetesAirflowDags{
				Source: &KubernetesAirflowDags_GitSync{GitSync: &KubernetesAirflowGitSync{
					Repo:              "git@github.com:example/dags.git",
					CredentialsSecret: "dags-git-token",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("git-sync with both credential mechanisms should fail", func() {
			input.Spec.Dags = &KubernetesAirflowDags{
				Source: &KubernetesAirflowDags_GitSync{GitSync: &KubernetesAirflowGitSync{
					Repo:              "https://github.com/example/dags.git",
					CredentialsSecret: "dags-git-token",
					SshKeySecret:      "dags-ssh-key",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("git-sync without a repo should fail", func() {
			input.Spec.Dags = &KubernetesAirflowDags{
				Source: &KubernetesAirflowDags_GitSync{GitSync: &KubernetesAirflowGitSync{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("KEDA with min above max should fail", func() {
			input.Spec.Executor = strPtr("CeleryExecutor")
			input.Spec.Broker = &KubernetesAirflowBroker{
				Backend: &KubernetesAirflowBroker_BundledRedis{BundledRedis: &KubernetesAirflowBundledRedis{}},
			}
			input.Spec.Components = &KubernetesAirflowComponents{
				Workers: &KubernetesAirflowWorkers{
					Keda: &KubernetesAirflowWorkersKeda{
						Enabled:     true,
						MinReplicas: int32Ptr(5),
						MaxReplicas: int32Ptr(2),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_replicas"))
		})

		ginkgo.It("a broker database number above 15 should fail", func() {
			input.Spec.Executor = strPtr("CeleryExecutor")
			input.Spec.Broker = &KubernetesAirflowBroker{
				Backend: &KubernetesAirflowBroker_Valkey{Valkey: &KubernetesAirflowValkeyBroker{
					Host:           literal("valkey"),
					DatabaseNumber: int32Ptr(16),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a log-backend username without a password Secret should fail", func() {
			input.Spec.Logging = &KubernetesAirflowLogging{
				RemoteRead: &KubernetesAirflowLogging_Opensearch{Opensearch: &KubernetesAirflowLogSearchBackend{
					Host:     literal("logs"),
					Username: "airflow",
				}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("password_secret"))
		})

		ginkgo.It("a malformed persistence size should fail", func() {
			input.Spec.Components = &KubernetesAirflowComponents{
				Workers: &KubernetesAirflowWorkers{
					PersistenceSize: strPtr("10 gigs"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an admin password Secret without a name should fail", func() {
			input.Spec.AdminUser = &KubernetesAirflowAdminUser{
				PasswordSecret: &KubernetesAirflowExistingSecretRef{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a wrong kind constant should fail", func() {
			input.Kind = "KubernetesAirFlow"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a wrong api version should fail", func() {
			input.ApiVersion = "kubernetes.planton.dev/v2"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
