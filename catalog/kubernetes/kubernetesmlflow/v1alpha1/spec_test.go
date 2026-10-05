package kubernetesmlflowv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesMlflow(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesMlflow Suite")
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

func testMlflowPostgres() *KubernetesMlflowPostgres {
	return &KubernetesMlflowPostgres{
		Host: literal("mlflow-pg-rw"),
		PasswordSecret: &KubernetesMlflowPasswordSecret{
			SecretName: literal("mlflow-pg-app"),
		},
	}
}

func testS3Compatible() *KubernetesMlflowS3CompatibleArtifacts {
	return &KubernetesMlflowS3CompatibleArtifacts{
		Endpoint: literal("http://seaweedfs-s3.mlflow.svc.cluster.local:8333"),
		Bucket:   "mlflow-artifacts",
		CredentialsSecret: &KubernetesMlflowS3CredentialsSecret{
			SecretName: literal("seaweedfs-s3-secret"),
		},
	}
}

var _ = ginkgo.Describe("KubernetesMlflow Validation Tests", func() {
	var input *KubernetesMlflow

	ginkgo.BeforeEach(func() {
		input = &KubernetesMlflow{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesMlflow",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "mlflow",
			},
			Spec: &KubernetesMlflowSpec{
				Namespace: literal("mlflow"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (namespace only — sqlite + PVC artifacts + auth defaults) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "mlflow", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("explicit sqlite backend with sizing should be valid", func() {
			input.Spec.BackendStore = &KubernetesMlflowBackendStore{
				Backend: &KubernetesMlflowBackendStore_SqlitePvc{SqlitePvc: &KubernetesMlflowSqlitePvc{
					StorageSize:  strPtr("10Gi"),
					StorageClass: "fast-ssd",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("postgres backend composed from a KubernetesPostgres should be valid", func() {
			pg := testMlflowPostgres()
			pg.Host = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "mlflow-pg", "status.outputs.rw_service")
			pg.PasswordSecret.SecretName = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "mlflow-pg", "status.outputs.password_secret.name")
			input.Spec.BackendStore = &KubernetesMlflowBackendStore{
				Backend: &KubernetesMlflowBackendStore_Postgres{Postgres: pg},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("postgres tuning (port, database, user) should be valid", func() {
			pg := testMlflowPostgres()
			pg.Port = int32Ptr(5433)
			pg.DatabaseName = strPtr("mlflow_prod")
			pg.Username = strPtr("mlflow_svc")
			input.Spec.BackendStore = &KubernetesMlflowBackendStore{
				Backend: &KubernetesMlflowBackendStore_Postgres{Postgres: pg},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("mysql backend arm should be valid", func() {
			input.Spec.BackendStore = &KubernetesMlflowBackendStore{
				Backend: &KubernetesMlflowBackendStore_Mysql{Mysql: &KubernetesMlflowMysql{
					Host: literal("mlflow-mysql"),
					PasswordSecret: &KubernetesMlflowMysqlPasswordSecret{
						SecretName: literal("mlflow-mysql-auth"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("s3-compatible artifacts composed from a KubernetesSeaweedFs should be valid", func() {
			s3 := testS3Compatible()
			s3.Endpoint = valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "artifacts", "status.outputs.s3_endpoint")
			s3.CredentialsSecret.SecretName = valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "artifacts", "status.outputs.s3_credentials_secret_name")
			s3.Prefix = "runs"
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_S3Compatible{S3Compatible: s3},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("PVC artifacts with sizing should be valid", func() {
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_Pvc{Pvc: &KubernetesMlflowPvcArtifacts{
					StorageSize:  strPtr("50Gi"),
					StorageClass: "standard",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("AWS S3 artifacts keyless (ambient identity) should be valid", func() {
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_AwsS3{AwsS3: &KubernetesMlflowAwsS3Artifacts{
					Bucket: "ml-artifacts",
					Region: "us-west-2",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("AWS S3 artifacts with declared keys should be valid", func() {
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_AwsS3{AwsS3: &KubernetesMlflowAwsS3Artifacts{
					Bucket: "ml-artifacts",
					Region: "us-west-2",
					CredentialsSecret: &KubernetesMlflowAwsCredentialsSecret{
						SecretName: "aws-keys",
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("GCS artifacts with a key Secret should be valid", func() {
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_Gcs{Gcs: &KubernetesMlflowGcsArtifacts{
					Bucket: "ml-artifacts",
					CredentialsSecret: &KubernetesMlflowGcpCredentialsSecret{
						SecretName: "gcp-sa-key",
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("Azure Blob artifacts should be valid", func() {
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_AzureBlob{AzureBlob: &KubernetesMlflowAzureBlobArtifacts{
					StorageAccount: "mlartifacts",
					Container:      "mlflow",
					CredentialsSecret: &KubernetesMlflowAzureCredentialsSecret{
						SecretName: "azure-storage-key",
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple replicas with postgres + object artifacts should be valid", func() {
			input.Spec.Server = &KubernetesMlflowServer{Replicas: int32Ptr(3)}
			input.Spec.BackendStore = &KubernetesMlflowBackendStore{
				Backend: &KubernetesMlflowBackendStore_Postgres{Postgres: testMlflowPostgres()},
			}
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_S3Compatible{S3Compatible: testS3Compatible()},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("auth tuning with an existing admin Secret should be valid", func() {
			input.Spec.Auth = &KubernetesMlflowAuth{
				Enabled:       boolPtr(true),
				AdminUsername: strPtr("mlops"),
				AdminPasswordSecret: &KubernetesMlflowSecretKeyRef{
					SecretName: "mlops-admin",
					SecretKey:  strPtr("pass"),
				},
				DefaultPermission: strPtr("NO_PERMISSIONS"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("auth disabled should be valid", func() {
			input.Spec.Auth = &KubernetesMlflowAuth{Enabled: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gc schedule and retention should be valid", func() {
			input.Spec.Gc = &KubernetesMlflowGc{
				Enabled:   true,
				Schedule:  strPtr("30 2 * * 0"),
				OlderThan: strPtr("72h"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("server image, workers and resources should be valid", func() {
			input.Spec.Server = &KubernetesMlflowServer{
				Image:   &KubernetesMlflowImage{Repository: strPtr("mirror.example.com/mlflow/mlflow"), Tag: strPtr("v3.15.0")},
				Workers: int32Ptr(8),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("metrics with a ServiceMonitor should be valid", func() {
			input.Spec.Metrics = &KubernetesMlflowMetrics{Enabled: true, ServiceMonitorEnabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("service exposure and extra env should be valid", func() {
			input.Spec.Service = &KubernetesMlflowService{
				Type:        strPtr("LoadBalancer"),
				Annotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-type": "nlb"},
			}
			input.Spec.ExtraEnv = map[string]string{"MLFLOW_SERVE_ARTIFACTS": "true"}
			input.Spec.ExtraEnvFromSecret = map[string]*KubernetesMlflowSecretKeyRef{
				"LDAP_BIND_PASSWORD": {SecretName: "ldap-bind", SecretKey: strPtr("password")},
			}
			input.Spec.ExtraArgs = []string{"--gunicorn-opts=--timeout 120"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("empty backend-store oneof should fail", func() {
			input.Spec.BackendStore = &KubernetesMlflowBackendStore{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("empty artifact-store oneof should fail", func() {
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("postgres without a password secret should fail", func() {
			input.Spec.BackendStore = &KubernetesMlflowBackendStore{
				Backend: &KubernetesMlflowBackendStore_Postgres{Postgres: &KubernetesMlflowPostgres{
					Host: literal("mlflow-pg-rw"),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("postgres with an invalid database name should fail", func() {
			pg := testMlflowPostgres()
			pg.DatabaseName = strPtr("bad-name!")
			input.Spec.BackendStore = &KubernetesMlflowBackendStore{
				Backend: &KubernetesMlflowBackendStore_Postgres{Postgres: pg},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("multiple replicas on the sqlite backend should fail", func() {
			input.Spec.Server = &KubernetesMlflowServer{Replicas: int32Ptr(2)}
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_S3Compatible{S3Compatible: testS3Compatible()},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("multiple replicas on PVC artifacts should fail", func() {
			input.Spec.Server = &KubernetesMlflowServer{Replicas: int32Ptr(2)}
			input.Spec.BackendStore = &KubernetesMlflowBackendStore{
				Backend: &KubernetesMlflowBackendStore_Postgres{Postgres: testMlflowPostgres()},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("s3-compatible artifacts without a bucket should fail", func() {
			s3 := testS3Compatible()
			s3.Bucket = ""
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_S3Compatible{S3Compatible: s3},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("s3-compatible artifacts without credentials should fail", func() {
			s3 := testS3Compatible()
			s3.CredentialsSecret = nil
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_S3Compatible{S3Compatible: s3},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("AWS S3 artifacts without a region should fail", func() {
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_AwsS3{AwsS3: &KubernetesMlflowAwsS3Artifacts{
					Bucket: "ml-artifacts",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("Azure Blob artifacts without credentials should fail", func() {
			input.Spec.ArtifactStore = &KubernetesMlflowArtifactStore{
				Backend: &KubernetesMlflowArtifactStore_AzureBlob{AzureBlob: &KubernetesMlflowAzureBlobArtifacts{
					StorageAccount: "mlartifacts",
					Container:      "mlflow",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("default permission outside the allowed set should fail", func() {
			input.Spec.Auth = &KubernetesMlflowAuth{DefaultPermission: strPtr("ADMIN")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("gc with a malformed schedule should fail", func() {
			input.Spec.Gc = &KubernetesMlflowGc{Enabled: true, Schedule: strPtr("whenever")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("gc with a malformed retention should fail", func() {
			input.Spec.Gc = &KubernetesMlflowGc{Enabled: true, OlderThan: strPtr("30 days")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("ServiceMonitor without metrics should fail", func() {
			input.Spec.Metrics = &KubernetesMlflowMetrics{ServiceMonitorEnabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("service type outside the allowed set should fail", func() {
			input.Spec.Service = &KubernetesMlflowService{Type: strPtr("ExternalName")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("workers above the ceiling should fail", func() {
			input.Spec.Server = &KubernetesMlflowServer{Workers: int32Ptr(100)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("wrong kind constant should fail", func() {
			input.Kind = "Mlflow"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
