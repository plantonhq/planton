package openstackloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackLoadBalancerSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidLoadBalancer() *OpenStackLoadBalancer {
	return &OpenStackLoadBalancer{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackLoadBalancer",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-lb",
		},
		Spec: &OpenStackLoadBalancerSpec{
			VipSubnetId: newStringValueOrRef("e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d"),
		},
	}
}

var _ = ginkgo.Describe("OpenStackLoadBalancerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_load_balancer", func() {

			ginkgo.It("should not return a validation error for minimal valid load balancer", func() {
				input := minimalValidLoadBalancer()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for LB with description", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Description = "Production load balancer"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for LB with vip_address", func() {
				input := minimalValidLoadBalancer()
				input.Spec.VipAddress = "192.168.1.100"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for LB with admin_state_up false", func() {
				adminStateUp := false
				input := minimalValidLoadBalancer()
				input.Spec.AdminStateUp = &adminStateUp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for LB with flavor_id", func() {
				input := minimalValidLoadBalancer()
				input.Spec.FlavorId = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for LB with tags", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Tags = []string{"env:prod", "team:platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for LB with region", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for LB with value_from ref", func() {
				input := minimalValidLoadBalancer()
				input.Spec.VipSubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-subnet",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified LB", func() {
				adminStateUp := true
				input := &OpenStackLoadBalancer{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackLoadBalancer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-lb",
						Org:  "acme-corp",
						Env:  "production",
					},
					Spec: &OpenStackLoadBalancerSpec{
						VipSubnetId:  newStringValueOrRef("e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d"),
						VipAddress:   "10.0.0.100",
						Description:  "Production Octavia LB",
						AdminStateUp: &adminStateUp,
						FlavorId:     "flavor-uuid",
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
		ginkgo.Context("openstack_load_balancer", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidLoadBalancer()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidLoadBalancer()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidLoadBalancer()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackLoadBalancer{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackLoadBalancer",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-lb"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vip_subnet_id is missing", func() {
				input := minimalValidLoadBalancer()
				input.Spec.VipSubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Tags = []string{"env:prod", "env:prod"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
