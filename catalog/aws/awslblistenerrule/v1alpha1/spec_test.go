package awslblistenerrulev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsLbListenerRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsLbListenerRuleSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	listenerArn    = "arn:aws:elasticloadbalancing:us-west-2:123456789012:listener/app/demo/50dc6c495c0c9188/f2f7dc8efc522ab2"
	targetGroupArn = "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/api/943f017f100becff"
)

func pathCondition(patterns ...string) *AwsLbListenerRuleCondition {
	return &AwsLbListenerRuleCondition{
		PathPattern: &AwsLbListenerRulePathPatternCondition{Values: patterns},
	}
}

func forwardAction() *AwsLbListenerRuleAction {
	return &AwsLbListenerRuleAction{
		Type: "forward",
		Forward: &AwsLbListenerRuleActionForward{
			TargetGroups: []*AwsLbListenerRuleActionForwardTargetGroup{
				{Arn: literalRef(targetGroupArn)},
			},
		},
	}
}

// minimalValidRule is the common case: a path-based forward rule -- the shape
// every per-service route uses.
func minimalValidRule() *AwsLbListenerRule {
	return &AwsLbListenerRule{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsLbListenerRule",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "api-route",
		},
		Spec: &AwsLbListenerRuleSpec{
			Region:      "us-west-2",
			ListenerArn: literalRef(listenerArn),
			Conditions:  []*AwsLbListenerRuleCondition{pathCondition("/api/*")},
			Actions:     []*AwsLbListenerRuleAction{forwardAction()},
		},
	}
}

