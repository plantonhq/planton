package azurecomputegalleryimagev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureComputeGalleryImageSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureComputeGalleryImageSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testSnapshotId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Compute/snapshots/golden-base"

// validResource returns a valid image definition that individual cases
// mutate into the shape under test.
func validResource() *AzureComputeGalleryImage {
	return &AzureComputeGalleryImage{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureComputeGalleryImage",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-gallery-image",
		},
		Spec: &AzureComputeGalleryImageSpec{
			ResourceGroup: literal("app-rg"),
			GalleryName:   literal("platform.images"),
			Name:          "ubuntu-base",
			Region:        "eastus",
			Identifier: &AzureComputeGalleryImageIdentifier{
				Publisher: "acme",
				Offer:     "ubuntu",
				Sku:       "22-04-lts",
			},
			OsType: "Linux",
		},
	}
}

// validVersion returns a valid snapshot-sourced version that cases
// mutate into the shape under test.
func validVersion() *AzureComputeGalleryImageVersion {
	return &AzureComputeGalleryImageVersion{
		Name:             "1.0.0",
		OsDiskSnapshotId: literal(testSnapshotId),
		TargetRegions: []*AzureComputeGalleryImageVersionTargetRegion{
			{Name: "eastus", RegionalReplicaCount: 1},
		},
	}
}

var _ = ginkgo.Describe("AzureComputeGalleryImageSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_compute_gallery_image", func() {

			ginkgo.It("should not return a validation error for the minimal definition", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept both OS types and both explicit architectures", func() {
				input := validResource()
				input.Spec.OsType = "Windows"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Architecture = "x64"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Architecture = "Arm64"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept each security flag on its own", func() {
				for _, mutate := range []func(*AzureComputeGalleryImageSpec){
					func(s *AzureComputeGalleryImageSpec) { s.TrustedLaunchSupported = true },
					func(s *AzureComputeGalleryImageSpec) { s.TrustedLaunchEnabled = true },
					func(s *AzureComputeGalleryImageSpec) { s.ConfidentialVmSupported = true },
					func(s *AzureComputeGalleryImageSpec) { s.ConfidentialVmEnabled = true },
				} {
					input := validResource()
					mutate(input.Spec)
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept an ordered recommended-sizing range", func() {
				input := validResource()
				input.Spec.MinRecommendedVcpuCount = proto.Int32(2)
				input.Spec.MaxRecommendedVcpuCount = proto.Int32(2)
				input.Spec.MinRecommendedMemoryInGb = proto.Int32(4)
				input.Spec.MaxRecommendedMemoryInGb = proto.Int32(64)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a snapshot-sourced version", func() {
				input := validResource()
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{validVersion()}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a blob-sourced version with its storage account", func() {
				input := validResource()
				version := validVersion()
				version.OsDiskSnapshotId = nil
				version.BlobUri = "https://acmeimages.blob.core.windows.net/vhds/base.vhd"
				version.StorageAccountId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/acmeimages")
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a managed-image-sourced version", func() {
				input := validResource()
				version := validVersion()
				version.OsDiskSnapshotId = nil
				version.ManagedImageId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/images/legacy")
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept the latest and recent literal version names", func() {
				input := validResource()
				version := validVersion()
				version.Name = "latest"
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				version.Name = "recent"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every target-region storage account type", func() {
				input := validResource()
				version := validVersion()
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				for _, t := range []string{"", "Premium_LRS", "Standard_LRS", "Standard_ZRS"} {
					version.TargetRegions[0].StorageAccountType = t
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "expected %q to be accepted", t)
				}
			})

			ginkgo.It("should accept a Full-replication version with a disk encryption set", func() {
				input := validResource()
				version := validVersion()
				version.ReplicationMode = "Full"
				version.TargetRegions[0].DiskEncryptionSetId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/diskEncryptionSets/des")
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an RFC3339 end-of-life date on the image and the version", func() {
				input := validResource()
				input.Spec.EndOfLifeDate = "2030-01-01T00:00:00Z"
				version := validVersion()
				version.EndOfLifeDate = "2029-06-30T12:00:00+05:30"
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_compute_gallery_image", func() {

			ginkgo.It("should reject a missing identifier or os_type", func() {
				input := validResource()
				input.Spec.Identifier = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.OsType = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject unknown os_type, architecture, and hyper_v_generation tokens", func() {
				input := validResource()
				input.Spec.OsType = "linux"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Architecture = "arm64"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.HyperVGeneration = "V3"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject any two security flags together", func() {
				input := validResource()
				input.Spec.TrustedLaunchSupported = true
				input.Spec.TrustedLaunchEnabled = true
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.TrustedLaunchEnabled = true
				input.Spec.ConfidentialVmEnabled = true
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.ConfidentialVmSupported = true
				input.Spec.ConfidentialVmEnabled = true
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an inverted recommended-sizing range", func() {
				input := validResource()
				input.Spec.MinRecommendedVcpuCount = proto.Int32(8)
				input.Spec.MaxRecommendedVcpuCount = proto.Int32(2)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.MinRecommendedMemoryInGb = proto.Int32(64)
				input.Spec.MaxRecommendedMemoryInGb = proto.Int32(4)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject recommended counts outside the provider bounds", func() {
				input := validResource()
				input.Spec.MaxRecommendedVcpuCount = proto.Int32(81)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.MaxRecommendedMemoryInGb = proto.Int32(641)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject identifier parts that end with a dot or exceed their caps", func() {
				input := validResource()
				input.Spec.Identifier.Publisher = "acme."
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Identifier.Publisher = strings.Repeat("a", 129)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Identifier.Offer = strings.Repeat("a", 65)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Identifier.Sku = strings.Repeat("a", 65)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject disallowed disk types outside the provider vocabulary", func() {
				input := validResource()
				input.Spec.DiskTypesNotAllowed = []string{"StandardSSD_LRS"}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a non-RFC3339 end-of-life date", func() {
				input := validResource()
				input.Spec.EndOfLifeDate = "2030-01-01"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a version with no source", func() {
				input := validResource()
				version := validVersion()
				version.OsDiskSnapshotId = nil
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a version with two sources", func() {
				input := validResource()
				version := validVersion()
				version.ManagedImageId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/images/legacy")
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a blob source without its storage account (and vice versa)", func() {
				input := validResource()
				version := validVersion()
				version.OsDiskSnapshotId = nil
				version.BlobUri = "https://acmeimages.blob.core.windows.net/vhds/base.vhd"
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				version = validVersion()
				version.StorageAccountId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/acmeimages")
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject version names outside the numeric triple / latest / recent forms", func() {
				input := validResource()
				version := validVersion()
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				for _, name := range []string{"1.2", "v1.0.0", "1.0.0.0", "newest", ""} {
					version.Name = name
					gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "expected %q to be rejected", name)
				}
			})

			ginkgo.It("should reject a version without target regions", func() {
				input := validResource()
				version := validVersion()
				version.TargetRegions = nil
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a zero regional replica count", func() {
				input := validResource()
				version := validVersion()
				version.TargetRegions[0].RegionalReplicaCount = 0
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a Shallow version carrying a disk encryption set", func() {
				input := validResource()
				version := validVersion()
				version.ReplicationMode = "Shallow"
				version.TargetRegions[0].DiskEncryptionSetId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/diskEncryptionSets/des")
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{version}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate version names", func() {
				input := validResource()
				first := validVersion()
				second := validVersion()
				input.Spec.Versions = []*AzureComputeGalleryImageVersion{first, second}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
