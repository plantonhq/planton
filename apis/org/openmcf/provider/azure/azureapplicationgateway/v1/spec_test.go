package azureapplicationgatewayv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzureApplicationGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureApplicationGatewaySpec Validation Tests")
}

// helper to create a StringValueOrRef with a literal value
func svr(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: val,
		},
	}
}

// helper to build a minimal valid AzureApplicationGateway
func minimalAppGateway() *AzureApplicationGateway {
	return &AzureApplicationGateway{
		ApiVersion: "azure.openmcf.org/v1",
		Kind:       "AzureApplicationGateway",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-agw",
		},
		Spec: &AzureApplicationGatewaySpec{
			Region:        "eastus",
			ResourceGroup: svr("my-rg"),
			Name:          "test-agw",
			SubnetId:      svr("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/agw-subnet"),
			PublicIpId:    svr("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/agw-pip"),
			Sku:           "Standard_v2",
			BackendAddressPools: []*AzureBackendAddressPool{
				{Name: "default", Fqdns: []string{"backend.contoso.com"}},
			},
			BackendHttpSettings: []*AzureBackendHttpSettings{
				{Name: "http-settings", Port: 80, Protocol: "Http"},
			},
			HttpListeners: []*AzureHttpListener{
				{Name: "http-listener", Port: 80, Protocol: "Http"},
			},
			RequestRoutingRules: []*AzureRequestRoutingRule{
				{
					Name:                    "http-rule",
					HttpListenerName:        "http-listener",
					BackendAddressPoolName:  "default",
					BackendHttpSettingsName: "http-settings",
					Priority:                100,
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureApplicationGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_application_gateway", func() {

			ginkgo.It("should not return a validation error for a minimal HTTP Application Gateway", func() {
				input := minimalAppGateway()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with WAF_v2 SKU and WAF enabled", func() {
				input := minimalAppGateway()
				input.Spec.Sku = "WAF_v2"
				wafEnabled := true
				input.Spec.WafEnabled = &wafEnabled
				wafMode := "Prevention"
				input.Spec.WafMode = &wafMode
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with WAF mode Detection", func() {
				input := minimalAppGateway()
				input.Spec.Sku = "WAF_v2"
				wafEnabled := true
				input.Spec.WafEnabled = &wafEnabled
				wafMode := "Detection"
				input.Spec.WafMode = &wafMode
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HTTPS listener and SSL certificate", func() {
				input := minimalAppGateway()
				input.Spec.SslCertificates = []*AzureSslCertificate{
					{
						Name:              "wildcard-cert",
						KeyVaultSecretId:  "https://my-vault.vault.azure.net/secrets/wildcard-cert",
					},
				}
				input.Spec.IdentityIds = []*foreignkeyv1.StringValueOrRef{
					svr("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/agw-identity"),
				}
				input.Spec.HttpListeners = append(input.Spec.HttpListeners, &AzureHttpListener{
					Name:               "https-listener",
					Port:               443,
					Protocol:           "Https",
					SslCertificateName: "wildcard-cert",
				})
				input.Spec.RequestRoutingRules = append(input.Spec.RequestRoutingRules, &AzureRequestRoutingRule{
					Name:                    "https-rule",
					HttpListenerName:        "https-listener",
					BackendAddressPoolName:  "default",
					BackendHttpSettingsName: "http-settings",
					Priority:                200,
				})
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with custom health probes", func() {
				input := minimalAppGateway()
				interval := int32(15)
				timeout := int32(10)
				threshold := int32(5)
				input.Spec.Probes = []*AzureHealthProbe{
					{
						Name:               "api-health",
						Protocol:           "Http",
						Path:               "/health",
						Host:               "api.contoso.com",
						Interval:           &interval,
						Timeout:            &timeout,
						UnhealthyThreshold: &threshold,
					},
				}
				input.Spec.BackendHttpSettings[0].ProbeName = "api-health"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with autoscale configuration", func() {
				input := minimalAppGateway()
				input.Spec.Capacity = nil // clear fixed capacity
				maxCap := int32(10)
				input.Spec.Autoscale = &AzureApplicationGatewayAutoscale{
					MinCapacity: 2,
					MaxCapacity: &maxCap,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple backend pools", func() {
				input := minimalAppGateway()
				input.Spec.BackendAddressPools = []*AzureBackendAddressPool{
					{Name: "api-pool", Fqdns: []string{"api.contoso.com"}},
					{Name: "web-pool", IpAddresses: []string{"10.0.1.4", "10.0.1.5"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with host-based routing", func() {
				input := minimalAppGateway()
				input.Spec.HttpListeners = []*AzureHttpListener{
					{Name: "api-listener", Port: 80, Protocol: "Http", HostName: "api.contoso.com"},
					{Name: "web-listener", Port: 80, Protocol: "Http", HostName: "www.contoso.com"},
				}
				input.Spec.BackendAddressPools = []*AzureBackendAddressPool{
					{Name: "api-pool", Fqdns: []string{"api-backend.internal"}},
					{Name: "web-pool", Fqdns: []string{"web-backend.internal"}},
				}
				input.Spec.BackendHttpSettings = []*AzureBackendHttpSettings{
					{Name: "http-settings", Port: 80, Protocol: "Http"},
				}
				input.Spec.RequestRoutingRules = []*AzureRequestRoutingRule{
					{Name: "api-rule", HttpListenerName: "api-listener", BackendAddressPoolName: "api-pool", BackendHttpSettingsName: "http-settings", Priority: 100},
					{Name: "web-rule", HttpListenerName: "web-listener", BackendAddressPoolName: "web-pool", BackendHttpSettingsName: "http-settings", Priority: 200},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backend HTTPS settings", func() {
				input := minimalAppGateway()
				requestTimeout := int32(60)
				pickHost := true
				input.Spec.BackendHttpSettings = []*AzureBackendHttpSettings{
					{
						Name:                             "https-backend",
						Port:                             443,
						Protocol:                         "Https",
						RequestTimeout:                   &requestTimeout,
						PickHostNameFromBackendAddress:   &pickHost,
					},
				}
				input.Spec.RequestRoutingRules[0].BackendHttpSettingsName = "https-backend"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HTTP/2 enabled", func() {
				input := minimalAppGateway()
				http2 := true
				input.Spec.EnableHttp2 = &http2
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with cookie-based affinity enabled", func() {
				input := minimalAppGateway()
				affinity := "Enabled"
				input.Spec.BackendHttpSettings[0].CookieBasedAffinity = &affinity
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with fixed capacity", func() {
				input := minimalAppGateway()
				cap := int32(5)
				input.Spec.Capacity = &cap
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom resource_group", func() {
				input := minimalAppGateway()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureResourceGroup,
							Name:      "prod-rg",
							FieldPath: "status.outputs.resource_group_name",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with Https health probe", func() {
				input := minimalAppGateway()
				input.Spec.Probes = []*AzureHealthProbe{
					{Name: "https-probe", Protocol: "Https", Path: "/healthz"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_application_gateway", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalAppGateway()
				input.ApiVersion = "wrong/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalAppGateway()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalAppGateway()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := minimalAppGateway()
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalAppGateway()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := minimalAppGateway()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is empty", func() {
				input := minimalAppGateway()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 80 characters", func() {
				input := minimalAppGateway()
				longName := ""
				for i := 0; i < 81; i++ {
					longName += "a"
				}
				input.Spec.Name = longName
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := minimalAppGateway()
				input.Spec.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when public_ip_id is missing", func() {
				input := minimalAppGateway()
				input.Spec.PublicIpId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sku is invalid", func() {
				input := minimalAppGateway()
				input.Spec.Sku = "Standard"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sku is empty", func() {
				input := minimalAppGateway()
				input.Spec.Sku = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when capacity exceeds 125", func() {
				input := minimalAppGateway()
				cap := int32(126)
				input.Spec.Capacity = &cap
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when capacity is 0", func() {
				input := minimalAppGateway()
				cap := int32(0)
				input.Spec.Capacity = &cap
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when autoscale min_capacity exceeds 100", func() {
				input := minimalAppGateway()
				input.Spec.Capacity = nil
				input.Spec.Autoscale = &AzureApplicationGatewayAutoscale{
					MinCapacity: 101,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when autoscale max_capacity exceeds 125", func() {
				input := minimalAppGateway()
				input.Spec.Capacity = nil
				maxCap := int32(126)
				input.Spec.Autoscale = &AzureApplicationGatewayAutoscale{
					MinCapacity: 2,
					MaxCapacity: &maxCap,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend_address_pools is empty", func() {
				input := minimalAppGateway()
				input.Spec.BackendAddressPools = []*AzureBackendAddressPool{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend pool name is empty", func() {
				input := minimalAppGateway()
				input.Spec.BackendAddressPools[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend_http_settings is empty", func() {
				input := minimalAppGateway()
				input.Spec.BackendHttpSettings = []*AzureBackendHttpSettings{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend HTTP settings name is empty", func() {
				input := minimalAppGateway()
				input.Spec.BackendHttpSettings[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend HTTP settings port is 0", func() {
				input := minimalAppGateway()
				input.Spec.BackendHttpSettings[0].Port = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend HTTP settings port exceeds 65535", func() {
				input := minimalAppGateway()
				input.Spec.BackendHttpSettings[0].Port = 65536
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend HTTP settings protocol is invalid", func() {
				input := minimalAppGateway()
				input.Spec.BackendHttpSettings[0].Protocol = "TCP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cookie_based_affinity is invalid", func() {
				input := minimalAppGateway()
				badAffinity := "SomeValue"
				input.Spec.BackendHttpSettings[0].CookieBasedAffinity = &badAffinity
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when request_timeout exceeds 86400", func() {
				input := minimalAppGateway()
				timeout := int32(86401)
				input.Spec.BackendHttpSettings[0].RequestTimeout = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when http_listeners is empty", func() {
				input := minimalAppGateway()
				input.Spec.HttpListeners = []*AzureHttpListener{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener name is empty", func() {
				input := minimalAppGateway()
				input.Spec.HttpListeners[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener port is 0", func() {
				input := minimalAppGateway()
				input.Spec.HttpListeners[0].Port = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener protocol is invalid", func() {
				input := minimalAppGateway()
				input.Spec.HttpListeners[0].Protocol = "TCP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when request_routing_rules is empty", func() {
				input := minimalAppGateway()
				input.Spec.RequestRoutingRules = []*AzureRequestRoutingRule{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when routing rule name is empty", func() {
				input := minimalAppGateway()
				input.Spec.RequestRoutingRules[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when routing rule http_listener_name is empty", func() {
				input := minimalAppGateway()
				input.Spec.RequestRoutingRules[0].HttpListenerName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when routing rule backend_address_pool_name is empty", func() {
				input := minimalAppGateway()
				input.Spec.RequestRoutingRules[0].BackendAddressPoolName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when routing rule backend_http_settings_name is empty", func() {
				input := minimalAppGateway()
				input.Spec.RequestRoutingRules[0].BackendHttpSettingsName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when routing rule priority is 0", func() {
				input := minimalAppGateway()
				input.Spec.RequestRoutingRules[0].Priority = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when routing rule priority exceeds 20000", func() {
				input := minimalAppGateway()
				input.Spec.RequestRoutingRules[0].Priority = 20001
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health probe name is empty", func() {
				input := minimalAppGateway()
				input.Spec.Probes = []*AzureHealthProbe{
					{Name: "", Protocol: "Http", Path: "/health"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health probe protocol is invalid", func() {
				input := minimalAppGateway()
				input.Spec.Probes = []*AzureHealthProbe{
					{Name: "bad-probe", Protocol: "Tcp", Path: "/health"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health probe path is empty", func() {
				input := minimalAppGateway()
				input.Spec.Probes = []*AzureHealthProbe{
					{Name: "bad-probe", Protocol: "Http", Path: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health probe unhealthy_threshold exceeds 20", func() {
				input := minimalAppGateway()
				threshold := int32(21)
				input.Spec.Probes = []*AzureHealthProbe{
					{Name: "bad-probe", Protocol: "Http", Path: "/health", UnhealthyThreshold: &threshold},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when SSL certificate name is empty", func() {
				input := minimalAppGateway()
				input.Spec.SslCertificates = []*AzureSslCertificate{
					{Name: "", KeyVaultSecretId: "https://vault.vault.azure.net/secrets/cert"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when SSL certificate key_vault_secret_id is empty", func() {
				input := minimalAppGateway()
				input.Spec.SslCertificates = []*AzureSslCertificate{
					{Name: "my-cert", KeyVaultSecretId: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when waf_mode is invalid", func() {
				input := minimalAppGateway()
				input.Spec.Sku = "WAF_v2"
				wafEnabled := true
				input.Spec.WafEnabled = &wafEnabled
				badMode := "Monitor"
				input.Spec.WafMode = &badMode
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
