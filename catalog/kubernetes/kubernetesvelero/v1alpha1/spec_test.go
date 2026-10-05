package kubernetesvelerov1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesVelero(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesVelero Suite")
}

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

// s3Backend returns a minimal valid S3 arm — the base every test mutates
// from (bucket and region are the only required fields).
func s3Backend() *KubernetesVeleroBackupStorage {
	return &KubernetesVeleroBackupStorage{
		Backend: &KubernetesVeleroBackupStorage_S3{
			S3: &KubernetesVeleroS3Backend{
				Bucket: "velero-backups",
				Region: "us-east-1",
			},
		},
	}
}

func gcsBackend(gcs *KubernetesVeleroGcsBackend) *KubernetesVeleroBackupStorage {
	return &KubernetesVeleroBackupStorage{
		Backend: &KubernetesVeleroBackupStorage_Gcs{Gcs: gcs},
	}
}

func azureBackend(azure *KubernetesVeleroAzureBlobBackend) *KubernetesVeleroBackupStorage {
	return &KubernetesVeleroBackupStorage{
		Backend: &KubernetesVeleroBackupStorage_AzureBlob{AzureBlob: azure},
	}
}

// validAzureBlob returns the required Azure identity fields; credential
// posture is added per test.
func validAzureBlob() *KubernetesVeleroAzureBlobBackend {
	return &KubernetesVeleroAzureBlobBackend{
		StorageAccount: "velerobackups",
		Container:      "backups",
		ResourceGroup:  "rg-velero",
		SubscriptionId: "00000000-1111-2222-3333-444444444444",
	}
}