var _ = ginkgo.Describe("AwsLbListenerRuleSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_lb_listener_rule", func() {

			ginkgo.It("should not return a validation error for a minimal path-forward rule", func() {
				err := protovalidate.Validate(minimalValidRule())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for host AND path conditions with a priority", func() {
				input := minimalValidRule()
				input.Spec.Priority = 10
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{
					{HostHeader: &AwsLbListenerRuleHostHeaderCondition{Values: []string{"api.example.com"}}},
					pathCondition("/v1/*"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a regex host condition", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{
					{HostHeader: &AwsLbListenerRuleHostHeaderCondition{RegexValues: []string{"^(api|www)\\.example\\.com$"}}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for header, method, query, and source-ip conditions", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{
					{HttpHeader: &AwsLbListenerRuleHttpHeaderCondition{
						HttpHeaderName: "X-Canary",
						Values:         []string{"true"},
					}},
					{HttpRequestMethod: &AwsLbListenerRuleHttpRequestMethodCondition{Values: []string{"GET", "POST"}}},
					{QueryString: &AwsLbListenerRuleQueryStringCondition{
						Pairs: []*AwsLbListenerRuleQueryStringPair{{Key: "version", Value: "beta"}},
					}},
					{SourceIp: &AwsLbListenerRuleSourceIpCondition{Values: []string{"10.0.0.0/8"}}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an IPv6 source-ip condition", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{
					{SourceIp: &AwsLbListenerRuleSourceIpCondition{
						Values: []string{"2001:db8::/32", "203.0.113.7/32"},
					}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a weighted canary forward", func() {
				input := minimalValidRule()
				input.Spec.Actions = []*AwsLbListenerRuleAction{{
					Type: "forward",
					Forward: &AwsLbListenerRuleActionForward{
						TargetGroups: []*AwsLbListenerRuleActionForwardTargetGroup{
							{Arn: literalRef(targetGroupArn), Weight: 95},
							{Arn: literalRef(targetGroupArn), Weight: 5},
						},
						Stickiness: &AwsLbListenerRuleActionForwardStickiness{
							Enabled:         true,
							DurationSeconds: 600,
						},
					},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for stickiness off without a duration", func() {
				input := minimalValidRule()
				input.Spec.Actions[0].Forward.Stickiness = &AwsLbListenerRuleActionForwardStickiness{
					Enabled: false,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a redirect rule", func() {
				input := minimalValidRule()
				input.Spec.Actions = []*AwsLbListenerRuleAction{{
					Type: "redirect",
					Redirect: &AwsLbListenerRuleActionRedirect{
						StatusCode: "HTTP_301",
						Host:       "new.example.com",
					},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a cognito-then-forward chain", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{pathCondition("/admin/*")}
				input.Spec.Actions = []*AwsLbListenerRuleAction{
					{
						Type: "authenticate-cognito",
						AuthenticateCognito: &AwsLbListenerRuleActionAuthenticateCognito{
							UserPoolArn:      literalRef("arn:aws:cognito-idp:us-west-2:123456789012:userpool/us-west-2_ABC"),
							UserPoolClientId: literalRef("client-id"),
							UserPoolDomain:   literalRef("my-app"),
						},
					},
					forwardAction(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a jwt-validation-then-forward chain", func() {
				input := minimalValidRule()
				input.Spec.Actions = []*AwsLbListenerRuleAction{
					{
						Type: "jwt-validation",
						JwtValidation: &AwsLbListenerRuleActionJwtValidation{
							Issuer:       "https://issuer.example.com",
							JwksEndpoint: "https://issuer.example.com/.well-known/jwks.json",
						},
					},
					forwardAction(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for url-rewrite and host-header-rewrite transforms", func() {
				input := minimalValidRule()
				input.Spec.Transforms = []*AwsLbListenerRuleTransform{
					{
						Type: "url-rewrite",
						UrlRewrite: &AwsLbListenerRuleRewrite{
							Regex:   "^/v1/(.*)$",
							Replace: "/v2/$1",
						},
					},
					{
						Type: "host-header-rewrite",
						HostHeaderRewrite: &AwsLbListenerRuleRewrite{
							Regex:   "^legacy\\.example\\.com$",
							Replace: "api.example.com",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_lb_listener_rule", func() {

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidRule()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the listener reference is missing", func() {
				input := minimalValidRule()
				input.Spec.ListenerArn = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range priority", func() {
				input := minimalValidRule()
				input.Spec.Priority = 50001
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a header name with illegal characters", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{
					{HttpHeader: &AwsLbListenerRuleHttpHeaderCondition{
						HttpHeaderName: "X Canary",
						Values:         []string{"true"},
					}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a bare IP in a source-ip condition", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{
					{SourceIp: &AwsLbListenerRuleSourceIpCondition{Values: []string{"203.0.113.7"}}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when conditions are empty", func() {
				input := minimalValidRule()
				input.Spec.Conditions = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for six condition blocks", func() {
				input := minimalValidRule()
				conditions := make([]*AwsLbListenerRuleCondition, 6)
				for i := range conditions {
					conditions[i] = pathCondition("/api/*")
				}
				input.Spec.Conditions = conditions
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a condition with no criterion", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{{}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a condition with two criteria", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{{
					HostHeader:  &AwsLbListenerRuleHostHeaderCondition{Values: []string{"api.example.com"}},
					PathPattern: &AwsLbListenerRulePathPatternCondition{Values: []string{"/api/*"}},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a host condition with no patterns", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{{
					HostHeader: &AwsLbListenerRuleHostHeaderCondition{},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a header condition missing the header name", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{{
					HttpHeader: &AwsLbListenerRuleHttpHeaderCondition{Values: []string{"true"}},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid HTTP method pattern", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{{
					HttpRequestMethod: &AwsLbListenerRuleHttpRequestMethodCondition{Values: []string{"GET/POST"}},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a query-string pair without a value", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{{
					QueryString: &AwsLbListenerRuleQueryStringCondition{
						Pairs: []*AwsLbListenerRuleQueryStringPair{{Key: "version"}},
					},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a source-ip condition with no blocks", func() {
				input := minimalValidRule()
				input.Spec.Conditions = []*AwsLbListenerRuleCondition{{
					SourceIp: &AwsLbListenerRuleSourceIpCondition{},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when actions are empty", func() {
				input := minimalValidRule()
				input.Spec.Actions = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a forward type without forward config", func() {
				input := minimalValidRule()
				input.Spec.Actions[0].Forward = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a fixed-response config on a forward action", func() {
				input := minimalValidRule()
				input.Spec.Actions[0].FixedResponse = &AwsLbListenerRuleActionFixedResponse{
					ContentType: "text/plain",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid redirect status code", func() {
				input := minimalValidRule()
				input.Spec.Actions = []*AwsLbListenerRuleAction{{
					Type:     "redirect",
					Redirect: &AwsLbListenerRuleActionRedirect{StatusCode: "HTTP_308"},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range forward weight", func() {
				input := minimalValidRule()
				input.Spec.Actions[0].Forward.TargetGroups[0].Weight = 1000
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for enabled stickiness without a duration", func() {
				input := minimalValidRule()
				input.Spec.Actions[0].Forward.Stickiness = &AwsLbListenerRuleActionForwardStickiness{
					Enabled: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for three transforms", func() {
				input := minimalValidRule()
				rewrite := &AwsLbListenerRuleRewrite{Regex: "^/a$", Replace: "/b"}
				input.Spec.Transforms = []*AwsLbListenerRuleTransform{
					{Type: "url-rewrite", UrlRewrite: rewrite},
					{Type: "url-rewrite", UrlRewrite: rewrite},
					{Type: "url-rewrite", UrlRewrite: rewrite},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a transform whose config mismatches its type", func() {
				input := minimalValidRule()
				input.Spec.Transforms = []*AwsLbListenerRuleTransform{{
					Type:       "host-header-rewrite",
					UrlRewrite: &AwsLbListenerRuleRewrite{Regex: "^/a$", Replace: "/b"},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a rewrite without a regex", func() {
				input := minimalValidRule()
				input.Spec.Transforms = []*AwsLbListenerRuleTransform{{
					Type:       "url-rewrite",
					UrlRewrite: &AwsLbListenerRuleRewrite{Replace: "/b"},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
