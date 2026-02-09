package openstackloadbalancermemberv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackLoadBalancerMemberSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackLoadBalancerMemberSpec Validation Tests")
}

// newStringValueOrRef is a helper to create a literal StringValueOrRef.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// minimalValidMember returns a minimal valid OpenStackLoadBalancerMember for test scaffolding.
func minimalValidMember() *OpenStackLoadBalancerMember {
	return &OpenStackLoadBalancerMember{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackLoadBalancerMember",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-member",
		},
		Spec: &OpenStackLoadBalancerMemberSpec{
			PoolId:       newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
			Address:      "10.0.0.10",
			ProtocolPort: 8080,
		},
	}
}

var _ = ginkgo.Describe("OpenStackLoadBalancerMemberSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_lb_member", func() {

			ginkgo.It("should not return a validation error for minimal valid member", func() {
				input := minimalValidMember()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for member with subnet_id FK", func() {
				input := minimalValidMember()
				input.Spec.SubnetId = newStringValueOrRef("b2c3d4e5-f6a7-8901-bcde-f12345678901")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for member with weight", func() {
				weight := int32(10)
				input := minimalValidMember()
				input.Spec.Weight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for member with admin_state_up false", func() {
				adminStateUp := false
				input := minimalValidMember()
				input.Spec.AdminStateUp = &adminStateUp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for member with tags", func() {
				input := minimalValidMember()
				input.Spec.Tags = []string{"team:platform", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for member with region", func() {
				input := minimalValidMember()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for member with value_from refs", func() {
				input := minimalValidMember()
				input.Spec.PoolId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-pool",
						},
					},
				}
				input.Spec.SubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-subnet",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for member with weight 0 (drain)", func() {
				weight := int32(0)
				input := minimalValidMember()
				input.Spec.Weight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for member with weight 256 (max)", func() {
				weight := int32(256)
				input := minimalValidMember()
				input.Spec.Weight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified member", func() {
				adminStateUp := true
				weight := int32(10)
				input := &OpenStackLoadBalancerMember{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackLoadBalancerMember",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-member",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackLoadBalancerMemberSpec{
						PoolId:       newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
						Address:      "10.0.0.10",
						ProtocolPort: 8080,
						SubnetId:     newStringValueOrRef("b2c3d4e5-f6a7-8901-bcde-f12345678901"),
						Weight:       &weight,
						AdminStateUp: &adminStateUp,
						Tags:         []string{"production", "managed"},
						Region:       "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_lb_member", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidMember()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidMember()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidMember()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackLoadBalancerMember{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackLoadBalancerMember",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-member",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when pool_id is missing", func() {
				input := minimalValidMember()
				input.Spec.PoolId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when address is missing", func() {
				input := minimalValidMember()
				input.Spec.Address = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when protocol_port is 0", func() {
				input := minimalValidMember()
				input.Spec.ProtocolPort = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when protocol_port is 65536", func() {
				input := minimalValidMember()
				input.Spec.ProtocolPort = 65536
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when protocol_port is -1", func() {
				input := minimalValidMember()
				input.Spec.ProtocolPort = -1
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when weight is -1", func() {
				weight := int32(-1)
				input := minimalValidMember()
				input.Spec.Weight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when weight is 257", func() {
				weight := int32(257)
				input := minimalValidMember()
				input.Spec.Weight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidMember()
				input.Spec.Tags = []string{"env:dev", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
