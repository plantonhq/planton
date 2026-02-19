package hetznercloudsnapshotv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestHetznerCloudSnapshotSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudSnapshotSpec Validation Suite")
}

var _ = Describe("HetznerCloudSnapshotSpec validations", func() {

	Context("with valid specs", func() {
		It("should accept a minimal spec (server_id only)", func() {
			spec := &HetznerCloudSnapshotSpec{
				ServerId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "12345678",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with server_id and description", func() {
			spec := &HetznerCloudSnapshotSpec{
				ServerId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "87654321",
					},
				},
				Description: "pre-upgrade baseline 2026-02-19",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("with invalid specs", func() {
		It("should reject a missing server_id (nil)", func() {
			spec := &HetznerCloudSnapshotSpec{
				Description: "orphan snapshot",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})
})
