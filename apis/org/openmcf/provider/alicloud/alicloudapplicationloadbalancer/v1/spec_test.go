package alicloudapplicationloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAlicloudApplicationLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudApplicationLoadBalancerSpec Validation Tests")
}

func strRef(s string) *fkv1.StringValueOrRef {
	return &fkv1.StringValueOrRef{
		LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: s},
	}
}

func minimalValidSpec() *AlicloudApplicationLoadBalancerSpec {
	return &AlicloudApplicationLoadBalancerSpec{
		Region: "cn-hangzhou",
		VpcId:  strRef("vpc-abc123"),
		ZoneMappings: []*AlicloudApplicationLoadBalancerZoneMapping{
			{ZoneId: "cn-hangzhou-a", VswitchId: strRef("vsw-aaa")},
			{ZoneId: "cn-hangzhou-b", VswitchId: strRef("vsw-bbb")},
		},
	}
}

func minimalValidInput() *AlicloudApplicationLoadBalancer {
	return &AlicloudApplicationLoadBalancer{
		ApiVersion: "alicloud.openmcf.org/v1",
		Kind:       "AlicloudApplicationLoadBalancer",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-alb"},
		Spec:       minimalValidSpec(),
	}
}

var _ = ginkgo.Describe("AlicloudApplicationLoadBalancerSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			err := protovalidate.Validate(minimalValidInput())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all top-level optional fields populated", func() {
			input := minimalValidInput()
			input.Spec.LoadBalancerName = "prod-alb"
			input.Spec.AddressType = proto.String("Internet")
			input.Spec.LoadBalancerEdition = proto.String("Standard")
			input.Spec.ResourceGroupId = "rg-abc123"
			input.Spec.Tags = map[string]string{"team": "platform", "env": "prod"}
			input.Spec.AccessLogConfig = &AlicloudApplicationLoadBalancerAccessLogConfig{
				LogProject: "my-sls-project",
				LogStore:   "alb-access-log",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Intranet address type", func() {
			input := minimalValidInput()
			input.Spec.AddressType = proto.String("Intranet")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Basic edition", func() {
			input := minimalValidInput()
			input.Spec.LoadBalancerEdition = proto.String("Basic")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with StandardWithWaf edition", func() {
			input := minimalValidInput()
			input.Spec.LoadBalancerEdition = proto.String("StandardWithWaf")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with server groups and listeners", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name: "web-backend",
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			input.Spec.Listeners = []*AlicloudApplicationLoadBalancerListener{
				{
					ListenerPort:                 80,
					ListenerProtocol:             "HTTP",
					DefaultActionServerGroupName: "web-backend",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with HTTPS listener and certificate", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name:     "api-backend",
					Protocol: proto.String("HTTPS"),
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled:  true,
						HealthCheckProtocol: proto.String("HTTPS"),
						HealthCheckPath:     "/health",
					},
				},
			}
			input.Spec.Listeners = []*AlicloudApplicationLoadBalancerListener{
				{
					ListenerPort:                 443,
					ListenerProtocol:             "HTTPS",
					DefaultActionServerGroupName: "api-backend",
					CertificateId:               "cas-abc123",
					SecurityPolicyId:             "tls_cipher_policy_1_2_strict",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with GRPC server group", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name:     "grpc-backend",
					Protocol: proto.String("GRPC"),
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled:  true,
						HealthCheckProtocol: proto.String("GRPC"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with sticky session config", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name: "sticky-backend",
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
					StickySessionConfig: &AlicloudApplicationLoadBalancerStickySessionConfig{
						StickySessionEnabled: true,
						StickySessionType:    proto.String("Insert"),
						CookieTimeout:        proto.Int32(3600),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all health check fields populated", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name: "full-hc",
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled:     true,
						HealthCheckProtocol:    proto.String("HTTP"),
						HealthCheckPath:        "/healthz",
						HealthCheckHost:        "backend.internal",
						HealthCheckMethod:      proto.String("GET"),
						HealthCheckConnectPort: proto.Int32(8080),
						HealthCheckInterval:    proto.Int32(5),
						HealthCheckTimeout:     proto.Int32(10),
						HealthyThreshold:       proto.Int32(5),
						UnhealthyThreshold:     proto.Int32(2),
						HealthCheckCodes:       []string{"http_2xx", "http_3xx"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Wlc scheduler", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name:      "wlc-backend",
					Scheduler: proto.String("Wlc"),
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: false,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with listener timeouts", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name: "timeout-backend",
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			input.Spec.Listeners = []*AlicloudApplicationLoadBalancerListener{
				{
					ListenerPort:                 80,
					ListenerProtocol:             "HTTP",
					DefaultActionServerGroupName: "timeout-backend",
					GzipEnabled:                  proto.Bool(false),
					IdleTimeout:                  proto.Int32(30),
					RequestTimeout:               proto.Int32(120),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := minimalValidInput()
			input.ApiVersion = "wrong/v1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := minimalValidInput()
			input.Kind = "WrongKind"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := minimalValidInput()
			input.Metadata = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AlicloudApplicationLoadBalancer{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudApplicationLoadBalancer",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := minimalValidInput()
			input.Spec.Region = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_id is missing", func() {
			input := minimalValidInput()
			input.Spec.VpcId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when zone_mappings has fewer than 2 entries", func() {
			input := minimalValidInput()
			input.Spec.ZoneMappings = []*AlicloudApplicationLoadBalancerZoneMapping{
				{ZoneId: "cn-hangzhou-a", VswitchId: strRef("vsw-aaa")},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when zone_mappings is empty", func() {
			input := minimalValidInput()
			input.Spec.ZoneMappings = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when zone_mapping zone_id is empty", func() {
			input := minimalValidInput()
			input.Spec.ZoneMappings[0].ZoneId = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when zone_mapping vswitch_id is missing", func() {
			input := minimalValidInput()
			input.Spec.ZoneMappings[0].VswitchId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when address_type is invalid", func() {
			input := minimalValidInput()
			input.Spec.AddressType = proto.String("External")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when load_balancer_edition is invalid", func() {
			input := minimalValidInput()
			input.Spec.LoadBalancerEdition = proto.String("Enterprise")
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when server group name is too short", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name: "x",
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when server group health check is missing", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{Name: "no-hc"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when server group protocol is invalid", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name:     "bad-proto",
					Protocol: proto.String("TCP"),
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when scheduler is invalid", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name:      "bad-sched",
					Scheduler: proto.String("Random"),
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check protocol is invalid", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name: "bad-hc-proto",
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled:  true,
						HealthCheckProtocol: proto.String("FTP"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when healthy_threshold is out of range", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name: "bad-threshold",
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
						HealthyThreshold:   proto.Int32(15),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when listener port is out of range", func() {
			input := minimalValidInput()
			input.Spec.Listeners = []*AlicloudApplicationLoadBalancerListener{
				{
					ListenerPort:                 0,
					ListenerProtocol:             "HTTP",
					DefaultActionServerGroupName: "backend",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when listener protocol is invalid", func() {
			input := minimalValidInput()
			input.Spec.Listeners = []*AlicloudApplicationLoadBalancerListener{
				{
					ListenerPort:                 80,
					ListenerProtocol:             "TCP",
					DefaultActionServerGroupName: "backend",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when listener default_action_server_group_name is empty", func() {
			input := minimalValidInput()
			input.Spec.Listeners = []*AlicloudApplicationLoadBalancerListener{
				{
					ListenerPort:                 80,
					ListenerProtocol:             "HTTP",
					DefaultActionServerGroupName: "",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when idle_timeout is out of range", func() {
			input := minimalValidInput()
			input.Spec.Listeners = []*AlicloudApplicationLoadBalancerListener{
				{
					ListenerPort:                 80,
					ListenerProtocol:             "HTTP",
					DefaultActionServerGroupName: "backend",
					IdleTimeout:                  proto.Int32(120),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when request_timeout is out of range", func() {
			input := minimalValidInput()
			input.Spec.Listeners = []*AlicloudApplicationLoadBalancerListener{
				{
					ListenerPort:                 80,
					ListenerProtocol:             "HTTP",
					DefaultActionServerGroupName: "backend",
					RequestTimeout:               proto.Int32(300),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cookie_timeout is out of range", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name: "bad-cookie",
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
					StickySessionConfig: &AlicloudApplicationLoadBalancerStickySessionConfig{
						StickySessionEnabled: true,
						StickySessionType:    proto.String("Insert"),
						CookieTimeout:        proto.Int32(100000),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when sticky_session_type is invalid", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudApplicationLoadBalancerServerGroup{
				{
					Name: "bad-sticky",
					HealthCheckConfig: &AlicloudApplicationLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
					StickySessionConfig: &AlicloudApplicationLoadBalancerStickySessionConfig{
						StickySessionEnabled: true,
						StickySessionType:    proto.String("Custom"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when access_log_config log_project is empty", func() {
			input := minimalValidInput()
			input.Spec.AccessLogConfig = &AlicloudApplicationLoadBalancerAccessLogConfig{
				LogProject: "",
				LogStore:   "alb-log",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when access_log_config log_store is empty", func() {
			input := minimalValidInput()
			input.Spec.AccessLogConfig = &AlicloudApplicationLoadBalancerAccessLogConfig{
				LogProject: "my-project",
				LogStore:   "",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
