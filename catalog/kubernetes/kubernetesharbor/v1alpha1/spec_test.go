package kubernetesharborv1alpha1

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

func TestKubernetesHarbor(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesHarbor Suite")
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

// A minimal valid spec: internal database + internal redis +
// filesystem storage (the zero-dependency evaluation shape).
func baseInput() *KubernetesHarbor {
	return &KubernetesHarbor{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesHarbor",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "registry",
		},
		Spec: &KubernetesHarborSpec{
			Namespace:   literal("harbor"),
			ExternalUrl: "http://localhost:8080",
			Database: &KubernetesHarborDatabase{
				Engine: &KubernetesHarborDatabase_Internal{
					Internal: &KubernetesHarborInternalDatabase{},
				},
			},
			Cache: &KubernetesHarborCache{
				Engine: &KubernetesHarborCache_Internal{
					Internal: &KubernetesHarborInternalRedis{},
				},
			},
			Storage: &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_Filesystem{
					Filesystem: &KubernetesHarborFilesystemStorage{},
				},
			},
		},
	}
}

func externalDatabase() *KubernetesHarborExternalDatabase {
	return &KubernetesHarborExternalDatabase{
		Host:               literal("harbor-pg-rw"),
		Username:           "harbor",
		PasswordSecretName: literal("harbor-pg-app"),
	}
}

