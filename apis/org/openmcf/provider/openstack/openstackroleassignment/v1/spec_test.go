package openstackroleassignmentv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackRoleAssignmentSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenStackRoleAssignmentSpec Validation Suite")
}

// newStringValueOrRef creates a literal StringValueOrRef from a plain string.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// newStringValueOrRefFrom creates a value_from StringValueOrRef for InfraChart DAG wiring.
func newStringValueOrRefFrom(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Name: name,
			},
		},
	}
}

var _ = Describe("OpenStackRoleAssignmentSpec validations", func() {

	Context("positive cases", func() {

		It("should accept a valid user+project assignment with literal project_id", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:    "role-uuid-123",
				ProjectId: newStringValueOrRef("project-uuid-456"),
				UserId:    "user-uuid-789",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a valid user+project assignment with value_from", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:    "role-uuid-123",
				ProjectId: newStringValueOrRefFrom("my-project"),
				UserId:    "user-uuid-789",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a valid group+project assignment", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:    "role-uuid-123",
				ProjectId: newStringValueOrRef("project-uuid-456"),
				GroupId:   "group-uuid-789",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a valid user+domain assignment", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:   "role-uuid-123",
				DomainId: "domain-uuid-456",
				UserId:   "user-uuid-789",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a valid group+domain assignment", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:   "role-uuid-123",
				DomainId: "domain-uuid-456",
				GroupId:  "group-uuid-789",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with region override", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:    "role-uuid-123",
				ProjectId: newStringValueOrRef("project-uuid-456"),
				UserId:    "user-uuid-789",
				Region:    "RegionTwo",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("negative cases - role_id required", func() {

		It("should reject missing role_id", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:    "",
				ProjectId: newStringValueOrRef("project-uuid-456"),
				UserId:    "user-uuid-789",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})

	Context("negative cases - scope XOR (project_id vs domain_id)", func() {

		It("should reject neither project_id nor domain_id set", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId: "role-uuid-123",
				UserId: "user-uuid-789",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject both project_id and domain_id set", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:    "role-uuid-123",
				ProjectId: newStringValueOrRef("project-uuid-456"),
				DomainId:  "domain-uuid-789",
				UserId:    "user-uuid-abc",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})

	Context("negative cases - principal XOR (user_id vs group_id)", func() {

		It("should reject neither user_id nor group_id set", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:    "role-uuid-123",
				ProjectId: newStringValueOrRef("project-uuid-456"),
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject both user_id and group_id set", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:    "role-uuid-123",
				ProjectId: newStringValueOrRef("project-uuid-456"),
				UserId:    "user-uuid-789",
				GroupId:   "group-uuid-abc",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})

	Context("negative cases - both XOR violations simultaneously", func() {

		It("should reject when both scope and principal have no fields set", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId: "role-uuid-123",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject when both scope and principal have both fields set", func() {
			spec := &OpenStackRoleAssignmentSpec{
				RoleId:    "role-uuid-123",
				ProjectId: newStringValueOrRef("project-uuid-456"),
				DomainId:  "domain-uuid-789",
				UserId:    "user-uuid-abc",
				GroupId:   "group-uuid-def",
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})
})
