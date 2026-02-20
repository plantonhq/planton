package alicloudnetworkloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAlicloudNetworkLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudNetworkLoadBalancerSpec Validation Tests")
}

func strRef(s string) *fkv1.StringValueOrRef {
	return &fkv1.StringValueOrRef{
		LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: s},
	}
}

func minimalValidSpec() *AlicloudNetworkLoadBalancerSpec {
	return &AlicloudNetworkLoadBalancerSpec{
		Region: "cn-hangzhou",
		VpcId:  strRef("vpc-abc123"),
		ZoneMappings: []*AlicloudNetworkLoadBalancerZoneMapping{
			{ZoneId: "cn-hangzhou-a", VswitchId: strRef("vsw-aaa")},
			{ZoneId: "cn-hangzhou-b", VswitchId: strRef("vsw-bbb")},
		},
	}
}

func minimalValidInput() *AlicloudNetworkLoadBalancer {
	return &AlicloudNetworkLoadBalancer{
		ApiVersion: "alicloud.openmcf.org/v1",
		Kind:       "AlicloudNetworkLoadBalancer",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-nlb"},
		Spec:       minimalValidSpec(),
	}
}

var _ = ginkgo.Describe("AlicloudNetworkLoadBalancerSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			err := protovalidate.Validate(minimalValidInput())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all top-level optional fields populated", func() {
			input := minimalValidInput()
			input.Spec.LoadBalancerName = "prod-nlb"
			input.Spec.AddressType = proto.String("Internet")
			input.Spec.ResourceGroupId = "rg-abc123"
			input.Spec.CrossZoneEnabled = proto.Bool(true)
			input.Spec.Tags = map[string]string{"team": "platform", "env": "prod"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Intranet address type", func() {
			input := minimalValidInput()
			input.Spec.AddressType = proto.String("Intranet")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with cross-zone disabled", func() {
			input := minimalValidInput()
			input.Spec.CrossZoneEnabled = proto.Bool(false)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with zone mapping allocation_id for EIP binding", func() {
			input := minimalValidInput()
			input.Spec.ZoneMappings[0].AllocationId = strRef("eip-abc123")
			input.Spec.ZoneMappings[1].AllocationId = strRef("eip-def456")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with TCP server group and listener", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name: "tcp-backend",
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			input.Spec.Listeners = []*AlicloudNetworkLoadBalancerListener{
				{
					ListenerPort:     80,
					ListenerProtocol: "TCP",
					ServerGroupName:  "tcp-backend",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with UDP server group and listener", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name:     "udp-backend",
					Protocol: proto.String("UDP"),
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
						HealthCheckType:    proto.String("UDP"),
					},
				},
			}
			input.Spec.Listeners = []*AlicloudNetworkLoadBalancerListener{
				{
					ListenerPort:     53,
					ListenerProtocol: "UDP",
					ServerGroupName:  "udp-backend",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with TCPSSL listener and certificates", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name:     "ssl-backend",
					Protocol: proto.String("TCPSSL"),
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			input.Spec.Listeners = []*AlicloudNetworkLoadBalancerListener{
				{
					ListenerPort:     443,
					ListenerProtocol: "TCPSSL",
					ServerGroupName:  "ssl-backend",
					CertificateIds:   []string{"cas-abc123"},
					SecurityPolicyId: "tls_cipher_policy_1_2_strict",
					CaCertificateIds: []string{"ca-abc123"},
					CaEnabled:        proto.Bool(true),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with connection drain enabled", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name:                   "drain-backend",
					ConnectionDrainEnabled: proto.Bool(true),
					ConnectionDrainTimeout: proto.Int32(300),
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all scheduler values", func() {
			schedulers := []string{"Wrr", "Rr", "Sch", "Tch", "Qch", "Wlc"}
			for _, sched := range schedulers {
				input := minimalValidInput()
				input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
					{
						Name:      "sched-backend",
						Scheduler: proto.String(sched),
						HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
							HealthCheckEnabled: false,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should pass with all health check fields populated for HTTP type", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name: "full-hc",
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled:        true,
						HealthCheckType:           proto.String("HTTP"),
						HealthCheckConnectPort:    proto.Int32(8080),
						HealthCheckConnectTimeout: proto.Int32(10),
						HealthCheckInterval:       proto.Int32(15),
						HealthyThreshold:          proto.Int32(5),
						UnhealthyThreshold:        proto.Int32(3),
						HealthCheckUrl:            "/healthz",
						HealthCheckDomain:         "backend.internal",
						HttpCheckMethod:           proto.String("GET"),
						HealthCheckHttpCodes:      []string{"http_2xx", "http_3xx"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with proxy protocol enabled", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name: "proxy-backend",
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			input.Spec.Listeners = []*AlicloudNetworkLoadBalancerListener{
				{
					ListenerPort:         80,
					ListenerProtocol:     "TCP",
					ServerGroupName:      "proxy-backend",
					ProxyProtocolEnabled: proto.Bool(true),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with custom idle timeout", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name: "timeout-backend",
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			input.Spec.Listeners = []*AlicloudNetworkLoadBalancerListener{
				{
					ListenerPort:     80,
					ListenerProtocol: "TCP",
					ServerGroupName:  "timeout-backend",
					IdleTimeout:      proto.Int32(300),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with preserve_client_ip disabled", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name:                    "no-preserve",
					PreserveClientIpEnabled: proto.Bool(false),
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
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
			input := &AlicloudNetworkLoadBalancer{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNetworkLoadBalancer",
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
			input.Spec.ZoneMappings = []*AlicloudNetworkLoadBalancerZoneMapping{
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

		ginkgo.It("should fail when server group name is too short", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name: "x",
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when server group health check is missing", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{Name: "no-hc"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when server group protocol is invalid", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name:     "bad-proto",
					Protocol: proto.String("HTTP"),
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when scheduler is invalid", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name:      "bad-sched",
					Scheduler: proto.String("Random"),
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when connection_drain_timeout is below minimum", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name:                   "bad-drain",
					ConnectionDrainTimeout: proto.Int32(5),
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when connection_drain_timeout exceeds maximum", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name:                   "bad-drain-max",
					ConnectionDrainTimeout: proto.Int32(1000),
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check type is invalid", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name: "bad-hc-type",
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
						HealthCheckType:    proto.String("GRPC"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check interval is below minimum", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name: "bad-interval",
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled:  true,
						HealthCheckInterval: proto.Int32(2),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when healthy_threshold is out of range", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name: "bad-threshold",
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
						HealthyThreshold:   proto.Int32(15),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when http_check_method is invalid", func() {
			input := minimalValidInput()
			input.Spec.ServerGroups = []*AlicloudNetworkLoadBalancerServerGroup{
				{
					Name: "bad-method",
					HealthCheck: &AlicloudNetworkLoadBalancerHealthCheckConfig{
						HealthCheckEnabled: true,
						HttpCheckMethod:    proto.String("POST"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when listener protocol is invalid", func() {
			input := minimalValidInput()
			input.Spec.Listeners = []*AlicloudNetworkLoadBalancerListener{
				{
					ListenerPort:     80,
					ListenerProtocol: "HTTP",
					ServerGroupName:  "backend",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when listener server_group_name is empty", func() {
			input := minimalValidInput()
			input.Spec.Listeners = []*AlicloudNetworkLoadBalancerListener{
				{
					ListenerPort:     80,
					ListenerProtocol: "TCP",
					ServerGroupName:  "",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when idle_timeout exceeds maximum", func() {
			input := minimalValidInput()
			input.Spec.Listeners = []*AlicloudNetworkLoadBalancerListener{
				{
					ListenerPort:     80,
					ListenerProtocol: "TCP",
					ServerGroupName:  "backend",
					IdleTimeout:      proto.Int32(1000),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when listener port exceeds maximum", func() {
			input := minimalValidInput()
			input.Spec.Listeners = []*AlicloudNetworkLoadBalancerListener{
				{
					ListenerPort:     70000,
					ListenerProtocol: "TCP",
					ServerGroupName:  "backend",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when load_balancer_name is 1 character", func() {
			input := minimalValidInput()
			input.Spec.LoadBalancerName = "x"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
