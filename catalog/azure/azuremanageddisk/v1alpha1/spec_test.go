package azuremanageddiskv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

const testDesId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/sec-rg/providers/Microsoft.Compute/diskEncryptionSets/cmk"

// validSpec returns a minimal valid empty Standard SSD disk the failure
// cases mutate one field at a time.
func validSpec() *AzureManagedDiskSpec {
	return &AzureManagedDiskSpec{
		Region:             "eastus",
		ResourceGroup:      stringRef("test-rg"),
		Name:               "data-disk",
		StorageAccountType: AzureManagedDiskStorageAccountType_STANDARD_SSD_LRS,
		CreateOption:       AzureManagedDiskCreateOption_EMPTY,
		DiskSizeGb:         proto.Int32(64),
	}
}

func validInput(spec *AzureManagedDiskSpec) *AzureManagedDisk {
	return &AzureManagedDisk{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureManagedDisk",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-managed-disk",
		},
		Spec: spec,
	}
}

func TestAzureManagedDiskSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureManagedDiskSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureManagedDiskSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal empty disk", func() {
			err := protovalidate.Validate(validInput(validSpec()))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a zonal premium disk with a tier and bursting", func() {
			spec := validSpec()
			spec.StorageAccountType = AzureManagedDiskStorageAccountType_PREMIUM_LRS
			spec.Zone = "1"
			spec.Tier = "P30"
			spec.DiskSizeGb = proto.Int32(1024)
			spec.OnDemandBurstingEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a PremiumV2 disk with independent performance dials", func() {
			spec := validSpec()
			spec.StorageAccountType = AzureManagedDiskStorageAccountType_PREMIUM_V2_LRS
			spec.Zone = "1"
			spec.DiskIopsReadWrite = proto.Int32(8000)
			spec.DiskMbpsReadWrite = proto.Int32(300)
			spec.LogicalSectorSize = proto.Int32(4096)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a shared Ultra disk with read-only dials", func() {
			spec := validSpec()
			spec.StorageAccountType = AzureManagedDiskStorageAccountType_ULTRA_SSD_LRS
			spec.Zone = "1"
			spec.MaxShares = proto.Int32(5)
			spec.DiskIopsReadWrite = proto.Int32(16000)
			spec.DiskIopsReadOnly = proto.Int32(4000)
			spec.DiskMbpsReadOnly = proto.Int32(100)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a COPY disk cloning a snapshot", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_COPY
			spec.DiskSizeGb = nil
			spec.SourceResourceId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/snapshots/nightly"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a FROM_IMAGE OS disk with trusted launch", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_FROM_IMAGE
			spec.DiskSizeGb = nil
			spec.ImageReferenceId = "/subscriptions/s/providers/Microsoft.Compute/locations/eastus/publishers/canonical/artifacttypes/vmimage/offers/ubuntu-24_04-lts/skus/server/versions/latest"
			spec.OsType = AzureManagedDiskOsType_LINUX
			spec.HyperVGeneration = AzureManagedDiskHyperVGeneration_V2
			spec.TrustedLaunchEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an IMPORT disk wrapping a VHD blob", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_IMPORT
			spec.SourceUri = "https://account.blob.core.windows.net/vhds/data.vhd"
			spec.StorageAccountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/account"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an UPLOAD target with an exact byte size", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_UPLOAD
			spec.DiskSizeGb = nil
			spec.UploadSizeBytes = 68719477248
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept CMK encryption via a disk encryption set", func() {
			spec := validSpec()
			spec.DiskEncryptionSetId = stringRef(testDesId)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a confidential-VM customer-key OS disk", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_FROM_IMAGE
			spec.DiskSizeGb = nil
			spec.GalleryImageReferenceId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/galleries/g/images/i/versions/1.0.0"
			spec.SecurityType = AzureManagedDiskSecurityType_CONFIDENTIAL_VM_DISK_ENCRYPTED_WITH_CUSTOMER_KEY
			spec.SecureVmDiskEncryptionSetId = stringRef(testDesId)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a private-export posture", func() {
			spec := validSpec()
			spec.NetworkAccessPolicy = AzureManagedDiskNetworkAccessPolicy_ALLOW_PRIVATE
			spec.DiskAccessId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/diskAccesses/private"
			spec.PublicNetworkAccessEnabled = proto.Bool(false)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing storage_account_type", func() {
			spec := validSpec()
			spec.StorageAccountType = AzureManagedDiskStorageAccountType_azure_managed_disk_storage_account_type_unspecified
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing create_option", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_azure_managed_disk_create_option_unspecified
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an EMPTY disk without a size", func() {
			spec := validSpec()
			spec.DiskSizeGb = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a COPY disk without a source", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_COPY
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an IMPORT disk missing its storage account", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_IMPORT
			spec.SourceUri = "https://account.blob.core.windows.net/vhds/data.vhd"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject FROM_IMAGE with both image sources", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_FROM_IMAGE
			spec.DiskSizeGb = nil
			spec.ImageReferenceId = "/subscriptions/s/.../versions/latest"
			spec.GalleryImageReferenceId = "/subscriptions/s/.../versions/1.0.0"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an image source on a non-FROM_IMAGE disk", func() {
			spec := validSpec()
			spec.ImageReferenceId = "/subscriptions/s/.../versions/latest"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an UPLOAD target without a byte size", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_UPLOAD
			spec.DiskSizeGb = nil
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject performance dials on a fixed-performance SKU", func() {
			spec := validSpec()
			spec.DiskIopsReadWrite = proto.Int32(8000)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject read-only dials without a shared disk", func() {
			spec := validSpec()
			spec.StorageAccountType = AzureManagedDiskStorageAccountType_ULTRA_SSD_LRS
			spec.DiskIopsReadOnly = proto.Int32(4000)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a tier on a non-premium SKU", func() {
			spec := validSpec()
			spec.Tier = "P30"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject both encryption sets together", func() {
			spec := validSpec()
			spec.DiskEncryptionSetId = stringRef(testDesId)
			spec.SecureVmDiskEncryptionSetId = stringRef(testDesId)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a secure-VM encryption set without the matching security type", func() {
			spec := validSpec()
			spec.SecureVmDiskEncryptionSetId = stringRef(testDesId)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject confidential customer-key security on an EMPTY disk", func() {
			spec := validSpec()
			spec.SecurityType = AzureManagedDiskSecurityType_CONFIDENTIAL_VM_DISK_ENCRYPTED_WITH_CUSTOMER_KEY
			spec.SecureVmDiskEncryptionSetId = stringRef(testDesId)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject trusted launch combined with a security type", func() {
			spec := validSpec()
			spec.CreateOption = AzureManagedDiskCreateOption_FROM_IMAGE
			spec.DiskSizeGb = nil
			spec.ImageReferenceId = "/subscriptions/s/.../versions/latest"
			spec.TrustedLaunchEnabled = true
			spec.SecurityType = AzureManagedDiskSecurityType_CONFIDENTIAL_VM_DISK_ENCRYPTED_WITH_PLATFORM_KEY
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject trusted launch on an EMPTY disk", func() {
			spec := validSpec()
			spec.TrustedLaunchEnabled = true
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a disk_access_id without ALLOW_PRIVATE", func() {
			spec := validSpec()
			spec.DiskAccessId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/diskAccesses/private"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject ALLOW_PRIVATE without a disk_access_id", func() {
			spec := validSpec()
			spec.NetworkAccessPolicy = AzureManagedDiskNetworkAccessPolicy_ALLOW_PRIVATE
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a zone pin on a ZRS SKU", func() {
			spec := validSpec()
			spec.StorageAccountType = AzureManagedDiskStorageAccountType_STANDARD_SSD_ZRS
			spec.Zone = "1"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid zone value", func() {
			spec := validSpec()
			spec.Zone = "4"
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject max_shares outside 2-10", func() {
			spec := validSpec()
			spec.MaxShares = proto.Int32(1)
			err := protovalidate.Validate(validInput(spec))
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
