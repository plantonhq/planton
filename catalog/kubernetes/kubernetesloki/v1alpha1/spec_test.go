package kuberneteslokiv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesLoki(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesLoki Suite")
}

func int32Ptr(i int32) *int32 { return &i }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind cloudresourcekind.CloudResourceKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
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

// testS3 returns an S3-compatible storage backend shaped like an
// in-cluster SeaweedFS composition.
func testS3() *KubernetesLokiStorage {
	return &KubernetesLokiStorage{
		Backend: &KubernetesLokiStorage_S3{
			S3: &KubernetesLokiS3Storage{
				Bucket:         "loki-chunks",
				Endpoint:       "http://objects-filer.storage.svc.cluster.local:8333",
				ForcePathStyle: true,
				Insecure:       true,
			},
		},
	}
}

// testR2 returns an R2 backend wired the way a composition wires it: the
// account and bucket by reference to a CloudflareR2Bucket, the key pair
// from dashboard-minted secrets.
func testR2() *KubernetesLokiStorage {
	return &KubernetesLokiStorage{
		Backend: &KubernetesLokiStorage_R2{
			R2: &KubernetesLokiR2Storage{
				AccountId: valueFrom(cloudresourcekind.CloudResourceKind_CloudflareR2Bucket, "logs-bucket", "status.outputs.account_id"),
				Bucket:    valueFrom(cloudresourcekind.CloudResourceKind_CloudflareR2Bucket, "logs-bucket", "status.outputs.bucket_name"),
				Credentials: &KubernetesLokiR2Credentials{
					AccessKeyId:     literal("$secret/logs-writer-access-key-id"),
					SecretAccessKey: literal("$secret/logs-writer-secret-access-key"),
				},
			},
		},
	}
}

// bcryptHash is a syntactically valid htpasswd bcrypt hash (the chart's
// own documented example shape).
const bcryptHash = "$2y$10$7O40CaY1yz7fu9O24k2/u.ct/wELYHRBsn25v/7AyuQ8E8hrLqpva"

