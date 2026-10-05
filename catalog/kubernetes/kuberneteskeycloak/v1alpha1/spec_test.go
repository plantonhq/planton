package kuberneteskeycloakv1alpha1

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

func TestKubernetesKeycloak(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKeycloak Suite")
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

func testDb() *KubernetesKeycloakDb {
	return &KubernetesKeycloakDb{
		Vendor:   "postgres",
		Host:     literal("keycloak-pg-rw"),
		Database: "keycloak",
		UsernameSecret: &KubernetesKeycloakSecretSelector{
			Name: literal("keycloak-pg-app"),
			Key:  "username",
		},
		PasswordSecret: &KubernetesKeycloakSecretSelector{
			Name: literal("keycloak-pg-app"),
			Key:  "password",
		},
	}
}

var _ = ginkgo.Describe("KubernetesKeycloak Validation Tests", func() {
	var input *KubernetesKeycloak

	ginkgo.BeforeEach(func() {
		input = &KubernetesKeycloak{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKeycloak",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "keycloak",
			},
			Spec: &KubernetesKeycloakSpec{
				Namespace: literal("keycloak"),
				Db:        testDb(),
				Http: &KubernetesKeycloakHttp{
					TlsSecretName: literal("keycloak-tls"),
				},
				Hostname: &KubernetesKeycloakHostname{
					Hostname: "https://auth.example.com",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec (postgres db, TLS, strict hostname) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "keycloak", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a maximal spec (every block populated) should be valid", func() {
			db := testDb()
			db.Host = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "keycloak-pg", "status.outputs.rw_service")
			db.Port = int32Ptr(5432)
			db.Schema = "public"
			db.PoolMinSize = int32Ptr(5)
			db.PoolMaxSize = int32Ptr(50)
			input.Spec.CreateNamespace = true
			input.Spec.Instances = int32Ptr(3)
			input.Spec.Image = "quay.io/keycloak/keycloak:26.0.7"
			input.Spec.StartOptimized = true
			input.Spec.Db = db
			input.Spec.Http = &KubernetesKeycloakHttp{
				TlsSecretName: valueFrom(catalogkind.CatalogKind_KubernetesCertificate, "keycloak-cert", "status.outputs.secret_name"),
				HttpEnabled:   true,
				HttpPort:      int32Ptr(8080),
				HttpsPort:     int32Ptr(8443),
			}
			input.Spec.Hostname = &KubernetesKeycloakHostname{
				Hostname:           "https://auth.example.com",
				Admin:              "https://keycloak-admin.internal.example.com",
				Strict:             boolPtr(true),
				BackchannelDynamic: true,
			}
			input.Spec.ProxyHeaders = "xforwarded"
			input.Spec.Features = &KubernetesKeycloakFeatures{
				Enabled:  []string{"token-exchange", "recovery-codes"},
				Disabled: []string{"docker"},
			}
			input.Spec.TransactionXaEnabled = true
			input.Spec.CacheConfig = &KubernetesKeycloakCacheConfig{
				ConfigMapName: "keycloak-cache-ispn",
				Key:           strPtr("cache-ispn.xml"),
			}
			input.Spec.TruststoreSecretNames = []string{"private-ca-bundle"}
			input.Spec.AdditionalOptions = []*KubernetesKeycloakAdditionalOption{
				{Name: "log-level", Value: "INFO"},
				{Name: "spi-connections-http-client-default-connection-pool-size", Secret: &KubernetesKeycloakOptionSecretSelector{
					Name: literal("keycloak-extra-options"),
					Key:  "pool-size",
				}},
			}
			input.Spec.BootstrapAdminSecretName = "keycloak-bootstrap-admin"
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "250m", Memory: "768Mi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "2", Memory: "2Gi"},
			}
			input.Spec.Scheduling = &KubernetesKeycloakScheduling{
				NodeSelector:      map[string]string{"workload": "identity"},
				PriorityClassName: "platform-critical",
			}
			input.Spec.Probes = &KubernetesKeycloakProbes{
				LivenessFailureThreshold:  int32Ptr(3),
				LivenessPeriodSeconds:     int32Ptr(10),
				ReadinessFailureThreshold: int32Ptr(3),
				ReadinessPeriodSeconds:    int32Ptr(10),
				StartupFailureThreshold:   int32Ptr(600),
				StartupPeriodSeconds:      int32Ptr(1),
			}
			input.Spec.HttpManagementPort = int32Ptr(9000)
			input.Spec.NetworkPolicyEnabled = boolPtr(true)
			input.Spec.ServiceMonitorEnabled = boolPtr(true)
			input.Spec.Update = &KubernetesKeycloakUpdate{
				Strategy: strPtr("Explicit"),
				Revision: "v2",
			}
			input.Spec.Tracing = &KubernetesKeycloakTracing{
				Enabled:      true,
				Endpoint:     "http://otel-collector.observability:4317",
				Protocol:     strPtr("http/protobuf"),
				SamplerRatio: strPtr("0.25"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("plain HTTP only (no TLS, behind a proxy) should be valid", func() {
			input.Spec.Http = &KubernetesKeycloakHttp{HttpEnabled: true}
			input.Spec.ProxyHeaders = "forwarded"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("strict disabled with no hostname should be valid", func() {
			input.Spec.Hostname = &KubernetesKeycloakHostname{Strict: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("backchannel_dynamic with a full-URL hostname should be valid", func() {
			input.Spec.Hostname = &KubernetesKeycloakHostname{
				Hostname:           "https://auth.example.com",
				BackchannelDynamic: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an admin URL with a full-URL hostname should be valid", func() {
			input.Spec.Hostname = &KubernetesKeycloakHostname{
				Hostname: "https://auth.example.com",
				Admin:    "https://auth-admin.internal.example.com",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the dev-file sandbox vendor with a single instance should be valid", func() {
			input.Spec.Db = &KubernetesKeycloakDb{Vendor: "dev-file"}
			input.Spec.Instances = int32Ptr(1)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the dev-mem sandbox vendor with no connection details should be valid", func() {
			input.Spec.Db = &KubernetesKeycloakDb{Vendor: "dev-mem"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a real vendor with only a jdbc_url should be valid", func() {
			input.Spec.Db = &KubernetesKeycloakDb{
				Vendor:  "postgres",
				JdbcUrl: "jdbc:postgresql://pg-primary,pg-standby/keycloak?targetServerType=primary",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should be invalid", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing db block should be invalid", func() {
			input.Spec.Db = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing http block should be invalid", func() {
			input.Spec.Http = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing hostname block should be invalid", func() {
			input.Spec.Hostname = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("neither TLS nor plain HTTP should be invalid (server refuses to start)", func() {
			input.Spec.Http = &KubernetesKeycloakHttp{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("strict hostname resolution (the default) with no hostname should be invalid", func() {
			input.Spec.Hostname = &KubernetesKeycloakHostname{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("explicit strict=true with no hostname should be invalid", func() {
			input.Spec.Hostname = &KubernetesKeycloakHostname{Strict: boolPtr(true)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("backchannel_dynamic with no hostname should be invalid (server refuses to start)", func() {
			input.Spec.Hostname = &KubernetesKeycloakHostname{
				Strict:             boolPtr(false),
				BackchannelDynamic: true,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("backchannel_dynamic with a bare (non-URL) hostname should be invalid (server refuses to start)", func() {
			input.Spec.Hostname = &KubernetesKeycloakHostname{
				Hostname:           "auth.example.com",
				BackchannelDynamic: true,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an admin URL with a bare (non-URL) hostname should be invalid (server refuses to start)", func() {
			input.Spec.Hostname = &KubernetesKeycloakHostname{
				Hostname: "auth.example.com",
				Admin:    "https://auth-admin.internal.example.com",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("the dev-file sandbox vendor with two instances should be invalid", func() {
			input.Spec.Db = &KubernetesKeycloakDb{Vendor: "dev-file"}
			input.Spec.Instances = int32Ptr(2)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a real vendor without a host should be invalid", func() {
			db := testDb()
			db.Host = nil
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a real vendor without a database name should be invalid", func() {
			db := testDb()
			db.Database = ""
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a real vendor without a password secret should be invalid", func() {
			db := testDb()
			db.PasswordSecret = nil
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown db vendor should be invalid", func() {
			db := testDb()
			db.Vendor = "h2"
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an empty db vendor should be invalid", func() {
			db := testDb()
			db.Vendor = ""
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero instances should be invalid", func() {
			input.Spec.Instances = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("instances above 20 should be invalid", func() {
			input.Spec.Instances = int32Ptr(21)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a db port out of range should be invalid", func() {
			db := testDb()
			db.Port = int32Ptr(70000)
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a negative pool floor should be invalid", func() {
			db := testDb()
			db.PoolMinSize = int32Ptr(-1)
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a zero pool ceiling should be invalid", func() {
			db := testDb()
			db.PoolMaxSize = int32Ptr(0)
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a secret selector without a name should be invalid", func() {
			db := testDb()
			db.PasswordSecret = &KubernetesKeycloakSecretSelector{Key: "password"}
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a secret selector without a key should be invalid", func() {
			db := testDb()
			db.PasswordSecret = &KubernetesKeycloakSecretSelector{Name: literal("keycloak-pg-app")}
			input.Spec.Db = db
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown proxy_headers value should be invalid", func() {
			input.Spec.ProxyHeaders = "x-forwarded"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an http port out of range should be invalid", func() {
			input.Spec.Http.HttpPort = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an https port out of range should be invalid", func() {
			input.Spec.Http.HttpsPort = int32Ptr(70000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a management port out of range should be invalid", func() {
			input.Spec.HttpManagementPort = int32Ptr(70000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a cache config without a ConfigMap name should be invalid", func() {
			input.Spec.CacheConfig = &KubernetesKeycloakCacheConfig{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an additional option without a name should be invalid", func() {
			input.Spec.AdditionalOptions = []*KubernetesKeycloakAdditionalOption{
				{Value: "INFO"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an additional option with BOTH an inline value and a secret should be invalid", func() {
			input.Spec.AdditionalOptions = []*KubernetesKeycloakAdditionalOption{
				{
					Name:  "log-level",
					Value: "INFO",
					Secret: &KubernetesKeycloakOptionSecretSelector{
						Name: literal("keycloak-extra-options"),
						Key:  "log-level",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero liveness failure threshold should be invalid", func() {
			input.Spec.Probes = &KubernetesKeycloakProbes{LivenessFailureThreshold: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown update strategy should be invalid", func() {
			input.Spec.Update = &KubernetesKeycloakUpdate{Strategy: strPtr("Rolling")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("the Explicit update strategy without a revision should be invalid", func() {
			input.Spec.Update = &KubernetesKeycloakUpdate{Strategy: strPtr("Explicit")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("tracing enabled without an endpoint should be invalid", func() {
			input.Spec.Tracing = &KubernetesKeycloakTracing{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown tracing protocol should be invalid", func() {
			input.Spec.Tracing = &KubernetesKeycloakTracing{Protocol: strPtr("http")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a sampler ratio above 1.0 should be invalid", func() {
			input.Spec.Tracing = &KubernetesKeycloakTracing{SamplerRatio: strPtr("2.0")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
