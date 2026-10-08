package awslblistenerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsLbListenerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsLbListenerSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	albArn         = "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/demo/50dc6c495c0c9188"
	targetGroupArn = "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/api/943f017f100becff"
	certArn        = "arn:aws:acm:us-west-2:123456789012:certificate/12345678-1234-1234-1234-123456789012"
)

func forwardAction() *AwsLbListenerAction {
	return &AwsLbListenerAction{
		Type: "forward",
		Forward: &AwsLbListenerActionForward{
			TargetGroups: []*AwsLbListenerActionForwardTargetGroup{
				{Arn: literalRef(targetGroupArn)},
			},
		},
	}
}

// minimalValidListener is the common case: an HTTP listener whose default
// action forwards to a target group.
func minimalValidListener() *AwsLbListener {
	return &AwsLbListener{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsLbListener",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "http-80",
		},
		Spec: &AwsLbListenerSpec{
			Region:          "us-west-2",
			LoadBalancerArn: literalRef(albArn),
			Port:            80,
			Protocol:        "HTTP",
			DefaultActions:  []*AwsLbListenerAction{forwardAction()},
		},
	}
}

var _ = ginkgo.Describe("AwsLbListenerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_lb_listener", func() {

			ginkgo.It("should not return a validation error for a minimal HTTP forward listener", func() {
				err := protovalidate.Validate(minimalValidListener())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an HTTPS listener with TLS material", func() {
				input := minimalValidListener()
				input.Spec.Port = 443
				input.Spec.Protocol = "HTTPS"
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.AdditionalCertificateArns = []*foreignkeyv1.StringValueOrRef{literalRef(certArn)}
				input.Spec.SslPolicy = "ELBSecurityPolicy-TLS13-1-2-2021-06"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an HTTP-to-HTTPS redirect listener", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions = []*AwsLbListenerAction{{
					Type: "redirect",
					Redirect: &AwsLbListenerActionRedirect{
						StatusCode: "HTTP_301",
						Protocol:   "HTTPS",
						Port:       "443",
					},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a fixed-response default action", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions = []*AwsLbListenerAction{{
					Type: "fixed-response",
					FixedResponse: &AwsLbListenerActionFixedResponse{
						ContentType: "text/plain",
						StatusCode:  "404",
						MessageBody: "not found",
					},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an NLB TCP listener with idle timeout", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "TCP"
				input.Spec.Port = 5432
				input.Spec.TcpIdleTimeoutSeconds = 600
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an NLB TLS listener with ALPN", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "TLS"
				input.Spec.Port = 443
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.AlpnPolicy = "HTTP2Preferred"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an NLB QUIC listener", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "QUIC"
				input.Spec.Port = 443
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an NLB TCP_QUIC listener", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "TCP_QUIC"
				input.Spec.Port = 443
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an OIDC-then-forward action chain", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.Port = 443
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.DefaultActions = []*AwsLbListenerAction{
					{
						Type: "authenticate-oidc",
						AuthenticateOidc: &AwsLbListenerActionAuthenticateOidc{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
							TokenEndpoint:         "https://oauth2.googleapis.com/token",
							UserInfoEndpoint:      "https://openidconnect.googleapis.com/v1/userinfo",
							ClientId:              "client-id",
							ClientSecret:          "client-secret",
						},
					},
					forwardAction(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a jwt-validation-then-forward chain", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.Port = 443
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.DefaultActions = []*AwsLbListenerAction{
					{
						Type: "jwt-validation",
						JwtValidation: &AwsLbListenerActionJwtValidation{
							Issuer:       "https://issuer.example.com",
							JwksEndpoint: "https://issuer.example.com/.well-known/jwks.json",
							AdditionalClaims: []*AwsLbListenerActionJwtClaim{
								{Name: "scope", Format: "space-separated-values", Values: []string{"api:read"}},
							},
						},
					},
					forwardAction(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for mutual TLS with a trust store", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.Port = 443
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.MutualAuthentication = &AwsLbListenerMutualAuthentication{
					Mode:          "verify",
					TrustStoreArn: literalRef("arn:aws:elasticloadbalancing:us-west-2:123456789012:truststore/demo/abc123"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for weighted forward with stickiness", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions = []*AwsLbListenerAction{{
					Type: "forward",
					Forward: &AwsLbListenerActionForward{
						TargetGroups: []*AwsLbListenerActionForwardTargetGroup{
							{Arn: literalRef(targetGroupArn), Weight: 90},
							{Arn: literalRef(targetGroupArn), Weight: 10},
						},
						Stickiness: &AwsLbListenerActionForwardStickiness{
							Enabled:         true,
							DurationSeconds: 3600,
						},
					},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for stickiness off without a duration", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions[0].Forward.Stickiness = &AwsLbListenerActionForwardStickiness{
					Enabled: false,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for HTTP header handling on HTTPS", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.Port = 443
				input.Spec.CertificateArn = literalRef(certArn)
				serverEnabled := false
				input.Spec.HttpHeaders = &AwsLbListenerHttpHeaders{
					Request: &AwsLbListenerHttpRequestHeaders{
						TlsVersionHeaderName: "X-Amzn-Tls-Version",
					},
					Response: &AwsLbListenerHttpResponseHeaders{
						StrictTransportSecurity: "max-age=31536000; includeSubDomains",
						XFrameOptions:           "DENY",
						ServerEnabled:           &serverEnabled,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_lb_listener", func() {

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidListener()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the load balancer reference is missing", func() {
				input := minimalValidListener()
				input.Spec.LoadBalancerArn = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range port", func() {
				input := minimalValidListener()
				input.Spec.Port = 70000
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid protocol", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "GENEVE"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when default actions are empty", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for ssl_policy on an HTTP listener", func() {
				input := minimalValidListener()
				input.Spec.SslPolicy = "ELBSecurityPolicy-TLS13-1-2-2021-06"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for alpn_policy on an HTTPS listener", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.AlpnPolicy = "HTTP2Preferred"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid alpn_policy", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "TLS"
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.AlpnPolicy = "HTTP3"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for mutual authentication on an HTTP listener", func() {
				input := minimalValidListener()
				input.Spec.MutualAuthentication = &AwsLbListenerMutualAuthentication{Mode: "passthrough"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid mutual authentication mode", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.MutualAuthentication = &AwsLbListenerMutualAuthentication{Mode: "strict"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for ignore-expiry outside verify mode", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.MutualAuthentication = &AwsLbListenerMutualAuthentication{
					Mode:                          "passthrough",
					IgnoreClientCertificateExpiry: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for CA-name advertising outside verify mode", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.MutualAuthentication = &AwsLbListenerMutualAuthentication{
					Mode:                       "passthrough",
					AdvertiseTrustStoreCaNames: "on",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for tcp_idle_timeout on an HTTP listener", func() {
				input := minimalValidListener()
				input.Spec.TcpIdleTimeoutSeconds = 600
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range tcp idle timeout", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "TCP"
				input.Spec.TcpIdleTimeoutSeconds = 30
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for http_headers on a TCP listener", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "TCP"
				input.Spec.HttpHeaders = &AwsLbListenerHttpHeaders{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid action type", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions[0].Type = "proxy"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a forward type without forward config", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions[0].Forward = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a redirect config on a forward action", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions[0].Redirect = &AwsLbListenerActionRedirect{StatusCode: "HTTP_301"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range action order", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions[0].Order = 50001
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a forward with no target groups", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions[0].Forward.TargetGroups = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a forward with six target groups", func() {
				input := minimalValidListener()
				groups := make([]*AwsLbListenerActionForwardTargetGroup, 6)
				for i := range groups {
					groups[i] = &AwsLbListenerActionForwardTargetGroup{Arn: literalRef(targetGroupArn)}
				}
				input.Spec.DefaultActions[0].Forward.TargetGroups = groups
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range target group weight", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions[0].Forward.TargetGroups[0].Weight = 1000
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for enabled stickiness without a duration", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions[0].Forward.Stickiness = &AwsLbListenerActionForwardStickiness{
					Enabled: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid redirect status code", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions = []*AwsLbListenerAction{{
					Type:     "redirect",
					Redirect: &AwsLbListenerActionRedirect{StatusCode: "HTTP_307"},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid fixed-response content type", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions = []*AwsLbListenerAction{{
					Type: "fixed-response",
					FixedResponse: &AwsLbListenerActionFixedResponse{
						ContentType: "application/xml",
					},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid fixed-response status code", func() {
				input := minimalValidListener()
				input.Spec.DefaultActions = []*AwsLbListenerAction{{
					Type: "fixed-response",
					FixedResponse: &AwsLbListenerActionFixedResponse{
						ContentType: "text/plain",
						StatusCode:  "301",
					},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an OIDC action missing its client secret", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.DefaultActions = []*AwsLbListenerAction{
					{
						Type: "authenticate-oidc",
						AuthenticateOidc: &AwsLbListenerActionAuthenticateOidc{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
							TokenEndpoint:         "https://oauth2.googleapis.com/token",
							UserInfoEndpoint:      "https://openidconnect.googleapis.com/v1/userinfo",
							ClientId:              "client-id",
						},
					},
					forwardAction(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid on_unauthenticated_request", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.DefaultActions = []*AwsLbListenerAction{
					{
						Type: "authenticate-cognito",
						AuthenticateCognito: &AwsLbListenerActionAuthenticateCognito{
							UserPoolArn:              literalRef("arn:aws:cognito-idp:us-west-2:123456789012:userpool/us-west-2_ABC"),
							UserPoolClientId:         literalRef("client-id"),
							UserPoolDomain:           literalRef("my-app"),
							OnUnauthenticatedRequest: "reject",
						},
					},
					forwardAction(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid jwt claim format", func() {
				input := minimalValidListener()
				input.Spec.Protocol = "HTTPS"
				input.Spec.CertificateArn = literalRef(certArn)
				input.Spec.DefaultActions = []*AwsLbListenerAction{
					{
						Type: "jwt-validation",
						JwtValidation: &AwsLbListenerActionJwtValidation{
							Issuer:       "https://issuer.example.com",
							JwksEndpoint: "https://issuer.example.com/.well-known/jwks.json",
							AdditionalClaims: []*AwsLbListenerActionJwtClaim{
								{Name: "aud", Format: "csv", Values: []string{"api"}},
							},
						},
					},
					forwardAction(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
