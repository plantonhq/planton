package openstackvolumev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOpenStackVolumeSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackVolumeSpec Validation Tests")
}

// newStringValueOrRef is a helper to create a literal StringValueOrRef.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// newValueFromRef is a helper to create a value_from StringValueOrRef.
func newValueFromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Name: name,
			},
		},
	}
}

// minimalValidVolume returns a minimal valid OpenStackVolume for test scaffolding.
func minimalValidVolume() *OpenStackVolume {
	return &OpenStackVolume{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackVolume",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-volume",
		},
		Spec: &OpenStackVolumeSpec{
			Size: 10,
		},
	}
}

var _ = ginkgo.Describe("OpenStackVolumeSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_volume", func() {

			ginkgo.It("should not return a validation error for minimal valid volume", func() {
				input := minimalValidVolume()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume with description", func() {
				input := minimalValidVolume()
				input.Spec.Description = "Database data volume for PostgreSQL"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume with volume_type", func() {
				input := minimalValidVolume()
				input.Spec.VolumeType = "SSD"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume with availability_zone", func() {
				input := minimalValidVolume()
				input.Spec.AvailabilityZone = "az-1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume with snapshot_id", func() {
				input := minimalValidVolume()
				input.Spec.SnapshotId = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume with source_vol_id", func() {
				input := minimalValidVolume()
				input.Spec.SourceVolId = "b2c3d4e5-f6a7-8901-bcde-f12345678901"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume with image_id literal", func() {
				input := minimalValidVolume()
				input.Spec.ImageId = newStringValueOrRef("c3d4e5f6-a7b8-9012-cdef-123456789012")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume with image_id via value_from ref", func() {
				input := minimalValidVolume()
				input.Spec.ImageId = newValueFromRef("my-ubuntu-image")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume with metadata", func() {
				input := minimalValidVolume()
				input.Spec.Metadata = map[string]string{
					"team":    "platform",
					"purpose": "database",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume with region override", func() {
				input := minimalValidVolume()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for large volume size", func() {
				input := minimalValidVolume()
				input.Spec.Size = 1000
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified blank volume", func() {
				input := &OpenStackVolume{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-volume",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackVolumeSpec{
						Description:      "Production database volume",
						Size:             100,
						VolumeType:       "SSD",
						AvailabilityZone: "az-1",
						Metadata: map[string]string{
							"backup":  "daily",
							"purpose": "database",
						},
						Region: "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_volume", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidVolume()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidVolume()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidVolume()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackVolume{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackVolume",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-volume",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size is zero", func() {
				input := minimalValidVolume()
				input.Spec.Size = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size is negative", func() {
				input := minimalValidVolume()
				input.Spec.Size = -1
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when snapshot_id and source_vol_id are both set", func() {
				input := minimalValidVolume()
				input.Spec.SnapshotId = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
				input.Spec.SourceVolId = "b2c3d4e5-f6a7-8901-bcde-f12345678901"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when snapshot_id and image_id are both set", func() {
				input := minimalValidVolume()
				input.Spec.SnapshotId = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
				input.Spec.ImageId = newStringValueOrRef("c3d4e5f6-a7b8-9012-cdef-123456789012")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when source_vol_id and image_id are both set", func() {
				input := minimalValidVolume()
				input.Spec.SourceVolId = "b2c3d4e5-f6a7-8901-bcde-f12345678901"
				input.Spec.ImageId = newStringValueOrRef("c3d4e5f6-a7b8-9012-cdef-123456789012")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when all three sources are set", func() {
				input := minimalValidVolume()
				input.Spec.SnapshotId = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
				input.Spec.SourceVolId = "b2c3d4e5-f6a7-8901-bcde-f12345678901"
				input.Spec.ImageId = newStringValueOrRef("c3d4e5f6-a7b8-9012-cdef-123456789012")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
