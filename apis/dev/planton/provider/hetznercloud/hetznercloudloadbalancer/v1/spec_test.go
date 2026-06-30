package hetznercloudloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestHetznerCloudLoadBalancerSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudLoadBalancerSpec Validation Suite")
}

func strRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

var _ = Describe("HetznerCloudLoadBalancerSpec validations", func() {

	Context("with valid specs", func() {
		It("should accept a minimal spec with one HTTP service", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with an HTTPS service and certificates", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{
						Protocol:   HetznerCloudLoadBalancerSpec_https,
						ListenPort: proto.Int32(443),
						Http: &HetznerCloudLoadBalancerSpec_HttpConfig{
							CertificateIds: []*foreignkeyv1.StringValueOrRef{
								strRef("100"),
							},
							RedirectHttp: true,
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with a TCP service and explicit ports", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "nbg1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{
						Protocol:        HetznerCloudLoadBalancerSpec_tcp,
						ListenPort:      proto.Int32(5432),
						DestinationPort: proto.Int32(5432),
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with multiple services", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb21",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{
						Protocol:        HetznerCloudLoadBalancerSpec_https,
						ListenPort:      proto.Int32(443),
						DestinationPort: proto.Int32(8080),
						Http: &HetznerCloudLoadBalancerSpec_HttpConfig{
							CertificateIds: []*foreignkeyv1.StringValueOrRef{strRef("100")},
						},
					},
					{
						Protocol:   HetznerCloudLoadBalancerSpec_http,
						ListenPort: proto.Int32(80),
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with server targets", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				ServerTargets: []*HetznerCloudLoadBalancerSpec_ServerTarget{
					{ServerId: strRef("111")},
					{ServerId: strRef("222"), UsePrivateIp: true},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with label selector targets", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				LabelSelectorTargets: []*HetznerCloudLoadBalancerSpec_LabelSelectorTarget{
					{Selector: "env=production,role=web", UsePrivateIp: true},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with IP targets", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				IpTargets: []*HetznerCloudLoadBalancerSpec_IpTarget{
					{Ip: "1.2.3.4"},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with mixed target types", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				ServerTargets: []*HetznerCloudLoadBalancerSpec_ServerTarget{
					{ServerId: strRef("111")},
				},
				LabelSelectorTargets: []*HetznerCloudLoadBalancerSpec_LabelSelectorTarget{
					{Selector: "role=worker"},
				},
				IpTargets: []*HetznerCloudLoadBalancerSpec_IpTarget{
					{Ip: "10.0.0.50"},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with a network attachment", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				Network: &HetznerCloudLoadBalancerSpec_NetworkAttachment{
					NetworkId: strRef("500"),
					Ip:        "10.0.1.100",
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a network attachment with enable_public_interface set to false", func() {
			enablePublic := false
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				Network: &HetznerCloudLoadBalancerSpec_NetworkAttachment{
					NetworkId:             strRef("500"),
					EnablePublicInterface: &enablePublic,
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with a health check", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{
						Protocol: HetznerCloudLoadBalancerSpec_http,
						HealthCheck: &HetznerCloudLoadBalancerSpec_HealthCheck{
							Protocol: HetznerCloudLoadBalancerSpec_http,
							Port:     proto.Int32(80),
							Interval: proto.Int32(15),
							Timeout:  proto.Int32(10),
							Retries:  proto.Int32(3),
							Http: &HetznerCloudLoadBalancerSpec_HealthCheckHttp{
								Path: "/health",
							},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a health check with minimal fields", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{
						Protocol: HetznerCloudLoadBalancerSpec_http,
						HealthCheck: &HetznerCloudLoadBalancerSpec_HealthCheck{
							Http: &HetznerCloudLoadBalancerSpec_HealthCheckHttp{
								Path: "/ready",
							},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with algorithm set to least_connections", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Algorithm:        HetznerCloudLoadBalancerSpec_least_connections,
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with HTTP sticky sessions", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{
						Protocol: HetznerCloudLoadBalancerSpec_http,
						Http: &HetznerCloudLoadBalancerSpec_HttpConfig{
							StickySessions: true,
							CookieName:     "MYSESSION",
							CookieLifetime: 600,
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a fully populated spec", func() {
			enablePublic := true
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb31",
				Location:         "fsn1",
				Algorithm:        HetznerCloudLoadBalancerSpec_least_connections,
				DeleteProtection: true,
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{
						Protocol:        HetznerCloudLoadBalancerSpec_https,
						ListenPort:      proto.Int32(443),
						DestinationPort: proto.Int32(8080),
						Http: &HetznerCloudLoadBalancerSpec_HttpConfig{
							CertificateIds: []*foreignkeyv1.StringValueOrRef{strRef("100"), strRef("200")},
							RedirectHttp:   true,
							StickySessions: true,
							CookieName:     "APPSESSION",
							CookieLifetime: 3600,
						},
						HealthCheck: &HetznerCloudLoadBalancerSpec_HealthCheck{
							Protocol: HetznerCloudLoadBalancerSpec_http,
							Port:     proto.Int32(8080),
							Interval: proto.Int32(10),
							Timeout:  proto.Int32(5),
							Retries:  proto.Int32(5),
							Http: &HetznerCloudLoadBalancerSpec_HealthCheckHttp{
								Path:        "/health",
								Domain:      "app.example.com",
								Response:    "ok",
								Tls:         true,
								StatusCodes: []string{"200", "204"},
							},
						},
					},
					{
						Protocol:        HetznerCloudLoadBalancerSpec_tcp,
						ListenPort:      proto.Int32(5432),
						DestinationPort: proto.Int32(5432),
						Proxyprotocol:   true,
					},
				},
				ServerTargets: []*HetznerCloudLoadBalancerSpec_ServerTarget{
					{ServerId: strRef("111"), UsePrivateIp: true},
					{ServerId: strRef("222"), UsePrivateIp: true},
				},
				LabelSelectorTargets: []*HetznerCloudLoadBalancerSpec_LabelSelectorTarget{
					{Selector: "role=web", UsePrivateIp: true},
				},
				IpTargets: []*HetznerCloudLoadBalancerSpec_IpTarget{
					{Ip: "203.0.113.50"},
				},
				Network: &HetznerCloudLoadBalancerSpec_NetworkAttachment{
					NetworkId:             strRef("500"),
					Ip:                    "10.0.1.100",
					EnablePublicInterface: &enablePublic,
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("with invalid specs", func() {
		It("should reject an empty load_balancer_type", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject an empty location", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a spec with no services", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services:         []*HetznerCloudLoadBalancerSpec_Service{},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a service with unspecified protocol", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_service_protocol_unspecified},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a TCP service without listen_port", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{
						Protocol:        HetznerCloudLoadBalancerSpec_tcp,
						DestinationPort: proto.Int32(5432),
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a TCP service without destination_port", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{
						Protocol:   HetznerCloudLoadBalancerSpec_tcp,
						ListenPort: proto.Int32(5432),
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a server target without server_id", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				ServerTargets: []*HetznerCloudLoadBalancerSpec_ServerTarget{
					{UsePrivateIp: true},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a label selector target with empty selector", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				LabelSelectorTargets: []*HetznerCloudLoadBalancerSpec_LabelSelectorTarget{
					{Selector: ""},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject an IP target with empty ip", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				IpTargets: []*HetznerCloudLoadBalancerSpec_IpTarget{
					{Ip: ""},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a network attachment without network_id", func() {
			spec := &HetznerCloudLoadBalancerSpec{
				LoadBalancerType: "lb11",
				Location:         "fsn1",
				Services: []*HetznerCloudLoadBalancerSpec_Service{
					{Protocol: HetznerCloudLoadBalancerSpec_http},
				},
				Network: &HetznerCloudLoadBalancerSpec_NetworkAttachment{
					Ip: "10.0.1.100",
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})
})
