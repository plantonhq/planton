package hetznercloudplacementgroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHetznerCloudPlacementGroupSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudPlacementGroupSpec Validation Suite")
}

var _ = Describe("HetznerCloudPlacementGroupSpec validations", func() {

	Context("with a valid spec", func() {
		It("should accept a spec with type unset (middleware applies default)", func() {
			spec := &HetznerCloudPlacementGroupSpec{}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with type explicitly set to spread", func() {
			spread := HetznerCloudPlacementGroupSpec_spread
			spec := &HetznerCloudPlacementGroupSpec{
				Type: &spread,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})
})
