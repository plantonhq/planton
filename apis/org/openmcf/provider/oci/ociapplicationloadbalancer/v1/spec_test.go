package ociapplicationloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciApplicationLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciApplicationLoadBalancerSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidLoadBalancer() *OciApplicationLoadBalancer {
	return &OciApplicationLoadBalancer{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciApplicationLoadBalancer",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-lb",
		},
		Spec: &OciApplicationLoadBalancerSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			Shape:         "flexible",
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				newStringValueOrRef("ocid1.subnet.oc1.iad.example"),
			},
			BackendSets: []*OciApplicationLoadBalancerSpec_BackendSet{
				{
					Name:   "web-backend",
					Policy: OciApplicationLoadBalancerSpec_BackendSet_round_robin,
					HealthChecker: &OciApplicationLoadBalancerSpec_HealthChecker{
						Protocol: OciApplicationLoadBalancerSpec_HealthChecker_http,
						UrlPath:  "/health",
					},
				},
			},
			Listeners: []*OciApplicationLoadBalancerSpec_Listener{
				{
					Name:                  "http-listener",
					Port:                  80,
					Protocol:              OciApplicationLoadBalancerSpec_Listener_http,
					DefaultBackendSetName: "web-backend",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("OciApplicationLoadBalancerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_load_balancer", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidLoadBalancer()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display name set", func() {
				input := minimalValidLoadBalancer()
				input.Spec.DisplayName = "My Production LB"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for flexible shape with shape details", func() {
				input := minimalValidLoadBalancer()
				input.Spec.ShapeDetails = &OciApplicationLoadBalancerSpec_ShapeDetails{
					MinimumBandwidthInMbps: 10,
					MaximumBandwidthInMbps: 100,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for private load balancer with NSGs", func() {
				input := minimalValidLoadBalancer()
				input.Spec.IsPrivate = true
				input.Spec.NetworkSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.example1"),
					newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.example2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple subnets for HA", func() {
				input := minimalValidLoadBalancer()
				input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.subnet.oc1.iad.ad1"),
					newStringValueOrRef("ocid1.subnet.oc1.iad.ad2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for HTTPS listener with certificate", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Certificates = []*OciApplicationLoadBalancerSpec_Certificate{
					{
						CertificateName:   "my-cert",
						PublicCertificate: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
						PrivateKey:        "-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----",
					},
				}
				input.Spec.Listeners = []*OciApplicationLoadBalancerSpec_Listener{
					{
						Name:                  "https-listener",
						Port:                  443,
						Protocol:              OciApplicationLoadBalancerSpec_Listener_http,
						DefaultBackendSetName: "web-backend",
						SslConfiguration: &OciApplicationLoadBalancerSpec_SslConfiguration{
							CertificateName: "my-cert",
							Protocols:       []string{"TLSv1.2", "TLSv1.3"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple backend sets", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets = []*OciApplicationLoadBalancerSpec_BackendSet{
					{
						Name:   "web-backend",
						Policy: OciApplicationLoadBalancerSpec_BackendSet_round_robin,
						HealthChecker: &OciApplicationLoadBalancerSpec_HealthChecker{
							Protocol: OciApplicationLoadBalancerSpec_HealthChecker_http,
							UrlPath:  "/health",
						},
					},
					{
						Name:   "api-backend",
						Policy: OciApplicationLoadBalancerSpec_BackendSet_least_connections,
						HealthChecker: &OciApplicationLoadBalancerSpec_HealthChecker{
							Protocol: OciApplicationLoadBalancerSpec_HealthChecker_tcp,
							Port:     8080,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backends in a backend set", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].Backends = []*OciApplicationLoadBalancerSpec_Backend{
					{IpAddress: "10.0.1.10", Port: 8080, Weight: 1},
					{IpAddress: "10.0.1.11", Port: 8080, Weight: 2},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backend drain and backup flags", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].Backends = []*OciApplicationLoadBalancerSpec_Backend{
					{IpAddress: "10.0.1.10", Port: 8080, Weight: 1},
					{IpAddress: "10.0.1.11", Port: 8080, Backup: true},
					{IpAddress: "10.0.1.12", Port: 8080, Drain: true},
					{IpAddress: "10.0.1.13", Port: 8080, Offline: true},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with LB cookie session persistence", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].SessionPersistenceConfig = &OciApplicationLoadBalancerSpec_BackendSet_LbCookieSessionPersistence{
					LbCookieSessionPersistence: &OciApplicationLoadBalancerSpec_LbCookieSessionPersistenceConfig{
						CookieName:      "X-Oracle-BMC-LBS-Route",
						IsHttpOnly:      true,
						IsSecure:        true,
						MaxAgeInSeconds: 3600,
						Path:            "/",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with app cookie session persistence", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].SessionPersistenceConfig = &OciApplicationLoadBalancerSpec_BackendSet_AppCookieSessionPersistence{
					AppCookieSessionPersistence: &OciApplicationLoadBalancerSpec_SessionPersistenceConfig{
						CookieName: "JSESSIONID",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backend set SSL configuration", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].SslConfiguration = &OciApplicationLoadBalancerSpec_SslConfiguration{
					CertificateIds:        []string{"ocid1.certificate.oc1.iad.example"},
					CipherSuiteName:       "oci-default-ssl-cipher-suite-v1",
					Protocols:             []string{"TLSv1.2"},
					VerifyDepth:           3,
					VerifyPeerCertificate: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with hostnames and listener hostname binding", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Hostnames = []*OciApplicationLoadBalancerSpec_Hostname{
					{Name: "app-host", Hostname: "app.example.com"},
					{Name: "api-host", Hostname: "api.example.com"},
				}
				input.Spec.Listeners[0].HostnameNames = []string{"app-host", "api-host"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HTTP redirect rule set", func() {
				input := minimalValidLoadBalancer()
				input.Spec.RuleSets = []*OciApplicationLoadBalancerSpec_RuleSet{
					{
						Name: "http-to-https",
						Items: []*OciApplicationLoadBalancerSpec_RuleSetItem{
							{
								Action:       OciApplicationLoadBalancerSpec_RuleSetItem_redirect,
								ResponseCode: 301,
								RedirectUri: &OciApplicationLoadBalancerSpec_RedirectUri{
									Protocol: "HTTPS",
									Host:     "{host}",
									Port:     443,
									Path:     "{path}",
									Query:    "{query}",
								},
							},
						},
					},
				}
				input.Spec.Listeners[0].RuleSetNames = []string{"http-to-https"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with header manipulation rule set", func() {
				input := minimalValidLoadBalancer()
				input.Spec.RuleSets = []*OciApplicationLoadBalancerSpec_RuleSet{
					{
						Name: "security-headers",
						Items: []*OciApplicationLoadBalancerSpec_RuleSetItem{
							{
								Action: OciApplicationLoadBalancerSpec_RuleSetItem_add_http_response_header,
								Header: "X-Frame-Options",
								Value:  "DENY",
							},
							{
								Action: OciApplicationLoadBalancerSpec_RuleSetItem_add_http_response_header,
								Header: "Strict-Transport-Security",
								Value:  "max-age=31536000; includeSubDomains",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with delete protection enabled", func() {
				input := minimalValidLoadBalancer()
				input.Spec.IsDeleteProtectionEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with request ID enabled", func() {
				input := minimalValidLoadBalancer()
				input.Spec.IsRequestIdEnabled = true
				input.Spec.RequestIdHeader = "X-Custom-Request-Id"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with reserved IPs", func() {
				input := minimalValidLoadBalancer()
				input.Spec.ReservedIps = []*OciApplicationLoadBalancerSpec_ReservedIp{
					{Id: "ocid1.publicip.oc1.iad.example"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with listener connection configuration", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Listeners[0].ConnectionConfiguration = &OciApplicationLoadBalancerSpec_ConnectionConfiguration{
					IdleTimeoutInSeconds: 300,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HTTP2 listener protocol", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Listeners[0].Protocol = OciApplicationLoadBalancerSpec_Listener_http2
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with gRPC listener protocol", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Listeners[0].Protocol = OciApplicationLoadBalancerSpec_Listener_grpc
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidLoadBalancer()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with subnet_ids via value_from ref", func() {
				input := minimalValidLoadBalancer()
				input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
					{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{
								Name: "my-subnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ip_hash policy", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].Policy = OciApplicationLoadBalancerSpec_BackendSet_ip_hash
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backend max connections", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].BackendMaxConnections = 1000
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ip_mode set", func() {
				input := minimalValidLoadBalancer()
				input.Spec.IpMode = "IPV4"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with health checker custom settings", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].HealthChecker = &OciApplicationLoadBalancerSpec_HealthChecker{
					Protocol:          OciApplicationLoadBalancerSpec_HealthChecker_http,
					Port:              8080,
					UrlPath:           "/ready",
					ReturnCode:        200,
					ResponseBodyRegex: ".*OK.*",
					IntervalMs:        10000,
					TimeoutInMillis:   5000,
					Retries:           5,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with allow rule and conditions", func() {
				input := minimalValidLoadBalancer()
				input.Spec.RuleSets = []*OciApplicationLoadBalancerSpec_RuleSet{
					{
						Name: "allow-internal",
						Items: []*OciApplicationLoadBalancerSpec_RuleSetItem{
							{
								Action: OciApplicationLoadBalancerSpec_RuleSetItem_allow,
								Conditions: []*OciApplicationLoadBalancerSpec_RuleSetItemCondition{
									{
										AttributeName:  "SOURCE_IP_ADDRESS",
										AttributeValue: "10.0.0.0/8",
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_load_balancer", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidLoadBalancer()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidLoadBalancer()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidLoadBalancer()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciApplicationLoadBalancer{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciApplicationLoadBalancer",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-lb"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidLoadBalancer()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Shape = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_ids is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.SubnetIds = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend_sets is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listeners is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Listeners = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend set name is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend set policy is unspecified", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].Policy = OciApplicationLoadBalancerSpec_BackendSet_policy_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health checker is missing", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].HealthChecker = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health checker protocol is unspecified", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].HealthChecker.Protocol = OciApplicationLoadBalancerSpec_HealthChecker_protocol_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener name is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Listeners[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener port is zero", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Listeners[0].Port = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener port exceeds 65535", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Listeners[0].Port = 70000
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener protocol is unspecified", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Listeners[0].Protocol = OciApplicationLoadBalancerSpec_Listener_protocol_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener default_backend_set_name is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Listeners[0].DefaultBackendSetName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend ip_address is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].Backends = []*OciApplicationLoadBalancerSpec_Backend{
					{IpAddress: "", Port: 8080},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend port is zero", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].Backends = []*OciApplicationLoadBalancerSpec_Backend{
					{IpAddress: "10.0.1.10", Port: 0},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when certificate name is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Certificates = []*OciApplicationLoadBalancerSpec_Certificate{
					{CertificateName: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when hostname name is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Hostnames = []*OciApplicationLoadBalancerSpec_Hostname{
					{Name: "", Hostname: "example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when hostname value is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.Hostnames = []*OciApplicationLoadBalancerSpec_Hostname{
					{Name: "my-host", Hostname: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule set name is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.RuleSets = []*OciApplicationLoadBalancerSpec_RuleSet{
					{
						Name: "",
						Items: []*OciApplicationLoadBalancerSpec_RuleSetItem{
							{Action: OciApplicationLoadBalancerSpec_RuleSetItem_redirect},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule set items is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.RuleSets = []*OciApplicationLoadBalancerSpec_RuleSet{
					{Name: "empty-rules", Items: nil},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rule set item action is unspecified", func() {
				input := minimalValidLoadBalancer()
				input.Spec.RuleSets = []*OciApplicationLoadBalancerSpec_RuleSet{
					{
						Name: "bad-action",
						Items: []*OciApplicationLoadBalancerSpec_RuleSetItem{
							{Action: OciApplicationLoadBalancerSpec_RuleSetItem_action_unspecified},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when app cookie session persistence cookie name is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.BackendSets[0].SessionPersistenceConfig = &OciApplicationLoadBalancerSpec_BackendSet_AppCookieSessionPersistence{
					AppCookieSessionPersistence: &OciApplicationLoadBalancerSpec_SessionPersistenceConfig{
						CookieName: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when reserved IP id is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.ReservedIps = []*OciApplicationLoadBalancerSpec_ReservedIp{
					{Id: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when condition attribute_name is empty", func() {
				input := minimalValidLoadBalancer()
				input.Spec.RuleSets = []*OciApplicationLoadBalancerSpec_RuleSet{
					{
						Name: "bad-condition",
						Items: []*OciApplicationLoadBalancerSpec_RuleSetItem{
							{
								Action: OciApplicationLoadBalancerSpec_RuleSetItem_allow,
								Conditions: []*OciApplicationLoadBalancerSpec_RuleSetItemCondition{
									{AttributeName: "", AttributeValue: "10.0.0.0/8"},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape_details bandwidth exceeds maximum", func() {
				input := minimalValidLoadBalancer()
				input.Spec.ShapeDetails = &OciApplicationLoadBalancerSpec_ShapeDetails{
					MinimumBandwidthInMbps: 10,
					MaximumBandwidthInMbps: 10000,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape_details bandwidth below minimum", func() {
				input := minimalValidLoadBalancer()
				input.Spec.ShapeDetails = &OciApplicationLoadBalancerSpec_ShapeDetails{
					MinimumBandwidthInMbps: 5,
					MaximumBandwidthInMbps: 100,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
