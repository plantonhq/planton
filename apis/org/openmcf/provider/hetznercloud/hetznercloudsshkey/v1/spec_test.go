package hetznercloudsshkeyv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHetznerCloudSshKeySpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudSshKeySpec Validation Suite")
}

var _ = Describe("HetznerCloudSshKeySpec validations", func() {

	Context("with a valid spec", func() {
		It("should accept a minimal valid spec", func() {
			spec := &HetznerCloudSshKeySpec{
				PublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample user@host",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept an RSA public key", func() {
			spec := &HetznerCloudSshKeySpec{
				PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQExample user@host",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("with an invalid spec", func() {
		It("should reject an empty public_key", func() {
			spec := &HetznerCloudSshKeySpec{
				PublicKey: "",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a missing public_key (zero value)", func() {
			spec := &HetznerCloudSshKeySpec{}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})
})
