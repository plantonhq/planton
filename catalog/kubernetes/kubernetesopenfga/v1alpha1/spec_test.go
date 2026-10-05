package kubernetesopenfgav1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesOpenFga(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesOpenFga Suite")
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

func testPostgres() *KubernetesOpenFgaPostgres {
	return &KubernetesOpenFgaPostgres{
		Host:     literal("openfga-pg-rw"),
		Database: "openfga",
		Username: "openfga",
		PasswordSecret: &KubernetesOpenFgaPasswordSecret{
			SecretName: literal("openfga-pg-app"),
		},
	}
}

var _ = ginkgo.Describe("KubernetesOpenFga Validation Tests", func() {
	var input *KubernetesOpenFga

	ginkgo.BeforeEach(func() {
		input = &KubernetesOpenFga{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesOpenFga",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "authz",
			},
			Spec: &KubernetesOpenFgaSpec{
				Namespace: literal("openfga"),
				Datastore: &KubernetesOpenFgaDatastore{
					Engine: &KubernetesOpenFgaDatastore_Postgres{Postgres: testPostgres()},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec (postgres engine, everything else defaulted) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "openfga", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a maximal spec (every block populated) should be valid", func() {
			pg := testPostgres()
			pg.Host = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "openfga-pg", "status.outputs.rw_service")
			pg.Port = int32Ptr(5432)
			pg.PasswordSecret.SecretName = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "openfga-pg", "status.outputs.password_secret.name")
			pg.PasswordSecret.SecretKey = strPtr("password")
			pg.SslMode = strPtr("verify-full")
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = strPtr("0.3.10")
			input.Spec.Replicas = int32Ptr(3)
			input.Spec.Datastore = &KubernetesOpenFgaDatastore{
				Engine:           &KubernetesOpenFgaDatastore_Postgres{Postgres: pg},
				MigrationTimeout: strPtr("5m"),
				MaxOpenConns:     int32Ptr(30),
				MaxIdleConns:     int32Ptr(10),
				ConnMaxIdleTime:  strPtr("30s"),
				ConnMaxLifetime:  strPtr("1h"),
			}
			input.Spec.Authn = &KubernetesOpenFgaAuthn{
				Method: &KubernetesOpenFgaAuthn_Oidc{
					Oidc: &KubernetesOpenFgaOidcAuthn{
						Issuer:   "https://auth.example.com/realms/platform",
						Audience: "openfga",
					},
				},
			}
			input.Spec.Metrics = &KubernetesOpenFgaMetrics{
				Enabled:               boolPtr(true),
				ServiceMonitorEnabled: true,
				EnableRpcHistograms:   true,
			}
			input.Spec.Tracing = &KubernetesOpenFgaTracing{
				Enabled:      true,
				OtlpEndpoint: "otel-collector.observability:4317",
				SampleRatio:  strPtr("0.5"),
			}
			input.Spec.Log = &KubernetesOpenFgaLog{
				Level:  strPtr("debug"),
				Format: strPtr("text"),
			}
			input.Spec.Tuning = &KubernetesOpenFgaTuning{
				MaxTuplesPerWrite:             int32Ptr(200),
				MaxTypesPerAuthorizationModel: int32Ptr(150),
				MaxChecksPerBatchCheck:        int32Ptr(100),
				ListObjectsDeadline:           strPtr("5s"),
				ListObjectsMaxResults:         int32Ptr(2000),
				ListUsersDeadline:             strPtr("5s"),
				ListUsersMaxResults:           int32Ptr(2000),
				RequestTimeout:                strPtr("10s"),
				CheckQueryCache: &KubernetesOpenFgaCheckQueryCache{
					Enabled: true,
					Limit:   int32Ptr(50000),
					Ttl:     strPtr("30s"),
				},
				Experimentals: []string{"enable-check-optimizations"},
			}
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "128Mi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "1", Memory: "512Mi"},
			}
			input.Spec.Hpa = &KubernetesOpenFgaHpa{
				Enabled:                        true,
				MinReplicas:                    int32Ptr(2),
				MaxReplicas:                    int32Ptr(10),
				TargetCpuUtilizationPercent:    int32Ptr(75),
				TargetMemoryUtilizationPercent: int32Ptr(80),
			}
			input.Spec.Scheduling = &KubernetesOpenFgaScheduling{
				NodeSelector: map[string]string{"workload": "authz"},
			}
			input.Spec.ServiceAccountAnnotations = map[string]string{"iam.gke.io/gcp-service-account": "openfga@my-project.iam.gserviceaccount.com"}
			input.Spec.HelmValues = "extraEnvVars:\n  - name: OPENFGA_LOG_TIMESTAMP_FORMAT\n    value: ISO8601\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the mysql engine should be valid", func() {
			input.Spec.Datastore.Engine = &KubernetesOpenFgaDatastore_Mysql{
				Mysql: &KubernetesOpenFgaMysql{
					Host:     literal("openfga-mysql-primary"),
					Database: "openfga",
					Username: "openfga",
					PasswordSecret: &KubernetesOpenFgaPasswordSecret{
						SecretName: literal("openfga-mysql-secrets"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the memory engine should be valid", func() {
			input.Spec.Datastore = &KubernetesOpenFgaDatastore{
				Engine: &KubernetesOpenFgaDatastore_Memory{Memory: &KubernetesOpenFgaMemory{}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("preshared authn with declared keys only should be valid", func() {
			input.Spec.Authn = &KubernetesOpenFgaAuthn{
				Method: &KubernetesOpenFgaAuthn_Preshared{
					Preshared: &KubernetesOpenFgaPresharedAuthn{Keys: []string{"key-one", "key-two"}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("preshared authn with an existing secret only should be valid", func() {
			input.Spec.Authn = &KubernetesOpenFgaAuthn{
				Method: &KubernetesOpenFgaAuthn_Preshared{
					Preshared: &KubernetesOpenFgaPresharedAuthn{ExistingKeysSecretName: "openfga-api-keys"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("hpa with equal min and max replicas should be valid", func() {
			input.Spec.Hpa = &KubernetesOpenFgaHpa{
				Enabled:     true,
				MinReplicas: int32Ptr(3),
				MaxReplicas: int32Ptr(3),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("chart-known experimentals should be valid", func() {
			input.Spec.Tuning = &KubernetesOpenFgaTuning{
				Experimentals: []string{"enable-check-optimizations", "pipeline_list_objects"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should be invalid", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing datastore should be invalid", func() {
			input.Spec.Datastore = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a datastore without an engine arm should be invalid", func() {
			input.Spec.Datastore = &KubernetesOpenFgaDatastore{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should be invalid", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("replicas above 50 should be invalid", func() {
			input.Spec.Replicas = int32Ptr(51)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres engine without a host should be invalid", func() {
			pg := testPostgres()
			pg.Host = nil
			input.Spec.Datastore.Engine = &KubernetesOpenFgaDatastore_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres engine without a database should be invalid", func() {
			pg := testPostgres()
			pg.Database = ""
			input.Spec.Datastore.Engine = &KubernetesOpenFgaDatastore_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres engine without a username should be invalid", func() {
			pg := testPostgres()
			pg.Username = ""
			input.Spec.Datastore.Engine = &KubernetesOpenFgaDatastore_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres engine without a password secret should be invalid", func() {
			pg := testPostgres()
			pg.PasswordSecret = nil
			input.Spec.Datastore.Engine = &KubernetesOpenFgaDatastore_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a password secret without a secret name should be invalid", func() {
			pg := testPostgres()
			pg.PasswordSecret = &KubernetesOpenFgaPasswordSecret{}
			input.Spec.Datastore.Engine = &KubernetesOpenFgaDatastore_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a postgres port out of range should be invalid", func() {
			pg := testPostgres()
			pg.Port = int32Ptr(70000)
			input.Spec.Datastore.Engine = &KubernetesOpenFgaDatastore_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown postgres sslmode should be invalid", func() {
			pg := testPostgres()
			pg.SslMode = strPtr("prefer")
			input.Spec.Datastore.Engine = &KubernetesOpenFgaDatastore_Postgres{Postgres: pg}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a mysql engine without a host should be invalid", func() {
			input.Spec.Datastore.Engine = &KubernetesOpenFgaDatastore_Mysql{
				Mysql: &KubernetesOpenFgaMysql{
					Database: "openfga",
					Username: "openfga",
					PasswordSecret: &KubernetesOpenFgaPasswordSecret{
						SecretName: literal("openfga-mysql-secrets"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a migration timeout that is not a Go duration should be invalid", func() {
			input.Spec.Datastore.MigrationTimeout = strPtr("3 minutes")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero max open connections should be invalid", func() {
			input.Spec.Datastore.MaxOpenConns = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad conn_max_idle_time should be invalid", func() {
			input.Spec.Datastore.ConnMaxIdleTime = strPtr("30sec")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad conn_max_lifetime should be invalid", func() {
			input.Spec.Datastore.ConnMaxLifetime = strPtr("1 hour")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("preshared authn with BOTH keys and an existing secret should be invalid", func() {
			input.Spec.Authn = &KubernetesOpenFgaAuthn{
				Method: &KubernetesOpenFgaAuthn_Preshared{
					Preshared: &KubernetesOpenFgaPresharedAuthn{
						Keys:                   []string{"key-one"},
						ExistingKeysSecretName: "openfga-api-keys",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("preshared authn with NEITHER keys nor an existing secret should be invalid", func() {
			input.Spec.Authn = &KubernetesOpenFgaAuthn{
				Method: &KubernetesOpenFgaAuthn_Preshared{
					Preshared: &KubernetesOpenFgaPresharedAuthn{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("oidc authn without an issuer should be invalid", func() {
			input.Spec.Authn = &KubernetesOpenFgaAuthn{
				Method: &KubernetesOpenFgaAuthn_Oidc{
					Oidc: &KubernetesOpenFgaOidcAuthn{Audience: "openfga"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("oidc authn with a non-URI issuer should be invalid", func() {
			input.Spec.Authn = &KubernetesOpenFgaAuthn{
				Method: &KubernetesOpenFgaAuthn_Oidc{
					Oidc: &KubernetesOpenFgaOidcAuthn{Issuer: "not a uri", Audience: "openfga"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("oidc authn without an audience should be invalid", func() {
			input.Spec.Authn = &KubernetesOpenFgaAuthn{
				Method: &KubernetesOpenFgaAuthn_Oidc{
					Oidc: &KubernetesOpenFgaOidcAuthn{Issuer: "https://auth.example.com"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("tracing enabled without an OTLP endpoint should be invalid", func() {
			input.Spec.Tracing = &KubernetesOpenFgaTracing{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a sample ratio above 1.0 should be invalid", func() {
			input.Spec.Tracing = &KubernetesOpenFgaTracing{SampleRatio: strPtr("1.5")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown log level should be invalid", func() {
			input.Spec.Log = &KubernetesOpenFgaLog{Level: strPtr("verbose")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown log format should be invalid", func() {
			input.Spec.Log = &KubernetesOpenFgaLog{Format: strPtr("yaml")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero max tuples per write should be invalid", func() {
			input.Spec.Tuning = &KubernetesOpenFgaTuning{MaxTuplesPerWrite: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad list_objects_deadline should be invalid", func() {
			input.Spec.Tuning = &KubernetesOpenFgaTuning{ListObjectsDeadline: strPtr("3sec")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("negative list_objects_max_results should be invalid", func() {
			input.Spec.Tuning = &KubernetesOpenFgaTuning{ListObjectsMaxResults: int32Ptr(-1)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad request_timeout should be invalid", func() {
			input.Spec.Tuning = &KubernetesOpenFgaTuning{RequestTimeout: strPtr("10 seconds")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a zero check-query-cache limit should be invalid", func() {
			input.Spec.Tuning = &KubernetesOpenFgaTuning{
				CheckQueryCache: &KubernetesOpenFgaCheckQueryCache{Limit: int32Ptr(0)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a bad check-query-cache ttl should be invalid", func() {
			input.Spec.Tuning = &KubernetesOpenFgaTuning{
				CheckQueryCache: &KubernetesOpenFgaCheckQueryCache{Ttl: strPtr("ten-seconds")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("hpa min replicas above max replicas should be invalid", func() {
			input.Spec.Hpa = &KubernetesOpenFgaHpa{
				Enabled:     true,
				MinReplicas: int32Ptr(5),
				MaxReplicas: int32Ptr(2),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero hpa min replicas should be invalid", func() {
			input.Spec.Hpa = &KubernetesOpenFgaHpa{MinReplicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a CPU utilization target above 100 percent should be invalid", func() {
			input.Spec.Hpa = &KubernetesOpenFgaHpa{TargetCpuUtilizationPercent: int32Ptr(101)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a zero memory utilization target should be invalid", func() {
			input.Spec.Hpa = &KubernetesOpenFgaHpa{TargetMemoryUtilizationPercent: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("hpa enabled on the memory datastore should be invalid", func() {
			input.Spec.Datastore = &KubernetesOpenFgaDatastore{
				Engine: &KubernetesOpenFgaDatastore_Memory{Memory: &KubernetesOpenFgaMemory{}},
			}
			input.Spec.Hpa = &KubernetesOpenFgaHpa{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an experimental the chart schema rejects should be invalid", func() {
			input.Spec.Tuning = &KubernetesOpenFgaTuning{Experimentals: []string{"shadow_check"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
