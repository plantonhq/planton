package openstackloadbalancerpoolv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackLoadBalancerPoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackLoadBalancerPoolSpec Validation Tests")
}

// newStringValueOrRef is a helper to create a literal StringValueOrRef.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// minimalValidPool returns a minimal valid OpenStackLoadBalancerPool for test scaffolding.
func minimalValidPool() *OpenStackLoadBalancerPool {
	return &OpenStackLoadBalancerPool{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackLoadBalancerPool",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-pool",
		},
		Spec: &OpenStackLoadBalancerPoolSpec{
			ListenerId: newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
			Protocol:   "HTTP",
			LbMethod:   "ROUND_ROBIN",
		},
	}
}

var _ = ginkgo.Describe("OpenStackLoadBalancerPoolSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_lb_pool", func() {

			ginkgo.It("should not return a validation error for minimal valid pool", func() {
				input := minimalValidPool()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with description", func() {
				input := minimalValidPool()
				input.Spec.Description = "Backend pool for web application"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with tags", func() {
				input := minimalValidPool()
				input.Spec.Tags = []string{"team:platform", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with region override", func() {
				input := minimalValidPool()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with persistence SOURCE_IP", func() {
				input := minimalValidPool()
				input.Spec.Persistence = &SessionPersistence{
					Type: "SOURCE_IP",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with persistence HTTP_COOKIE", func() {
				input := minimalValidPool()
				input.Spec.Persistence = &SessionPersistence{
					Type: "HTTP_COOKIE",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with persistence APP_COOKIE and cookie_name", func() {
				input := minimalValidPool()
				input.Spec.Persistence = &SessionPersistence{
					Type:       "APP_COOKIE",
					CookieName: "JSESSIONID",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with protocol HTTP", func() {
				input := minimalValidPool()
				input.Spec.Protocol = "HTTP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with protocol HTTPS", func() {
				input := minimalValidPool()
				input.Spec.Protocol = "HTTPS"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with protocol TCP", func() {
				input := minimalValidPool()
				input.Spec.Protocol = "TCP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with protocol UDP", func() {
				input := minimalValidPool()
				input.Spec.Protocol = "UDP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with protocol PROXY", func() {
				input := minimalValidPool()
				input.Spec.Protocol = "PROXY"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with lb_method ROUND_ROBIN", func() {
				input := minimalValidPool()
				input.Spec.LbMethod = "ROUND_ROBIN"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with lb_method LEAST_CONNECTIONS", func() {
				input := minimalValidPool()
				input.Spec.LbMethod = "LEAST_CONNECTIONS"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with lb_method SOURCE_IP", func() {
				input := minimalValidPool()
				input.Spec.LbMethod = "SOURCE_IP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with lb_method SOURCE_IP_PORT", func() {
				input := minimalValidPool()
				input.Spec.LbMethod = "SOURCE_IP_PORT"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool with listener_id via value_from ref", func() {
				input := minimalValidPool()
				input.Spec.ListenerId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-listener",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified pool", func() {
				adminStateUp := true
				input := &OpenStackLoadBalancerPool{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackLoadBalancerPool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-pool",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackLoadBalancerPoolSpec{
						ListenerId: newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
						Protocol:   "HTTP",
						LbMethod:   "LEAST_CONNECTIONS",
						Persistence: &SessionPersistence{
							Type:       "APP_COOKIE",
							CookieName: "JSESSIONID",
						},
						Description:  "Production backend pool for ACME Corp",
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
		ginkgo.Context("openstack_lb_pool", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidPool()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidPool()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidPool()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackLoadBalancerPool{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackLoadBalancerPool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener_id is missing", func() {
				input := minimalValidPool()
				input.Spec.ListenerId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when protocol is missing", func() {
				input := minimalValidPool()
				input.Spec.Protocol = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when lb_method is missing", func() {
				input := minimalValidPool()
				input.Spec.LbMethod = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when protocol is invalid", func() {
				input := minimalValidPool()
				input.Spec.Protocol = "INVALID"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when lb_method is invalid", func() {
				input := minimalValidPool()
				input.Spec.LbMethod = "INVALID"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when persistence has cookie_name but type is not APP_COOKIE", func() {
				input := minimalValidPool()
				input.Spec.Persistence = &SessionPersistence{
					Type:       "SOURCE_IP",
					CookieName: "my-cookie",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidPool()
				input.Spec.Tags = []string{"env:dev", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
