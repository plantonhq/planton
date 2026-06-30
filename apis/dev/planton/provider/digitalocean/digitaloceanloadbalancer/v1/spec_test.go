package digitaloceanloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestDigitalOceanLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanLoadBalancerSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanLoadBalancerSpec validations", func() {

	// Helper function to create a minimal valid HTTP load balancer spec with tag-based targeting
	makeValidHTTPSpec := func() *DigitalOceanLoadBalancerSpec {
		return &DigitalOceanLoadBalancerSpec{
			LoadBalancerName: "prod-web-lb",
			Region:           digitalocean.DigitalOceanRegion_nyc3,
			Vpc: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-vpc-id"},
			},
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

	// Helper function to create a valid HTTPS load balancer spec with SSL termination
	makeValidHTTPSSpec := func() *DigitalOceanLoadBalancerSpec {
		return &DigitalOceanLoadBalancerSpec{
			LoadBalancerName: "prod-https-lb",
			Region:           digitalocean.DigitalOceanRegion_sfo3,
			Vpc: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-vpc-id"},
			},
			ForwardingRules: []*DigitalOceanLoadBalancerForwardingRule{
				{
					EntryPort:       443,
					EntryProtocol:   DigitalOceanLoadBalancerProtocol_https,
					TargetPort:      80,
					TargetProtocol:  DigitalOceanLoadBalancerProtocol_http,
					CertificateName: "my-le-cert-name",
				},
			},
			HealthCheck: &DigitalOceanLoadBalancerHealthCheck{
				Port:             80,
				Protocol:         DigitalOceanLoadBalancerProtocol_http,
				Path:             "/health",
				CheckIntervalSec: 10,
			},
			DropletTag:           "web-prod",
			EnableStickySessions: true,
		}
	}

	// Helper function to create a valid TCP load balancer spec for database
	makeValidTCPSpec := func() *DigitalOceanLoadBalancerSpec {
		return &DigitalOceanLoadBalancerSpec{
			LoadBalancerName: "prod-db-lb",
			Region:           digitalocean.DigitalOceanRegion_fra1,
			Vpc: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-vpc-id"},
			},
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
			DropletTag: "db-prod",
		}
	}

	// Helper function to create a valid spec with droplet IDs instead of tag
	makeValidDropletIdSpec := func() *DigitalOceanLoadBalancerSpec {
		return &DigitalOceanLoadBalancerSpec{
			LoadBalancerName: "dev-lb",
			Region:           digitalocean.DigitalOceanRegion_nyc3,
			Vpc: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-vpc-id"},
			},
			ForwardingRules: []*DigitalOceanLoadBalancerForwardingRule{
				{
					EntryPort:      80,
					EntryProtocol:  DigitalOceanLoadBalancerProtocol_http,
					TargetPort:     8080,
					TargetProtocol: DigitalOceanLoadBalancerProtocol_http,
				},
			},
			HealthCheck: &DigitalOceanLoadBalancerHealthCheck{
				Port:     8080,
				Protocol: DigitalOceanLoadBalancerProtocol_http,
				Path:     "/",
			},
			DropletIds: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "123456"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "789012"}},
			},
		}
	}

	ginkgo.Context("Valid configurations", func() {

		ginkgo.It("should accept minimal valid HTTP load balancer with tag-based targeting", func() {
			spec := makeValidHTTPSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept valid HTTPS load balancer with SSL certificate", func() {
			spec := makeValidHTTPSSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept valid TCP load balancer for database", func() {
			spec := makeValidTCPSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept valid load balancer with droplet IDs", func() {
			spec := makeValidDropletIdSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept multi-port forwarding rules (HTTP + HTTPS)", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules = append(spec.ForwardingRules, &DigitalOceanLoadBalancerForwardingRule{
				EntryPort:       443,
				EntryProtocol:   DigitalOceanLoadBalancerProtocol_https,
				TargetPort:      80,
				TargetProtocol:  DigitalOceanLoadBalancerProtocol_http,
				CertificateName: "my-cert",
			})
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Required fields", func() {

		ginkgo.It("should reject spec with missing load_balancer_name", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject spec with missing region", func() {
			spec := makeValidHTTPSpec()
			spec.Region = digitalocean.DigitalOceanRegion_digital_ocean_region_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject spec with missing VPC", func() {
			spec := makeValidHTTPSpec()
			spec.Vpc = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject spec with empty forwarding_rules", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules = []*DigitalOceanLoadBalancerForwardingRule{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject spec with nil forwarding_rules", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Load balancer name validation", func() {

		ginkgo.It("should reject name that is too short (empty)", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name that is too long (>64 characters)", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = "this-is-a-very-long-load-balancer-name-that-exceeds-sixty-four-characters"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name with uppercase letters", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = "Prod-Web-LB"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject name with special characters", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = "prod_web_lb"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept valid lowercase alphanumeric name with hyphens", func() {
			spec := makeValidHTTPSpec()
			spec.LoadBalancerName = "prod-web-lb-2024"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Forwarding rule validation", func() {

		ginkgo.It("should reject forwarding rule with port 0", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryPort = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject forwarding rule with port > 65535", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryPort = 70000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept forwarding rule with valid port 1", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryPort = 1
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept forwarding rule with valid port 65535", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryPort = 65535
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject forwarding rule with unspecified entry protocol", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].EntryProtocol = DigitalOceanLoadBalancerProtocol_digitalocean_load_balancer_protocol_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject forwarding rule with unspecified target protocol", func() {
			spec := makeValidHTTPSpec()
			spec.ForwardingRules[0].TargetProtocol = DigitalOceanLoadBalancerProtocol_digitalocean_load_balancer_protocol_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept certificate_name with valid length", func() {
			spec := makeValidHTTPSSpec()
			spec.ForwardingRules[0].CertificateName = "my-certificate-name"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject certificate_name that is too long (>255 characters)", func() {
			spec := makeValidHTTPSSpec()
			longName := ""
			for i := 0; i < 260; i++ {
				longName += "a"
			}
			spec.ForwardingRules[0].CertificateName = longName
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Health check validation", func() {

		ginkgo.It("should reject health check with port 0", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.Port = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject health check with port > 65535", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.Port = 70000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept health check with valid port range", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.Port = 8080
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject health check with unspecified protocol", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.Protocol = DigitalOceanLoadBalancerProtocol_digitalocean_load_balancer_protocol_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept health check with path for HTTP protocol", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.Path = "/healthz"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept health check without path for TCP protocol", func() {
			spec := makeValidTCPSpec()
			spec.HealthCheck.Path = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept health check with check_interval_sec", func() {
			spec := makeValidHTTPSpec()
			spec.HealthCheck.CheckIntervalSec = 10
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Backend targeting validation", func() {

		ginkgo.It("should accept spec with only droplet_tag", func() {
			spec := makeValidHTTPSpec()
			spec.DropletTag = "web-prod"
			spec.DropletIds = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept spec with only droplet_ids", func() {
			spec := makeValidDropletIdSpec()
			spec.DropletIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "123456"}},
			}
			spec.DropletTag = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept spec with both droplet_ids and droplet_tag (mutually exclusive handled by application logic)", func() {
			spec := makeValidHTTPSpec()
			spec.DropletTag = "web-prod"
			spec.DropletIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "123456"}},
			}
			// Note: Proto validation allows both, but application logic should enforce mutual exclusivity
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject droplet_tag that is too long (>255 characters)", func() {
			spec := makeValidHTTPSpec()
			longTag := ""
			for i := 0; i < 260; i++ {
				longTag += "a"
			}
			spec.DropletTag = longTag
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("Full DigitalOceanLoadBalancer resource validation", func() {

		ginkgo.It("should accept complete valid load balancer resource", func() {
			input := &DigitalOceanLoadBalancer{
				ApiVersion: "digital-ocean.planton.dev/v1",
				Kind:       "DigitalOceanLoadBalancer",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-lb",
				},
				Spec: makeValidHTTPSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept complete valid HTTPS load balancer resource with sticky sessions", func() {
			input := &DigitalOceanLoadBalancer{
				ApiVersion: "digital-ocean.planton.dev/v1",
				Kind:       "DigitalOceanLoadBalancer",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-https-lb",
				},
				Spec: makeValidHTTPSSpec(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
