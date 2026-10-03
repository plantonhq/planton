package kubernetessupersetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesSuperset(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSuperset Suite")
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

func testMetadataDatabase() *KubernetesSupersetMetadataDatabase {
	return &KubernetesSupersetMetadataDatabase{
		Host: literal("superset-pg-rw"),
		PasswordSecret: &KubernetesSupersetPostgresPasswordSecret{
			SecretName: literal("superset-pg-app"),
		},
	}
}

func testCache() *KubernetesSupersetCache {
	return &KubernetesSupersetCache{
		Host: literal("superset-valkey"),
	}
}

var _ = ginkgo.Describe("KubernetesSuperset Validation Tests", func() {
	var input *KubernetesSuperset

	ginkgo.BeforeEach(func() {
		input = &KubernetesSuperset{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesSuperset",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "superset",
			},
			Spec: &KubernetesSupersetSpec{
				Namespace:        literal("superset"),
				MetadataDatabase: testMetadataDatabase(),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (namespace + metadata database — web-only, no cache) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("metadata database with FK references should be valid", func() {
			input.Spec.MetadataDatabase = &KubernetesSupersetMetadataDatabase{
				Host: valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "superset-pg", "status.outputs.rw_service"),
				PasswordSecret: &KubernetesSupersetPostgresPasswordSecret{
					SecretName: valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "superset-pg", "status.outputs.password_secret.name"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a cache composing a KubernetesValkey with its password Secret should be valid", func() {
			input.Spec.Cache = &KubernetesSupersetCache{
				Host: valueFrom(catalogkind.CatalogKind_KubernetesValkey, "superset-cache", "status.outputs.service"),
				PasswordSecret: &KubernetesSupersetCachePasswordSecret{
					SecretName: valueFrom(catalogkind.CatalogKind_KubernetesValkey, "superset-cache", "status.outputs.password_secret.name"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("workers with the cache set should be valid", func() {
			input.Spec.Cache = testCache()
			input.Spec.Worker = &KubernetesSupersetWorker{Replicas: int32Ptr(2)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an explicitly disabled worker without a cache should be valid", func() {
			input.Spec.Worker = &KubernetesSupersetWorker{Enabled: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("beat with worker and cache should be valid", func() {
			input.Spec.Cache = testCache()
			input.Spec.Worker = &KubernetesSupersetWorker{}
			input.Spec.Beat = &KubernetesSupersetBeat{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("flower with the cache set should be valid", func() {
			input.Spec.Cache = testCache()
			input.Spec.Flower = &KubernetesSupersetFlower{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("websockets with the cache set should be valid", func() {
			input.Spec.Cache = testCache()
			input.Spec.Websockets = &KubernetesSupersetWebsockets{
				Enabled: true,
				Image:   &KubernetesSupersetWsImage{Tag: strPtr("0.1.0")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("mcp with the cache set should be valid", func() {
			input.Spec.Cache = testCache()
			input.Spec.Mcp = &KubernetesSupersetMcp{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("web HPA should be valid", func() {
			input.Spec.Web = &KubernetesSupersetWeb{
				Hpa: &KubernetesSupersetHpa{MaxReplicas: 5},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("feature flags and config overrides should be valid", func() {
			input.Spec.FeatureFlags = map[string]bool{
				"DASHBOARD_RBAC": true,
				"ALERT_REPORTS":  true,
			}
			input.Spec.ConfigOverrides = map[string]string{
				"row_limit": "ROW_LIMIT = 100000",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an existing admin password Secret should be valid", func() {
			input.Spec.Init = &KubernetesSupersetInit{
				Admin: &KubernetesSupersetAdminUser{
					Username: strPtr("bi-admin"),
					PasswordSecret: &KubernetesSupersetSecretKeyRef{
						SecretName: "bi-admin-auth",
						SecretKey:  "password",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an existing SECRET_KEY Secret should be valid", func() {
			input.Spec.SecretKeySecret = &KubernetesSupersetSecretKeyRef{
				SecretName: "superset-signing",
				SecretKey:  "key",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("database SSL should be valid", func() {
			input.Spec.MetadataDatabase.Ssl = &KubernetesSupersetDatabaseSsl{
				Enabled: true,
				Mode:    strPtr("verify-full"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a bootstrap script installing an extra driver should be valid", func() {
			input.Spec.BootstrapScript = "#!/bin/bash\npip install trino\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("extra env from secret should be valid", func() {
			input.Spec.ExtraEnvFromSecret = map[string]*KubernetesSupersetSecretKeyRef{
				"MAPBOX_API_KEY": {SecretName: "mapbox", SecretKey: "api_key"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing metadata database should fail — Superset cannot run without it", func() {
			input.Spec.MetadataDatabase = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a metadata database without a host should fail", func() {
			input.Spec.MetadataDatabase.Host = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a metadata database without a password Secret should fail", func() {
			input.Spec.MetadataDatabase.PasswordSecret = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("wrong api_version should fail", func() {
			input.ApiVersion = "v1"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("wrong kind should fail", func() {
			input.Kind = "Superset"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a database name with a dash should fail the identifier pattern", func() {
			input.Spec.MetadataDatabase.DatabaseName = strPtr("superset-meta")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an out-of-range database port should fail", func() {
			input.Spec.MetadataDatabase.Port = int32Ptr(70000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("workers without a cache should fail the broker rule", func() {
			input.Spec.Worker = &KubernetesSupersetWorker{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("beat without a worker should fail", func() {
			input.Spec.Cache = testCache()
			input.Spec.Worker = &KubernetesSupersetWorker{Enabled: boolPtr(false)}
			input.Spec.Beat = &KubernetesSupersetBeat{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("beat without a cache should fail", func() {
			input.Spec.Beat = &KubernetesSupersetBeat{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("flower without a cache should fail", func() {
			input.Spec.Flower = &KubernetesSupersetFlower{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("websockets without a cache should fail", func() {
			input.Spec.Websockets = &KubernetesSupersetWebsockets{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("mcp without a cache should fail", func() {
			input.Spec.Mcp = &KubernetesSupersetMcp{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown SSL mode should fail", func() {
			input.Spec.MetadataDatabase.Ssl = &KubernetesSupersetDatabaseSsl{
				Enabled: true,
				Mode:    strPtr("prefer"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a cache db number above 15 should fail", func() {
			input.Spec.Cache = testCache()
			input.Spec.Cache.CacheDb = int32Ptr(16)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an HPA without max_replicas should fail", func() {
			input.Spec.Web = &KubernetesSupersetWeb{Hpa: &KubernetesSupersetHpa{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero web replicas should fail", func() {
			input.Spec.Web = &KubernetesSupersetWeb{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown service type should fail", func() {
			input.Spec.Service = &KubernetesSupersetService{Type: strPtr("ExternalName")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an admin password Secret reference without a key should fail", func() {
			input.Spec.Init = &KubernetesSupersetInit{
				Admin: &KubernetesSupersetAdminUser{
					PasswordSecret: &KubernetesSupersetSecretKeyRef{SecretName: "x"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an extra_env_from_secret entry without a Secret name should fail", func() {
			input.Spec.ExtraEnvFromSecret = map[string]*KubernetesSupersetSecretKeyRef{
				"X": {SecretKey: "k"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
