package ociapigatewayv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestOciApiGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciApiGatewaySpec Validation Tests")
}

func svr(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func svrRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func minimalValid() *OciApiGateway {
	return &OciApiGateway{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciApiGateway",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-api-gw"},
		Spec: &OciApiGatewaySpec{
			CompartmentId: svr("ocid1.compartment.oc1..example"),
			EndpointType:  OciApiGatewaySpec_public,
			SubnetId:      svr("ocid1.subnet.oc1..example"),
			Deployment: &OciApiGatewaySpec_Deployment{
				PathPrefix: "/api",
				Routes: []*OciApiGatewaySpec_Route{
					{
						Path: "/health",
						Backend: &OciApiGatewaySpec_Backend{
							Type:   OciApiGatewaySpec_stock_response,
							Status: 200,
							Body:   `{"status":"ok"}`,
						},
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("OciApiGatewaySpec Validation Tests", func() {

	// ── Valid Scenarios ──────────────────────────────────────────────

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gateway configuration", func() {

			ginkgo.It("should accept minimal gateway with stock response route", func() {
				gomega.Expect(protovalidate.Validate(minimalValid())).To(gomega.BeNil())
			})

			ginkgo.It("should accept public endpoint type", func() {
				input := minimalValid()
				input.Spec.EndpointType = OciApiGatewaySpec_public
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept private endpoint type", func() {
				input := minimalValid()
				input.Spec.EndpointType = OciApiGatewaySpec_private
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept display_name on gateway", func() {
				input := minimalValid()
				input.Spec.DisplayName = "production-gateway"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept certificate_id", func() {
				input := minimalValid()
				input.Spec.CertificateId = "ocid1.certificate.oc1..example"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept network security group IDs", func() {
				input := minimalValid()
				input.Spec.NetworkSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					svr("ocid1.networksecuritygroup.oc1..example1"),
					svr("ocid1.networksecuritygroup.oc1..example2"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept compartment_id via valueFrom ref", func() {
				input := minimalValid()
				input.Spec.CompartmentId = svrRef("my-compartment")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept subnet_id via valueFrom ref", func() {
				input := minimalValid()
				input.Spec.SubnetId = svrRef("my-subnet")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("deployment configuration", func() {

			ginkgo.It("should accept deployment display_name", func() {
				input := minimalValid()
				input.Spec.Deployment.DisplayName = "api-v1"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept logging policies", func() {
				input := minimalValid()
				input.Spec.Deployment.LoggingPolicies = &OciApiGatewaySpec_LoggingPolicies{
					AccessLog:    &OciApiGatewaySpec_AccessLog{IsEnabled: true},
					ExecutionLog: &OciApiGatewaySpec_ExecutionLog{IsEnabled: true, LogLevel: OciApiGatewaySpec_info},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept execution log with warn level", func() {
				input := minimalValid()
				input.Spec.Deployment.LoggingPolicies = &OciApiGatewaySpec_LoggingPolicies{
					ExecutionLog: &OciApiGatewaySpec_ExecutionLog{IsEnabled: true, LogLevel: OciApiGatewaySpec_warn},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept execution log with error level", func() {
				input := minimalValid()
				input.Spec.Deployment.LoggingPolicies = &OciApiGatewaySpec_LoggingPolicies{
					ExecutionLog: &OciApiGatewaySpec_ExecutionLog{IsEnabled: true, LogLevel: OciApiGatewaySpec_error},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("backend types", func() {

			ginkgo.It("should accept HTTP backend with URL", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes = []*OciApiGatewaySpec_Route{
					{
						Path:    "/proxy",
						Methods: []string{"GET", "POST"},
						Backend: &OciApiGatewaySpec_Backend{
							Type: OciApiGatewaySpec_http,
							Url:  "https://backend.example.com:8080",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept Oracle Functions backend", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes = []*OciApiGatewaySpec_Route{
					{
						Path: "/invoke",
						Backend: &OciApiGatewaySpec_Backend{
							Type:       OciApiGatewaySpec_oracle_functions,
							FunctionId: "ocid1.fnfunc.oc1..example",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept stock response backend", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes = []*OciApiGatewaySpec_Route{
					{
						Path: "/status",
						Backend: &OciApiGatewaySpec_Backend{
							Type:   OciApiGatewaySpec_stock_response,
							Status: 200,
							Body:   `{"status":"healthy"}`,
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept HTTP backend with timeouts and headers", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes = []*OciApiGatewaySpec_Route{
					{
						Path: "/slow-api",
						Backend: &OciApiGatewaySpec_Backend{
							Type:                    OciApiGatewaySpec_http,
							Url:                     "https://backend.example.com",
							ConnectTimeoutInSeconds: proto.Float32(5.0),
							ReadTimeoutInSeconds:    proto.Float32(30.0),
							SendTimeoutInSeconds:    proto.Float32(15.0),
							IsSslVerifyDisabled:     proto.Bool(true),
							Headers: []*OciApiGatewaySpec_BackendHeader{
								{Name: "X-Custom-Header", Value: "custom-value"},
								{Name: "X-Request-Source", Value: "api-gateway"},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("CORS policy", func() {

			ginkgo.It("should accept CORS with all origins", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Cors: &OciApiGatewaySpec_CorsPolicy{
						AllowedOrigins: []string{"*"},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept full CORS configuration", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Cors: &OciApiGatewaySpec_CorsPolicy{
						AllowedOrigins:            []string{"https://app.example.com", "https://admin.example.com"},
						AllowedMethods:            []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
						AllowedHeaders:            []string{"Content-Type", "Authorization", "X-Request-ID"},
						ExposedHeaders:            []string{"X-Total-Count", "X-Rate-Limit-Remaining"},
						IsAllowCredentialsEnabled: proto.Bool(true),
						MaxAgeInSeconds:           proto.Int32(3600),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("rate limiting", func() {

			ginkgo.It("should accept rate limiting with client_ip", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					RateLimiting: &OciApiGatewaySpec_RateLimiting{
						RateInRequestsPerSecond: 100,
						RateKey:                 OciApiGatewaySpec_client_ip,
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept rate limiting with total", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					RateLimiting: &OciApiGatewaySpec_RateLimiting{
						RateInRequestsPerSecond: 1000,
						RateKey:                 OciApiGatewaySpec_total,
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("JWT authentication", func() {

			ginkgo.It("should accept JWT auth with remote JWKS", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						Issuers:   []string{"https://idcs.example.com/"},
						Audiences: []string{"https://api.example.com"},
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_remote_jwks,
							Uri:  "https://idcs.example.com/.well-known/jwks.json",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept JWT auth with remote JWKS and cache duration", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						Issuers:   []string{"https://auth0.example.com/"},
						Audiences: []string{"https://api.example.com"},
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type:                    OciApiGatewaySpec_remote_jwks,
							Uri:                     "https://auth0.example.com/.well-known/jwks.json",
							IsSslVerifyDisabled:     proto.Bool(false),
							MaxCacheDurationInHours: proto.Int32(24),
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept JWT auth with static PEM key", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						Issuers:   []string{"https://test.example.com/"},
						Audiences: []string{"test-api"},
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_static_keys,
							Keys: []*OciApiGatewaySpec_StaticKey{
								{
									Kid:    "key-1",
									Format: OciApiGatewaySpec_pem,
									Key:    "-----BEGIN PUBLIC KEY-----\nMIIBIjANBg...\n-----END PUBLIC KEY-----",
								},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept JWT auth with static JWK key", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						Issuers:   []string{"https://test.example.com/"},
						Audiences: []string{"test-api"},
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_static_keys,
							Keys: []*OciApiGatewaySpec_StaticKey{
								{
									Kid:    "rsa-key-1",
									Format: OciApiGatewaySpec_json_web_key,
									Kty:    "RSA",
									Alg:    "RS256",
									N:      "0vx7agoebGc...base64url",
									E:      "AQAB",
									Use:    "sig",
								},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept JWT auth with verify claims", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						Issuers:   []string{"https://idcs.example.com/"},
						Audiences: []string{"https://api.example.com"},
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_remote_jwks,
							Uri:  "https://idcs.example.com/.well-known/jwks.json",
						},
						VerifyClaims: []*OciApiGatewaySpec_VerifyClaim{
							{Key: "email", IsRequired: proto.Bool(true)},
							{Key: "groups", Values: []string{"admin", "users"}},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept JWT auth with anonymous access allowed", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						IsAnonymousAccessAllowed: proto.Bool(true),
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_remote_jwks,
							Uri:  "https://idcs.example.com/.well-known/jwks.json",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept JWT auth with token header and scheme", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						TokenHeader:     "X-API-Token",
						TokenAuthScheme: "Token",
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_remote_jwks,
							Uri:  "https://idcs.example.com/.well-known/jwks.json",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept JWT auth with token query param", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						TokenQueryParam: "access_token",
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_remote_jwks,
							Uri:  "https://idcs.example.com/.well-known/jwks.json",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept JWT auth with max clock skew", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						MaxClockSkewInSeconds: proto.Float32(60),
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_remote_jwks,
							Uri:  "https://idcs.example.com/.well-known/jwks.json",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("route authorization", func() {

			ginkgo.It("should accept anonymous route authorization", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Authorization = &OciApiGatewaySpec_RouteAuthorization{
					Type: OciApiGatewaySpec_anonymous,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept authentication_only route authorization", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Authorization = &OciApiGatewaySpec_RouteAuthorization{
					Type: OciApiGatewaySpec_authentication_only,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept any_of route authorization with scopes", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Authorization = &OciApiGatewaySpec_RouteAuthorization{
					Type:         OciApiGatewaySpec_any_of,
					AllowedScope: []string{"read:users", "admin"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("per-route logging", func() {

			ginkgo.It("should accept per-route logging override", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].LoggingPolicies = &OciApiGatewaySpec_LoggingPolicies{
					AccessLog:    &OciApiGatewaySpec_AccessLog{IsEnabled: true},
					ExecutionLog: &OciApiGatewaySpec_ExecutionLog{IsEnabled: true, LogLevel: OciApiGatewaySpec_error},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("multiple routes", func() {

			ginkgo.It("should accept multiple routes with different backend types", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes = []*OciApiGatewaySpec_Route{
					{
						Path:    "/health",
						Methods: []string{"GET"},
						Backend: &OciApiGatewaySpec_Backend{
							Type:   OciApiGatewaySpec_stock_response,
							Status: 200,
							Body:   `{"status":"ok"}`,
						},
						Authorization: &OciApiGatewaySpec_RouteAuthorization{
							Type: OciApiGatewaySpec_anonymous,
						},
					},
					{
						Path:    "/users/{userId}",
						Methods: []string{"GET", "PUT"},
						Backend: &OciApiGatewaySpec_Backend{
							Type: OciApiGatewaySpec_http,
							Url:  "https://users-service.internal:8080",
						},
						Authorization: &OciApiGatewaySpec_RouteAuthorization{
							Type:         OciApiGatewaySpec_any_of,
							AllowedScope: []string{"user:read", "user:write"},
						},
					},
					{
						Path:    "/process",
						Methods: []string{"POST"},
						Backend: &OciApiGatewaySpec_Backend{
							Type:       OciApiGatewaySpec_oracle_functions,
							FunctionId: "ocid1.fnfunc.oc1..processor",
						},
						Authorization: &OciApiGatewaySpec_RouteAuthorization{
							Type: OciApiGatewaySpec_authentication_only,
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("full configuration", func() {

			ginkgo.It("should accept full configuration with all features", func() {
				input := minimalValid()
				input.Spec.DisplayName = "production-api-gateway"
				input.Spec.CertificateId = "ocid1.certificate.oc1..example"
				input.Spec.NetworkSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					svr("ocid1.networksecuritygroup.oc1..example"),
				}
				input.Spec.Deployment.DisplayName = "api-v1"
				input.Spec.Deployment.PathPrefix = "/api/v1"
				input.Spec.Deployment.LoggingPolicies = &OciApiGatewaySpec_LoggingPolicies{
					AccessLog:    &OciApiGatewaySpec_AccessLog{IsEnabled: true},
					ExecutionLog: &OciApiGatewaySpec_ExecutionLog{IsEnabled: true, LogLevel: OciApiGatewaySpec_info},
				}
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						Issuers:                  []string{"https://idcs.example.com/"},
						Audiences:                []string{"https://api.example.com"},
						TokenHeader:              "Authorization",
						TokenAuthScheme:          "Bearer",
						IsAnonymousAccessAllowed: proto.Bool(false),
						MaxClockSkewInSeconds:    proto.Float32(30),
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type:                    OciApiGatewaySpec_remote_jwks,
							Uri:                     "https://idcs.example.com/.well-known/jwks.json",
							MaxCacheDurationInHours: proto.Int32(12),
						},
						VerifyClaims: []*OciApiGatewaySpec_VerifyClaim{
							{Key: "email", IsRequired: proto.Bool(true)},
						},
					},
					Cors: &OciApiGatewaySpec_CorsPolicy{
						AllowedOrigins:            []string{"https://app.example.com"},
						AllowedMethods:            []string{"GET", "POST", "PUT", "DELETE"},
						AllowedHeaders:            []string{"Content-Type", "Authorization"},
						IsAllowCredentialsEnabled: proto.Bool(true),
						MaxAgeInSeconds:           proto.Int32(3600),
					},
					RateLimiting: &OciApiGatewaySpec_RateLimiting{
						RateInRequestsPerSecond: 500,
						RateKey:                 OciApiGatewaySpec_client_ip,
					},
				}
				input.Spec.Deployment.Routes = []*OciApiGatewaySpec_Route{
					{
						Path:    "/health",
						Methods: []string{"GET"},
						Backend: &OciApiGatewaySpec_Backend{
							Type:   OciApiGatewaySpec_stock_response,
							Status: 200,
							Body:   `{"status":"ok"}`,
						},
						Authorization: &OciApiGatewaySpec_RouteAuthorization{Type: OciApiGatewaySpec_anonymous},
					},
					{
						Path:    "/users",
						Methods: []string{"GET", "POST"},
						Backend: &OciApiGatewaySpec_Backend{
							Type:                    OciApiGatewaySpec_http,
							Url:                     "https://users.internal:8080",
							ConnectTimeoutInSeconds: proto.Float32(5),
							ReadTimeoutInSeconds:    proto.Float32(30),
						},
						Authorization: &OciApiGatewaySpec_RouteAuthorization{
							Type:         OciApiGatewaySpec_any_of,
							AllowedScope: []string{"user:read", "user:write"},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	// ── Invalid Scenarios ────────────────────────────────────────────

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("api-level validation", func() {

			ginkgo.It("should reject wrong api_version", func() {
				input := minimalValid()
				input.ApiVersion = "wrong.openmcf.org/v1"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject wrong kind", func() {
				input := minimalValid()
				input.Kind = "WrongKind"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject missing metadata", func() {
				input := minimalValid()
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject missing spec", func() {
				input := &OciApiGateway{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciApiGateway",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("gateway field validation", func() {

			ginkgo.It("should reject missing compartment_id", func() {
				input := minimalValid()
				input.Spec.CompartmentId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject unspecified endpoint_type", func() {
				input := minimalValid()
				input.Spec.EndpointType = OciApiGatewaySpec_unspecified
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject missing subnet_id", func() {
				input := minimalValid()
				input.Spec.SubnetId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("deployment validation", func() {

			ginkgo.It("should reject missing deployment", func() {
				input := minimalValid()
				input.Spec.Deployment = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject empty path_prefix", func() {
				input := minimalValid()
				input.Spec.Deployment.PathPrefix = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject path_prefix without leading slash", func() {
				input := minimalValid()
				input.Spec.Deployment.PathPrefix = "api/v1"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject empty routes list", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes = []*OciApiGatewaySpec_Route{}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject nil routes", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("route and backend validation", func() {

			ginkgo.It("should reject route with empty path", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Path = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject route with missing backend", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Backend = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject backend with unspecified type", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Backend.Type = OciApiGatewaySpec_backend_unspecified
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject HTTP backend without URL", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Backend = &OciApiGatewaySpec_Backend{
					Type: OciApiGatewaySpec_http,
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject Functions backend without function_id", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Backend = &OciApiGatewaySpec_Backend{
					Type: OciApiGatewaySpec_oracle_functions,
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject backend header with empty name", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Backend = &OciApiGatewaySpec_Backend{
					Type: OciApiGatewaySpec_http,
					Url:  "https://example.com",
					Headers: []*OciApiGatewaySpec_BackendHeader{
						{Name: "", Value: "value"},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("authentication validation", func() {

			ginkgo.It("should reject authentication without public_keys", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						Issuers:   []string{"https://idcs.example.com/"},
						Audiences: []string{"https://api.example.com"},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject public_keys with unspecified type", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_public_key_unspecified,
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject remote_jwks without URI", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_remote_jwks,
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject static_keys with empty keys list", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_static_keys,
							Keys: []*OciApiGatewaySpec_StaticKey{},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject static key with empty kid", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_static_keys,
							Keys: []*OciApiGatewaySpec_StaticKey{
								{Kid: "", Format: OciApiGatewaySpec_pem, Key: "-----BEGIN PUBLIC KEY-----\nMIIB...\n-----END PUBLIC KEY-----"},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject static key with unspecified format", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_static_keys,
							Keys: []*OciApiGatewaySpec_StaticKey{
								{Kid: "key-1", Format: OciApiGatewaySpec_key_format_unspecified, Key: "some-key"},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject PEM static key without key value", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_static_keys,
							Keys: []*OciApiGatewaySpec_StaticKey{
								{Kid: "key-1", Format: OciApiGatewaySpec_pem},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject JWK static key without kty", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Authentication: &OciApiGatewaySpec_Authentication{
						PublicKeys: &OciApiGatewaySpec_PublicKeys{
							Type: OciApiGatewaySpec_static_keys,
							Keys: []*OciApiGatewaySpec_StaticKey{
								{Kid: "key-1", Format: OciApiGatewaySpec_json_web_key, Alg: "RS256"},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("CORS and rate limiting validation", func() {

			ginkgo.It("should reject CORS with empty allowed_origins", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					Cors: &OciApiGatewaySpec_CorsPolicy{
						AllowedOrigins: []string{},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject rate limiting with zero rate", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					RateLimiting: &OciApiGatewaySpec_RateLimiting{
						RateInRequestsPerSecond: 0,
						RateKey:                 OciApiGatewaySpec_client_ip,
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject rate limiting with unspecified rate_key", func() {
				input := minimalValid()
				input.Spec.Deployment.RequestPolicies = &OciApiGatewaySpec_RequestPolicies{
					RateLimiting: &OciApiGatewaySpec_RateLimiting{
						RateInRequestsPerSecond: 100,
						RateKey:                 OciApiGatewaySpec_rate_key_unspecified,
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("route authorization validation", func() {

			ginkgo.It("should reject any_of authorization without allowed_scope", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Authorization = &OciApiGatewaySpec_RouteAuthorization{
					Type:         OciApiGatewaySpec_any_of,
					AllowedScope: []string{},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject any_of authorization with nil allowed_scope", func() {
				input := minimalValid()
				input.Spec.Deployment.Routes[0].Authorization = &OciApiGatewaySpec_RouteAuthorization{
					Type: OciApiGatewaySpec_any_of,
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