var _ = ginkgo.Describe("KubernetesLoki Validation Tests", func() {
	var input *KubernetesLoki

	ginkgo.BeforeEach(func() {
		input = &KubernetesLoki{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesLoki",
			Metadata: &shared.CloudResourceMetadata{
				Name: "logs",
			},
			Spec: &KubernetesLokiSpec{
				Namespace: literal("observability"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "observability", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a single monolithic replica on filesystem storage should be valid", func() {
			input.Spec.Mode = &KubernetesLokiSpec_Monolithic{
				Monolithic: &KubernetesLokiMonolithic{Replicas: int32Ptr(1)},
			}
			input.Spec.Storage = &KubernetesLokiStorage{
				Backend: &KubernetesLokiStorage_Filesystem{Filesystem: &KubernetesLokiFilesystemStorage{}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple monolithic replicas WITH object storage should be valid", func() {
			input.Spec.Mode = &KubernetesLokiSpec_Monolithic{
				Monolithic: &KubernetesLokiMonolithic{Replicas: int32Ptr(3)},
			}
			input.Spec.Storage = testS3()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("simple_scalable with s3 storage should be valid", func() {
			input.Spec.Mode = &KubernetesLokiSpec_SimpleScalable{
				SimpleScalable: &KubernetesLokiSimpleScalable{
					WriteReplicas:   int32Ptr(3),
					ReadReplicas:    int32Ptr(2),
					BackendReplicas: int32Ptr(3),
				},
			}
			input.Spec.Storage = testS3()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("r2 by reference should be valid", func() {
			input.Spec.Storage = testR2()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("r2 with a literal account and a us jurisdiction should be valid", func() {
			storage := testR2()
			storage.GetR2().AccountId = literal("074755a78d8e8f77c119a90a125e8a06")
			storage.GetR2().Jurisdiction = literal("us")
			input.Spec.Storage = storage
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple monolithic replicas on r2 should be valid (r2 counts as object storage)", func() {
			input.Spec.Mode = &KubernetesLokiSpec_Monolithic{
				Monolithic: &KubernetesLokiMonolithic{Replicas: int32Ptr(3)},
			}
			input.Spec.Storage = testR2()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("simple_scalable on r2 should be valid (r2 counts as object storage)", func() {
			input.Spec.Mode = &KubernetesLokiSpec_SimpleScalable{
				SimpleScalable: &KubernetesLokiSimpleScalable{},
			}
			input.Spec.Storage = testR2()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("s3 with declared credentials should be valid", func() {
			s3 := testS3()
			s3.GetS3().Credentials = &KubernetesLokiObjectStoreCredentials{
				AccessKeyIdSecret:     &KubernetesLokiSecretKeyRef{Name: "loki-s3", Key: "access-key-id"},
				SecretAccessKeySecret: &KubernetesLokiSecretKeyRef{Name: "loki-s3", Key: "secret-access-key"},
			}
			input.Spec.Storage = s3
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gcs keyless (ambient workload identity) should be valid", func() {
			input.Spec.Storage = &KubernetesLokiStorage{
				Backend: &KubernetesLokiStorage_Gcs{
					Gcs: &KubernetesLokiGcsStorage{Bucket: "acme-loki-chunks"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("azure with an account-key secret should be valid", func() {
			input.Spec.Storage = &KubernetesLokiStorage{
				Backend: &KubernetesLokiStorage_Azure{
					Azure: &KubernetesLokiAzureStorage{
						AccountName:      "acmelogs",
						Container:        "loki",
						AccountKeySecret: &KubernetesLokiSecretKeyRef{Name: "loki-azure", Key: "account-key"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("retention in hours and days should be valid", func() {
			input.Spec.RetentionPeriod = "744h"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.RetentionPeriod = "31d"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a schema_from_date override should be valid", func() {
			input.Spec.SchemaFromDate = "2024-04-01"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("limits should be valid", func() {
			input.Spec.Limits = &KubernetesLokiLimits{
				IngestionRateMb:         int32Ptr(8),
				IngestionBurstSizeMb:    int32Ptr(12),
				MaxGlobalStreamsPerUser: int32Ptr(10000),
				MaxQueryLookback:        "720h",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multi-tenancy with hash-declared tenants should be valid", func() {
			input.Spec.MultiTenancy = &KubernetesLokiMultiTenancy{
				Enabled: true,
				Tenants: []*KubernetesLokiTenant{
					{Name: "team-a", PasswordHash: bcryptHash},
					{Name: "team-b", PasswordHash: bcryptHash},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multi-tenancy with an existing htpasswd secret should be valid", func() {
			input.Spec.MultiTenancy = &KubernetesLokiMultiTenancy{
				Enabled:                true,
				ExistingHtpasswdSecret: "loki-gateway-auth",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gateway and caching tuning should be valid", func() {
			enabled := true
			input.Spec.Gateway = &KubernetesLokiGateway{Enabled: &enabled, Replicas: int32Ptr(2)}
			input.Spec.Caching = &KubernetesLokiCaching{
				ChunksCacheMemoryMb:  int32Ptr(1024),
				ResultsCacheMemoryMb: int32Ptr(256),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a ruler with an alertmanager reference should be valid", func() {
			input.Spec.Ruler = &KubernetesLokiRuler{
				Enabled:         true,
				AlertmanagerUrl: valueFrom(cloudresourcekind.CloudResourceKind_KubernetesKubePrometheusStack, "monitoring", "status.outputs.alertmanager_endpoint"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface should be valid", func() {
			canary := false
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = func() *string { s := "18.5.4"; return &s }()
			input.Spec.Mode = &KubernetesLokiSpec_SimpleScalable{
				SimpleScalable: &KubernetesLokiSimpleScalable{
					WriteReplicas: int32Ptr(3),
					DiskSize:      func() *string { s := "20Gi"; return &s }(),
					StorageClass:  literal("gp3"),
				},
			}
			input.Spec.Storage = testS3()
			input.Spec.RetentionPeriod = "744h"
			input.Spec.CanaryEnabled = &canary
			input.Spec.ServiceMonitorEnabled = true
			input.Spec.ImageRegistry = "mirror.example.com"
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			input.Spec.Scheduling = &KubernetesLokiScheduling{
				NodeSelector: map[string]string{"workload": "observability"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a namespace-less spec should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("simple_scalable WITHOUT storage should fail (tiers rendezvous in the object store)", func() {
			input.Spec.Mode = &KubernetesLokiSpec_SimpleScalable{
				SimpleScalable: &KubernetesLokiSimpleScalable{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("simple_scalable with filesystem storage should fail", func() {
			input.Spec.Mode = &KubernetesLokiSpec_SimpleScalable{
				SimpleScalable: &KubernetesLokiSimpleScalable{},
			}
			input.Spec.Storage = &KubernetesLokiStorage{
				Backend: &KubernetesLokiStorage_Filesystem{Filesystem: &KubernetesLokiFilesystemStorage{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("multiple monolithic replicas WITHOUT object storage should fail", func() {
			input.Spec.Mode = &KubernetesLokiSpec_Monolithic{
				Monolithic: &KubernetesLokiMonolithic{Replicas: int32Ptr(3)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero monolithic replicas should fail", func() {
			input.Spec.Mode = &KubernetesLokiSpec_Monolithic{
				Monolithic: &KubernetesLokiMonolithic{Replicas: int32Ptr(0)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a malformed schema_from_date should fail", func() {
			input.Spec.SchemaFromDate = "April 1, 2024"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a malformed retention_period should fail", func() {
			input.Spec.RetentionPeriod = "744"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			input.Spec.RetentionPeriod = "4w"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("tenants without multi_tenancy.enabled should fail", func() {
			input.Spec.MultiTenancy = &KubernetesLokiMultiTenancy{
				Tenants: []*KubernetesLokiTenant{{Name: "team-a", PasswordHash: bcryptHash}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an htpasswd secret without multi_tenancy.enabled should fail", func() {
			input.Spec.MultiTenancy = &KubernetesLokiMultiTenancy{
				ExistingHtpasswdSecret: "loki-gateway-auth",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("declaring BOTH tenants and an htpasswd secret should fail", func() {
			input.Spec.MultiTenancy = &KubernetesLokiMultiTenancy{
				Enabled:                true,
				Tenants:                []*KubernetesLokiTenant{{Name: "team-a", PasswordHash: bcryptHash}},
				ExistingHtpasswdSecret: "loki-gateway-auth",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a tenant with a plaintext password instead of a bcrypt hash should fail", func() {
			input.Spec.MultiTenancy = &KubernetesLokiMultiTenancy{
				Enabled: true,
				Tenants: []*KubernetesLokiTenant{{Name: "team-a", PasswordHash: "hunter2"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a tenant without a name should fail", func() {
			input.Spec.MultiTenancy = &KubernetesLokiMultiTenancy{
				Enabled: true,
				Tenants: []*KubernetesLokiTenant{{PasswordHash: bcryptHash}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("s3 without a bucket should fail", func() {
			s3 := testS3()
			s3.GetS3().Bucket = ""
			input.Spec.Storage = s3
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("r2 without credentials should fail", func() {
			storage := testR2()
			storage.GetR2().Credentials = nil
			input.Spec.Storage = storage
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("r2 without a bucket should fail", func() {
			storage := testR2()
			storage.GetR2().Bucket = nil
			input.Spec.Storage = storage
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("r2 with a malformed account id should fail", func() {
			storage := testR2()
			storage.GetR2().AccountId = literal("not-an-account")
			input.Spec.Storage = storage
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("r2 with an unknown jurisdiction should fail", func() {
			storage := testR2()
			storage.GetR2().Jurisdiction = literal("apac")
			input.Spec.Storage = storage
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("gcs without a bucket should fail", func() {
			input.Spec.Storage = &KubernetesLokiStorage{
				Backend: &KubernetesLokiStorage_Gcs{Gcs: &KubernetesLokiGcsStorage{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("azure without an account name should fail", func() {
			input.Spec.Storage = &KubernetesLokiStorage{
				Backend: &KubernetesLokiStorage_Azure{
					Azure: &KubernetesLokiAzureStorage{Container: "loki"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("declared credentials missing the secret key should fail", func() {
			s3 := testS3()
			s3.GetS3().Credentials = &KubernetesLokiObjectStoreCredentials{
				AccessKeyIdSecret:     &KubernetesLokiSecretKeyRef{Name: "loki-s3"},
				SecretAccessKeySecret: &KubernetesLokiSecretKeyRef{Name: "loki-s3", Key: "secret-access-key"},
			}
			input.Spec.Storage = s3
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a zero ingestion rate should fail", func() {
			input.Spec.Limits = &KubernetesLokiLimits{IngestionRateMb: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a malformed max_query_lookback should fail", func() {
			input.Spec.Limits = &KubernetesLokiLimits{MaxQueryLookback: "yesterday"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a malformed disk_size should fail", func() {
			input.Spec.Mode = &KubernetesLokiSpec_Monolithic{
				Monolithic: &KubernetesLokiMonolithic{
					DiskSize: func() *string { s := "ten gigs"; return &s }(),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an undersized chunks-cache memory should fail", func() {
			input.Spec.Caching = &KubernetesLokiCaching{ChunksCacheMemoryMb: int32Ptr(16)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
