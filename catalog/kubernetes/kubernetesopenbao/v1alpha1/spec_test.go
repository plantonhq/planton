package kubernetesopenbaov1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesOpenBao(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesOpenBao Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

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

var _ = ginkgo.Describe("KubernetesOpenBao Validation Tests", func() {
	var input *KubernetesOpenBao

	ginkgo.BeforeEach(func() {
		input = &KubernetesOpenBao{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesOpenBao",
			Metadata: &shared.CloudResourceMetadata{
				Name: "openbao",
			},
			Spec: &KubernetesOpenBaoSpec{
				Namespace: literal("openbao"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec (namespace only, chart defaults) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "openbao", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a maximal spec (every block populated) should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = strPtr("0.28.6")
			input.Spec.Server = &KubernetesOpenBaoServer{
				Mode: &KubernetesOpenBaoServer_Ha{Ha: &KubernetesOpenBaoHaMode{Replicas: int32Ptr(3)}},
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "256Mi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "1", Memory: "1Gi"},
				},
				DataStorage: &KubernetesOpenBaoStorage{
					Size:         strPtr("20Gi"),
					StorageClass: literal("fast-ssd"),
				},
				AuditStorage: &KubernetesOpenBaoStorage{Size: strPtr("5Gi")},
				LogLevel:     strPtr("debug"),
				LogFormat:    strPtr("json"),
				Scheduling: &KubernetesOpenBaoScheduling{
					NodeSelector: map[string]string{"workload": "secrets"},
				},
			}
			input.Spec.Tls = &KubernetesOpenBaoTls{
				Enabled:        true,
				CertSecretName: literal("openbao-server-tls"),
			}
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_AwsKms{
					AwsKms: &KubernetesOpenBaoAwsKmsSeal{
						Region:   "us-west-2",
						KmsKeyId: "alias/openbao-unseal",
					},
				},
			}
			input.Spec.Injector = &KubernetesOpenBaoInjector{
				Enabled:       true,
				Replicas:      int32Ptr(2),
				FailurePolicy: strPtr("Fail"),
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "50m", Memory: "64Mi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "250m", Memory: "256Mi"},
				},
			}
			input.Spec.UiEnabled = boolPtr(true)
			input.Spec.NetworkPolicyEnabled = true
			input.Spec.Metrics = &KubernetesOpenBaoMetrics{
				Enabled:               true,
				ServiceMonitorEnabled: true,
			}
			input.Spec.Backup = &KubernetesOpenBaoBackup{
				Schedule:      strPtr("0 */6 * * *"),
				RetentionDays: int32Ptr(30),
				ObjectStore: &KubernetesOpenBaoBackupObjectStore{
					Prefix: "openbao/prod",
					Backend: &KubernetesOpenBaoBackupObjectStore_S3{S3: &KubernetesOpenBaoS3ObjectStore{
						Bucket:  "openbao-snapshots",
						Region:  "us-west-2",
						Keyless: true,
					}},
				},
				WorkloadIdentity: &kubernetes.KubernetesWorkloadIdentity{
					Provider: &kubernetes.KubernetesWorkloadIdentity_Eks{Eks: &kubernetes.KubernetesWorkloadIdentityEksIrsa{
						RoleArn: literal("arn:aws:iam::123456789012:role/openbao-backup"),
					}},
				},
				Auth:   &KubernetesOpenBaoBackupAuth{MountPath: strPtr("kubernetes"), Role: "openbao-backup"},
				Images: &KubernetesOpenBaoBackupImages{Rclone: &kubernetes.ContainerImage{Repo: "mirror.example.com/rclone/rclone", Tag: "1.75.1"}},
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "50m", Memory: "64Mi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "500m", Memory: "512Mi"},
				},
			}
			input.Spec.Restore = &KubernetesOpenBaoRestore{
				Source:    &KubernetesOpenBaoRestore_Latest{Latest: true},
				RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init", Key: "root_token"},
			}
			input.Spec.ServiceAccount = &KubernetesOpenBaoServiceAccount{
				Annotations:          map[string]string{"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/openbao"},
				AuthDelegatorEnabled: boolPtr(true),
			}
			input.Spec.HelmValues = "server:\n  extraLabels:\n    team: platform\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("dev mode should be valid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{
				Mode: &KubernetesOpenBaoServer_Dev{Dev: &KubernetesOpenBaoDevMode{}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("standalone mode should be valid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{
				Mode: &KubernetesOpenBaoServer_Standalone{Standalone: &KubernetesOpenBaoStandaloneMode{}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a single-replica HA cluster (Raft cluster of one) should be valid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{
				Mode: &KubernetesOpenBaoServer_Ha{Ha: &KubernetesOpenBaoHaMode{Replicas: int32Ptr(1)}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a TLS cert secret from a Certificate reference should be valid", func() {
			input.Spec.Tls = &KubernetesOpenBaoTls{
				Enabled:        true,
				CertSecretName: valueFrom(cloudresourcekind.CloudResourceKind_KubernetesCertificate, "openbao-cert", "status.outputs.secret_name"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a GCP KMS seal should be valid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
					GcpKms: &KubernetesOpenBaoGcpKmsSeal{
						Project:   literal("my-project"),
						Region:    "global",
						KeyRing:   literal("openbao-ring"),
						CryptoKey: literal("openbao-unseal"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an Azure Key Vault seal should be valid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_AzureKeyVault{
					AzureKeyVault: &KubernetesOpenBaoAzureKeyVaultSeal{
						VaultName: "openbao-vault",
						KeyName:   "unseal-key",
						TenantId:  "00000000-0000-0000-0000-000000000000",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a transit seal should be valid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_Transit{
					Transit: &KubernetesOpenBaoTransitSeal{
						Address:   "https://bao.example.com:8200",
						KeyName:   "autounseal",
						MountPath: strPtr("transit/"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("metrics without a ServiceMonitor should be valid", func() {
			input.Spec.Metrics = &KubernetesOpenBaoMetrics{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Backup", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Server = raftServer()
		})

		ginkgo.It("an s3 store with access keys and no identity should be valid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("")})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an S3-compatible endpoint by SeaweedFS reference with access keys should be valid", func() {
			s3 := s3WithKeys("")
			s3.EndpointUrl = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesSeaweedFs, "object-store", "status.outputs.s3_endpoint")
			s3.ForcePathStyle = true
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a keyless s3 store with an EKS identity should be valid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: &KubernetesOpenBaoS3ObjectStore{Bucket: "b", Region: "us-west-2", Keyless: true}})
			input.Spec.Backup.WorkloadIdentity = eksIdentity()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a keyless gcs store with a GKE identity by reference should be valid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_Gcs{Gcs: &KubernetesOpenBaoGcsObjectStore{
				Bucket:  valueFrom(cloudresourcekind.CloudResourceKind_GcpGcsBucket, "backups", "status.outputs.bucket_name"),
				Keyless: true,
			}})
			input.Spec.Backup.WorkloadIdentity = &kubernetes.KubernetesWorkloadIdentity{
				Provider: &kubernetes.KubernetesWorkloadIdentity_Gke{Gke: &kubernetes.KubernetesWorkloadIdentityGke{
					ServiceAccountEmail: valueFrom(cloudresourcekind.CloudResourceKind_GcpServiceAccount, "openbao-backup", "status.outputs.email"),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a gcs store with a service-account key by reference should be valid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_Gcs{Gcs: &KubernetesOpenBaoGcsObjectStore{
				Bucket:            literal("openbao-backups"),
				ServiceAccountKey: valueFrom(cloudresourcekind.CloudResourceKind_GcpServiceAccount, "openbao-backup", "status.outputs.key_base64"),
			}})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an azure_blob store with account and key should be valid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_AzureBlob{AzureBlob: &KubernetesOpenBaoAzureBlobObjectStore{
				StorageAccount: "openbaobackups", Container: "snapshots", StorageKey: "c2VjcmV0",
			}})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an azure_blob store with a connection string should be valid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_AzureBlob{AzureBlob: &KubernetesOpenBaoAzureBlobObjectStore{
				Container: "snapshots", ConnectionString: "DefaultEndpointsProtocol=https;AccountName=x;AccountKey=y;EndpointSuffix=core.windows.net",
			}})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a keyless azure_blob store with an AKS identity should be valid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_AzureBlob{AzureBlob: &KubernetesOpenBaoAzureBlobObjectStore{
				StorageAccount: "openbaobackups", Container: "snapshots", Keyless: true,
			}})
			input.Spec.Backup.WorkloadIdentity = &kubernetes.KubernetesWorkloadIdentity{
				Provider: &kubernetes.KubernetesWorkloadIdentity_Aks{Aks: &kubernetes.KubernetesWorkloadIdentityAks{
					ClientId: literal("11111111-2222-3333-4444-555555555555"),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an r2 store wired entirely by reference should be valid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_R2{R2: r2ByReference()})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an r2 store with literal account id and jurisdiction should be valid", func() {
			r2 := r2ByReference()
			r2.AccountId = literal("0123456789abcdef0123456789abcdef")
			r2.Jurisdiction = literal("eu")
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_R2{R2: r2})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a single-replica Raft cluster with backups should be valid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{Mode: &KubernetesOpenBaoServer_Ha{Ha: &KubernetesOpenBaoHaMode{Replicas: int32Ptr(1)}}}
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("")})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("zero retention days (keep forever) should be valid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("")})
			input.Spec.Backup.RetentionDays = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a backup on standalone file storage should be invalid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{Mode: &KubernetesOpenBaoServer_Standalone{Standalone: &KubernetesOpenBaoStandaloneMode{}}}
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("")})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a backup with no server mode (chart-default standalone) should be invalid", func() {
			input.Spec.Server = nil
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("")})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a backup in dev mode should be invalid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{Mode: &KubernetesOpenBaoServer_Dev{Dev: &KubernetesOpenBaoDevMode{}}}
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("")})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a backup with the auth delegator disabled should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("")})
			input.Spec.ServiceAccount = &KubernetesOpenBaoServiceAccount{AuthDelegatorEnabled: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a backup without an object store should be invalid", func() {
			input.Spec.Backup = &KubernetesOpenBaoBackup{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an object store without a backend should be invalid", func() {
			input.Spec.Backup = &KubernetesOpenBaoBackup{ObjectStore: &KubernetesOpenBaoBackupObjectStore{Prefix: "x"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a keyless s3 store without a workload identity should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: &KubernetesOpenBaoS3ObjectStore{Bucket: "b", Region: "us-west-2", Keyless: true}})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a keyless gcs store without a workload identity should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_Gcs{Gcs: &KubernetesOpenBaoGcsObjectStore{Bucket: literal("b"), Keyless: true}})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a keyless azure_blob store without a workload identity should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_AzureBlob{AzureBlob: &KubernetesOpenBaoAzureBlobObjectStore{StorageAccount: "a", Container: "c", Keyless: true}})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an s3 store that is both keyless and keyed should be invalid", func() {
			s3 := s3WithKeys("")
			s3.Keyless = true
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3})
			input.Spec.Backup.WorkloadIdentity = eksIdentity()
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an s3 store with neither keys nor keyless should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: &KubernetesOpenBaoS3ObjectStore{Bucket: "b", Region: "us-west-2"}})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a keyless s3 store against an S3-compatible endpoint should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: &KubernetesOpenBaoS3ObjectStore{
				Bucket: "b", Keyless: true, EndpointUrl: literal("http://main-s3.object-store.svc.cluster.local:8333"),
			}})
			input.Spec.Backup.WorkloadIdentity = eksIdentity()
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an s3 endpoint without an http(s) scheme should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("main-s3.object-store.svc.cluster.local:8333")})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an s3 store without a bucket should be invalid", func() {
			s3 := s3WithKeys("")
			s3.Bucket = ""
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("access keys without a secret access key should be invalid", func() {
			s3 := s3WithKeys("")
			s3.AccessKeys.SecretAccessKey = ""
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a gcs store with both keyless and a key should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_Gcs{Gcs: &KubernetesOpenBaoGcsObjectStore{
				Bucket: literal("b"), Keyless: true, ServiceAccountKey: literal("e30="),
			}})
			input.Spec.Backup.WorkloadIdentity = eksIdentity()
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a gcs store without a bucket should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_Gcs{Gcs: &KubernetesOpenBaoGcsObjectStore{ServiceAccountKey: literal("e30=")}})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an azure_blob store with two credential postures should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_AzureBlob{AzureBlob: &KubernetesOpenBaoAzureBlobObjectStore{
				StorageAccount: "a", Container: "c", StorageKey: "k", ConnectionString: "cs",
			}})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an azure_blob storage key without a storage account should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_AzureBlob{AzureBlob: &KubernetesOpenBaoAzureBlobObjectStore{Container: "c", StorageKey: "k"}})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a keyless azure_blob store without a storage account should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_AzureBlob{AzureBlob: &KubernetesOpenBaoAzureBlobObjectStore{Container: "c", Keyless: true}})
			input.Spec.Backup.WorkloadIdentity = eksIdentity()
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an azure_blob store without a container should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_AzureBlob{AzureBlob: &KubernetesOpenBaoAzureBlobObjectStore{StorageAccount: "a", StorageKey: "k"}})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an r2 store without credentials should be invalid", func() {
			r2 := r2ByReference()
			r2.Credentials = nil
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_R2{R2: r2})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an r2 store without a bucket should be invalid", func() {
			r2 := r2ByReference()
			r2.Bucket = nil
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_R2{R2: r2})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an r2 account id that is not 32 hex characters should be invalid", func() {
			r2 := r2ByReference()
			r2.AccountId = literal("not-an-account-id")
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_R2{R2: r2})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown r2 jurisdiction should be invalid", func() {
			r2 := r2ByReference()
			r2.Jurisdiction = literal("mars")
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_R2{R2: r2})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("r2 credentials without a secret access key should be invalid", func() {
			r2 := r2ByReference()
			r2.Credentials.SecretAccessKey = nil
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_R2{R2: r2})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a schedule that is not five cron fields should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("")})
			input.Spec.Backup.Schedule = strPtr("hourly")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("negative retention days should be invalid", func() {
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_S3{S3: s3WithKeys("")})
			input.Spec.Backup.RetentionDays = int32Ptr(-1)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Restore", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Server = raftServer()
			input.Spec.Backup = backupWith(&KubernetesOpenBaoBackupObjectStore_R2{R2: r2ByReference()})
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_Transit{Transit: &KubernetesOpenBaoTransitSeal{Address: "http://key-holder.openbao.svc:8200", KeyName: "autounseal"}},
			}
		})

		ginkgo.It("restoring the latest snapshot with a transit seal should be valid", func() {
			input.Spec.Restore = &KubernetesOpenBaoRestore{
				Source:    &KubernetesOpenBaoRestore_Latest{Latest: true},
				RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init", Key: "root_token"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("restoring a named snapshot with a GCP KMS seal should be valid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{GcpKms: &KubernetesOpenBaoGcpKmsSeal{
					Project: literal("my-project"), Region: "global", KeyRing: literal("openbao-ring"), CryptoKey: literal("openbao-unseal"),
				}},
			}
			input.Spec.Restore = &KubernetesOpenBaoRestore{
				Source:    &KubernetesOpenBaoRestore_SnapshotKey{SnapshotKey: "openbao/prod/prod-20260101T020000Z.snap"},
				RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init", Key: "root_token"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a restore without a backup block should be invalid", func() {
			input.Spec.Backup = nil
			input.Spec.Restore = &KubernetesOpenBaoRestore{
				Source:    &KubernetesOpenBaoRestore_Latest{Latest: true},
				RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init", Key: "root_token"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a restore without auto-unseal (Shamir) should be invalid", func() {
			input.Spec.AutoUnseal = nil
			input.Spec.Restore = &KubernetesOpenBaoRestore{
				Source:    &KubernetesOpenBaoRestore_Latest{Latest: true},
				RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init", Key: "root_token"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an auto_unseal block with no seal arm should not satisfy a restore", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{}
			input.Spec.Restore = &KubernetesOpenBaoRestore{
				Source:    &KubernetesOpenBaoRestore_Latest{Latest: true},
				RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init", Key: "root_token"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a restore without a source should be invalid", func() {
			input.Spec.Restore = &KubernetesOpenBaoRestore{RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init", Key: "root_token"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an empty snapshot key should be invalid", func() {
			input.Spec.Restore = &KubernetesOpenBaoRestore{
				Source:    &KubernetesOpenBaoRestore_SnapshotKey{SnapshotKey: ""},
				RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init", Key: "root_token"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("latest set to false should be invalid", func() {
			input.Spec.Restore = &KubernetesOpenBaoRestore{
				Source:    &KubernetesOpenBaoRestore_Latest{Latest: false},
				RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init", Key: "root_token"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a restore without a root token should be invalid", func() {
			input.Spec.Restore = &KubernetesOpenBaoRestore{Source: &KubernetesOpenBaoRestore_Latest{Latest: true}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a root token Secret without a key should be invalid", func() {
			input.Spec.Restore = &KubernetesOpenBaoRestore{
				Source:    &KubernetesOpenBaoRestore_Latest{Latest: true},
				RootToken: &kubernetes.KubernetesSecretKey{Name: "openbao-init"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should be invalid", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown server log level should be invalid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{LogLevel: strPtr("verbose")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown server log format should be invalid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{LogFormat: strPtr("yaml")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero HA replicas should be invalid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{
				Mode: &KubernetesOpenBaoServer_Ha{Ha: &KubernetesOpenBaoHaMode{Replicas: int32Ptr(0)}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("HA replicas above 11 should be invalid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{
				Mode: &KubernetesOpenBaoServer_Ha{Ha: &KubernetesOpenBaoHaMode{Replicas: int32Ptr(12)}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a storage size without a unit suffix should be invalid", func() {
			input.Spec.Server = &KubernetesOpenBaoServer{
				DataStorage: &KubernetesOpenBaoStorage{Size: strPtr("10GB")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("TLS enabled without a cert secret should be invalid", func() {
			input.Spec.Tls = &KubernetesOpenBaoTls{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("TLS enabled with an empty literal cert secret name should be invalid", func() {
			input.Spec.Tls = &KubernetesOpenBaoTls{Enabled: true, CertSecretName: literal("")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an AWS KMS seal without a region should be invalid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_AwsKms{
					AwsKms: &KubernetesOpenBaoAwsKmsSeal{KmsKeyId: "alias/openbao-unseal"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an AWS KMS seal without a key id should be invalid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_AwsKms{
					AwsKms: &KubernetesOpenBaoAwsKmsSeal{Region: "us-west-2"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a GCP KMS seal without a project should be invalid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
					GcpKms: &KubernetesOpenBaoGcpKmsSeal{
						Region:    "global",
						KeyRing:   literal("openbao-ring"),
						CryptoKey: literal("openbao-unseal"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a GCP KMS seal without a key ring should be invalid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
					GcpKms: &KubernetesOpenBaoGcpKmsSeal{
						Project:   literal("my-project"),
						Region:    "global",
						CryptoKey: literal("openbao-unseal"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a GCP KMS seal without a crypto key should be invalid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
					GcpKms: &KubernetesOpenBaoGcpKmsSeal{
						Project: literal("my-project"),
						Region:  "global",
						KeyRing: literal("openbao-ring"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an Azure Key Vault seal without a tenant id should be invalid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_AzureKeyVault{
					AzureKeyVault: &KubernetesOpenBaoAzureKeyVaultSeal{
						VaultName: "openbao-vault",
						KeyName:   "unseal-key",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a transit seal without an address should be invalid", func() {
			input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
				Seal: &KubernetesOpenBaoAutoUnseal_Transit{
					Transit: &KubernetesOpenBaoTransitSeal{KeyName: "autounseal"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero injector replicas should be invalid", func() {
			input.Spec.Injector = &KubernetesOpenBaoInjector{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("injector replicas above 5 should be invalid", func() {
			input.Spec.Injector = &KubernetesOpenBaoInjector{Replicas: int32Ptr(6)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a lowercase injector failure policy should be invalid", func() {
			input.Spec.Injector = &KubernetesOpenBaoInjector{FailurePolicy: strPtr("ignore")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a ServiceMonitor without metrics enabled should be invalid", func() {
			input.Spec.Metrics = &KubernetesOpenBaoMetrics{ServiceMonitorEnabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

	})
})

// raftServer is the one server mode backups are legal on: a Raft cluster
// (here of one replica, the laboratory shape).
func raftServer() *KubernetesOpenBaoServer {
	return &KubernetesOpenBaoServer{Mode: &KubernetesOpenBaoServer_Ha{Ha: &KubernetesOpenBaoHaMode{Replicas: int32Ptr(1)}}}
}

// backupWith wraps one object-store arm in an otherwise-default backup block.
func backupWith(backend isKubernetesOpenBaoBackupObjectStore_Backend) *KubernetesOpenBaoBackup {
	return &KubernetesOpenBaoBackup{ObjectStore: &KubernetesOpenBaoBackupObjectStore{Prefix: "openbao/test", Backend: backend}}
}

// s3WithKeys is the declared-key S3 posture; endpoint names an
// S3-compatible store when non-empty (a literal, so the format rule runs).
func s3WithKeys(endpoint string) *KubernetesOpenBaoS3ObjectStore {
	s3 := &KubernetesOpenBaoS3ObjectStore{
		Bucket:     "openbao-snapshots",
		Region:     "us-west-2",
		AccessKeys: &KubernetesOpenBaoS3AccessKeys{AccessKeyId: "AKIAEXAMPLE", SecretAccessKey: "secret"},
	}
	if endpoint != "" {
		s3.EndpointUrl = literal(endpoint)
	}
	return s3
}

func eksIdentity() *kubernetes.KubernetesWorkloadIdentity {
	return &kubernetes.KubernetesWorkloadIdentity{
		Provider: &kubernetes.KubernetesWorkloadIdentity_Eks{Eks: &kubernetes.KubernetesWorkloadIdentityEksIrsa{
			RoleArn: literal("arn:aws:iam::123456789012:role/openbao-backup"),
		}},
	}
}

// r2ByReference is the R2 arm exactly as a chart wires it: every field a
// reference onto the catalog's bucket and token kinds.
func r2ByReference() *KubernetesOpenBaoR2ObjectStore {
	return &KubernetesOpenBaoR2ObjectStore{
		Bucket:       valueFrom(cloudresourcekind.CloudResourceKind_CloudflareR2Bucket, "backups", "status.outputs.bucket_name"),
		AccountId:    valueFrom(cloudresourcekind.CloudResourceKind_CloudflareR2Bucket, "backups", "status.outputs.account_id"),
		Jurisdiction: valueFrom(cloudresourcekind.CloudResourceKind_CloudflareR2Bucket, "backups", "status.outputs.jurisdiction"),
		Credentials: &KubernetesOpenBaoR2Credentials{
			AccessKeyId:     valueFrom(cloudresourcekind.CloudResourceKind_CloudflareAccountApiToken, "backups-writer", "status.outputs.r2_access_key_id"),
			SecretAccessKey: valueFrom(cloudresourcekind.CloudResourceKind_CloudflareAccountApiToken, "backups-writer", "status.outputs.r2_secret_access_key"),
		},
	}
}
