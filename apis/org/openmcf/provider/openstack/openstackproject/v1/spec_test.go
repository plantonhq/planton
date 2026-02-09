package openstackprojectv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOpenStackProjectSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenStackProjectSpec Validation Suite")
}

func boolPtr(b bool) *bool {
	return &b
}

var _ = Describe("OpenStackProjectSpec validations", func() {

	Context("positive cases", func() {

		It("should accept a minimal valid spec (all defaults)", func() {
			spec := &OpenStackProjectSpec{}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with description only", func() {
			spec := &OpenStackProjectSpec{
				Description: "Developer sandbox for team alpha",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with domain_id", func() {
			spec := &OpenStackProjectSpec{
				DomainId: "default",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with enabled explicitly set to true", func() {
			spec := &OpenStackProjectSpec{
				Enabled: boolPtr(true),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with enabled explicitly set to false", func() {
			spec := &OpenStackProjectSpec{
				Enabled: boolPtr(false),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with parent_id for nested project hierarchy", func() {
			spec := &OpenStackProjectSpec{
				ParentId: "abcdef12-3456-7890-abcd-ef1234567890",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with tags", func() {
			spec := &OpenStackProjectSpec{
				Tags: []string{"team:alpha", "env:dev", "cost-center:engineering"},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with region override", func() {
			spec := &OpenStackProjectSpec{
				Region: "RegionOne",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a fully populated spec", func() {
			spec := &OpenStackProjectSpec{
				Description: "Production workloads for platform team",
				DomainId:    "default",
				Enabled:     boolPtr(true),
				ParentId:    "parent-project-uuid",
				Tags:        []string{"production", "platform"},
				Region:      "RegionTwo",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with empty tags list", func() {
			spec := &OpenStackProjectSpec{
				Tags: []string{},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with enabled unset (nil pointer -- middleware applies default true)", func() {
			spec := &OpenStackProjectSpec{
				Description: "Test project",
			}
			// enabled is nil here, middleware will apply default "true"
			Expect(spec.Enabled).To(BeNil())
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	// Note: OpenStackProjectSpec has no required fields, no CEL validations,
	// and no constrained string enums. The only field option is the default
	// on 'enabled'. Since all fields are optional and unconstrained, there
	// are limited negative validation cases. The tests below verify that the
	// spec structure is well-formed by testing various valid combinations.

	Context("edge cases", func() {

		It("should accept description with special characters", func() {
			spec := &OpenStackProjectSpec{
				Description: "ARM Ltd. — Developer Environment (Phase 1) @2026",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a very long description", func() {
			longDesc := ""
			for i := 0; i < 500; i++ {
				longDesc += "a"
			}
			spec := &OpenStackProjectSpec{
				Description: longDesc,
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})
})
