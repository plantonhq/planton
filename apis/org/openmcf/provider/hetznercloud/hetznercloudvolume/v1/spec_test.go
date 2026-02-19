package hetznercloudvolumev1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHetznerCloudVolumeSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudVolumeSpec Validation Suite")
}

var _ = Describe("HetznerCloudVolumeSpec validations", func() {

	Context("with valid specs", func() {
		It("should accept a minimal spec (size + location only)", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     10,
				Location: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept size at the upper bound (10240 GB)", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     10240,
				Location: "nbg1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with ext4 format", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     50,
				Location: "hel1",
				Format:   HetznerCloudVolumeSpec_ext4,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with xfs format", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     100,
				Location: "ash",
				Format:   HetznerCloudVolumeSpec_xfs,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept format_unspecified (raw volume)", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     10,
				Location: "fsn1",
				Format:   HetznerCloudVolumeSpec_format_unspecified,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with delete_protection enabled", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:             100,
				Location:         "fsn1",
				DeleteProtection: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with automount", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:      50,
				Location:  "fsn1",
				Format:    HetznerCloudVolumeSpec_ext4,
				Automount: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a fully populated spec", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:             100,
				Location:         "fsn1",
				Format:           HetznerCloudVolumeSpec_ext4,
				Automount:        true,
				DeleteProtection: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("with invalid specs", func() {
		It("should reject size below minimum (9 GB)", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     9,
				Location: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject size of zero", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     0,
				Location: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject negative size", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     -1,
				Location: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject size above maximum (10241 GB)", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     10241,
				Location: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject an empty location", func() {
			spec := &HetznerCloudVolumeSpec{
				Size:     50,
				Location: "",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a missing location (zero value)", func() {
			spec := &HetznerCloudVolumeSpec{
				Size: 50,
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})
})
