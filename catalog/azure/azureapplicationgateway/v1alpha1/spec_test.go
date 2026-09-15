package azureapplicationgatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureApplicationGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureApplicationGatewaySpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimal valid spec: a Standard_v2 gateway with one public HTTP
// listener routed to one pool.
func minimalSpec() *AzureApplicationGateway {
	capacity := int32(1)
	return &AzureApplicationGateway{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureApplicationGateway",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-agw",
		},
		Spec: &AzureApplicationGatewaySpec{
			Region:        "eastus",
			ResourceGroup: literal("my-rg"),
			Name:          "test-agw",
			SubnetId:      literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/agw"),
			Sku:           AzureApplicationGatewaySku_STANDARD_V2,
			Capacity:      &capacity,
			FrontendIpConfigurations: []*AzureApplicationGatewayFrontendIpConfiguration{
				{
					Name:              "public",
					PublicIpAddressId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/agw-pip"),
				},
			},
			FrontendPorts: []*AzureApplicationGatewayFrontendPort{
				{Name: "http", Port: 80},
			},
			BackendAddressPools: []*AzureApplicationGatewayBackendAddressPool{
				{Name: "web", IpAddresses: []string{"10.0.1.4"}},
			},
			BackendHttpSettings: []*AzureApplicationGatewayBackendHttpSettings{
				{Name: "http-settings", Port: 8080, Protocol: AzureApplicationGatewayProtocol_HTTP},
			},
			HttpListeners: []*AzureApplicationGatewayHttpListener{
				{
					Name:                        "http-listener",
					FrontendIpConfigurationName: "public",
					FrontendPortName:            "http",
					Protocol:                    AzureApplicationGatewayProtocol_HTTP,
				},
			},
			RequestRoutingRules: []*AzureApplicationGatewayRequestRoutingRule{
				{
					Name:                    "http-rule",
					RuleType:                AzureApplicationGatewayRequestRoutingRuleType_BASIC_ROUTING,
					HttpListenerName:        "http-listener",
					Priority:                100,
					BackendAddressPoolName:  "web",
					BackendHttpSettingsName: "http-settings",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureApplicationGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal Standard_v2 gateway", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept autoscale instead of capacity", func() {
			input := minimalSpec()
			input.Spec.Capacity = nil
			input.Spec.Autoscale = &AzureApplicationGatewayAutoscale{MinCapacity: proto.Int32(2)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept zones and HTTP/2", func() {
			http2 := true
			input := minimalSpec()
			input.Spec.Zones = []string{"1", "2", "3"}
			input.Spec.Http2Enabled = &http2
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a private frontend with a static address", func() {
			input := minimalSpec()
			input.Spec.FrontendIpConfigurations = append(input.Spec.FrontendIpConfigurations,
				&AzureApplicationGatewayFrontendIpConfiguration{
					Name:                       "internal",
					SubnetId:                   literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/agw"),
					PrivateIpAddress:           "10.0.2.10",
					PrivateIpAddressAllocation: AzureApplicationGatewayIpAllocation_STATIC,
				})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an HTTPS listener with a Key Vault certificate and identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureApplicationGatewayIdentity{
				Type:        AzureApplicationGatewayIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/agw")},
			}
			input.Spec.SslCertificates = []*AzureApplicationGatewaySslCertificate{
				{Name: "tls", KeyVaultSecretId: literal("https://vault.vault.azure.net/secrets/tls-cert")},
			}
			input.Spec.FrontendPorts = append(input.Spec.FrontendPorts, &AzureApplicationGatewayFrontendPort{Name: "https", Port: 443})
			input.Spec.HttpListeners = append(input.Spec.HttpListeners, &AzureApplicationGatewayHttpListener{
				Name:                        "https-listener",
				FrontendIpConfigurationName: "public",
				FrontendPortName:            "https",
				Protocol:                    AzureApplicationGatewayProtocol_HTTPS,
				SslCertificateName:          "tls",
				HostNames:                   []string{"www.contoso.com", "*.contoso.com"},
				RequireSni:                  true,
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an inline PFX certificate with a password", func() {
			input := minimalSpec()
			input.Spec.SslCertificates = []*AzureApplicationGatewaySslCertificate{
				{Name: "inline-tls", Data: "bWljcm9zb2Z0LXBmeA==", Password: "changeit"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept path-based routing with a url path map", func() {
			input := minimalSpec()
			input.Spec.UrlPathMaps = []*AzureApplicationGatewayUrlPathMap{
				{
					Name:                           "paths",
					DefaultBackendAddressPoolName:  "web",
					DefaultBackendHttpSettingsName: "http-settings",
					PathRules: []*AzureApplicationGatewayPathRule{
						{
							Name:                    "api",
							Paths:                   []string{"/api/*"},
							BackendAddressPoolName:  "web",
							BackendHttpSettingsName: "http-settings",
						},
					},
				},
			}
			input.Spec.RequestRoutingRules[0].RuleType = AzureApplicationGatewayRequestRoutingRuleType_PATH_BASED_ROUTING
			input.Spec.RequestRoutingRules[0].BackendAddressPoolName = ""
			input.Spec.RequestRoutingRules[0].BackendHttpSettingsName = ""
			input.Spec.RequestRoutingRules[0].UrlPathMapName = "paths"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an HTTP -> HTTPS redirect rule", func() {
			input := minimalSpec()
			input.Spec.RedirectConfigurations = []*AzureApplicationGatewayRedirectConfiguration{
				{
					Name:               "to-https",
					RedirectType:       AzureApplicationGatewayRedirectType_PERMANENT,
					TargetListenerName: "http-listener",
					IncludePath:        true,
					IncludeQueryString: true,
				},
			}
			input.Spec.RequestRoutingRules[0].BackendAddressPoolName = ""
			input.Spec.RequestRoutingRules[0].BackendHttpSettingsName = ""
			input.Spec.RequestRoutingRules[0].RedirectConfigurationName = "to-https"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept rewrite rule sets with URL rewriting on v2", func() {
			input := minimalSpec()
			input.Spec.RewriteRuleSets = []*AzureApplicationGatewayRewriteRuleSet{
				{
					Name: "rewrites",
					RewriteRules: []*AzureApplicationGatewayRewriteRule{
						{
							Name:         "strip-prefix",
							RuleSequence: 100,
							Conditions: []*AzureApplicationGatewayRewriteRuleCondition{
								{Variable: "var_uri_path", Pattern: "^/legacy/(.*)", IgnoreCase: true},
							},
							RequestHeaderConfigurations: []*AzureApplicationGatewayRewriteRuleHeaderConfiguration{
								{HeaderName: "X-Forwarded-Host", HeaderValue: "{var_host}"},
							},
							Url: &AzureApplicationGatewayRewriteRuleUrl{Path: "/{var_uri_path_1}", Reroute: true},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept probes with match criteria and TCP probes", func() {
			input := minimalSpec()
			input.Spec.Probes = []*AzureApplicationGatewayProbe{
				{
					Name:               "http-health",
					Protocol:           AzureApplicationGatewayProtocol_HTTP,
					Path:               "/healthz",
					Host:               "app.internal",
					Interval:           30,
					Timeout:            10,
					UnhealthyThreshold: 3,
					Match: &AzureApplicationGatewayProbeMatch{
						StatusCodes: []string{"200-399"},
					},
				},
				{
					Name:                       "tcp-health",
					Protocol:                   AzureApplicationGatewayProtocol_TCP,
					Interval:                   30,
					Timeout:                    10,
					UnhealthyThreshold:         3,
					ProxyProtocolHeaderEnabled: true,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept mTLS via an SSL profile with a custom policy", func() {
			input := minimalSpec()
			input.Spec.TrustedClientCertificates = []*AzureApplicationGatewayTrustedClientCertificate{
				{Name: "client-ca", Data: "Y2xpZW50LWNh"},
			}
			input.Spec.SslProfiles = []*AzureApplicationGatewaySslProfile{
				{
					Name:                              "mtls",
					TrustedClientCertificateNames:     []string{"client-ca"},
					VerifyClientCertificateIssuerDn:   true,
					VerifyClientCertificateRevocation: AzureApplicationGatewayClientRevocationCheck_OCSP,
					SslPolicy: &AzureApplicationGatewaySslPolicy{
						PolicyType:         AzureApplicationGatewaySslPolicyType_CUSTOM_V2,
						MinProtocolVersion: AzureApplicationGatewayTlsProtocol_TLS_V1_3,
						CipherSuites:       []string{"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a gateway-wide predefined SSL policy", func() {
			input := minimalSpec()
			input.Spec.SslPolicy = &AzureApplicationGatewaySslPolicy{
				PolicyType: AzureApplicationGatewaySslPolicyType_PREDEFINED,
				PolicyName: "AppGwSslPolicy20220101S",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept disabling TLS versions without a policy type", func() {
			input := minimalSpec()
			input.Spec.SslPolicy = &AzureApplicationGatewaySslPolicy{
				DisabledProtocols: []AzureApplicationGatewayTlsProtocol{
					AzureApplicationGatewayTlsProtocol_TLS_V1_0,
					AzureApplicationGatewayTlsProtocol_TLS_V1_1,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a WAF policy on the WAF_V2 SKU", func() {
			input := minimalSpec()
			input.Spec.Sku = AzureApplicationGatewaySku_WAF_V2
			input.Spec.FirewallPolicyId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies/waf")
			input.Spec.ForceFirewallPolicyAssociation = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a layer-4 TLS proxy trio", func() {
			input := minimalSpec()
			input.Spec.SslCertificates = []*AzureApplicationGatewaySslCertificate{
				{Name: "tls", Data: "cGZ4"},
			}
			input.Spec.FrontendPorts = append(input.Spec.FrontendPorts, &AzureApplicationGatewayFrontendPort{Name: "amqps", Port: 5671})
			input.Spec.Listeners = []*AzureApplicationGatewayLayer4Listener{
				{
					Name:                        "amqps-listener",
					FrontendIpConfigurationName: "public",
					FrontendPortName:            "amqps",
					Protocol:                    AzureApplicationGatewayProtocol_TLS,
					SslCertificateName:          "tls",
				},
			}
			input.Spec.Backends = []*AzureApplicationGatewayLayer4BackendSettings{
				{Name: "amqps-backend", Port: 5671, Protocol: AzureApplicationGatewayProtocol_TCP, ClientIpPreservationEnabled: true},
			}
			input.Spec.RoutingRules = []*AzureApplicationGatewayLayer4RoutingRule{
				{
					Name:                   "amqps-rule",
					ListenerName:           "amqps-listener",
					BackendAddressPoolName: "web",
					BackendSettingsName:    "amqps-backend",
					Priority:               200,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept private link, custom errors, buffering, and connection draining", func() {
			off := false
			input := minimalSpec()
			input.Spec.FrontendIpConfigurations[0].PrivateLinkConfigurationName = "pl"
			input.Spec.PrivateLinkConfigurations = []*AzureApplicationGatewayPrivateLinkConfiguration{
				{
					Name: "pl",
					IpConfigurations: []*AzureApplicationGatewayPrivateLinkIpConfiguration{
						{
							Name:                       "nat",
							SubnetId:                   literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/agw"),
							PrivateIpAddressAllocation: AzureApplicationGatewayIpAllocation_DYNAMIC,
							Primary:                    true,
						},
					},
				},
			}
			input.Spec.CustomErrorConfigurations = []*AzureApplicationGatewayCustomErrorConfiguration{
				{StatusCode: AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_502, CustomErrorPageUrl: "https://errors.contoso.com/502.html"},
			}
			input.Spec.GlobalConfiguration = &AzureApplicationGatewayGlobalConfiguration{
				RequestBufferingEnabled:  &off,
				ResponseBufferingEnabled: &off,
			}
			input.Spec.BackendHttpSettings[0].ConnectionDraining = &AzureApplicationGatewayConnectionDraining{
				Enabled:         true,
				DrainTimeoutSec: 60,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a BASIC gateway inside its limits", func() {
			capacity := int32(2)
			input := minimalSpec()
			input.Spec.Sku = AzureApplicationGatewaySku_BASIC
			input.Spec.Capacity = &capacity
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should accept an autoscale floor of 0 stated explicitly", func() {
			input := minimalSpec()
			input.Spec.Capacity = nil
			input.Spec.Autoscale = &AzureApplicationGatewayAutoscale{MinCapacity: proto.Int32(0)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject autoscale that does not state its floor", func() {
			input := minimalSpec()
			input.Spec.Capacity = nil
			input.Spec.Autoscale = &AzureApplicationGatewayAutoscale{MaxCapacity: proto.Int32(10)}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("min_capacity"))
		})

		ginkgo.It("should reject both capacity and autoscale", func() {
			input := minimalSpec()
			input.Spec.Autoscale = &AzureApplicationGatewayAutoscale{MinCapacity: proto.Int32(2)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject neither capacity nor autoscale", func() {
			input := minimalSpec()
			input.Spec.Capacity = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject autoscale on the BASIC SKU", func() {
			input := minimalSpec()
			input.Spec.Sku = AzureApplicationGatewaySku_BASIC
			input.Spec.Capacity = nil
			input.Spec.Autoscale = &AzureApplicationGatewayAutoscale{MinCapacity: proto.Int32(2)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject BASIC capacity above 2", func() {
			capacity := int32(3)
			input := minimalSpec()
			input.Spec.Sku = AzureApplicationGatewaySku_BASIC
			input.Spec.Capacity = &capacity
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject mTLS certificates on the BASIC SKU", func() {
			capacity := int32(2)
			input := minimalSpec()
			input.Spec.Sku = AzureApplicationGatewaySku_BASIC
			input.Spec.Capacity = &capacity
			input.Spec.TrustedClientCertificates = []*AzureApplicationGatewayTrustedClientCertificate{
				{Name: "ca", Data: "Y2E="},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject URL rewriting on the BASIC SKU", func() {
			capacity := int32(2)
			input := minimalSpec()
			input.Spec.Sku = AzureApplicationGatewaySku_BASIC
			input.Spec.Capacity = &capacity
			input.Spec.RewriteRuleSets = []*AzureApplicationGatewayRewriteRuleSet{
				{
					Name: "rw",
					RewriteRules: []*AzureApplicationGatewayRewriteRule{
						{Name: "r", RuleSequence: 1, Url: &AzureApplicationGatewayRewriteRuleUrl{Path: "/x"}},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a WAF policy on a non-WAF SKU", func() {
			input := minimalSpec()
			input.Spec.FirewallPolicyId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies/waf")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a frontend that is both public and private", func() {
			input := minimalSpec()
			input.Spec.FrontendIpConfigurations[0].SubnetId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/agw")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a static private frontend without an address", func() {
			input := minimalSpec()
			input.Spec.FrontendIpConfigurations = []*AzureApplicationGatewayFrontendIpConfiguration{
				{
					Name:                       "internal",
					SubnetId:                   literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/agw"),
					PrivateIpAddressAllocation: AzureApplicationGatewayIpAllocation_STATIC,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an HTTPS listener without a certificate", func() {
			input := minimalSpec()
			input.Spec.HttpListeners[0].Protocol = AzureApplicationGatewayProtocol_HTTPS
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a TCP protocol on an HTTP listener", func() {
			input := minimalSpec()
			input.Spec.HttpListeners[0].Protocol = AzureApplicationGatewayProtocol_TCP
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a routing rule with both a backend and a redirect", func() {
			input := minimalSpec()
			input.Spec.RedirectConfigurations = []*AzureApplicationGatewayRedirectConfiguration{
				{Name: "r", RedirectType: AzureApplicationGatewayRedirectType_PERMANENT, TargetUrl: "https://contoso.com"},
			}
			input.Spec.RequestRoutingRules[0].RedirectConfigurationName = "r"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a PATH_BASED_ROUTING rule without a url path map", func() {
			input := minimalSpec()
			input.Spec.RequestRoutingRules[0].RuleType = AzureApplicationGatewayRequestRoutingRuleType_PATH_BASED_ROUTING
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a routing rule without a priority", func() {
			input := minimalSpec()
			input.Spec.RequestRoutingRules[0].Priority = 0
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a url path map defaulting to both backend and redirect", func() {
			input := minimalSpec()
			input.Spec.UrlPathMaps = []*AzureApplicationGatewayUrlPathMap{
				{
					Name:                             "paths",
					DefaultBackendAddressPoolName:    "web",
					DefaultBackendHttpSettingsName:   "http-settings",
					DefaultRedirectConfigurationName: "r",
					PathRules: []*AzureApplicationGatewayPathRule{
						{Name: "api", Paths: []string{"/api/*"}, BackendAddressPoolName: "web", BackendHttpSettingsName: "http-settings"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a certificate with both Key Vault and inline sources", func() {
			input := minimalSpec()
			input.Spec.SslCertificates = []*AzureApplicationGatewaySslCertificate{
				{Name: "tls", KeyVaultSecretId: literal("https://v.vault.azure.net/secrets/c"), Data: "cGZ4"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a certificate password without inline data", func() {
			input := minimalSpec()
			input.Spec.SslCertificates = []*AzureApplicationGatewaySslCertificate{
				{Name: "tls", KeyVaultSecretId: literal("https://v.vault.azure.net/secrets/c"), Password: "x"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a probe whose timeout exceeds its interval", func() {
			input := minimalSpec()
			input.Spec.Probes = []*AzureApplicationGatewayProbe{
				{Name: "p", Protocol: AzureApplicationGatewayProtocol_HTTP, Path: "/h", Host: "h", Interval: 10, Timeout: 30, UnhealthyThreshold: 3},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an HTTP probe without a path", func() {
			input := minimalSpec()
			input.Spec.Probes = []*AzureApplicationGatewayProbe{
				{Name: "p", Protocol: AzureApplicationGatewayProtocol_HTTP, Host: "h", Interval: 30, Timeout: 10, UnhealthyThreshold: 3},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a TCP probe carrying HTTP fields", func() {
			input := minimalSpec()
			input.Spec.Probes = []*AzureApplicationGatewayProbe{
				{Name: "p", Protocol: AzureApplicationGatewayProtocol_TCP, Path: "/h", Interval: 30, Timeout: 10, UnhealthyThreshold: 3},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an SSL policy mixing disabled protocols with a policy type", func() {
			input := minimalSpec()
			input.Spec.SslPolicy = &AzureApplicationGatewaySslPolicy{
				PolicyType:        AzureApplicationGatewaySslPolicyType_PREDEFINED,
				PolicyName:        "AppGwSslPolicy20220101S",
				DisabledProtocols: []AzureApplicationGatewayTlsProtocol{AzureApplicationGatewayTlsProtocol_TLS_V1_0},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a CUSTOM SSL policy without cipher suites", func() {
			input := minimalSpec()
			input.Spec.SslPolicy = &AzureApplicationGatewaySslPolicy{
				PolicyType:         AzureApplicationGatewaySslPolicyType_CUSTOM,
				MinProtocolVersion: AzureApplicationGatewayTlsProtocol_TLS_V1_2,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a redirect with both listener and URL targets", func() {
			input := minimalSpec()
			input.Spec.RedirectConfigurations = []*AzureApplicationGatewayRedirectConfiguration{
				{Name: "r", RedirectType: AzureApplicationGatewayRedirectType_PERMANENT, TargetListenerName: "http-listener", TargetUrl: "https://contoso.com"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a TLS layer-4 listener without a certificate", func() {
			input := minimalSpec()
			input.Spec.Listeners = []*AzureApplicationGatewayLayer4Listener{
				{
					Name:                        "l4",
					FrontendIpConfigurationName: "public",
					FrontendPortName:            "http",
					Protocol:                    AzureApplicationGatewayProtocol_TLS,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject host names on a TCP layer-4 listener", func() {
			input := minimalSpec()
			input.Spec.Listeners = []*AzureApplicationGatewayLayer4Listener{
				{
					Name:                        "l4",
					FrontendIpConfigurationName: "public",
					FrontendPortName:            "http",
					Protocol:                    AzureApplicationGatewayProtocol_TCP,
					HostNames:                   []string{"x.contoso.com"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a host_name on TCP layer-4 backend settings", func() {
			input := minimalSpec()
			input.Spec.Backends = []*AzureApplicationGatewayLayer4BackendSettings{
				{Name: "b", Port: 5671, Protocol: AzureApplicationGatewayProtocol_TCP, HostName: "x"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a buffering block missing one field", func() {
			on := true
			input := minimalSpec()
			input.Spec.GlobalConfiguration = &AzureApplicationGatewayGlobalConfiguration{
				RequestBufferingEnabled: &on,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a gateway without frontends or ports", func() {
			input := minimalSpec()
			input.Spec.FrontendIpConfigurations = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())

			input = minimalSpec()
			input.Spec.FrontendPorts = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a gateway without listeners or routing rules", func() {
			input := minimalSpec()
			input.Spec.HttpListeners = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())

			input = minimalSpec()
			input.Spec.RequestRoutingRules = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid zone", func() {
			input := minimalSpec()
			input.Spec.Zones = []string{"4"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject host_name together with pick_host_name on backend settings", func() {
			input := minimalSpec()
			input.Spec.BackendHttpSettings[0].HostName = "app.internal"
			input.Spec.BackendHttpSettings[0].PickHostNameFromBackendAddress = true
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
