package openstackvolumeattachv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackVolumeAttachSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackVolumeAttachSpec Validation Tests")
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

// minimalValidVolumeAttach returns a minimal valid OpenStackVolumeAttach.
// Both required FKs (instance_id and volume_id) are set with literal values.
func minimalValidVolumeAttach() *OpenStackVolumeAttach {
	return &OpenStackVolumeAttach{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackVolumeAttach",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-va",
		},
		Spec: &OpenStackVolumeAttachSpec{
			InstanceId: newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
			VolumeId:   newStringValueOrRef("b2c3d4e5-f6a7-8901-bcde-f12345678901"),
		},
	}
}

var _ = ginkgo.Describe("OpenStackVolumeAttachSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_volume_attach", func() {

			ginkgo.It("should not return a validation error for minimal valid attachment", func() {
				input := minimalValidVolumeAttach()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance_id via value_from ref", func() {
				input := minimalValidVolumeAttach()
				input.Spec.InstanceId = newValueFromRef("my-instance")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for volume_id via value_from ref", func() {
				input := minimalValidVolumeAttach()
				input.Spec.VolumeId = newValueFromRef("my-volume")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for both FKs via value_from refs", func() {
				input := &OpenStackVolumeAttach{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackVolumeAttach",
					Metadata: &shared.CloudResourceMetadata{
						Name: "app-data-attach",
					},
					Spec: &OpenStackVolumeAttachSpec{
						InstanceId: newValueFromRef("app-server"),
						VolumeId:   newValueFromRef("app-data"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for attachment with device", func() {
				input := minimalValidVolumeAttach()
				input.Spec.Device = "/dev/vdb"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for attachment with region override", func() {
				input := minimalValidVolumeAttach()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified attachment", func() {
				input := &OpenStackVolumeAttach{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackVolumeAttach",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-va",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackVolumeAttachSpec{
						InstanceId: newValueFromRef("prod-server"),
						VolumeId:   newValueFromRef("prod-data"),
						Device:     "/dev/vdc",
						Region:     "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_volume_attach", func() {

			ginkgo.It("should return a validation error when instance_id is missing", func() {
				input := minimalValidVolumeAttach()
				input.Spec.InstanceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume_id is missing", func() {
				input := minimalValidVolumeAttach()
				input.Spec.VolumeId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both FKs are missing", func() {
				input := &OpenStackVolumeAttach{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackVolumeAttach",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-va",
					},
					Spec: &OpenStackVolumeAttachSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidVolumeAttach()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidVolumeAttach()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidVolumeAttach()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackVolumeAttach{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackVolumeAttach",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-va",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
