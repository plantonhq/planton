package azuredataprotectionbackupinstancev1alpha1

import (
	"fmt"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureDataProtectionBackupInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataProtectionBackupInstanceSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testVaultId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/backup-rg/providers/Microsoft.DataProtection/backupVaults/prod-vault"
	testPolicyId = testVaultId + "/backupPolicies/daily-disk"
	testDiskId   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Compute/disks/app-data"
	testSaId     = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Storage/storageAccounts/appdata"
	testAksId    = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ContainerService/managedClusters/prod-aks"
	testServerId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DBforMySQL/flexibleServers/prod-mysql"
)

// validResource returns a minimal valid disk-variant instance that
// individual cases mutate into the shape under test.
func validResource() *AzureDataProtectionBackupInstance {
	return &AzureDataProtectionBackupInstance{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataProtectionBackupInstance",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-backup-instance",
		},
		Spec: &AzureDataProtectionBackupInstanceSpec{
			VaultId:        literal(testVaultId),
			Name:           "app-data-disk",
			Region:         "eastus",
			BackupPolicyId: literal(testPolicyId),
			Disk: &AzureDataProtectionBackupInstanceDisk{
				DiskId:                    literal(testDiskId),
				SnapshotResourceGroupName: literal("backup-rg"),
			},
		},
	}
}

var _ = ginkgo.Describe("AzureDataProtectionBackupInstanceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_data_protection_backup_instance", func() {

			ginkgo.It("should not return a validation error for a minimal disk instance", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a disk instance with a cross-subscription snapshot home", func() {
				input := validResource()
				snapshotSubscription := "11111111-2222-3333-4444-555555555555"
				input.Spec.Disk.SnapshotSubscriptionId = &snapshotSubscription
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a blob instance without containers (operational-only)", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupInstanceBlobStorage{
					StorageAccountId: literal(testSaId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a blob instance with container names", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupInstanceBlobStorage{
					StorageAccountId:             literal(testSaId),
					StorageAccountContainerNames: []string{"orders", "invoices"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a kubernetes instance without datasource parameters", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.KubernetesCluster = &AzureDataProtectionBackupInstanceKubernetesCluster{
					KubernetesClusterId:       literal(testAksId),
					SnapshotResourceGroupName: literal("backup-rg"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a kubernetes instance with full datasource parameters", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.KubernetesCluster = &AzureDataProtectionBackupInstanceKubernetesCluster{
					KubernetesClusterId:       literal(testAksId),
					SnapshotResourceGroupName: literal("backup-rg"),
					BackupDatasourceParameters: &AzureDataProtectionBackupInstanceKubernetesClusterDatasourceParameters{
						IncludedNamespaces:            []string{"commerce", "payments"},
						ExcludedNamespaces:            []string{"kube-system"},
						IncludedResourceTypes:         []string{"deployments.apps"},
						ExcludedResourceTypes:         []string{"events"},
						LabelSelectors:                []string{"backup=true"},
						ClusterScopedResourcesEnabled: proto.Bool(true),
						VolumeSnapshotEnabled:         proto.Bool(true),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept datasource parameters that state one switch off and leave the other unset", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.KubernetesCluster = &AzureDataProtectionBackupInstanceKubernetesCluster{
					KubernetesClusterId:       literal(testAksId),
					SnapshotResourceGroupName: literal("backup-rg"),
					BackupDatasourceParameters: &AzureDataProtectionBackupInstanceKubernetesClusterDatasourceParameters{
						VolumeSnapshotEnabled: proto.Bool(false),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				gomega.Expect(input.Spec.KubernetesCluster.BackupDatasourceParameters.ClusterScopedResourcesEnabled).To(gomega.BeNil(), "the unset switch stays unset at the schema layer; the platform fills its declared default before the module reads it")
			})

			ginkgo.It("should accept a mysql flexible-server instance", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.MysqlFlexibleServer = &AzureDataProtectionBackupInstanceMysqlFlexibleServer{
					ServerId: literal(testServerId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a postgresql flexible-server instance", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.PostgresqlFlexibleServer = &AzureDataProtectionBackupInstancePostgresqlFlexibleServer{
					ServerId: literal(testServerId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a data-lake instance with one container", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupInstanceDataLakeStorage{
					StorageAccountId:      literal(testSaId),
					StorageContainerNames: []string{"raw-events"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a data-lake container name at the 63-character bound", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupInstanceDataLakeStorage{
					StorageAccountId:      literal(testSaId),
					StorageContainerNames: []string{"c" + strings.Repeat("a", 62)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_data_protection_backup_instance", func() {

			ginkgo.It("should reject an instance with no variant set", func() {
				input := validResource()
				input.Spec.Disk = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an instance with two variants set", func() {
				input := validResource()
				input.Spec.BlobStorage = &AzureDataProtectionBackupInstanceBlobStorage{
					StorageAccountId: literal(testSaId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing vault reference", func() {
				input := validResource()
				input.Spec.VaultId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing policy reference", func() {
				input := validResource()
				input.Spec.BackupPolicyId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a disk instance without a disk reference", func() {
				input := validResource()
				input.Spec.Disk.DiskId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a disk instance without a snapshot resource group", func() {
				input := validResource()
				input.Spec.Disk.SnapshotResourceGroupName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a non-UUID snapshot subscription", func() {
				input := validResource()
				notAUuid := "prod-subscription"
				input.Spec.Disk.SnapshotSubscriptionId = &notAUuid
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a blob instance without a storage account reference", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupInstanceBlobStorage{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a kubernetes instance without a cluster reference", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.KubernetesCluster = &AzureDataProtectionBackupInstanceKubernetesCluster{
					SnapshotResourceGroupName: literal("backup-rg"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a kubernetes instance without a snapshot resource group", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.KubernetesCluster = &AzureDataProtectionBackupInstanceKubernetesCluster{
					KubernetesClusterId: literal(testAksId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a mysql instance without a server reference", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.MysqlFlexibleServer = &AzureDataProtectionBackupInstanceMysqlFlexibleServer{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a postgresql instance without a server reference", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.PostgresqlFlexibleServer = &AzureDataProtectionBackupInstancePostgresqlFlexibleServer{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a data-lake instance without containers", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupInstanceDataLakeStorage{
					StorageAccountId: literal(testSaId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a data-lake container name under 3 characters", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupInstanceDataLakeStorage{
					StorageAccountId:      literal(testSaId),
					StorageContainerNames: []string{"ab"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a data-lake container name with uppercase letters", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupInstanceDataLakeStorage{
					StorageAccountId:      literal(testSaId),
					StorageContainerNames: []string{"RawEvents"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a data-lake container name starting with a hyphen", func() {
				input := validResource()
				input.Spec.Disk = nil
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupInstanceDataLakeStorage{
					StorageAccountId:      literal(testSaId),
					StorageContainerNames: []string{"-raw-events"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a data-lake instance over the 1000-container bound", func() {
				input := validResource()
				input.Spec.Disk = nil
				containers := make([]string, 1001)
				for i := range containers {
					containers[i] = fmt.Sprintf("container-%04d", i)
				}
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupInstanceDataLakeStorage{
					StorageAccountId:      literal(testSaId),
					StorageContainerNames: containers,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
