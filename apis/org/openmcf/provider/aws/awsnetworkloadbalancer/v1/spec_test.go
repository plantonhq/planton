package awsnetworkloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsNetworkLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsNetworkLoadBalancerSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// minimalListener returns a minimal valid TCP listener with target group.
func minimalListener(name string, port int32) *AwsNetworkLoadBalancerListener {
	return &AwsNetworkLoadBalancerListener{
		Name:     name,
		Port:     port,
		Protocol: "TCP",
		TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
			Port:     port,
			Protocol: "TCP",
		},
	}
}

// minimalSubnetMapping returns a minimal valid subnet mapping.
func minimalSubnetMapping(subnetId string) *AwsNetworkLoadBalancerSubnetMapping {
	return &AwsNetworkLoadBalancerSubnetMapping{
		SubnetId: strRef(subnetId),
	}
}

var _ = ginkgo.Describe("AwsNetworkLoadBalancerSpec validations", func() {
	var spec *AwsNetworkLoadBalancerSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: one subnet, one TCP listener.
		spec = &AwsNetworkLoadBalancerSpec{
			Region: "us-west-2",
			SubnetMappings: []*AwsNetworkLoadBalancerSubnetMapping{
				minimalSubnetMapping("subnet-abc123"),
			},
			Listeners: []*AwsNetworkLoadBalancerListener{
				minimalListener("tcp-80", 80),
			},
		}
	})

	// =========================================================================
	// Happy path — Spec level
	// =========================================================================

	ginkgo.It("accepts a minimal valid spec (one subnet, one TCP listener)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multiple subnet mappings across AZs", func() {
		spec.SubnetMappings = []*AwsNetworkLoadBalancerSubnetMapping{
			minimalSubnetMapping("subnet-az1"),
			minimalSubnetMapping("subnet-az2"),
			minimalSubnetMapping("subnet-az3"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts subnet mapping with Elastic IP allocation", func() {
		spec.SubnetMappings = []*AwsNetworkLoadBalancerSubnetMapping{
			{
				SubnetId:     strRef("subnet-public1"),
				AllocationId: strRef("eipalloc-abc123"),
			},
			{
				SubnetId:     strRef("subnet-public2"),
				AllocationId: strRef("eipalloc-def456"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts internal NLB with private IP addresses", func() {
		spec.Internal = true
		spec.SubnetMappings = []*AwsNetworkLoadBalancerSubnetMapping{
			{
				SubnetId:           strRef("subnet-private1"),
				PrivateIpv4Address: "10.0.1.100",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts security groups", func() {
		spec.SecurityGroups = []*foreignkeyv1.StringValueOrRef{
			strRef("sg-abc123"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts cross-zone load balancing enabled", func() {
		spec.CrossZoneLoadBalancingEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts dualstack IP address type", func() {
		spec.IpAddressType = "dualstack"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts ipv4 IP address type", func() {
		spec.IpAddressType = "ipv4"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts availability_zone_affinity DNS routing policy", func() {
		spec.DnsRecordClientRoutingPolicy = "availability_zone_affinity"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts partial_availability_zone_affinity DNS routing policy", func() {
		spec.DnsRecordClientRoutingPolicy = "partial_availability_zone_affinity"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts delete protection enabled", func() {
		spec.DeleteProtectionEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Listener configurations
	// =========================================================================

	ginkgo.It("accepts a TLS listener with certificate", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name:     "tls-443",
				Port:     443,
				Protocol: "TLS",
				Tls: &AwsNetworkLoadBalancerTlsConfig{
					CertificateArn: strRef("arn:aws:acm:us-east-1:123456789012:certificate/abc-123"),
					SslPolicy:      "ELBSecurityPolicy-TLS13-1-2-2021-06",
				},
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     8080,
					Protocol: "TCP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a TLS listener with ALPN policy", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name:     "tls-443",
				Port:     443,
				Protocol: "TLS",
				Tls: &AwsNetworkLoadBalancerTlsConfig{
					CertificateArn: strRef("arn:aws:acm:us-east-1:123456789012:certificate/abc-123"),
				},
				AlpnPolicy: "HTTP2Preferred",
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     8080,
					Protocol: "TCP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a UDP listener", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name:     "udp-53",
				Port:     53,
				Protocol: "UDP",
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     53,
					Protocol: "UDP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a TCP_UDP listener", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name:     "tcpudp-53",
				Port:     53,
				Protocol: "TCP_UDP",
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     53,
					Protocol: "TCP_UDP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a TCP listener with idle timeout", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name:                  "tcp-80",
				Port:                  80,
				Protocol:              "TCP",
				TcpIdleTimeoutSeconds: 600,
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     80,
					Protocol: "TCP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multiple listeners", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			minimalListener("tcp-80", 80),
			{
				Name:     "tls-443",
				Port:     443,
				Protocol: "TLS",
				Tls: &AwsNetworkLoadBalancerTlsConfig{
					CertificateArn: strRef("arn:aws:acm:us-east-1:123456789012:certificate/abc-123"),
				},
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     8080,
					Protocol: "TCP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Target group configurations
	// =========================================================================

	ginkgo.It("accepts target group with IP target type", func() {
		spec.Listeners[0].TargetGroup.TargetType = "ip"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts target group with ALB target type", func() {
		spec.Listeners[0].TargetGroup.TargetType = "alb"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts target group with all options", func() {
		spec.Listeners[0].TargetGroup = &AwsNetworkLoadBalancerTargetGroup{
			Port:                       8080,
			Protocol:                   "TCP",
			TargetType:                 "ip",
			DeregistrationDelaySeconds: 60,
			PreserveClientIp:           true,
			ProxyProtocolV2:            true,
			ConnectionTermination:      true,
			StickinessEnabled:          true,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts target group with HTTP health check", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:           "HTTP",
			Port:               "traffic-port",
			Path:               "/health",
			HealthyThreshold:   5,
			UnhealthyThreshold: 3,
			IntervalSeconds:    30,
			TimeoutSeconds:     10,
			Matcher:            "200-299",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts target group with TCP health check", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:         "TCP",
			Port:             "8080",
			HealthyThreshold: 3,
			IntervalSeconds:  10,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts target group with HTTPS health check", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:         "HTTPS",
			Path:             "/healthz",
			HealthyThreshold: 2,
			IntervalSeconds:  10,
			Matcher:          "200",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — DNS
	// =========================================================================

	ginkgo.It("accepts DNS configuration", func() {
		spec.Dns = &AwsNetworkLoadBalancerDns{
			Enabled:       true,
			Route53ZoneId: strRef("Z0123456789ABCDEFGHIJ"),
			Hostnames:     []string{"api.example.com", "app.example.com"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Production-ready configuration
	// =========================================================================

	ginkgo.It("accepts a production-ready configuration", func() {
		spec = &AwsNetworkLoadBalancerSpec{
			Region: "us-west-2",
			SubnetMappings: []*AwsNetworkLoadBalancerSubnetMapping{
				{
					SubnetId:     strRef("subnet-public1"),
					AllocationId: strRef("eipalloc-abc123"),
				},
				{
					SubnetId:     strRef("subnet-public2"),
					AllocationId: strRef("eipalloc-def456"),
				},
			},
			SecurityGroups:                []*foreignkeyv1.StringValueOrRef{strRef("sg-nlb-prod")},
			DeleteProtectionEnabled:       true,
			CrossZoneLoadBalancingEnabled: true,
			IpAddressType:                 "ipv4",
			Listeners: []*AwsNetworkLoadBalancerListener{
				{
					Name:     "tls-443",
					Port:     443,
					Protocol: "TLS",
					Tls: &AwsNetworkLoadBalancerTlsConfig{
						CertificateArn: strRef("arn:aws:acm:us-east-1:123456789012:certificate/prod-cert"),
						SslPolicy:      "ELBSecurityPolicy-TLS13-1-2-2021-06",
					},
					AlpnPolicy: "HTTP2Preferred",
					TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
						Port:                       8443,
						Protocol:                   "TCP",
						TargetType:                 "ip",
						DeregistrationDelaySeconds: 60,
						PreserveClientIp:           true,
						ConnectionTermination:      true,
						HealthCheck: &AwsNetworkLoadBalancerHealthCheck{
							Protocol:           "HTTPS",
							Path:               "/healthz",
							HealthyThreshold:   3,
							UnhealthyThreshold: 3,
							IntervalSeconds:    10,
							TimeoutSeconds:     6,
							Matcher:            "200",
						},
					},
				},
			},
			Dns: &AwsNetworkLoadBalancerDns{
				Enabled:       true,
				Route53ZoneId: strRef("Z0123456789ABCDEFGHIJ"),
				Hostnames:     []string{"api.example.com"},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Spec level
	// =========================================================================

	ginkgo.It("rejects empty subnet_mappings", func() {
		spec.SubnetMappings = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("subnet_mappings"))
	})

	ginkgo.It("rejects empty listeners", func() {
		spec.Listeners = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("listeners"))
	})

	ginkgo.It("rejects invalid ip_address_type", func() {
		spec.IpAddressType = "triplestack"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("ip_address_type"))
	})

	ginkgo.It("rejects invalid dns_record_client_routing_policy", func() {
		spec.DnsRecordClientRoutingPolicy = "random_policy"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("dns_record_client_routing_policy"))
	})

	// =========================================================================
	// Failure — Subnet mapping
	// =========================================================================

	ginkgo.It("rejects subnet mapping without subnet_id", func() {
		spec.SubnetMappings = []*AwsNetworkLoadBalancerSubnetMapping{
			{}, // missing subnet_id
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("subnet_id"))
	})

	// =========================================================================
	// Failure — Listener validations
	// =========================================================================

	ginkgo.It("rejects listener without name", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Port:     80,
				Protocol: "TCP",
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     80,
					Protocol: "TCP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("name"))
	})

	ginkgo.It("rejects listener with invalid name format (uppercase)", func() {
		spec.Listeners[0].Name = "TCP-80"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("name"))
	})

	ginkgo.It("rejects listener with invalid name format (starts with number)", func() {
		spec.Listeners[0].Name = "80-tcp"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("name"))
	})

	ginkgo.It("rejects listener without port", func() {
		spec.Listeners[0].Port = 0
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("port"))
	})

	ginkgo.It("rejects listener with port above 65535", func() {
		spec.Listeners[0].Port = 70000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("port"))
	})

	ginkgo.It("rejects listener without protocol", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name: "tcp-80",
				Port: 80,
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     80,
					Protocol: "TCP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("protocol"))
	})

	ginkgo.It("rejects listener with invalid protocol", func() {
		spec.Listeners[0].Protocol = "HTTP"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("protocol"))
	})

	ginkgo.It("rejects TLS listener without tls configuration", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name:     "tls-443",
				Port:     443,
				Protocol: "TLS",
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     8080,
					Protocol: "TCP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("tls"))
	})

	ginkgo.It("rejects tcp_idle_timeout on non-TCP listener", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name:                  "udp-53",
				Port:                  53,
				Protocol:              "UDP",
				TcpIdleTimeoutSeconds: 120,
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     53,
					Protocol: "UDP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("tcp_idle_timeout"))
	})

	ginkgo.It("rejects tcp_idle_timeout below 60", func() {
		spec.Listeners[0].TcpIdleTimeoutSeconds = 30
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("tcp_idle_timeout"))
	})

	ginkgo.It("rejects tcp_idle_timeout above 6000", func() {
		spec.Listeners[0].TcpIdleTimeoutSeconds = 7000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("tcp_idle_timeout"))
	})

	ginkgo.It("rejects alpn_policy on non-TLS listener", func() {
		spec.Listeners[0].AlpnPolicy = "HTTP2Preferred"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("alpn_policy"))
	})

	ginkgo.It("rejects invalid alpn_policy", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name:     "tls-443",
				Port:     443,
				Protocol: "TLS",
				Tls: &AwsNetworkLoadBalancerTlsConfig{
					CertificateArn: strRef("arn:aws:acm:us-east-1:123456789012:certificate/abc"),
				},
				AlpnPolicy: "InvalidPolicy",
				TargetGroup: &AwsNetworkLoadBalancerTargetGroup{
					Port:     8080,
					Protocol: "TCP",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("alpn_policy"))
	})

	ginkgo.It("rejects listener without target_group", func() {
		spec.Listeners = []*AwsNetworkLoadBalancerListener{
			{
				Name:     "tcp-80",
				Port:     80,
				Protocol: "TCP",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("target_group"))
	})

	// =========================================================================
	// Failure — Target group validations
	// =========================================================================

	ginkgo.It("rejects target group without port", func() {
		spec.Listeners[0].TargetGroup.Port = 0
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("port"))
	})

	ginkgo.It("rejects target group with port above 65535", func() {
		spec.Listeners[0].TargetGroup.Port = 70000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("port"))
	})

	ginkgo.It("rejects target group without protocol", func() {
		spec.Listeners[0].TargetGroup.Protocol = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("protocol"))
	})

	ginkgo.It("rejects target group with invalid protocol", func() {
		spec.Listeners[0].TargetGroup.Protocol = "HTTPS"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("protocol"))
	})

	ginkgo.It("rejects target group with invalid target_type", func() {
		spec.Listeners[0].TargetGroup.TargetType = "lambda"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("target_type"))
	})

	// =========================================================================
	// Failure — Health check validations
	// =========================================================================

	ginkgo.It("rejects health check with invalid protocol", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol: "gRPC",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("protocol"))
	})

	ginkgo.It("rejects HTTP health check without path", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol: "HTTP",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("path"))
	})

	ginkgo.It("rejects matcher on TCP health check", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol: "TCP",
			Matcher:  "200",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("matcher"))
	})

	ginkgo.It("rejects healthy_threshold below 2", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:         "TCP",
			HealthyThreshold: 1,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("healthy_threshold"))
	})

	ginkgo.It("rejects healthy_threshold above 10", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:         "TCP",
			HealthyThreshold: 11,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("healthy_threshold"))
	})

	ginkgo.It("rejects unhealthy_threshold above 10", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:           "TCP",
			UnhealthyThreshold: 15,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("unhealthy_threshold"))
	})

	ginkgo.It("rejects interval_seconds below 5", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:        "TCP",
			IntervalSeconds: 2,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("interval_seconds"))
	})

	ginkgo.It("rejects interval_seconds above 300", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:        "TCP",
			IntervalSeconds: 500,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("interval_seconds"))
	})

	ginkgo.It("rejects timeout_seconds below 2", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:       "TCP",
			TimeoutSeconds: 1,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("timeout_seconds"))
	})

	ginkgo.It("rejects timeout_seconds above 120", func() {
		spec.Listeners[0].TargetGroup.HealthCheck = &AwsNetworkLoadBalancerHealthCheck{
			Protocol:       "TCP",
			TimeoutSeconds: 200,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("timeout_seconds"))
	})
})
