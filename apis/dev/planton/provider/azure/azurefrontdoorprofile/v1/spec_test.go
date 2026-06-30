package azurefrontdoorprofilev1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureFrontDoorProfileSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorProfileSpec Validation Tests")
}

// helper to create a minimal valid spec
func minimalSpec() *AzureFrontDoorProfile {
	return &AzureFrontDoorProfile{
		ApiVersion: "azure.planton.dev/v1",
		Kind:       "AzureFrontDoorProfile",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-fd",
		},
		Spec: &AzureFrontDoorProfileSpec{
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name: "my-frontdoor",
			Endpoints: []*AzureFrontDoorEndpoint{
				{Name: "my-endpoint"},
			},
			OriginGroups: []*AzureFrontDoorOriginGroup{
				{
					Name:          "my-origins",
					LoadBalancing: &AzureFrontDoorLoadBalancing{},
					Origins: []*AzureFrontDoorOrigin{
						{Name: "backend", HostName: "api.example.com"},
					},
				},
			},
			Routes: []*AzureFrontDoorRoute{
				{
					Name:               "default-route",
					EndpointName:       "my-endpoint",
					OriginGroupName:    "my-origins",
					PatternsToMatch:    []string{"/*"},
					SupportedProtocols: []string{"Https"},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureFrontDoorProfileSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_front_door_profile", func() {

			ginkgo.It("should not return a validation error for a minimal valid spec", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Standard_AzureFrontDoor SKU", func() {
				sku := "Standard_AzureFrontDoor"
				input := minimalSpec()
				input.Spec.Sku = &sku
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Premium_AzureFrontDoor SKU", func() {
				sku := "Premium_AzureFrontDoor"
				input := minimalSpec()
				input.Spec.Sku = &sku
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for response timeout at minimum boundary (16)", func() {
				timeout := int32(16)
				input := minimalSpec()
				input.Spec.ResponseTimeoutSeconds = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for response timeout at 120", func() {
				timeout := int32(120)
				input := minimalSpec()
				input.Spec.ResponseTimeoutSeconds = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for response timeout at maximum boundary (240)", func() {
				timeout := int32(240)
				input := minimalSpec()
				input.Spec.ResponseTimeoutSeconds = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for endpoint with enabled=false", func() {
				enabled := false
				input := minimalSpec()
				input.Spec.Endpoints[0].Enabled = &enabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for origin group with session_affinity_enabled=false", func() {
				sessionAffinity := false
				input := minimalSpec()
				input.Spec.OriginGroups[0].SessionAffinityEnabled = &sessionAffinity
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for full load balancing config", func() {
				sampleSize := int32(10)
				successfulSamples := int32(5)
				latency := int32(100)
				input := minimalSpec()
				input.Spec.OriginGroups[0].LoadBalancing = &AzureFrontDoorLoadBalancing{
					SampleSize:                      &sampleSize,
					SuccessfulSamplesRequired:       &successfulSamples,
					AdditionalLatencyInMilliseconds: &latency,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for health probe with all fields (Https, /health, GET, 30)", func() {
				path := "/health"
				requestType := "GET"
				input := minimalSpec()
				input.Spec.OriginGroups[0].HealthProbe = &AzureFrontDoorHealthProbe{
					Protocol:          "Https",
					Path:              &path,
					RequestType:       &requestType,
					IntervalInSeconds: 30,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for health probe with Http protocol", func() {
				input := minimalSpec()
				input.Spec.OriginGroups[0].HealthProbe = &AzureFrontDoorHealthProbe{
					Protocol:          "Http",
					IntervalInSeconds: 30,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for origin with all optional fields", func() {
				originHostHeader := "api.example.com"
				httpPort := int32(8080)
				httpsPort := int32(8443)
				priority := int32(2)
				weight := int32(700)
				enabled := false
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins = []*AzureFrontDoorOrigin{
					{
						Name:             "full-origin",
						HostName:         "api.example.com",
						OriginHostHeader: &originHostHeader,
						HttpPort:         &httpPort,
						HttpsPort:        &httpsPort,
						Priority:         &priority,
						Weight:           &weight,
						Enabled:          &enabled,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for origin priority at min boundary (1)", func() {
				priority := int32(1)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].Priority = &priority
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for origin priority at max boundary (5)", func() {
				priority := int32(5)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].Priority = &priority
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for origin weight at min boundary (1)", func() {
				weight := int32(1)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].Weight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for origin weight at max boundary (1000)", func() {
				weight := int32(1000)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].Weight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for origin port at min boundary (1)", func() {
				httpPort := int32(1)
				httpsPort := int32(1)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].HttpPort = &httpPort
				input.Spec.OriginGroups[0].Origins[0].HttpsPort = &httpsPort
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for origin port at max boundary (65535)", func() {
				httpPort := int32(65535)
				httpsPort := int32(65535)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].HttpPort = &httpPort
				input.Spec.OriginGroups[0].Origins[0].HttpsPort = &httpsPort
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for private link on origin", func() {
				targetType := "sites"
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].PrivateLink = &AzureFrontDoorPrivateLink{
					Location:            "eastus",
					PrivateLinkTargetId: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Web/sites/myapp",
					TargetType:          &targetType,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for route with cache enabled", func() {
				compressionEnabled := true
				qsBehavior := "UseQueryString"
				input := minimalSpec()
				input.Spec.Routes[0].Cache = &AzureFrontDoorRouteCache{
					CompressionEnabled:         &compressionEnabled,
					QueryStringCachingBehavior: &qsBehavior,
					QueryStrings:               []string{"utm_source", "utm_medium"},
					ContentTypesToCompress:     []string{"text/html", "application/json"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for route with forwarding_protocol HttpsOnly", func() {
				fwdProtocol := "HttpsOnly"
				input := minimalSpec()
				input.Spec.Routes[0].ForwardingProtocol = &fwdProtocol
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for route with forwarding_protocol HttpOnly", func() {
				fwdProtocol := "HttpOnly"
				input := minimalSpec()
				input.Spec.Routes[0].ForwardingProtocol = &fwdProtocol
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for route with https_redirect_enabled=false", func() {
				httpsRedirect := false
				input := minimalSpec()
				input.Spec.Routes[0].HttpsRedirectEnabled = &httpsRedirect
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for route with link_to_default_domain=false", func() {
				linkToDefault := false
				input := minimalSpec()
				input.Spec.Routes[0].LinkToDefaultDomain = &linkToDefault
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for multiple endpoints, origin groups, and routes", func() {
				input := minimalSpec()
				input.Spec.Endpoints = []*AzureFrontDoorEndpoint{
					{Name: "endpoint-a1"},
					{Name: "endpoint-b2"},
				}
				input.Spec.OriginGroups = []*AzureFrontDoorOriginGroup{
					{
						Name:          "group-api",
						LoadBalancing: &AzureFrontDoorLoadBalancing{},
						Origins: []*AzureFrontDoorOrigin{
							{Name: "api-primary", HostName: "api-primary.example.com"},
							{Name: "api-secondary", HostName: "api-secondary.example.com"},
						},
					},
					{
						Name:          "group-web",
						LoadBalancing: &AzureFrontDoorLoadBalancing{},
						Origins: []*AzureFrontDoorOrigin{
							{Name: "web-backend", HostName: "web.example.com"},
						},
					},
				}
				input.Spec.Routes = []*AzureFrontDoorRoute{
					{
						Name:               "api-route",
						EndpointName:       "endpoint-a1",
						OriginGroupName:    "group-api",
						PatternsToMatch:    []string{"/api/*"},
						SupportedProtocols: []string{"Https"},
					},
					{
						Name:               "web-route",
						EndpointName:       "endpoint-b2",
						OriginGroupName:    "group-web",
						PatternsToMatch:    []string{"/*"},
						SupportedProtocols: []string{"Http", "Https"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for profile name at min length (2 chars)", func() {
				input := minimalSpec()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for profile name at max length (46 chars)", func() {
				input := minimalSpec()
				// 46 chars: starts with letter, ends with number, contains hyphens
				input.Spec.Name = "a-very-long-front-door-profile-name-for-test01" // 46 chars
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for endpoint name at min length (2 chars)", func() {
				input := minimalSpec()
				input.Spec.Endpoints[0].Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for endpoint name at max length (46 chars)", func() {
				input := minimalSpec()
				input.Spec.Endpoints[0].Name = "a-very-long-front-door-endpoint-name-for-test1" // 46 chars
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for resource_group", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureResourceGroup,
							Name:      "shared-rg",
							FieldPath: "status.outputs.resource_group_name",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for each valid query_string_caching_behavior", func() {
				behaviors := []string{
					"IgnoreQueryString",
					"UseQueryString",
					"IgnoreSpecifiedQueryStrings",
					"IncludeSpecifiedQueryStrings",
				}
				for _, b := range behaviors {
					behavior := b
					input := minimalSpec()
					input.Spec.Routes[0].Cache = &AzureFrontDoorRouteCache{
						QueryStringCachingBehavior: &behavior,
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for load balancing at boundary values", func() {
				sampleSizeMin := int32(0)
				successfulMax := int32(255)
				latencyMax := int32(1000)
				input := minimalSpec()
				input.Spec.OriginGroups[0].LoadBalancing = &AzureFrontDoorLoadBalancing{
					SampleSize:                      &sampleSizeMin,
					SuccessfulSamplesRequired:       &successfulMax,
					AdditionalLatencyInMilliseconds: &latencyMax,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_front_door_profile", func() {

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := minimalSpec()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is too short (1 char)", func() {
				input := minimalSpec()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is too long (47 chars)", func() {
				input := minimalSpec()
				input.Spec.Name = "a" + strings.Repeat("b", 45) + "c" // 47 chars
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "-my-frontdoor"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name ends with hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "my-frontdoor-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with special char", func() {
				input := minimalSpec()
				input.Spec.Name = "@my-frontdoor"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sku is invalid", func() {
				invalidSku := "Enterprise"
				input := minimalSpec()
				input.Spec.Sku = &invalidSku
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when response_timeout_seconds is too low (15)", func() {
				timeout := int32(15)
				input := minimalSpec()
				input.Spec.ResponseTimeoutSeconds = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when response_timeout_seconds is too high (241)", func() {
				timeout := int32(241)
				input := minimalSpec()
				input.Spec.ResponseTimeoutSeconds = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health probe protocol is invalid", func() {
				input := minimalSpec()
				input.Spec.OriginGroups[0].HealthProbe = &AzureFrontDoorHealthProbe{
					Protocol:          "TCP",
					IntervalInSeconds: 30,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health probe request_type is invalid", func() {
				requestType := "POST"
				input := minimalSpec()
				input.Spec.OriginGroups[0].HealthProbe = &AzureFrontDoorHealthProbe{
					Protocol:          "Https",
					RequestType:       &requestType,
					IntervalInSeconds: 30,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health probe interval is too low (0)", func() {
				input := minimalSpec()
				input.Spec.OriginGroups[0].HealthProbe = &AzureFrontDoorHealthProbe{
					Protocol:          "Https",
					IntervalInSeconds: 0,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health probe interval is too high (256)", func() {
				input := minimalSpec()
				input.Spec.OriginGroups[0].HealthProbe = &AzureFrontDoorHealthProbe{
					Protocol:          "Https",
					IntervalInSeconds: 256,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when origin name is missing", func() {
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins = []*AzureFrontDoorOrigin{
					{Name: "", HostName: "api.example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when origin host_name is missing", func() {
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins = []*AzureFrontDoorOrigin{
					{Name: "backend", HostName: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when origin http_port is too low (0)", func() {
				httpPort := int32(0)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].HttpPort = &httpPort
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when origin http_port is too high (65536)", func() {
				httpPort := int32(65536)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].HttpPort = &httpPort
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when origin priority is too low (0)", func() {
				priority := int32(0)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].Priority = &priority
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when origin priority is too high (6)", func() {
				priority := int32(6)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].Priority = &priority
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when origin weight is too low (0)", func() {
				weight := int32(0)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].Weight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when origin weight is too high (1001)", func() {
				weight := int32(1001)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].Weight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when private link location is missing", func() {
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].PrivateLink = &AzureFrontDoorPrivateLink{
					Location:            "",
					PrivateLinkTargetId: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Web/sites/myapp",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when private link target_id is missing", func() {
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].PrivateLink = &AzureFrontDoorPrivateLink{
					Location:            "eastus",
					PrivateLinkTargetId: "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when private link request_message is too long (141 chars)", func() {
				longMessage := strings.Repeat("a", 141)
				input := minimalSpec()
				input.Spec.OriginGroups[0].Origins[0].PrivateLink = &AzureFrontDoorPrivateLink{
					Location:            "eastus",
					PrivateLinkTargetId: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Web/sites/myapp",
					RequestMessage:      &longMessage,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when forwarding_protocol is invalid", func() {
				fwdProtocol := "TCP"
				input := minimalSpec()
				input.Spec.Routes[0].ForwardingProtocol = &fwdProtocol
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when query_string_caching_behavior is invalid", func() {
				qsBehavior := "InvalidBehavior"
				input := minimalSpec()
				input.Spec.Routes[0].Cache = &AzureFrontDoorRouteCache{
					QueryStringCachingBehavior: &qsBehavior,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when load balancing sample_size is too high (256)", func() {
				sampleSize := int32(256)
				input := minimalSpec()
				input.Spec.OriginGroups[0].LoadBalancing = &AzureFrontDoorLoadBalancing{
					SampleSize: &sampleSize,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when load balancing additional_latency_in_milliseconds is too high (1001)", func() {
				latency := int32(1001)
				input := minimalSpec()
				input.Spec.OriginGroups[0].LoadBalancing = &AzureFrontDoorLoadBalancing{
					AdditionalLatencyInMilliseconds: &latency,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalSpec()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureFrontDoorProfile{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureFrontDoorProfile",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-fd",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := minimalSpec()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := minimalSpec()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when endpoint name is too short (1 char)", func() {
				input := minimalSpec()
				input.Spec.Endpoints[0].Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when endpoint name is too long (47 chars)", func() {
				input := minimalSpec()
				input.Spec.Endpoints[0].Name = "a" + strings.Repeat("b", 45) + "c" // 47 chars
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when route endpoint_name is missing", func() {
				input := minimalSpec()
				input.Spec.Routes[0].EndpointName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when route origin_group_name is missing", func() {
				input := minimalSpec()
				input.Spec.Routes[0].OriginGroupName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
