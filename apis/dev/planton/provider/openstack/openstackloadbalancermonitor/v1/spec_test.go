package openstackloadbalancermonitorv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOpenStackLoadBalancerMonitorSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackLoadBalancerMonitorSpec Validation Tests")
}

// newStringValueOrRef is a helper to create a literal StringValueOrRef.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// minimalValidMonitor returns a minimal valid OpenStackLoadBalancerMonitor for test scaffolding.
func minimalValidMonitor() *OpenStackLoadBalancerMonitor {
	return &OpenStackLoadBalancerMonitor{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackLoadBalancerMonitor",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-monitor",
		},
		Spec: &OpenStackLoadBalancerMonitorSpec{
			PoolId:     newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
			Type:       "HTTP",
			Delay:      5,
			Timeout:    10,
			MaxRetries: 3,
		},
	}
}

var _ = ginkgo.Describe("OpenStackLoadBalancerMonitorSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_lb_monitor", func() {

			ginkgo.It("should not return a validation error for minimal valid monitor", func() {
				input := minimalValidMonitor()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for monitor with url_path", func() {
				input := minimalValidMonitor()
				input.Spec.UrlPath = "/healthz"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for monitor with http_method GET", func() {
				input := minimalValidMonitor()
				input.Spec.HttpMethod = "GET"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for monitor with expected_codes", func() {
				input := minimalValidMonitor()
				input.Spec.ExpectedCodes = "200"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for monitor with max_retries_down", func() {
				maxRetriesDown := int32(5)
				input := minimalValidMonitor()
				input.Spec.MaxRetriesDown = &maxRetriesDown
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for monitor with admin_state_up false", func() {
				adminStateUp := false
				input := minimalValidMonitor()
				input.Spec.AdminStateUp = &adminStateUp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for monitor with region", func() {
				input := minimalValidMonitor()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for PING monitor", func() {
				input := minimalValidMonitor()
				input.Spec.Type = "PING"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TCP monitor", func() {
				input := minimalValidMonitor()
				input.Spec.Type = "TCP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TLS-HELLO monitor", func() {
				input := minimalValidMonitor()
				input.Spec.Type = "TLS-HELLO"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for UDP-CONNECT monitor", func() {
				input := minimalValidMonitor()
				input.Spec.Type = "UDP-CONNECT"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for HTTPS monitor", func() {
				input := minimalValidMonitor()
				input.Spec.Type = "HTTPS"
				input.Spec.UrlPath = "/health"
				input.Spec.HttpMethod = "GET"
				input.Spec.ExpectedCodes = "200-299"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified HTTP monitor", func() {
				adminStateUp := true
				maxRetriesDown := int32(5)
				input := &OpenStackLoadBalancerMonitor{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackLoadBalancerMonitor",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-monitor",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackLoadBalancerMonitorSpec{
						PoolId:         newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
						Type:           "HTTP",
						Delay:          5,
						Timeout:        10,
						MaxRetries:     3,
						MaxRetriesDown: &maxRetriesDown,
						UrlPath:        "/healthz",
						HttpMethod:     "GET",
						ExpectedCodes:  "200",
						AdminStateUp:   &adminStateUp,
						Region:         "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for pool_id via value_from ref", func() {
				input := minimalValidMonitor()
				input.Spec.PoolId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-pool",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_lb_monitor", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidMonitor()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidMonitor()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidMonitor()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackLoadBalancerMonitor{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackLoadBalancerMonitor",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-monitor",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when pool_id is missing", func() {
				input := minimalValidMonitor()
				input.Spec.PoolId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is missing", func() {
				input := minimalValidMonitor()
				input.Spec.Type = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is invalid", func() {
				input := minimalValidMonitor()
				input.Spec.Type = "INVALID"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when delay is missing", func() {
				input := minimalValidMonitor()
				input.Spec.Delay = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when timeout is missing", func() {
				input := minimalValidMonitor()
				input.Spec.Timeout = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_retries is missing", func() {
				input := minimalValidMonitor()
				input.Spec.MaxRetries = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_retries is 0", func() {
				input := minimalValidMonitor()
				input.Spec.MaxRetries = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_retries is 11", func() {
				input := minimalValidMonitor()
				input.Spec.MaxRetries = 11
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_retries_down is 0", func() {
				maxRetriesDown := int32(0)
				input := minimalValidMonitor()
				input.Spec.MaxRetriesDown = &maxRetriesDown
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_retries_down is 11", func() {
				maxRetriesDown := int32(11)
				input := minimalValidMonitor()
				input.Spec.MaxRetriesDown = &maxRetriesDown
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when url_path is set with type PING", func() {
				input := minimalValidMonitor()
				input.Spec.Type = "PING"
				input.Spec.UrlPath = "/healthz"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when http_method is set with type TCP", func() {
				input := minimalValidMonitor()
				input.Spec.Type = "TCP"
				input.Spec.HttpMethod = "GET"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when http_method is invalid", func() {
				input := minimalValidMonitor()
				input.Spec.HttpMethod = "INVALID"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
