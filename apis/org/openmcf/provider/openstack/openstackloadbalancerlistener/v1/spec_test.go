package openstackloadbalancerlistenerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackLoadBalancerListenerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackLoadBalancerListenerSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidListener() *OpenStackLoadBalancerListener {
	return &OpenStackLoadBalancerListener{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackLoadBalancerListener",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-listener",
		},
		Spec: &OpenStackLoadBalancerListenerSpec{
			LoadbalancerId: newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
			Protocol:       "HTTP",
			ProtocolPort:   80,
		},
	}
}

var _ = ginkgo.Describe("OpenStackLoadBalancerListenerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_load_balancer_listener", func() {

			ginkgo.It("should not return a validation error for minimal valid listener", func() {
				input := minimalValidListener()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for listener with description", func() {
				input := minimalValidListener()
				input.Spec.Description = "HTTP listener for web traffic"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for HTTPS protocol", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.ProtocolPort = 443
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TCP protocol", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "TCP"
				input.Spec.ProtocolPort = 3306
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for UDP protocol", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "UDP"
				input.Spec.ProtocolPort = 53
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TERMINATED_HTTPS with tls ref", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "TERMINATED_HTTPS"
				input.Spec.ProtocolPort = 443
				input.Spec.DefaultTlsContainerRef = "https://barbican.example.com/v1/secrets/abc-123"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for listener with connection_limit", func() {
				connLimit := int32(10000)
				input := minimalValidListener()
				input.Spec.ConnectionLimit = &connLimit
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for listener with connection_limit of -1", func() {
				connLimit := int32(-1)
				input := minimalValidListener()
				input.Spec.ConnectionLimit = &connLimit
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for listener with insert_headers", func() {
				input := minimalValidListener()
				input.Spec.InsertHeaders = map[string]string{
					"X-Forwarded-For":   "true",
					"X-Forwarded-Proto": "true",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for listener with allowed_cidrs", func() {
				input := minimalValidListener()
				input.Spec.AllowedCidrs = []string{"10.0.0.0/8", "192.168.0.0/16"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for listener with admin_state_up false", func() {
				adminStateUp := false
				input := minimalValidListener()
				input.Spec.AdminStateUp = &adminStateUp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for listener with tags", func() {
				input := minimalValidListener()
				input.Spec.Tags = []string{"env:prod", "team:platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for listener with region", func() {
				input := minimalValidListener()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for listener with value_from ref", func() {
				input := minimalValidListener()
				input.Spec.LoadbalancerId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-loadbalancer",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for protocol_port at lower bound", func() {
				input := minimalValidListener()
				input.Spec.ProtocolPort = 1
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for protocol_port at upper bound", func() {
				input := minimalValidListener()
				input.Spec.ProtocolPort = 65535
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified listener", func() {
				adminStateUp := true
				connLimit := int32(50000)
				input := &OpenStackLoadBalancerListener{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackLoadBalancerListener",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-https-listener",
						Org:  "acme-corp",
						Env:  "production",
					},
					Spec: &OpenStackLoadBalancerListenerSpec{
						LoadbalancerId:         newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
						Protocol:               "TERMINATED_HTTPS",
						ProtocolPort:            443,
						Description:             "Production HTTPS listener with TLS termination",
						ConnectionLimit:         &connLimit,
						DefaultTlsContainerRef: "https://barbican.example.com/v1/secrets/cert-abc-123",
						InsertHeaders: map[string]string{
							"X-Forwarded-For":   "true",
							"X-Forwarded-Proto": "true",
						},
						AllowedCidrs: []string{"10.0.0.0/8"},
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
		ginkgo.Context("openstack_load_balancer_listener", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidListener()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidListener()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidListener()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackLoadBalancerListener{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackLoadBalancerListener",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-listener"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when loadbalancer_id is missing", func() {
				input := minimalValidListener()
				input.Spec.LoadbalancerId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when protocol is invalid", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "INVALID_PROTOCOL"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when protocol is empty", func() {
				input := minimalValidListener()
				input.Spec.Protocol = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when protocol_port is 0", func() {
				input := minimalValidListener()
				input.Spec.ProtocolPort = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when protocol_port exceeds 65535", func() {
				input := minimalValidListener()
				input.Spec.ProtocolPort = 65536
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when TERMINATED_HTTPS is missing tls ref", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "TERMINATED_HTTPS"
				input.Spec.ProtocolPort = 443
				input.Spec.DefaultTlsContainerRef = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidListener()
				input.Spec.Tags = []string{"env:prod", "env:prod"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