var _ = ginkgo.Describe("KubernetesVelero Validation Tests", func() {
	var input *KubernetesVelero

	ginkgo.BeforeEach(func() {
		input = &KubernetesVelero{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesVelero",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-velero",
			},
			Spec: &KubernetesVeleroSpec{
				Namespace:     literal("velero"),
				BackupStorage: s3Backend(),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (s3, ambient credentials) should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "velero", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("s3 with IRSA should be valid", func() {
			input.Spec.BackupStorage.GetS3().IrsaRoleArn = literal("arn:aws:iam::123456789012:role/velero-backups")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("s3 IRSA role as an AwsIamRole reference should be valid", func() {
			// The one-run composition shape: the backup role deploys in the
			// same run and its ARN flows in by reference.
			input.Spec.BackupStorage.GetS3().IrsaRoleArn = valueFrom(
				catalogkind.CatalogKind_AwsIamRole, "velero-backups", "status.outputs.role_arn")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("s3-compatible endpoint with access keys (MinIO posture) should be valid", func() {
			s3 := input.Spec.BackupStorage.GetS3()
			s3.S3Url = "http://minio.minio.svc:9000"
			s3.ForcePathStyle = true
			s3.AccessKeys = &KubernetesVeleroS3AccessKeys{
				AccessKeyId:     "minio",
				SecretAccessKey: "minio123",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("s3 with KMS key and CA bundle should be valid", func() {
			s3 := input.Spec.BackupStorage.GetS3()
			s3.KmsKeyId = "alias/velero"
			s3.CaCert = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gcs with workload identity should be valid", func() {
			input.Spec.BackupStorage = gcsBackend(&KubernetesVeleroGcsBackend{
				Bucket:                              "velero-backups",
				WorkloadIdentityServiceAccountEmail: literal("velero@my-project.iam.gserviceaccount.com"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gcs workload identity as a GcpServiceAccount reference should be valid", func() {
			// The one-run composition shape: the backup identity deploys in
			// the same run and its email flows in by reference.
			input.Spec.BackupStorage = gcsBackend(&KubernetesVeleroGcsBackend{
				Bucket: "velero-backups",
				WorkloadIdentityServiceAccountEmail: valueFrom(
					catalogkind.CatalogKind_GcpServiceAccount, "velero-backups", "status.outputs.email"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gcs with a service-account key should be valid (key is optional posture)", func() {
			input.Spec.BackupStorage = gcsBackend(&KubernetesVeleroGcsBackend{
				Bucket:                "velero-backups",
				ServiceAccountKeyJson: `{"type":"service_account","project_id":"my-project"}`,
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gcs with neither credential posture (ambient) should be valid", func() {
			input.Spec.BackupStorage = gcsBackend(&KubernetesVeleroGcsBackend{Bucket: "velero-backups"})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("azure with a workload-identity client id should be valid", func() {
			azure := validAzureBlob()
			azure.WorkloadIdentityClientId = literal("55555555-6666-7777-8888-999999999999")
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("azure workload identity as an AzureUserAssignedIdentity reference should be valid", func() {
			// The one-run composition shape: the backup identity deploys in
			// the same run and its client id flows in by reference.
			azure := validAzureBlob()
			azure.WorkloadIdentityClientId = valueFrom(
				catalogkind.CatalogKind_AzureUserAssignedIdentity, "velero-backups", "status.outputs.client_id")
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("azure with a service principal should be valid", func() {
			azure := validAzureBlob()
			azure.ServicePrincipal = &KubernetesVeleroAzureServicePrincipal{
				TenantId:     "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				ClientId:     "11111111-2222-3333-4444-555555555555",
				ClientSecret: "sp-secret-value",
			}
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("schedules with distinct names, cron shorthand and ttl should be valid", func() {
			input.Spec.Schedules = []*KubernetesVeleroSchedule{
				{Name: "daily-cluster", Schedule: "0 2 * * *", Ttl: "720h"},
				{Name: "hourly-critical", Schedule: "@hourly", Ttl: "72h30m"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("csi snapshots with data movement should be valid when both flags are set", func() {
			input.Spec.VolumeSnapshots = &KubernetesVeleroVolumeSnapshots{
				EnableCsi:               true,
				DefaultSnapshotMoveData: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("fs-backup defaults with the node-agent deployed should be valid", func() {
			input.Spec.FsBackup = &KubernetesVeleroFsBackup{
				DeployNodeAgent:          true,
				DefaultVolumesToFsBackup: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("server tuning with valid durations and enums should be valid", func() {
			input.Spec.Server = &KubernetesVeleroServer{
				DefaultBackupTtl:            "720h",
				DefaultItemOperationTimeout: "4h",
				GarbageCollectionFrequency:  "1h30m",
				RestoreOnlyMode:             true,
				LogLevel:                    stringPtr("debug"),
				LogFormat:                   stringPtr("json"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("prometheus service monitor with metrics enabled should be valid", func() {
			input.Spec.Prometheus = &KubernetesVeleroPrometheus{
				Enabled:        boolPtr(true),
				ServiceMonitor: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("service monitor with enabled unset should be valid (defaults to enabled)", func() {
			input.Spec.Prometheus = &KubernetesVeleroPrometheus{ServiceMonitor: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("crds lifecycle flags should be valid", func() {
			input.Spec.Crds = &KubernetesVeleroCrds{
				UpgradeOnInstall:   boolPtr(false),
				CleanupOnUninstall: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("missing backup_storage should fail (required)", func() {
			input.Spec.BackupStorage = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backup_storage without a backend arm should fail (oneof required)", func() {
			input.Spec.BackupStorage = &KubernetesVeleroBackupStorage{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("s3 without a bucket should fail", func() {
			input.Spec.BackupStorage.GetS3().Bucket = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("s3 without a region should fail", func() {
			input.Spec.BackupStorage.GetS3().Region = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("s3 with both IRSA and access keys should fail (irsa_xor_keys)", func() {
			s3 := input.Spec.BackupStorage.GetS3()
			s3.IrsaRoleArn = literal("arn:aws:iam::123456789012:role/velero-backups")
			s3.AccessKeys = &KubernetesVeleroS3AccessKeys{
				AccessKeyId:     "AKIAEXAMPLE",
				SecretAccessKey: "secret",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("s3-compatible endpoint with IRSA should fail (compatible_needs_keys)", func() {
			s3 := input.Spec.BackupStorage.GetS3()
			s3.S3Url = "http://minio.minio.svc:9000"
			s3.IrsaRoleArn = literal("arn:aws:iam::123456789012:role/velero-backups")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("non-http s3_url should fail (url_format)", func() {
			input.Spec.BackupStorage.GetS3().S3Url = "minio.minio.svc:9000"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("IRSA role present but empty should fail (StringValueOrRef contract)", func() {
			// The shared StringValueOrRef rule: a present message must carry a
			// non-empty literal or a reference.
			input.Spec.BackupStorage.GetS3().IrsaRoleArn = &foreignkeyv1.StringValueOrRef{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("access keys without access_key_id should fail", func() {
			input.Spec.BackupStorage.GetS3().AccessKeys = &KubernetesVeleroS3AccessKeys{
				SecretAccessKey: "secret",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("access keys without secret_access_key should fail", func() {
			input.Spec.BackupStorage.GetS3().AccessKeys = &KubernetesVeleroS3AccessKeys{
				AccessKeyId: "AKIAEXAMPLE",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("gcs without a bucket should fail", func() {
			input.Spec.BackupStorage = gcsBackend(&KubernetesVeleroGcsBackend{
				WorkloadIdentityServiceAccountEmail: literal("velero@my-project.iam.gserviceaccount.com"),
			})
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("gcs with both workload identity and a key should fail (wi_xor_key)", func() {
			input.Spec.BackupStorage = gcsBackend(&KubernetesVeleroGcsBackend{
				Bucket:                              "velero-backups",
				WorkloadIdentityServiceAccountEmail: literal("velero@my-project.iam.gserviceaccount.com"),
				ServiceAccountKeyJson:               `{"type":"service_account"}`,
			})
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("gcs workload identity present but empty should fail (StringValueOrRef contract)", func() {
			// The shared StringValueOrRef rule: a present message must carry a
			// non-empty literal or a reference.
			input.Spec.BackupStorage = gcsBackend(&KubernetesVeleroGcsBackend{
				Bucket:                              "velero-backups",
				WorkloadIdentityServiceAccountEmail: &foreignkeyv1.StringValueOrRef{},
			})
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure missing storage_account should fail", func() {
			azure := validAzureBlob()
			azure.StorageAccount = ""
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure missing container should fail", func() {
			azure := validAzureBlob()
			azure.Container = ""
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure missing resource_group should fail", func() {
			azure := validAzureBlob()
			azure.ResourceGroup = ""
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure missing subscription_id should fail", func() {
			azure := validAzureBlob()
			azure.SubscriptionId = ""
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure with both a workload-identity client id and a service principal should fail (wi_xor_sp)", func() {
			azure := validAzureBlob()
			azure.WorkloadIdentityClientId = literal("55555555-6666-7777-8888-999999999999")
			azure.ServicePrincipal = &KubernetesVeleroAzureServicePrincipal{
				TenantId:     "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				ClientId:     "11111111-2222-3333-4444-555555555555",
				ClientSecret: "sp-secret-value",
			}
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("service principal without tenant_id should fail", func() {
			azure := validAzureBlob()
			azure.ServicePrincipal = &KubernetesVeleroAzureServicePrincipal{
				ClientId:     "11111111-2222-3333-4444-555555555555",
				ClientSecret: "sp-secret-value",
			}
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("service principal without client_id should fail", func() {
			azure := validAzureBlob()
			azure.ServicePrincipal = &KubernetesVeleroAzureServicePrincipal{
				TenantId:     "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				ClientSecret: "sp-secret-value",
			}
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("service principal without client_secret should fail", func() {
			azure := validAzureBlob()
			azure.ServicePrincipal = &KubernetesVeleroAzureServicePrincipal{
				TenantId: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				ClientId: "11111111-2222-3333-4444-555555555555",
			}
			input.Spec.BackupStorage = azureBackend(azure)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate schedule names should fail (unique_names)", func() {
			input.Spec.Schedules = []*KubernetesVeleroSchedule{
				{Name: "daily", Schedule: "0 2 * * *"},
				{Name: "daily", Schedule: "0 4 * * *"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("uppercase schedule name should fail (kubernetes name pattern)", func() {
			input.Spec.Schedules = []*KubernetesVeleroSchedule{
				{Name: "Daily", Schedule: "0 2 * * *"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed cron expression should fail (cron pattern)", func() {
			input.Spec.Schedules = []*KubernetesVeleroSchedule{
				{Name: "daily", Schedule: "not-a-cron"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed schedule ttl should fail (ttl_format)", func() {
			input.Spec.Schedules = []*KubernetesVeleroSchedule{
				{Name: "daily", Schedule: "0 2 * * *", Ttl: "30 days"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("snapshot data movement without CSI should fail (move_needs_csi)", func() {
			input.Spec.VolumeSnapshots = &KubernetesVeleroVolumeSnapshots{
				DefaultSnapshotMoveData: true,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("fs-backup defaults without the node-agent should fail (defaults_need_agent)", func() {
			input.Spec.FsBackup = &KubernetesVeleroFsBackup{
				DefaultVolumesToFsBackup: true,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown server log level should fail (closed enum)", func() {
			input.Spec.Server = &KubernetesVeleroServer{LogLevel: stringPtr("verbose")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown server log format should fail (closed enum)", func() {
			input.Spec.Server = &KubernetesVeleroServer{LogFormat: stringPtr("yaml")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed default_backup_ttl should fail (duration format)", func() {
			input.Spec.Server = &KubernetesVeleroServer{DefaultBackupTtl: "30d"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed default_item_operation_timeout should fail (duration format)", func() {
			input.Spec.Server = &KubernetesVeleroServer{DefaultItemOperationTimeout: "4 hours"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed garbage_collection_frequency should fail (duration format)", func() {
			input.Spec.Server = &KubernetesVeleroServer{GarbageCollectionFrequency: "hourly"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("service monitor with metrics explicitly disabled should fail (monitor_requires_enabled)", func() {
			input.Spec.Prometheus = &KubernetesVeleroPrometheus{
				Enabled:        boolPtr(false),
				ServiceMonitor: true,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
