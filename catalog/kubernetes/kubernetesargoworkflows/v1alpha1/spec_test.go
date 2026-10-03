package kubernetesargoworkflowsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesArgoWorkflows(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesArgoWorkflows Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }
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

func testArchive() *KubernetesArgoWorkflowsArchive {
	return &KubernetesArgoWorkflowsArchive{
		Engine:   KubernetesArgoWorkflowsArchiveEngine_postgres,
		Host:     literal("workflows-db-rw.data.svc.cluster.local"),
		Database: "argo_archive",
		CredentialsSecret: &KubernetesArgoWorkflowsArchiveCredentials{
			Name: literal("workflows-db-app"),
		},
	}
}

var _ = ginkgo.Describe("KubernetesArgoWorkflows Validation Tests", func() {
	var input *KubernetesArgoWorkflows

	ginkgo.BeforeEach(func() {
		input = &KubernetesArgoWorkflows{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesArgoWorkflows",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "pipelines",
			},
			Spec: &KubernetesArgoWorkflowsSpec{
				Namespace: literal("argo"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "argo", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("controller knobs should be valid", func() {
			input.Spec.Controller = &KubernetesArgoWorkflowsController{
				Replicas:             int32Ptr(2),
				WorkflowNamespaces:   []string{"team-a", "team-b"},
				InstanceId:           "platform",
				Parallelism:          int32Ptr(50),
				NamespaceParallelism: int32Ptr(10),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("server auth modes should be valid", func() {
			input.Spec.Server = &KubernetesArgoWorkflowsServer{
				AuthModes: []string{"client", "sso"},
				BaseHref:  "https://workflows.example.com/",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a disabled server should be valid", func() {
			input.Spec.Server = &KubernetesArgoWorkflowsServer{Enabled: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an s3 artifact repository with declared keys should be valid", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				ArchiveLogs: true,
				Backend: &KubernetesArgoWorkflowsArtifactRepository_S3{S3: &KubernetesArgoWorkflowsArtifactS3{
					Bucket:   "workflow-artifacts",
					Endpoint: literal("objects-s3.storage.svc.cluster.local:8333"),
					Insecure: true,
					CredentialsSecret: &KubernetesArgoWorkflowsS3CredentialsSecret{
						SecretName: literal("objects-s3-auth"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an s3 credentials Secret with explicit key names should be valid", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_S3{S3: &KubernetesArgoWorkflowsArtifactS3{
					Bucket:   "workflow-artifacts",
					Endpoint: literal("objects-s3.storage.svc.cluster.local:8333"),
					Insecure: true,
					CredentialsSecret: &KubernetesArgoWorkflowsS3CredentialsSecret{
						SecretName:         literal("generic-s3-auth"),
						AccessKeyIdKey:     stringPtr("accesskey"),
						SecretAccessKeyKey: stringPtr("secretkey"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an s3 endpoint and credentials Secret referencing a KubernetesSeaweedFs should be valid", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_S3{S3: &KubernetesArgoWorkflowsArtifactS3{
					Bucket:   "workflow-artifacts",
					Endpoint: valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "objects", "status.outputs.s3_endpoint"),
					Insecure: true,
					CredentialsSecret: &KubernetesArgoWorkflowsS3CredentialsSecret{
						SecretName: valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "objects", "status.outputs.s3_credentials_secret_name"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a keyless s3 repository (ambient pod identity) should be valid", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_S3{S3: &KubernetesArgoWorkflowsArtifactS3{
					Bucket:                "workflow-artifacts",
					Region:                "us-west-2",
					UseAmbientCredentials: true,
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a gcs repository, keyless and keyed, should be valid", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_Gcs{Gcs: &KubernetesArgoWorkflowsArtifactGcs{
					Bucket: "workflow-artifacts",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_Gcs{Gcs: &KubernetesArgoWorkflowsArtifactGcs{
					Bucket:                "workflow-artifacts",
					CredentialsSecretName: "gcs-sa-key",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an azure repository should be valid", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_Azure{Azure: &KubernetesArgoWorkflowsArtifactAzure{
					Endpoint:  "https://myaccount.blob.core.windows.net",
					Container: "artifacts",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a postgres archive with KubernetesPostgres host and credentials references should be valid", func() {
			archive := testArchive()
			archive.Host = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "workflows-db", "status.outputs.rw_service")
			archive.CredentialsSecret.Name = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "workflows-db", "status.outputs.password_secret.name")
			archive.SslMode = "require"
			input.Spec.Archive = archive
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a mysql archive with an explicit port should be valid", func() {
			archive := testArchive()
			archive.Engine = KubernetesArgoWorkflowsArchiveEngine_mysql
			archive.Port = int32Ptr(3306)
			input.Spec.Archive = archive
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("retention policy counts should be valid", func() {
			input.Spec.RetentionPolicy = &KubernetesArgoWorkflowsRetentionPolicy{
				Completed: int32Ptr(10), Failed: int32Ptr(3), Errored: int32Ptr(0),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("crds lifecycle incl. the air-gap arm should be valid", func() {
			input.Spec.Crds = &KubernetesArgoWorkflowsCrds{
				Install:    boolPtr(true),
				Keep:       boolPtr(true),
				FullSchema: boolPtr(false),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			input.Spec.Crds = &KubernetesArgoWorkflowsCrds{
				FullSchema: boolPtr(true),
				BaseUrl:    "https://mirror.internal/argo-crds",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("image override and scheduling should be valid", func() {
			input.Spec.Image = &KubernetesArgoWorkflowsImage{Registry: "my.registry.com", Tag: "v4.0.8", PullSecretName: "mirror-pull"}
			input.Spec.Scheduling = &KubernetesArgoWorkflowsScheduling{
				NodeSelector:      map[string]string{"role": "pipelines"},
				PriorityClassName: "batch-normal",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("helm values escape hatch should be valid", func() {
			input.Spec.HelmValues = "controller:\n  workflowDefaults:\n    spec:\n      ttlStrategy:\n        secondsAfterCompletion: 86400\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero controller replicas should fail", func() {
			input.Spec.Controller = &KubernetesArgoWorkflowsController{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero parallelism should fail (omit it for unlimited)", func() {
			input.Spec.Controller = &KubernetesArgoWorkflowsController{Parallelism: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unsupported auth mode should fail", func() {
			input.Spec.Server = &KubernetesArgoWorkflowsServer{AuthModes: []string{"token"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate auth modes should fail", func() {
			input.Spec.Server = &KubernetesArgoWorkflowsServer{AuthModes: []string{"client", "client"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an s3 repository without a bucket should fail", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_S3{S3: &KubernetesArgoWorkflowsArtifactS3{
					CredentialsSecret: &KubernetesArgoWorkflowsS3CredentialsSecret{
						SecretName: literal("objects-s3-auth"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an s3 credentials Secret without a name should fail", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_S3{S3: &KubernetesArgoWorkflowsArtifactS3{
					Bucket:            "workflow-artifacts",
					CredentialsSecret: &KubernetesArgoWorkflowsS3CredentialsSecret{},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an s3 repository with NO credential path should fail the exactly-one rule", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_S3{S3: &KubernetesArgoWorkflowsArtifactS3{
					Bucket: "workflow-artifacts",
				}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("credential"))
		})

		ginkgo.It("an s3 repository with BOTH credential paths should fail the exactly-one rule", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_S3{S3: &KubernetesArgoWorkflowsArtifactS3{
					Bucket:                "workflow-artifacts",
					UseAmbientCredentials: true,
					CredentialsSecret: &KubernetesArgoWorkflowsS3CredentialsSecret{
						SecretName: literal("objects-s3-auth"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an azure repository without a container should fail", func() {
			input.Spec.ArtifactRepository = &KubernetesArgoWorkflowsArtifactRepository{
				Backend: &KubernetesArgoWorkflowsArtifactRepository_Azure{Azure: &KubernetesArgoWorkflowsArtifactAzure{
					Endpoint: "https://myaccount.blob.core.windows.net",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an archive without an engine should fail", func() {
			archive := testArchive()
			archive.Engine = KubernetesArgoWorkflowsArchiveEngine_kubernetes_argo_workflows_archive_engine_unspecified
			input.Spec.Archive = archive
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an archive without credentials should fail", func() {
			archive := testArchive()
			archive.CredentialsSecret = nil
			input.Spec.Archive = archive
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("archive credentials without a Secret name should fail", func() {
			archive := testArchive()
			archive.CredentialsSecret = &KubernetesArgoWorkflowsArchiveCredentials{}
			input.Spec.Archive = archive
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an archive with an invalid ssl mode should fail", func() {
			archive := testArchive()
			archive.SslMode = "prefer"
			input.Spec.Archive = archive
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an archive port out of range should fail", func() {
			archive := testArchive()
			archive.Port = int32Ptr(0)
			input.Spec.Archive = archive
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a negative retention count should fail", func() {
			input.Spec.RetentionPolicy = &KubernetesArgoWorkflowsRetentionPolicy{Completed: int32Ptr(-1)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
