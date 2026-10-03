package awslbtargetgroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsLbTargetGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsLbTargetGroupSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidTargetGroup is the common case: an ALB HTTP target group for
// IP targets (the shape every ECS/EKS integration uses).
func minimalValidTargetGroup() *AwsLbTargetGroup {
	return &AwsLbTargetGroup{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsLbTargetGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "api",
		},
		Spec: &AwsLbTargetGroupSpec{
			Region:     "us-west-2",
			VpcId:      literalRef("vpc-0123456789abcdef0"),
			TargetType: "ip",
			Port:       8080,
			Protocol:   "HTTP",
		},
	}
}

var _ = ginkgo.Describe("AwsLbTargetGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_lb_target_group", func() {

			ginkgo.It("should not return a validation error for a minimal ALB target group", func() {
				err := protovalidate.Validate(minimalValidTargetGroup())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an NLB TCP target group with health checks", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "TCP"
				input.Spec.Port = 5432
				preserve := true
				input.Spec.PreserveClientIp = &preserve
				input.Spec.ProxyProtocolV2 = true
				input.Spec.ConnectionTermination = true
				input.Spec.HealthCheck = &AwsLbTargetGroupHealthCheck{
					Protocol:           "HTTP",
					Port:               "8081",
					Path:               "/healthz",
					HealthyThreshold:   3,
					UnhealthyThreshold: 3,
					IntervalSeconds:    10,
					TimeoutSeconds:     5,
					Matcher:            "200-299",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a lambda target group", func() {
				input := minimalValidTargetGroup()
				input.Spec.TargetType = "lambda"
				input.Spec.VpcId = nil
				input.Spec.Port = 0
				input.Spec.Protocol = ""
				input.Spec.LambdaMultiValueHeadersEnabled = true
				input.Spec.Targets = []*AwsLbTargetGroupTarget{
					{TargetId: literalRef("arn:aws:lambda:us-west-2:123456789012:function:handler")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a tuned ALB target group", func() {
				input := minimalValidTargetGroup()
				input.Spec.ProtocolVersion = "GRPC"
				input.Spec.IpAddressType = "ipv4"
				input.Spec.DeregistrationDelaySeconds = 60
				input.Spec.LoadBalancingAlgorithmType = "weighted_random"
				input.Spec.LoadBalancingAnomalyMitigation = "on"
				input.Spec.LoadBalancingCrossZoneEnabled = "use_load_balancer_configuration"
				input.Spec.Stickiness = &AwsLbTargetGroupStickiness{
					Type:                  "lb_cookie",
					CookieDurationSeconds: 3600,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for slow start without stickiness", func() {
				input := minimalValidTargetGroup()
				input.Spec.SlowStartSeconds = 120
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for app_cookie stickiness with a cookie name", func() {
				input := minimalValidTargetGroup()
				input.Spec.Stickiness = &AwsLbTargetGroupStickiness{
					Type:       "app_cookie",
					CookieName: "SESSIONID",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for group-level health policy and target health state", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "TCP"
				input.Spec.TargetGroupHealth = &AwsLbTargetGroupHealthPolicy{
					DnsFailover: &AwsLbTargetGroupDnsFailover{
						MinimumHealthyTargetsCount:      "2",
						MinimumHealthyTargetsPercentage: "off",
					},
					UnhealthyStateRouting: &AwsLbTargetGroupUnhealthyStateRouting{
						MinimumHealthyTargetsPercentage: "50",
					},
				}
				input.Spec.TargetHealthState = &AwsLbTargetGroupTargetHealthState{
					EnableUnhealthyConnectionTermination: false,
					UnhealthyDrainingIntervalSeconds:     300,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for static instance targets", func() {
				input := minimalValidTargetGroup()
				input.Spec.TargetType = "instance"
				input.Spec.Targets = []*AwsLbTargetGroupTarget{
					{TargetId: literalRef("i-0123456789abcdef0"), Port: 9090},
					{TargetId: literalRef("i-0fedcba9876543210")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a QUIC target group with a QUIC server id", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "QUIC"
				input.Spec.Port = 443
				input.Spec.Targets = []*AwsLbTargetGroupTarget{
					{TargetId: literalRef("10.0.1.15"), QuicServerId: "server-1"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a TCP_QUIC group with flow-hash stickiness", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "TCP_QUIC"
				input.Spec.Port = 443
				input.Spec.Stickiness = &AwsLbTargetGroupStickiness{Type: "source_ip_dest_ip"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for source_ip_dest_ip_proto stickiness", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "UDP"
				input.Spec.Stickiness = &AwsLbTargetGroupStickiness{Type: "source_ip_dest_ip_proto"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an ALB group with a target control port", func() {
				input := minimalValidTargetGroup()
				input.Spec.TargetControlPort = 9999
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_lb_target_group", func() {

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidTargetGroup()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidTargetGroup()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a non-lambda group has no port or protocol", func() {
				input := minimalValidTargetGroup()
				input.Spec.Port = 0
				input.Spec.Protocol = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a lambda group sets a port", func() {
				input := minimalValidTargetGroup()
				input.Spec.TargetType = "lambda"
				input.Spec.Protocol = ""
				input.Spec.Port = 8080
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range port", func() {
				input := minimalValidTargetGroup()
				input.Spec.Port = 70000
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid protocol", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "GENEVE"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid target type", func() {
				input := minimalValidTargetGroup()
				input.Spec.TargetType = "container"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for protocol_version on a TCP group", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "TCP"
				input.Spec.ProtocolVersion = "HTTP2"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid ip_address_type", func() {
				input := minimalValidTargetGroup()
				input.Spec.IpAddressType = "dualstack"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a deregistration delay above 3600", func() {
				input := minimalValidTargetGroup()
				input.Spec.DeregistrationDelaySeconds = 3601
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a slow start below 30 seconds", func() {
				input := minimalValidTargetGroup()
				input.Spec.SlowStartSeconds = 20
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for slow start on an NLB protocol", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "TCP"
				input.Spec.SlowStartSeconds = 60
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for slow start combined with stickiness", func() {
				input := minimalValidTargetGroup()
				input.Spec.SlowStartSeconds = 60
				input.Spec.Stickiness = &AwsLbTargetGroupStickiness{Type: "lb_cookie"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for slow start with least_outstanding_requests", func() {
				input := minimalValidTargetGroup()
				input.Spec.SlowStartSeconds = 60
				input.Spec.LoadBalancingAlgorithmType = "least_outstanding_requests"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for anomaly mitigation without weighted_random", func() {
				input := minimalValidTargetGroup()
				input.Spec.LoadBalancingAnomalyMitigation = "on"
				input.Spec.LoadBalancingAlgorithmType = "round_robin"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid cross-zone value", func() {
				input := minimalValidTargetGroup()
				input.Spec.LoadBalancingCrossZoneEnabled = "maybe"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for proxy_protocol_v2 on an HTTP group", func() {
				input := minimalValidTargetGroup()
				input.Spec.ProxyProtocolV2 = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for lambda_multi_value_headers on a non-lambda group", func() {
				input := minimalValidTargetGroup()
				input.Spec.LambdaMultiValueHeadersEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a health check path on a TCP probe", func() {
				input := minimalValidTargetGroup()
				input.Spec.HealthCheck = &AwsLbTargetGroupHealthCheck{
					Protocol: "TCP",
					Path:     "/healthz",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a health check matcher on a TCP probe", func() {
				input := minimalValidTargetGroup()
				input.Spec.HealthCheck = &AwsLbTargetGroupHealthCheck{
					Protocol: "TCP",
					Matcher:  "200",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a health check timeout not below the interval", func() {
				input := minimalValidTargetGroup()
				input.Spec.HealthCheck = &AwsLbTargetGroupHealthCheck{
					IntervalSeconds: 10,
					TimeoutSeconds:  10,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range healthy threshold", func() {
				input := minimalValidTargetGroup()
				input.Spec.HealthCheck = &AwsLbTargetGroupHealthCheck{
					HealthyThreshold: 11,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid stickiness type", func() {
				input := minimalValidTargetGroup()
				input.Spec.Stickiness = &AwsLbTargetGroupStickiness{Type: "sticky"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range target control port", func() {
				input := minimalValidTargetGroup()
				input.Spec.TargetControlPort = 70000
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a target control port on an NLB protocol", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "TCP"
				input.Spec.TargetControlPort = 9999
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for app_cookie stickiness without a cookie name", func() {
				input := minimalValidTargetGroup()
				input.Spec.Stickiness = &AwsLbTargetGroupStickiness{Type: "app_cookie"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a cookie name on lb_cookie stickiness", func() {
				input := minimalValidTargetGroup()
				input.Spec.Stickiness = &AwsLbTargetGroupStickiness{
					Type:       "lb_cookie",
					CookieName: "SESSIONID",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for cookie duration on source_ip stickiness", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "TCP"
				input.Spec.Stickiness = &AwsLbTargetGroupStickiness{
					Type:                  "source_ip",
					CookieDurationSeconds: 3600,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range draining interval", func() {
				input := minimalValidTargetGroup()
				input.Spec.Protocol = "TCP"
				input.Spec.TargetHealthState = &AwsLbTargetGroupTargetHealthState{
					UnhealthyDrainingIntervalSeconds: 360001,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a static target without a target id", func() {
				input := minimalValidTargetGroup()
				input.Spec.Targets = []*AwsLbTargetGroupTarget{{Port: 8080}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a static target with an out-of-range port", func() {
				input := minimalValidTargetGroup()
				input.Spec.Targets = []*AwsLbTargetGroupTarget{
					{TargetId: literalRef("i-0123456789abcdef0"), Port: 70000},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
