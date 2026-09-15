package digitaloceanloadbalancerv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestDigitalOceanLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanLoadBalancerSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanLoadBalancerSpec validations", func() {

	val := func(s string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
		}
	}

	// Minimal valid regional HTTP balancer with tag-based targeting. VPC is
	// intentionally absent: it is optional and DigitalOcean falls back to the
	// region's default VPC.
	makeValidHTTPSpec := func() *DigitalOceanLoadBalancerSpec {
		return &DigitalOceanLoadBalancerSpec{
			LoadBalancerName: "prod-web-lb",
			Region:           digitalocean.DigitalOceanRegion_nyc3,
			ForwardingRules: []*DigitalOceanLoadBalancerForwardingRule{
				{
					EntryPort:      80,
					EntryProtocol:  DigitalOceanLoadBalancerProtocol_http,
					TargetPort:     80,
					TargetProtocol: DigitalOceanLoadBalancerProtocol_http,
				},
			},
			HealthCheck: &DigitalOceanLoadBalancerHealthCheck{
				Port:     80,
				Protocol: DigitalOceanLoadBalancerProtocol_http,
				Path:     "/healthz",
			},
			DropletTag: "web-prod",
		}
	}

	// HTTPS balancer terminating TLS with a certificate reference, placed in
	// an explicit VPC, with cookie-based sticky sessions.
	makeValidHTTPSSpec := func() *DigitalOceanLoadBalancerSpec {
		return &DigitalOceanLoadBalancerSpec{
			LoadBalancerName: "prod-https-lb",
			Region:           digitalocean.DigitalOceanRegion_sfo3,
			Vpc:              val("test-vpc-id"),
			ForwardingRules: []*DigitalOceanLoadBalancerForwardingRule{
				{
					EntryPort:       443,
					EntryProtocol:   DigitalOceanLoadBalancerProtocol_https,
					TargetPort:      80,
					TargetProtocol:  DigitalOceanLoadBalancerProtocol_http,
					CertificateName: val("my-le-cert-name"),
				},
			},
			HealthCheck: &DigitalOceanLoadBalancerHealthCheck{
				Port:             80,
				Protocol:         DigitalOceanLoadBalancerProtocol_http,
				Path:             "/health",
				CheckIntervalSec: 10,
			},
			DropletTag: "web-prod",
			StickySessions: &DigitalOceanLoadBalancerStickySessions{
				Type:             "cookies",
				CookieName:       "DO-LB",
				CookieTtlSeconds: 300,
			},
		}
	}

	// TCP balancer fronting a database, targeting explicit Droplet IDs.
	makeValidTCPSpec := func() *DigitalOceanLoadBalancerSpec {
		return &DigitalOceanLoadBalancerSpec{
			LoadBalancerName: "prod-db-lb",
			Region:           digitalocean.DigitalOceanRegion_fra1,
			ForwardingRules: []*DigitalOceanLoadBalancerForwardingRule{
				{
					EntryPort:      3306,
					EntryProtocol:  DigitalOceanLoadBalancerProtocol_tcp,
					TargetPort:     3306,
					TargetProtocol: DigitalOceanLoadBalancerProtocol_tcp,
				},
			},
			HealthCheck: &DigitalOceanLoadBalancerHealthCheck{
				Port:     3306,
				Protocol: DigitalOceanLoadBalancerProtocol_tcp,
			},
			DropletIds: []*foreignkeyv1.StringValueOrRef{val("123456"), val("789012")},
		}
	}

	// GLOBAL balancer: no region, no forwarding rules; routed through
	// glb_settings, domains, and regional target balancers.
	makeValidGlobalSpec := func() *DigitalOceanLoadBalancerSpec {
		return &DigitalOceanLoadBalancerSpec{
			LoadBalancerName: "prod-global-lb",
			Type:             "GLOBAL",
			GlbSettings: &DigitalOceanLoadBalancerGlbSettings{
				TargetProtocol: "https",
				TargetPort:     443,
				RegionPriorities: map[string]uint32{
					"nyc3": 1,
					"fra1": 2,
				},
				FailoverThreshold: 50,
				Cdn:               &DigitalOceanLoadBalancerGlbCdn{IsEnabled: true},
			},
			Domains: []*DigitalOceanLoadBalancerDomain{
				{
					Name:            "www.example.com",
					IsManaged:       true,
					CertificateName: val("my-cert-name"),
				},
			},
			TargetLoadBalancerIds: []*foreignkeyv1.StringValueOrRef{val("regional-lb-uuid")},
		}
	}

	ginkgo.Context("Valid configurations", func() {

		ginkgo.It("should accept minimal valid HTTP load balancer with tag-based targeting", func() {
			gomega.Expect(protovalidate.Validate(makeValidHTTPSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept valid HTTPS load balancer with certificate reference and sticky sessions", func() {
			gomega.Expect(protovalidate.Validate(makeValidHTTPSSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept valid TCP load balancer with droplet IDs", func() {
			gomega.Expect(protovalidate.Validate(makeValidTCPSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept valid GLOBAL load balancer", func() {
			gomega.Expect(protovalidate.Validate(makeValidGlobalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a fully-loaded regional balancer", func() {
			spec := makeValidHTTPSSpec()
			spec.Type = "REGIONAL"
			spec.SizeUnit = 4
			spec.RedirectHttpToHttps = true
			spec.EnableProxyProtocol = true
			spec.EnableBackendKeepalive = true
			spec.DisableLetsEncryptDnsRecords = true
			spec.HttpIdleTimeoutSeconds = 120
			spec.TlsCipherPolicy = "STRONG"
			spec.Network = "EXTERNAL"
			spec.NetworkStack = "DUALSTACK"
			spec.ProjectId = val("0a4b1c2d-1111-2222-3333-444455556666")
			spec.SubnetUuid = "9f8e7d6c-1111-2222-3333-444455556666"
			spec.Ip = "203.0.113.10"
			spec.Firewall = &DigitalOceanLoadBalancerFirewall{
				Allow: []string{"ip:203.0.113.5", "cidr:10.0.0.0/8"},
				Deny:  []string{"cidr:192.0.2.0/24"},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should accept multi-port forwarding rules (HTTP + HTTPS)", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules = append(spec.ForwardingRules, &DigitalOceanLoadBalancerForwardingRule{
				EntryPort:       443,
				EntryProtocol:   DigitalOceanLoadBalancerProtocol_https,
				TargetPort:      80,
				TargetProtocol:  DigitalOceanLoadBalancerProtocol_http,
				CertificateName: val("my-cert"),
			})
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Required fields", func() {

		ginkgo.It("should reject spec with missing load_balancer_name", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = ""
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject spec with neither forwarding_rules nor glb_settings", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules = nil
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Load balancer name validation", func() {

		ginkgo.It("should reject name that is too long (>64 characters)", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = "this-is-a-very-long-load-balancer-name-that-exceeds-sixty-four-characters"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name with uppercase letters", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = "Prod-Web-LB"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name with special characters", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = "prod_web_lb"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("region_by_type coupling", func() {

		ginkgo.It("should reject a regional balancer without a region", func() {
			spec := makeValidHTTPSpec()
			spec.Region = digitalocean.DigitalOceanRegion_digital_ocean_region_unspecified
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an explicit REGIONAL balancer without a region", func() {
			spec := makeValidHTTPSpec()
			spec.Type = "REGIONAL"
			spec.Region = digitalocean.DigitalOceanRegion_digital_ocean_region_unspecified
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a GLOBAL balancer carrying a region", func() {
			spec := makeValidGlobalSpec()
			spec.Region = digitalocean.DigitalOceanRegion_nyc3
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid type token", func() {
			spec := makeValidHTTPSpec()
			spec.Type = "regional"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("forwarding_rules_xor_glb_settings coupling", func() {

		ginkgo.It("should reject a balancer with both forwarding_rules and glb_settings", func() {
			spec := makeValidGlobalSpec()
			spec.ForwardingRules = makeValidHTTPSpec().ForwardingRules
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a GLOBAL balancer with neither", func() {
			spec := makeValidGlobalSpec()
			spec.GlbSettings = nil
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Sizing", func() {

		ginkgo.It("should accept size slug alone", func() {
			spec := makeValidHTTPSpec()
			spec.Size = "lb-medium"
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should accept size_unit alone", func() {
			spec := makeValidHTTPSpec()
			spec.SizeUnit = 7
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should reject size and size_unit together", func() {
			spec := makeValidHTTPSpec()
			spec.Size = "lb-small"
			spec.SizeUnit = 1
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid size slug", func() {
			spec := makeValidHTTPSpec()
			spec.Size = "lb-xlarge"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject size_unit above 200", func() {
			spec := makeValidHTTPSpec()
			spec.SizeUnit = 201
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Placement and networking", func() {

		ginkgo.It("should reject subnet_uuid without vpc", func() {
			spec := makeValidHTTPSpec()
			spec.SubnetUuid = "9f8e7d6c-1111-2222-3333-444455556666"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept subnet_uuid with vpc", func() {
			spec := makeValidHTTPSpec()
			spec.Vpc = val("test-vpc-id")
			spec.SubnetUuid = "9f8e7d6c-1111-2222-3333-444455556666"
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid network token", func() {
			spec := makeValidHTTPSpec()
			spec.Network = "PUBLIC"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid network_stack token", func() {
			spec := makeValidHTTPSpec()
			spec.NetworkStack = "IPV6"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid tls_cipher_policy token", func() {
			spec := makeValidHTTPSpec()
			spec.TlsCipherPolicy = "MODERN"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a BYOIP value that is not an IP address", func() {
			spec := makeValidHTTPSpec()
			spec.Ip = "not-an-ip"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Forwarding rule validation", func() {

		ginkgo.It("should reject forwarding rule with port 0", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryPort = 0
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject forwarding rule with port > 65535", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryPort = 70000
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject forwarding rule with unspecified entry protocol", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryProtocol = DigitalOceanLoadBalancerProtocol_digitalocean_load_balancer_protocol_unspecified
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject forwarding rule with unspecified target protocol", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].TargetProtocol = DigitalOceanLoadBalancerProtocol_digitalocean_load_balancer_protocol_unspecified
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept http3 as an entry protocol", func() {
			spec := makeValidHTTPSSpec()
			spec.ForwardingRules[0].EntryProtocol = DigitalOceanLoadBalancerProtocol_http3
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should reject http3 as a target protocol", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].TargetProtocol = DigitalOceanLoadBalancerProtocol_http3
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept udp end to end", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryProtocol = DigitalOceanLoadBalancerProtocol_udp
			spec.ForwardingRules[0].TargetProtocol = DigitalOceanLoadBalancerProtocol_udp
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a TLS passthrough rule without a certificate", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryProtocol = DigitalOceanLoadBalancerProtocol_https
			spec.ForwardingRules[0].TargetProtocol = DigitalOceanLoadBalancerProtocol_https
			spec.ForwardingRules[0].TlsPassthrough = true
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a certificate expressed as a reference", func() {
			spec := makeValidHTTPSSpec()
			spec.ForwardingRules[0].CertificateName = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-certificate"},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Health check validation", func() {

		ginkgo.It("should reject health check with port 0", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.Port = 0
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject health check with unspecified protocol", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.Protocol = DigitalOceanLoadBalancerProtocol_digitalocean_load_balancer_protocol_unspecified
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a forwarding-rule-only protocol", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.Protocol = DigitalOceanLoadBalancerProtocol_udp
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an http health check without a path", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.Path = ""
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a tcp health check with a path", func() {
			spec := makeValidTCPSpec()
			spec.HealthCheck.Path = "/health"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept full threshold tuning inside provider ranges", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.CheckIntervalSec = 30
			spec.HealthCheck.ResponseTimeoutSeconds = 10
			spec.HealthCheck.UnhealthyThreshold = 5
			spec.HealthCheck.HealthyThreshold = 2
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should reject check_interval_sec below 3", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.CheckIntervalSec = 2
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject response_timeout_seconds above 300", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.ResponseTimeoutSeconds = 301
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject unhealthy_threshold outside 2-10", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.UnhealthyThreshold = 1
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject healthy_threshold outside 2-10", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.HealthyThreshold = 11
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Sticky sessions", func() {

		ginkgo.It("should accept type none with no cookie leaves", func() {
			spec := makeValidHTTPSpec()
			spec.StickySessions = &DigitalOceanLoadBalancerStickySessions{Type: "none"}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should reject type cookies without cookie_name", func() {
			spec := makeValidHTTPSpec()
			spec.StickySessions = &DigitalOceanLoadBalancerStickySessions{
				Type:             "cookies",
				CookieTtlSeconds: 300,
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject type cookies without cookie_ttl_seconds", func() {
			spec := makeValidHTTPSpec()
			spec.StickySessions = &DigitalOceanLoadBalancerStickySessions{
				Type:       "cookies",
				CookieName: "DO-LB",
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject type none carrying cookie leaves", func() {
			spec := makeValidHTTPSpec()
			spec.StickySessions = &DigitalOceanLoadBalancerStickySessions{
				Type:       "none",
				CookieName: "DO-LB",
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid affinity type", func() {
			spec := makeValidHTTPSpec()
			spec.StickySessions = &DigitalOceanLoadBalancerStickySessions{Type: "ip-hash"}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a cookie_name shorter than 2 characters", func() {
			spec := makeValidHTTPSpec()
			spec.StickySessions = &DigitalOceanLoadBalancerStickySessions{
				Type:             "cookies",
				CookieName:       "x",
				CookieTtlSeconds: 300,
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Firewall rules", func() {

		ginkgo.It("should accept well-formed ip and cidr rules", func() {
			spec := makeValidHTTPSpec()
			spec.Firewall = &DigitalOceanLoadBalancerFirewall{
				Allow: []string{"ip:203.0.113.5", "cidr:10.0.0.0/8"},
				Deny:  []string{"cidr:192.0.2.0/24"},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should reject a bare address without a prefix", func() {
			spec := makeValidHTTPSpec()
			spec.Firewall = &DigitalOceanLoadBalancerFirewall{Allow: []string{"203.0.113.5"}}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an ip rule whose value is not an address", func() {
			spec := makeValidHTTPSpec()
			spec.Firewall = &DigitalOceanLoadBalancerFirewall{Allow: []string{"ip:banana"}}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a cidr rule whose value is not a prefix", func() {
			spec := makeValidHTTPSpec()
			spec.Firewall = &DigitalOceanLoadBalancerFirewall{Deny: []string{"cidr:10.0.0.0"}}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Backend targeting", func() {

		ginkgo.It("should reject both droplet_ids and droplet_tag together", func() {
			spec := makeValidHTTPSpec()
			spec.DropletIds = []*foreignkeyv1.StringValueOrRef{val("123456")}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject droplet_tag that is too long (>255 characters)", func() {
			spec := makeValidHTTPSpec()
			spec.DropletTag = strings.Repeat("a", 260)
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Global load balancer surfaces", func() {

		ginkgo.It("should reject a domain without a name", func() {
			spec := makeValidGlobalSpec()
			spec.Domains[0].Name = ""
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a domain name that is not a hostname", func() {
			spec := makeValidGlobalSpec()
			spec.Domains[0].Name = "not a hostname"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject glb_settings without a target_protocol", func() {
			spec := makeValidGlobalSpec()
			spec.GlbSettings.TargetProtocol = ""
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject glb_settings with a tcp target_protocol", func() {
			spec := makeValidGlobalSpec()
			spec.GlbSettings.TargetProtocol = "tcp"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject glb_settings with a target_port other than 80 or 443", func() {
			spec := makeValidGlobalSpec()
			spec.GlbSettings.TargetPort = 8080
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a failover_threshold above 99", func() {
			spec := makeValidGlobalSpec()
			spec.GlbSettings.FailoverThreshold = 100
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Full DigitalOceanLoadBalancer resource validation", func() {

		ginkgo.It("should accept complete valid load balancer resource", func() {
			input := &DigitalOceanLoadBalancer{
				ApiVersion: "digital-ocean.planton.dev/v1alpha1",
				Kind:       "DigitalOceanLoadBalancer",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-lb",
				},
				Spec: makeValidHTTPSpec(),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept complete valid GLOBAL load balancer resource", func() {
			input := &DigitalOceanLoadBalancer{
				ApiVersion: "digital-ocean.planton.dev/v1alpha1",
				Kind:       "DigitalOceanLoadBalancer",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-global-lb",
				},
				Spec: makeValidGlobalSpec(),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})
})
