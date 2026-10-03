package kubernetestrinov1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesTrino(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTrino Suite")
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

func testPostgresCatalog(name string) *KubernetesTrinoPostgresCatalog {
	return &KubernetesTrinoPostgresCatalog{
		Name:     name,
		Host:     literal("analytics-pg-rw"),
		Database: "analytics",
		PasswordSecret: &KubernetesTrinoPostgresPasswordSecret{
			SecretName: literal("analytics-pg-app"),
		},
	}
}

func testMysqlCatalog(name string) *KubernetesTrinoMysqlCatalog {
	return &KubernetesTrinoMysqlCatalog{
		Name: name,
		Host: literal("orders-mysql"),
		PasswordSecret: &KubernetesTrinoMysqlPasswordSecret{
			SecretName: literal("orders-mysql-root"),
		},
	}
}

var _ = ginkgo.Describe("KubernetesTrino Validation Tests", func() {
	var input *KubernetesTrino

	ginkgo.BeforeEach(func() {
		input = &KubernetesTrino{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesTrino",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "trino",
			},
			Spec: &KubernetesTrinoSpec{
				Namespace: literal("trino"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (namespace only — 2 workers, sample catalogs, auth defaults) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a postgres catalog with FK references should be valid", func() {
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				Postgres: []*KubernetesTrinoPostgresCatalog{
					{
						Name:     "analytics",
						Host:     valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "analytics-pg", "status.outputs.rw_service"),
						Database: "analytics",
						PasswordSecret: &KubernetesTrinoPostgresPasswordSecret{
							SecretName: valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "analytics-pg", "status.outputs.password_secret.name"),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("postgres + mysql + custom catalogs with distinct names should be valid", func() {
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				Postgres: []*KubernetesTrinoPostgresCatalog{testPostgresCatalog("analytics")},
				Mysql:    []*KubernetesTrinoMysqlCatalog{testMysqlCatalog("orders")},
				Custom: map[string]string{
					"lake": "connector.name=memory",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("tpch as a custom catalog name should be valid once sample catalogs are disabled", func() {
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				SampleCatalogsEnabled: boolPtr(false),
				Custom: map[string]string{
					"tpch": "connector.name=tpch\ntpch.splits-per-node=8",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the single-node shape (workers 0 + coordinator scheduling) should be valid", func() {
			input.Spec.Workers = &KubernetesTrinoWorkers{Replicas: int32Ptr(0)}
			input.Spec.Coordinator = &KubernetesTrinoCoordinator{IncludeInScheduling: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("worker HPA autoscaling should be valid", func() {
			input.Spec.Workers = &KubernetesTrinoWorkers{
				Autoscaling: &KubernetesTrinoWorkers_Hpa{
					Hpa: &KubernetesTrinoWorkerHpa{MaxReplicas: 10},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("worker KEDA autoscaling with triggers should be valid", func() {
			input.Spec.Workers = &KubernetesTrinoWorkers{
				Autoscaling: &KubernetesTrinoWorkers_Keda{
					Keda: &KubernetesTrinoWorkerKeda{
						MaxReplicas: 20,
						Triggers:    "- type: prometheus\n  metadata:\n    serverAddress: http://prometheus:9090\n    query: sum(trino_required_workers)\n    threshold: \"1\"",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("graceful shutdown with a custom grace period should be valid", func() {
			input.Spec.Workers = &KubernetesTrinoWorkers{
				GracefulShutdown: &KubernetesTrinoGracefulShutdown{
					Enabled:            true,
					GracePeriodSeconds: int32Ptr(300),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("jvm percent-based heap sizing should be valid", func() {
			input.Spec.Coordinator = &KubernetesTrinoCoordinator{
				Jvm: &KubernetesTrinoJvm{MaxHeapPercent: int32Ptr(75)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("auth disabled should be valid (the field exists to be explicit)", func() {
			input.Spec.Auth = &KubernetesTrinoAuth{Enabled: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("bring-your-own password.db Secret should be valid", func() {
			input.Spec.Auth = &KubernetesTrinoAuth{
				ExistingPasswordDbSecret: &KubernetesTrinoExistingSecret{SecretName: "trino-passwords"},
				GroupsSecret:             &KubernetesTrinoExistingSecret{SecretName: "trino-groups"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("https with a keystore Secret should be valid", func() {
			input.Spec.Https = &KubernetesTrinoHttps{
				Enabled:        true,
				KeystoreSecret: &KubernetesTrinoKeystoreSecret{SecretName: "trino-keystore"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("fault-tolerant execution with an exchange manager should be valid", func() {
			input.Spec.FaultTolerantExecution = &KubernetesTrinoFaultTolerantExecution{
				RetryPolicy: "TASK",
				ExchangeManager: &KubernetesTrinoExchangeManager{
					BaseDirectories: []string{"s3://trino-exchange"},
					AdditionalProperties: []string{
						"exchange.s3.endpoint=http://seaweedfs-s3:8333",
						"exchange.s3.aws-access-key=${ENV:EXCHANGE_ACCESS_KEY}",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("metrics with ServiceMonitors should be valid", func() {
			input.Spec.Metrics = &KubernetesTrinoMetrics{
				Enabled:               true,
				ServiceMonitorEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("access control, resource groups and session properties documents should be valid", func() {
			input.Spec.AccessControlRules = `{"catalogs":[{"user":"trino","catalog":".*","allow":"all"}]}`
			input.Spec.ResourceGroupsConfig = `{"rootGroups":[{"name":"global","softMemoryLimit":"80%","hardConcurrencyLimit":100,"maxQueued":100}],"selectors":[{"group":"global"}]}`
			input.Spec.SessionPropertiesConfig = `[{"group":"global.*","sessionProperties":{"query_max_execution_time":"1h"}}]`
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("extra env from secret paired with ${ENV:} properties should be valid", func() {
			input.Spec.ExtraEnvFromSecret = map[string]*KubernetesTrinoSecretKeyRef{
				"LAKE_SECRET_KEY": {SecretName: "lake-creds", SecretKey: "secret_access_key"},
			}
			input.Spec.AdditionalConfigProperties = []string{"http-server.process-forwarded=true"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("wrong api_version should fail", func() {
			input.ApiVersion = "v1"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("wrong kind should fail", func() {
			input.Kind = "Trino"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("uppercase node_environment should fail (Trino rejects it at startup)", func() {
			input.Spec.NodeEnvironment = strPtr("Production")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("unknown log_level should fail", func() {
			input.Spec.LogLevel = strPtr("TRACE")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("kubernetes-style max_query_memory (4Gi) should fail — Trino wants SI data sizes (4GB)", func() {
			input.Spec.MaxQueryMemory = strPtr("4Gi")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("negative worker replicas should fail", func() {
			input.Spec.Workers = &KubernetesTrinoWorkers{Replicas: int32Ptr(-1)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("jvm max_heap_percent above 95 should fail", func() {
			input.Spec.Workers = &KubernetesTrinoWorkers{
				Jvm: &KubernetesTrinoJvm{MaxHeapPercent: int32Ptr(99)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("HPA without max_replicas should fail", func() {
			input.Spec.Workers = &KubernetesTrinoWorkers{
				Autoscaling: &KubernetesTrinoWorkers_Hpa{Hpa: &KubernetesTrinoWorkerHpa{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("KEDA without triggers should fail (it would scale nothing)", func() {
			input.Spec.Workers = &KubernetesTrinoWorkers{
				Autoscaling: &KubernetesTrinoWorkers_Keda{
					Keda: &KubernetesTrinoWorkerKeda{MaxReplicas: 10},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("https enabled without a keystore should fail the pairing rule", func() {
			input.Spec.Https = &KubernetesTrinoHttps{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase catalog name should fail the pattern", func() {
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				Postgres: []*KubernetesTrinoPostgresCatalog{testPostgresCatalog("Analytics")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate catalog names across postgres and mysql should fail", func() {
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				Postgres: []*KubernetesTrinoPostgresCatalog{testPostgresCatalog("shared")},
				Mysql:    []*KubernetesTrinoMysqlCatalog{testMysqlCatalog("shared")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a typed catalog colliding with a custom catalog name should fail", func() {
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				Postgres: []*KubernetesTrinoPostgresCatalog{testPostgresCatalog("lake")},
				Custom:   map[string]string{"lake": "connector.name=memory"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a catalog named system should fail (Trino's internal catalog)", func() {
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				Custom: map[string]string{"system": "connector.name=memory"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a catalog named tpch while sample catalogs are enabled should fail", func() {
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				Postgres: []*KubernetesTrinoPostgresCatalog{testPostgresCatalog("tpch")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres catalog without a password secret should fail", func() {
			catalog := testPostgresCatalog("analytics")
			catalog.PasswordSecret = nil
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				Postgres: []*KubernetesTrinoPostgresCatalog{catalog},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres catalog without a database should fail", func() {
			catalog := testPostgresCatalog("analytics")
			catalog.Database = ""
			input.Spec.Catalogs = &KubernetesTrinoCatalogs{
				Postgres: []*KubernetesTrinoPostgresCatalog{catalog},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a ServiceMonitor without metrics enabled should fail the pairing rule", func() {
			input.Spec.Metrics = &KubernetesTrinoMetrics{ServiceMonitorEnabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("fault-tolerant execution without an exchange manager should fail", func() {
			input.Spec.FaultTolerantExecution = &KubernetesTrinoFaultTolerantExecution{
				RetryPolicy: "TASK",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an exchange manager without base directories should fail", func() {
			input.Spec.FaultTolerantExecution = &KubernetesTrinoFaultTolerantExecution{
				RetryPolicy:     "QUERY",
				ExchangeManager: &KubernetesTrinoExchangeManager{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown retry policy should fail", func() {
			input.Spec.FaultTolerantExecution = &KubernetesTrinoFaultTolerantExecution{
				RetryPolicy: "ALWAYS",
				ExchangeManager: &KubernetesTrinoExchangeManager{
					BaseDirectories: []string{"s3://x"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown service type should fail", func() {
			input.Spec.Service = &KubernetesTrinoService{Type: strPtr("ExternalName")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase admin username should fail the pattern", func() {
			input.Spec.Auth = &KubernetesTrinoAuth{AdminUsername: strPtr("Admin")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an out-of-range https port should fail", func() {
			input.Spec.Https = &KubernetesTrinoHttps{
				Enabled:        true,
				Port:           int32Ptr(70000),
				KeystoreSecret: &KubernetesTrinoKeystoreSecret{SecretName: "ks"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an extra_env_from_secret entry without a key should fail", func() {
			input.Spec.ExtraEnvFromSecret = map[string]*KubernetesTrinoSecretKeyRef{
				"X": {SecretName: "creds"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
