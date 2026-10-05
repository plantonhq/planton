package digitaloceanvolumev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/plantonhq/planton/shared"
)

func TestDigitalOceanVolumeSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanVolumeSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanVolumeSpec validations", func() {

	// Helper function to create a minimal valid volume spec
	makeValidMinimalSpec := func() *DigitalOceanVolumeSpec {
		return &DigitalOceanVolumeSpec{
			VolumeName:     "test-volume",
			Region:         digitalocean.DigitalOceanRegion_nyc3,
			SizeGib:        10,
			FilesystemType: DigitalOceanVolumeFilesystemType_unformatted,
		}
	}

	// Helper function to create a production-ready volume spec
	makeValidProductionSpec := func() *DigitalOceanVolumeSpec {
		return &DigitalOceanVolumeSpec{
			VolumeName:             "prod-db-data",
			Description:            "PostgreSQL data volume for production",
			Region:                 digitalocean.DigitalOceanRegion_sfo3,
			SizeGib:                500,
			FilesystemType:         DigitalOceanVolumeFilesystemType_xfs,
			InitialFilesystemLabel: "pgdata",
			Tags:                   []string{"env:prod", "service:postgres", "tier:database"},
		}
	}

	// Helper function to create a volume from snapshot
	makeValidSnapshotSpec := func() *DigitalOceanVolumeSpec {
		return &DigitalOceanVolumeSpec{
			VolumeName:     "restored-volume",
			Description:    "Volume restored from snapshot",
			Region:         digitalocean.DigitalOceanRegion_nyc3,
			SizeGib:        100,
			FilesystemType: DigitalOceanVolumeFilesystemType_ext4,
			SnapshotId:     "snapshot-abc123",
		}
	}

	ginkgo.Context("Valid configurations", func() {

		ginkgo.It("should accept minimal valid configuration", func() {
			spec := makeValidMinimalSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept production configuration with all fields", func() {
			spec := makeValidProductionSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept volume from snapshot", func() {
			spec := makeValidSnapshotSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept ext4 filesystem type", func() {
			spec := makeValidMinimalSpec()
			spec.FilesystemType = DigitalOceanVolumeFilesystemType_ext4
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept xfs filesystem type", func() {
			spec := makeValidMinimalSpec()
			spec.FilesystemType = DigitalOceanVolumeFilesystemType_xfs
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Required fields", func() {

		ginkgo.It("should reject spec with missing volume_name", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject spec with missing region", func() {
			spec := makeValidMinimalSpec()
			spec.Region = digitalocean.DigitalOceanRegion_digital_ocean_region_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject spec with missing size_gib", func() {
			spec := makeValidMinimalSpec()
			spec.SizeGib = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Volume name validation", func() {

		ginkgo.It("should reject name with uppercase letters", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = "Test-Volume"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name with underscores", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = "test_volume"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name with spaces", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = "test volume"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name starting with hyphen", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = "-test-volume"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name ending with hyphen", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = "test-volume-"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name starting with number", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = "1test-volume"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name that is too long (>64 characters)", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = "this-is-a-very-long-volume-name-that-exceeds-the-maximum-allowed-length-of-sixty-four-characters"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept name with lowercase letters, numbers, and hyphens", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = "prod-db-vol-2024"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept single character name", func() {
			spec := makeValidMinimalSpec()
			spec.VolumeName = "v"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept 64 character name", func() {
			spec := makeValidMinimalSpec()
			// Exactly 64 characters
			spec.VolumeName = "a234567890123456789012345678901234567890123456789012345678901234"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Description validation", func() {

		ginkgo.It("should accept empty description", func() {
			spec := makeValidMinimalSpec()
			spec.Description = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a long description (the provider imposes no length cap)", func() {
			spec := makeValidMinimalSpec()
			spec.Description = "This volume holds the primary PostgreSQL data directory for the production deployment of the billing service, including WAL archives staged before shipping to object storage."
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Size validation", func() {

		ginkgo.It("should reject size of 0", func() {
			spec := makeValidMinimalSpec()
			spec.SizeGib = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept minimum size of 1 GiB", func() {
			spec := makeValidMinimalSpec()
			spec.SizeGib = 1
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept 16 TiB (DigitalOcean's documented maximum, enforced by the API)", func() {
			spec := makeValidMinimalSpec()
			spec.SizeGib = 16384
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept common production sizes", func() {
			testSizes := []uint32{10, 50, 100, 250, 500, 1000, 2000, 5000, 10000}
			for _, size := range testSizes {
				spec := makeValidMinimalSpec()
				spec.SizeGib = size
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})
	})

	ginkgo.Context("Tags validation", func() {

		ginkgo.It("should accept empty tags list", func() {
			spec := makeValidMinimalSpec()
			spec.Tags = []string{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept valid tags with letters, numbers, colons, dashes, underscores", func() {
			spec := makeValidMinimalSpec()
			spec.Tags = []string{"env:prod", "service-postgres", "tier_database", "version-1-2-3"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate tags", func() {
			spec := makeValidMinimalSpec()
			spec.Tags = []string{"env:prod", "service:postgres", "env:prod"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject tags with spaces", func() {
			spec := makeValidMinimalSpec()
			spec.Tags = []string{"env prod"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject tags with special characters", func() {
			spec := makeValidMinimalSpec()
			spec.Tags = []string{"env@prod"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept a tag of 255 characters (the provider limit)", func() {
			spec := makeValidMinimalSpec()
			tag := make([]byte, 255)
			for i := range tag {
				tag[i] = 'a'
			}
			spec.Tags = []string{string(tag)}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a tag longer than 255 characters", func() {
			spec := makeValidMinimalSpec()
			tag := make([]byte, 256)
			for i := range tag {
				tag[i] = 'a'
			}
			spec.Tags = []string{string(tag)}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Filesystem type validation", func() {

		ginkgo.It("should accept the unformatted filesystem type", func() {
			spec := makeValidMinimalSpec()
			spec.FilesystemType = DigitalOceanVolumeFilesystemType_unformatted
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept the ext4 filesystem type", func() {
			spec := makeValidMinimalSpec()
			spec.FilesystemType = DigitalOceanVolumeFilesystemType_ext4
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept the xfs filesystem type", func() {
			spec := makeValidMinimalSpec()
			spec.FilesystemType = DigitalOceanVolumeFilesystemType_xfs
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Initial filesystem label", func() {

		ginkgo.It("should accept a label alongside a filesystem type", func() {
			spec := makeValidMinimalSpec()
			spec.FilesystemType = DigitalOceanVolumeFilesystemType_ext4
			spec.InitialFilesystemLabel = "data"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an empty label", func() {
			spec := makeValidMinimalSpec()
			spec.InitialFilesystemLabel = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Snapshot ID validation", func() {

		ginkgo.It("should accept empty snapshot_id", func() {
			spec := makeValidMinimalSpec()
			spec.SnapshotId = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept valid snapshot_id", func() {
			spec := makeValidMinimalSpec()
			spec.SnapshotId = "snapshot-abc123"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept numeric snapshot_id", func() {
			spec := makeValidMinimalSpec()
			spec.SnapshotId = "123456789"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Full DigitalOceanVolume resource validation", func() {

		ginkgo.It("should accept complete valid volume resource", func() {
			input := &DigitalOceanVolume{
				ApiVersion: "digital-ocean.planton.dev/v1alpha1",
				Kind:       "DigitalOceanVolume",
				Metadata: &shared.CatalogObjectMetadata{
					Name: "test-volume",
				},
				Spec: makeValidMinimalSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept production volume resource with all fields", func() {
			input := &DigitalOceanVolume{
				ApiVersion: "digital-ocean.planton.dev/v1alpha1",
				Kind:       "DigitalOceanVolume",
				Metadata: &shared.CatalogObjectMetadata{
					Name: "prod-db-data",
					Labels: map[string]string{
						"app": "postgres",
						"env": "production",
					},
				},
				Spec: makeValidProductionSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