var _ = ginkgo.Describe("KubernetesHarbor Validation Tests", func() {
	var input *KubernetesHarbor

	ginkgo.BeforeEach(func() {
		input = baseInput()
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec (internal db/redis, filesystem storage) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "harbor", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an https external_url with host and port should be valid", func() {
			input.Spec.ExternalUrl = "https://harbor.example.com:8443"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a maximal composed spec (external postgres/valkey FKs, s3 storage, trivy, metrics) should be valid", func() {
			db := externalDatabase()
			db.Host = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "harbor-pg", "status.outputs.rw_service")
			db.PasswordSecretName = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "harbor-pg", "status.outputs.password_secret.name")
			db.Port = int32Ptr(5432)
			db.CoreDatabase = strPtr("registry")
			db.SslMode = strPtr("verify-full")
			input.Spec.ChartVersion = strPtr("1.19.1")
			input.Spec.CreateNamespace = true
			input.Spec.ExternalUrl = "https://harbor.example.com"
			input.Spec.Database = &KubernetesHarborDatabase{
				Engine: &KubernetesHarborDatabase_External{External: db},
			}
			input.Spec.Cache = &KubernetesHarborCache{
				Engine: &KubernetesHarborCache_External{
					External: &KubernetesHarborExternalRedis{
						Addr:               valueFrom(catalogkind.CatalogKind_KubernetesValkey, "harbor-valkey", "status.outputs.kube_endpoint"),
						ExistingSecretName: strPtr("harbor-redis-auth"),
					},
				},
			}
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_S3{
					S3: &KubernetesHarborS3Storage{
						Bucket:   "harbor-artifacts",
						Region:   "us-east-1",
						Endpoint: valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "harbor-objects", "status.outputs.s3_endpoint"),
						Credentials: &KubernetesHarborS3Credentials{
							ExistingSecretName: strPtr("harbor-s3-auth"),
						},
						DisableRedirect: true,
						Secure:          boolPtr(false),
					},
				},
			}
			input.Spec.Expose = &KubernetesHarborExpose{
				ServiceType: strPtr("LoadBalancer"),
				Tls:         &KubernetesHarborExposeTls{Enabled: true, CertSecretName: literal("harbor-tls")},
				ServiceAnnotations: map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
				},
				SourceRanges: []string{"10.0.0.0/8"},
			}
			input.Spec.AdminAuth = &KubernetesHarborAdminAuth{
				ExistingSecretName: strPtr("harbor-admin"),
			}
			input.Spec.Trivy = &KubernetesHarborTrivy{
				Enabled:       boolPtr(true),
				Replicas:      int32Ptr(2),
				SkipUpdate:    true,
				OfflineScan:   true,
				Severity:      strPtr("HIGH,CRITICAL"),
				IgnoreUnfixed: true,
			}
			input.Spec.Core = &KubernetesHarborComponent{Replicas: int32Ptr(2)}
			input.Spec.Portal = &KubernetesHarborComponent{Replicas: int32Ptr(2)}
			input.Spec.Registry = &KubernetesHarborComponent{Replicas: int32Ptr(2)}
			input.Spec.Jobservice = &KubernetesHarborJobservice{
				Replicas:      int32Ptr(2),
				MaxJobWorkers: int32Ptr(20),
				LogDiskSize:   strPtr("2Gi"),
			}
			input.Spec.Nginx = &KubernetesHarborComponent{Replicas: int32Ptr(2)}
			input.Spec.Metrics = &KubernetesHarborMetrics{
				Enabled:               true,
				ServiceMonitorEnabled: true,
				ServiceMonitorLabels:  map[string]string{"release": "monitoring"},
			}
			input.Spec.CacheLayer = &KubernetesHarborCacheLayer{Enabled: true, ExpireHours: int32Ptr(48)}
			input.Spec.OutboundProxy = &KubernetesHarborOutboundProxy{
				HttpProxy:  strPtr("http://proxy.internal:3128"),
				HttpsProxy: strPtr("http://proxy.internal:3128"),
			}
			input.Spec.LogLevel = strPtr("warning")
			input.Spec.ImageRegistry = strPtr("mirror.example.com")
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			input.Spec.CaBundleSecretName = strPtr("storage-ca")
			input.Spec.KeepVolumesOnUninstall = boolPtr(false)
			input.Spec.UpdateStrategy = strPtr("Recreate")
			input.Spec.Scheduling = &KubernetesHarborScheduling{
				NodeSelector: map[string]string{"pool": "apps"},
				Tolerations: []*kubernetes.WorkloadToleration{
					{Key: "dedicated", Operator: "Equal", Value: "harbor", Effect: "NoSchedule"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("two registry replicas on filesystem with ReadWriteMany should be valid", func() {
			input.Spec.Registry = &KubernetesHarborComponent{Replicas: int32Ptr(2)}
			input.Spec.Storage.GetFilesystem().AccessMode = strPtr("ReadWriteMany")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("two registry replicas on s3 should be valid without access-mode ceremony", func() {
			input.Spec.Registry = &KubernetesHarborComponent{Replicas: int32Ptr(2)}
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_S3{
					S3: &KubernetesHarborS3Storage{Bucket: "b", Region: "us-east-1"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gcs with workload identity (keyless) should be valid", func() {
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_Gcs{
					Gcs: &KubernetesHarborGcsStorage{Bucket: "b", UseWorkloadIdentity: true},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("azure with exactly one credential arm should be valid", func() {
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_Azure{
					Azure: &KubernetesHarborAzureStorage{
						AccountName: "acct",
						Container:   "harbor",
						AccountKey:  strPtr("base64key"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("external redis with a declared password (module-materialized) should be valid", func() {
			input.Spec.Cache = &KubernetesHarborCache{
				Engine: &KubernetesHarborCache_External{
					External: &KubernetesHarborExternalRedis{
						Addr:     literal("harbor-valkey:6379"),
						Password: strPtr("s3cret"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("external redis with sentinels should be valid", func() {
			input.Spec.Cache = &KubernetesHarborCache{
				Engine: &KubernetesHarborCache_External{
					External: &KubernetesHarborExternalRedis{
						Addr:              literal("s1:26379,s2:26379,s3:26379"),
						SentinelMasterSet: strPtr("mymaster"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("external_url contract", func() {
		ginkgo.It("should reject a trailing slash", func() {
			input.Spec.ExternalUrl = "https://harbor.example.com/"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a path suffix", func() {
			input.Spec.ExternalUrl = "https://harbor.example.com/registry"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a bare host with no scheme", func() {
			input.Spec.ExternalUrl = "harbor.example.com"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty external_url", func() {
			input.Spec.ExternalUrl = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("expose contract", func() {
		ginkgo.It("should reject the chart's ingress exposure type (exposure composes)", func() {
			input.Spec.Expose = &KubernetesHarborExpose{ServiceType: strPtr("Ingress")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an out-of-range node port", func() {
			input.Spec.Expose = &KubernetesHarborExpose{
				ServiceType: strPtr("NodePort"),
				NodePorts:   &KubernetesHarborNodePorts{Http: int32Ptr(20000)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("database contract", func() {
		ginkgo.It("should reject a database block with no engine", func() {
			input.Spec.Database = &KubernetesHarborDatabase{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an external database without a host", func() {
			db := externalDatabase()
			db.Host = nil
			input.Spec.Database = &KubernetesHarborDatabase{
				Engine: &KubernetesHarborDatabase_External{External: db},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an external database without a username", func() {
			db := externalDatabase()
			db.Username = ""
			input.Spec.Database = &KubernetesHarborDatabase{
				Engine: &KubernetesHarborDatabase_External{External: db},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an external database without a password secret", func() {
			db := externalDatabase()
			db.PasswordSecretName = nil
			input.Spec.Database = &KubernetesHarborDatabase{
				Engine: &KubernetesHarborDatabase_External{External: db},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid ssl_mode", func() {
			db := externalDatabase()
			db.SslMode = strPtr("prefer")
			input.Spec.Database = &KubernetesHarborDatabase{
				Engine: &KubernetesHarborDatabase_External{External: db},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed internal database disk size", func() {
			input.Spec.Database = &KubernetesHarborDatabase{
				Engine: &KubernetesHarborDatabase_Internal{
					Internal: &KubernetesHarborInternalDatabase{DiskSize: strPtr("10gigs")},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("cache contract", func() {
		ginkgo.It("should reject a cache block with no engine", func() {
			input.Spec.Cache = &KubernetesHarborCache{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an external redis without an address", func() {
			input.Spec.Cache = &KubernetesHarborCache{
				Engine: &KubernetesHarborCache_External{External: &KubernetesHarborExternalRedis{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject declaring both a password and an existing secret", func() {
			input.Spec.Cache = &KubernetesHarborCache{
				Engine: &KubernetesHarborCache_External{
					External: &KubernetesHarborExternalRedis{
						Addr:               literal("valkey:6379"),
						Password:           strPtr("x"),
						ExistingSecretName: strPtr("y"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("storage contract", func() {
		ginkgo.It("should reject a storage block with no backend", func() {
			input.Spec.Storage = &KubernetesHarborArtifactStorage{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject two registry replicas on filesystem without ReadWriteMany", func() {
			input.Spec.Registry = &KubernetesHarborComponent{Replicas: int32Ptr(2)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject s3 without a bucket", func() {
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_S3{
					S3: &KubernetesHarborS3Storage{Region: "us-east-1"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject s3 with an access key but no secret key", func() {
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_S3{
					S3: &KubernetesHarborS3Storage{
						Bucket: "b", Region: "us-east-1",
						Credentials: &KubernetesHarborS3Credentials{AccessKey: strPtr("AKIA...")},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject s3 with both declared keys and an existing secret", func() {
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_S3{
					S3: &KubernetesHarborS3Storage{
						Bucket: "b", Region: "us-east-1",
						Credentials: &KubernetesHarborS3Credentials{
							AccessKey:          strPtr("AKIA..."),
							SecretKey:          strPtr("shhh"),
							ExistingSecretName: strPtr("s3-auth"),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject gcs mixing workload identity with key material", func() {
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_Gcs{
					Gcs: &KubernetesHarborGcsStorage{
						Bucket:              "b",
						UseWorkloadIdentity: true,
						KeyData:             strPtr("base64"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject azure with no credential arm", func() {
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_Azure{
					Azure: &KubernetesHarborAzureStorage{AccountName: "acct", Container: "c"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject azure with both credential arms", func() {
			input.Spec.Storage = &KubernetesHarborArtifactStorage{
				Backend: &KubernetesHarborArtifactStorage_Azure{
					Azure: &KubernetesHarborAzureStorage{
						AccountName:        "acct",
						Container:          "c",
						AccountKey:         strPtr("k"),
						ExistingSecretName: strPtr("s"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("internal TLS contract", func() {
		ginkgo.It("should reject a cert-secrets block missing a component secret", func() {
			input.Spec.InternalTls = &KubernetesHarborInternalTls{
				Enabled: true,
				CertSecrets: &KubernetesHarborInternalTlsSecrets{
					CoreSecretName:       "core-tls",
					JobserviceSecretName: "js-tls",
					RegistrySecretName:   "reg-tls",
					// portal missing
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("misc field contracts", func() {
		ginkgo.It("should reject an invalid log level", func() {
			input.Spec.LogLevel = strPtr("verbose")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid update strategy", func() {
			input.Spec.UpdateStrategy = strPtr("BlueGreen")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject zero trivy replicas", func() {
			input.Spec.Trivy = &KubernetesHarborTrivy{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject zero cache-layer expire hours", func() {
			input.Spec.CacheLayer = &KubernetesHarborCacheLayer{Enabled: true, ExpireHours: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject zero jobservice max workers", func() {
			input.Spec.Jobservice = &KubernetesHarborJobservice{MaxJobWorkers: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
