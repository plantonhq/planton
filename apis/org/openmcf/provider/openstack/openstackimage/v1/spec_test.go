package openstackimagev1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOpenStackImageSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenStackImageSpec Validation Suite")
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

var _ = Describe("OpenStackImageSpec validations", func() {

	Context("positive cases", func() {

		It("should accept a minimal valid spec with only required fields", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with image_source_url", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				ImageSourceUrl:  "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with min_disk_gb and min_ram_mb", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				MinDiskGb:       10,
				MinRamMb:        512,
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with zero min_disk_gb and min_ram_mb", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				MinDiskGb:       0,
				MinRamMb:        0,
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with protected set to true", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				Protected:       boolPtr(true),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with hidden set to true", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				Hidden:          boolPtr(true),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with tags", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				Tags:            []string{"ubuntu", "22.04", "cloud-init"},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with visibility set to public", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				Visibility:      stringPtr("public"),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with visibility set to shared", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				Visibility:      stringPtr("shared"),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with visibility set to community", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				Visibility:      stringPtr("community"),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept all container_format values", func() {
			formats := []string{"bare", "ovf", "aki", "ari", "ami", "ova", "docker", "compressed"}
			for _, format := range formats {
				spec := &OpenStackImageSpec{
					ContainerFormat: format,
					DiskFormat:      "qcow2",
				}
				Expect(protovalidate.Validate(spec)).To(BeNil())
			}
		})

		It("should accept all disk_format values", func() {
			formats := []string{"raw", "vhd", "vhdx", "vmdk", "vdi", "iso", "ploop", "qcow2", "aki", "ari", "ami"}
			for _, format := range formats {
				spec := &OpenStackImageSpec{
					ContainerFormat: "bare",
					DiskFormat:      format,
				}
				Expect(protovalidate.Validate(spec)).To(BeNil())
			}
		})

		It("should accept a fully populated spec", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				ImageSourceUrl:  "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img",
				MinDiskGb:       20,
				MinRamMb:        2048,
				Protected:       boolPtr(true),
				Hidden:          boolPtr(false),
				Tags:            []string{"ubuntu", "jammy", "production"},
				Visibility:      stringPtr("private"),
				Region:          "RegionOne",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("negative cases", func() {

		It("should reject empty container_format", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "",
				DiskFormat:      "qcow2",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject invalid container_format", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "invalid",
				DiskFormat:      "qcow2",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject empty disk_format", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject invalid disk_format", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "invalid",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject negative min_disk_gb", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				MinDiskGb:       -1,
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject negative min_ram_mb", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				MinRamMb:        -1,
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject invalid visibility value", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "qcow2",
				Visibility:      stringPtr("internal"),
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject case-sensitive container_format", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "BARE",
				DiskFormat:      "qcow2",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject case-sensitive disk_format", func() {
			spec := &OpenStackImageSpec{
				ContainerFormat: "bare",
				DiskFormat:      "QCOW2",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})
})
