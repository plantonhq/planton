package civovolumev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestCivoVolumeSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoVolumeSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoVolumeSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("civo_volume with minimal valid fields", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with minimum size", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "tiny-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "tiny-volume",
						Region:     civo.CivoRegion_fra1,
						SizeGib:    1, // minimum allowed
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maximum size", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "huge-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "huge-volume",
						Region:     civo.CivoRegion_nyc1,
						SizeGib:    16000, // maximum allowed
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ext4 filesystem", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ext4-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName:     "ext4-volume",
						Region:         civo.CivoRegion_lon1,
						SizeGib:        50,
						FilesystemType: CivoVolumeFilesystemType_ext4,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with xfs filesystem", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "xfs-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName:     "xfs-volume",
						Region:         civo.CivoRegion_fra1,
						SizeGib:        100,
						FilesystemType: CivoVolumeFilesystemType_xfs,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with snapshot_id", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "restored-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "restored-volume",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    50,
						SnapshotId: "snapshot-12345",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with tags", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "tagged-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "tagged-volume",
						Region:     civo.CivoRegion_nyc1,
						SizeGib:    100,
						Tags:       []string{"env:prod", "team:backend", "criticality:high"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all fields", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-db-data",
					},
					Spec: &CivoVolumeSpec{
						VolumeName:     "prod-db-data",
						Region:         civo.CivoRegion_fra1,
						SizeGib:        1000,
						FilesystemType: CivoVolumeFilesystemType_xfs,
						SnapshotId:     "snapshot-67890",
						Tags:           []string{"env:prod", "backup:daily", "retention:90-days"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with volume_name at max length", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "max-length-volume-name",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "this-is-a-very-long-volume-name-with-exactly-sixty-four-chars1", // exactly 64 chars
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with single character volume_name", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "a",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "a", // single character is valid
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with numeric only volume_name", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "volume123",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "volume123",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("volume_name validation", func() {

			ginkgo.It("should return a validation error when volume_name is empty", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume_name is too long", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "this-is-a-very-long-volume-name-that-exceeds-the-maximum-allowed-length-of-sixty-four-characters-so-it-should-fail", // more than 64 chars
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume_name contains uppercase", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "Test-Volume", // uppercase not allowed
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume_name contains underscores", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test_volume", // underscores not allowed
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume_name starts with hyphen", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "-test-volume", // can't start with hyphen
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume_name ends with hyphen", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume-", // can't end with hyphen
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume_name starts with number", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "9test-volume", // can't start with number
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume_name contains special characters", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test@volume", // special chars not allowed
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("region validation", func() {

			ginkgo.It("should return a validation error when region is not set", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume",
						SizeGib:    10,
						// Region not set
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("size_gib validation", func() {

			ginkgo.It("should return a validation error when size_gib is 0", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    0, // below minimum
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size_gib exceeds maximum", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    16001, // above maximum
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size_gib is not set", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume",
						Region:     civo.CivoRegion_lon1,
						// SizeGib not set (defaults to 0)
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("tags validation", func() {

			ginkgo.It("should return a validation error when tags are not unique", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
						Tags:       []string{"env:prod", "env:prod"}, // duplicate tags
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tag exceeds max length", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
						Tags:       []string{"this-is-a-very-long-tag-that-exceeds-the-maximum-allowed-length-of-sixty-four-characters-and-should-fail"}, // more than 64 chars
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tag contains invalid characters", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
						Tags:       []string{"env@prod"}, // @ not allowed
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tag contains spaces", func() {
				input := &CivoVolume{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
					Spec: &CivoVolumeSpec{
						VolumeName: "test-volume",
						Region:     civo.CivoRegion_lon1,
						SizeGib:    10,
						Tags:       []string{"env prod"}, // spaces not allowed
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
