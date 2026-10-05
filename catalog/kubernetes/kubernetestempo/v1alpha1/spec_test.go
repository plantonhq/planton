package kubernetestempov1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesTempo(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTempo Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func stringPtr(s string) *string { return &s }

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

// testS3 returns an S3-compatible storage backend shaped like an
// in-cluster SeaweedFS composition.
// testR2 returns an R2 backend wired the way a composition wires it: the
// account and bucket by reference to a CloudflareR2Bucket, the key pair
// from dashboard-minted secrets.
func testR2() *KubernetesTempoStorage {
	return &KubernetesTempoStorage{
		Backend: &KubernetesTempoStorage_R2{
			R2: &KubernetesTempoR2Storage{
				AccountId: valueFrom(catalogkind.CatalogKind_CloudflareR2Bucket, "traces-bucket", "status.outputs.account_id"),
				Bucket:    valueFrom(catalogkind.CatalogKind_CloudflareR2Bucket, "traces-bucket", "status.outputs.bucket_name"),
				Credentials: &KubernetesTempoR2Credentials{
					AccessKeyId:     literal("$secret/traces-writer-access-key-id"),
					SecretAccessKey: literal("$secret/traces-writer-secret-access-key"),
				},
			},
		},
	}
}

func testS3() *KubernetesTempoStorage {
	return &KubernetesTempoStorage{
		Backend: &KubernetesTempoStorage_S3{
			S3: &KubernetesTempoS3Storage{
				Bucket:         "tempo-traces",
				Endpoint:       "objects-filer.storage.svc.cluster.local:8333",
				ForcePathStyle: true,
				Insecure:       true,
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesTempo Validation Tests", func() {
	var input *KubernetesTempo

	ginkgo.BeforeEach(func() {
		input = &KubernetesTempo{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesTempo",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "traces",
			},
			Spec: &KubernetesTempoSpec{
				Namespace: literal("observability"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "observability", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("explicit local storage on one replica should be valid", func() {
			input.Spec.Storage = &KubernetesTempoStorage{
				Backend: &KubernetesTempoStorage_Local{Local: &KubernetesTempoLocalStorage{}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("r2 by reference should be valid", func() {
			input.Spec.Storage = testR2()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple replicas on r2 should be valid (r2 counts as object storage)", func() {
			input.Spec.Replicas = int32Ptr(3)
			input.Spec.Storage = testR2()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple replicas WITH object storage should be valid", func() {
			input.Spec.Replicas = int32Ptr(3)
			input.Spec.Storage = testS3()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("s3 with declared credentials should be valid", func() {
			s3 := testS3()
			s3.GetS3().Credentials = &KubernetesTempoObjectStoreCredentials{
				AccessKeyIdSecret:     &KubernetesTempoSecretKeyRef{Name: "tempo-s3", Key: "access-key-id"},
				SecretAccessKeySecret: &KubernetesTempoSecretKeyRef{Name: "tempo-s3", Key: "secret-access-key"},
			}
			input.Spec.Storage = s3
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gcs keyless and azure with a key secret should be valid", func() {
			input.Spec.Storage = &KubernetesTempoStorage{
				Backend: &KubernetesTempoStorage_Gcs{
					Gcs: &KubernetesTempoGcsStorage{Bucket: "acme-tempo-traces"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.Storage = &KubernetesTempoStorage{
				Backend: &KubernetesTempoStorage_Azure{
					Azure: &KubernetesTempoAzureStorage{
						AccountName:      "acmetraces",
						Container:        "tempo",
						AccountKeySecret: &KubernetesTempoSecretKeyRef{Name: "tempo-azure", Key: "account-key"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("ephemeral with the middleware-defaulted disk size should be valid", func() {
			// The platform's defaulting middleware stamps disk_size "10Gi"
			// onto every manifest — an ephemeral manifest must stay
			// expressible with it present.
			input.Spec.Ephemeral = true
			input.Spec.DiskSize = stringPtr("10Gi")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("retention in minutes or hours should be valid", func() {
			input.Spec.Retention = stringPtr("168h")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.Retention = stringPtr("30m")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("jaeger receivers and multi-tenancy should be valid", func() {
			enabled := true
			input.Spec.JaegerReceiversEnabled = &enabled
			input.Spec.MultiTenancyEnabled = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a metrics generator with a stack reference should be valid", func() {
			input.Spec.MetricsGenerator = &KubernetesTempoMetricsGenerator{
				Enabled:        true,
				RemoteWriteUrl: valueFrom(catalogkind.CatalogKind_KubernetesKubePrometheusStack, "monitoring", "status.outputs.prometheus_endpoint"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("metrics-generator processors should accept the two defined values", func() {
			input.Spec.MetricsGenerator = &KubernetesTempoMetricsGenerator{
				Enabled:        true,
				RemoteWriteUrl: literal("http://monitoring-prometheus.observability.svc.cluster.local:9090"),
				Processors: []KubernetesTempoMetricsGeneratorProcessor{
					KubernetesTempoMetricsGeneratorProcessor_service_graphs,
					KubernetesTempoMetricsGeneratorProcessor_span_metrics,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = stringPtr("2.2.3")
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.Storage = testS3()
			input.Spec.DiskSize = stringPtr("20Gi")
			input.Spec.StorageClass = literal("gp3")
			input.Spec.Retention = stringPtr("336h")
			input.Spec.TempoQueryEnabled = true
			input.Spec.ServiceMonitorEnabled = true
			input.Spec.ImageRegistry = "mirror.example.com"
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			input.Spec.Scheduling = &KubernetesTempoScheduling{
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

		ginkgo.It("multiple replicas WITHOUT object storage should fail", func() {
			input.Spec.Replicas = int32Ptr(2)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("multiple replicas with LOCAL storage should fail", func() {
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.Storage = &KubernetesTempoStorage{
				Backend: &KubernetesTempoStorage_Local{Local: &KubernetesTempoLocalStorage{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("r2 without credentials should fail", func() {
			storage := testR2()
			storage.GetR2().Credentials = nil
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

		ginkgo.It("s3 without a bucket should fail", func() {
			s3 := testS3()
			s3.GetS3().Bucket = ""
			input.Spec.Storage = s3
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("s3 without an endpoint should fail (Tempo does not derive it from the region)", func() {
			s3 := testS3()
			s3.GetS3().Endpoint = ""
			input.Spec.Storage = s3
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("ephemeral with a storage class should fail", func() {
			input.Spec.Ephemeral = true
			input.Spec.StorageClass = literal("gp3")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("ephemeral with a non-default disk size should fail", func() {
			input.Spec.Ephemeral = true
			input.Spec.DiskSize = stringPtr("50Gi")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("retention in days should fail (Tempo parses Go durations — no day unit)", func() {
			input.Spec.Retention = stringPtr("7d")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a malformed retention should fail", func() {
			input.Spec.Retention = stringPtr("forever")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an enabled metrics generator WITHOUT a remote-write url should fail", func() {
			input.Spec.MetricsGenerator = &KubernetesTempoMetricsGenerator{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate metrics-generator processors should fail", func() {
			input.Spec.MetricsGenerator = &KubernetesTempoMetricsGenerator{
				Enabled:        true,
				RemoteWriteUrl: literal("http://prometheus:9090"),
				Processors: []KubernetesTempoMetricsGeneratorProcessor{
					KubernetesTempoMetricsGeneratorProcessor_service_graphs,
					KubernetesTempoMetricsGeneratorProcessor_service_graphs,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unspecified metrics-generator processor should fail", func() {
			input.Spec.MetricsGenerator = &KubernetesTempoMetricsGenerator{
				Enabled:        true,
				RemoteWriteUrl: literal("http://prometheus:9090"),
				Processors: []KubernetesTempoMetricsGeneratorProcessor{
					KubernetesTempoMetricsGeneratorProcessor_kubernetes_tempo_metrics_generator_processor_unspecified,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a malformed disk_size should fail", func() {
			input.Spec.DiskSize = stringPtr("ten gigs")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
