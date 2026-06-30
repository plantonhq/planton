package awswafwebaclv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsWafWebAclSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsWafWebAclSpec Validation Suite")
}

// helper to create a minimal valid AwsWafWebAcl wrapper.
func minimalAcl(spec *AwsWafWebAclSpec) *AwsWafWebAcl {
	return &AwsWafWebAcl{
		ApiVersion: "aws.planton.dev/v1",
		Kind:       "AwsWafWebAcl",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-acl"},
		Spec:       spec,
	}
}

// helper to create a minimal valid spec with allow default action.
func minimalSpec() *AwsWafWebAclSpec {
	return &AwsWafWebAclSpec{
		Region: "us-west-2",
		Scope:  "REGIONAL",
		DefaultAction: &AwsWafWebAclDefaultAction{
			Type: "allow",
		},
	}
}

// helper to create a managed rule group rule.
func managedRuleGroupRule(name string, priority int32, groupName string) *AwsWafWebAclRule {
	return &AwsWafWebAclRule{
		Name:     name,
		Priority: priority,
		Statement: &AwsWafWebAclRule_ManagedRuleGroup{
			ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{
				Name:       groupName,
				VendorName: "AWS",
			},
		},
		OverrideAction: "none",
	}
}

// helper to create a rate-based rule.
func rateBasedRule(name string, priority int32, limit int32) *AwsWafWebAclRule {
	return &AwsWafWebAclRule{
		Name:     name,
		Priority: priority,
		Statement: &AwsWafWebAclRule_RateBased{
			RateBased: &AwsWafWebAclRateBasedStatement{
				Limit: limit,
			},
		},
		Action: "block",
	}
}

// helper to create a geo match rule.
func geoMatchRule(name string, priority int32, countryCodes []string) *AwsWafWebAclRule {
	return &AwsWafWebAclRule{
		Name:     name,
		Priority: priority,
		Statement: &AwsWafWebAclRule_GeoMatch{
			GeoMatch: &AwsWafWebAclGeoMatchStatement{
				CountryCodes: countryCodes,
			},
		},
		Action: "block",
	}
}

// helper to create an IP set reference rule.
func ipSetRefRule(name string, priority int32, arn string) *AwsWafWebAclRule {
	return &AwsWafWebAclRule{
		Name:     name,
		Priority: priority,
		Statement: &AwsWafWebAclRule_IpSetReference{
			IpSetReference: &AwsWafWebAclIpSetReferenceStatement{
				Arn: arn,
			},
		},
		Action: "block",
	}
}

var _ = ginkgo.Describe("AwsWafWebAclSpec validations", func() {

	// =========================================================================
	// Happy path — Spec level
	// =========================================================================

	ginkgo.It("accepts a minimal spec with REGIONAL scope and allow default action", func() {
		input := minimalAcl(minimalSpec())
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a minimal spec with CLOUDFRONT scope", func() {
		spec := minimalSpec()
		spec.Scope = "CLOUDFRONT"
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with block default action", func() {
		spec := minimalSpec()
		spec.DefaultAction = &AwsWafWebAclDefaultAction{Type: "block"}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with description", func() {
		spec := minimalSpec()
		spec.Description = "Production Web ACL for API Gateway"
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with token_domains", func() {
		spec := minimalSpec()
		spec.TokenDomains = []string{"example.com", "api.example.com"}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with visibility_config override", func() {
		spec := minimalSpec()
		spec.VisibilityConfig = &AwsWafWebAclVisibilityConfig{
			CloudwatchMetricsEnabled: true,
			SampledRequestsEnabled:   true,
			MetricName:               "my-custom-metric",
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Managed rule group rules
	// =========================================================================

	ginkgo.It("accepts a managed rule group rule with override_action none", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			managedRuleGroupRule("aws-common", 1, "AWSManagedRulesCommonRuleSet"),
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a managed rule group rule with override_action count", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "aws-common-count",
				Priority: 1,
				Statement: &AwsWafWebAclRule_ManagedRuleGroup{
					ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{
						Name:       "AWSManagedRulesCommonRuleSet",
						VendorName: "AWS",
					},
				},
				OverrideAction: "count",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a managed rule group with version pinning", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "aws-common-pinned",
				Priority: 1,
				Statement: &AwsWafWebAclRule_ManagedRuleGroup{
					ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{
						Name:       "AWSManagedRulesCommonRuleSet",
						VendorName: "AWS",
						Version:    "Version_1.0",
					},
				},
				OverrideAction: "none",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a managed rule group with rule action overrides", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "aws-common-tuned",
				Priority: 1,
				Statement: &AwsWafWebAclRule_ManagedRuleGroup{
					ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{
						Name:       "AWSManagedRulesCommonRuleSet",
						VendorName: "AWS",
						RuleActionOverrides: []*AwsWafWebAclRuleActionOverride{
							{Name: "SizeRestrictions_BODY", Action: "count"},
							{Name: "NoUserAgent_HEADER", Action: "count"},
						},
					},
				},
				OverrideAction: "none",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Rate-based rules
	// =========================================================================

	ginkgo.It("accepts a rate-based rule with minimal config", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{rateBasedRule("rate-limit", 1, 2000)}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a rate-based rule with custom evaluation window", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "rate-limit-60s",
				Priority: 1,
				Statement: &AwsWafWebAclRule_RateBased{
					RateBased: &AwsWafWebAclRateBasedStatement{
						Limit:               500,
						EvaluationWindowSec: 60,
					},
				},
				Action: "block",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a rate-based rule with FORWARDED_IP and config", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "rate-forwarded",
				Priority: 1,
				Statement: &AwsWafWebAclRule_RateBased{
					RateBased: &AwsWafWebAclRateBasedStatement{
						Limit:            1000,
						AggregateKeyType: "FORWARDED_IP",
						ForwardedIpConfig: &AwsWafWebAclForwardedIpConfig{
							HeaderName:       "X-Forwarded-For",
							FallbackBehavior: "MATCH",
						},
					},
				},
				Action: "block",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Geo match rules
	// =========================================================================

	ginkgo.It("accepts a geo match rule blocking specific countries", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			geoMatchRule("block-countries", 1, []string{"RU", "CN", "KP"}),
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a geo match rule with forwarded IP config", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "geo-forwarded",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
						ForwardedIpConfig: &AwsWafWebAclForwardedIpConfig{
							HeaderName:       "X-Forwarded-For",
							FallbackBehavior: "NO_MATCH",
						},
					},
				},
				Action: "allow",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — IP set reference rules
	// =========================================================================

	ginkgo.It("accepts an IP set reference rule", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			ipSetRefRule("block-ips", 1, "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/bad-ips/abc123"),
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an IP set reference rule with forwarded IP config and position", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "ip-set-forwarded",
				Priority: 1,
				Statement: &AwsWafWebAclRule_IpSetReference{
					IpSetReference: &AwsWafWebAclIpSetReferenceStatement{
						Arn: "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/allow-ips/def456",
						ForwardedIpConfig: &AwsWafWebAclForwardedIpConfig{
							HeaderName:       "X-Forwarded-For",
							FallbackBehavior: "NO_MATCH",
							Position:         "FIRST",
						},
					},
				},
				Action: "allow",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Custom statement (escape hatch)
	// =========================================================================

	ginkgo.It("accepts a custom statement for SQL injection detection", func() {
		sqliStatement, _ := structpb.NewStruct(map[string]interface{}{
			"SqliMatchStatement": map[string]interface{}{
				"FieldToMatch": map[string]interface{}{
					"Body": map[string]interface{}{},
				},
				"TextTransformations": []interface{}{
					map[string]interface{}{
						"Priority": 0,
						"Type":     "URL_DECODE",
					},
				},
			},
		})

		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "block-sqli",
				Priority: 1,
				Statement: &AwsWafWebAclRule_CustomStatement{
					CustomStatement: sqliStatement,
				},
				Action: "block",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Multiple rules (production-like)
	// =========================================================================

	ginkgo.It("accepts a production-like Web ACL with multiple rule types", func() {
		spec := minimalSpec()
		spec.Description = "Production Web ACL"
		spec.Rules = []*AwsWafWebAclRule{
			managedRuleGroupRule("aws-common", 1, "AWSManagedRulesCommonRuleSet"),
			managedRuleGroupRule("aws-known-bad", 2, "AWSManagedRulesKnownBadInputsRuleSet"),
			managedRuleGroupRule("aws-sqli", 3, "AWSManagedRulesSQLiRuleSet"),
			rateBasedRule("rate-limit", 4, 2000),
			geoMatchRule("block-countries", 5, []string{"RU", "CN", "KP"}),
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Custom response bodies
	// =========================================================================

	ginkgo.It("accepts custom response bodies with block action referencing them", func() {
		spec := minimalSpec()
		spec.DefaultAction = &AwsWafWebAclDefaultAction{
			Type: "block",
			CustomResponse: &AwsWafWebAclCustomResponse{
				ResponseCode:          403,
				CustomResponseBodyKey: "forbidden-html",
			},
		}
		spec.CustomResponseBodies = []*AwsWafWebAclCustomResponseBody{
			{
				Key:         "forbidden-html",
				Content:     "<html><body><h1>Forbidden</h1></body></html>",
				ContentType: "TEXT_HTML",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Logging
	// =========================================================================

	ginkgo.It("accepts a spec with logging enabled", func() {
		spec := minimalSpec()
		spec.Logging = &AwsWafWebAclLoggingConfig{
			DestinationArn: strRef("arn:aws:logs:us-east-1:123456789012:log-group:aws-waf-logs-my-acl"),
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts logging with redacted headers", func() {
		spec := minimalSpec()
		spec.Logging = &AwsWafWebAclLoggingConfig{
			DestinationArn:      strRef("arn:aws:logs:us-east-1:123456789012:log-group:aws-waf-logs-my-acl"),
			RedactedHeaderNames: []string{"Authorization", "Cookie"},
			RedactUriPath:       true,
			RedactQueryString:   true,
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Happy path — Rule actions
	// =========================================================================

	ginkgo.It("accepts a rule with captcha action", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "captcha-suspicious",
				Priority: 1,
				Statement: &AwsWafWebAclRule_RateBased{
					RateBased: &AwsWafWebAclRateBasedStatement{Limit: 100},
				},
				Action: "captcha",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a rule with challenge action", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "challenge-bots",
				Priority: 1,
				Statement: &AwsWafWebAclRule_RateBased{
					RateBased: &AwsWafWebAclRateBasedStatement{Limit: 500},
				},
				Action: "challenge",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a rule with count action", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			geoMatchRule("count-geo", 1, []string{"CN"}),
		}
		spec.Rules[0].Action = "count"
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a block rule with custom response", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "block-with-custom-response",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{
						CountryCodes: []string{"RU"},
					},
				},
				Action: "block",
				CustomResponse: &AwsWafWebAclCustomResponse{
					ResponseCode: 403,
					ResponseHeaders: []*AwsWafWebAclCustomHeader{
						{Name: "x-blocked-reason", Value: "geo-restriction"},
					},
				},
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an allow rule with custom request headers", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "allow-with-headers",
				Priority: 1,
				Statement: &AwsWafWebAclRule_IpSetReference{
					IpSetReference: &AwsWafWebAclIpSetReferenceStatement{
						Arn: "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/trusted-ips/xyz",
					},
				},
				Action: "allow",
				CustomRequestHeaders: []*AwsWafWebAclCustomHeader{
					{Name: "x-waf-verified", Value: "true"},
				},
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Scope
	// =========================================================================

	ginkgo.It("fails with invalid scope", func() {
		spec := minimalSpec()
		spec.Scope = "INVALID"
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with empty scope", func() {
		spec := minimalSpec()
		spec.Scope = ""
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Default action
	// =========================================================================

	ginkgo.It("fails with invalid default action type", func() {
		spec := minimalSpec()
		spec.DefaultAction = &AwsWafWebAclDefaultAction{Type: "count"}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with missing default action", func() {
		spec := &AwsWafWebAclSpec{
			Scope: "REGIONAL",
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with custom_response on allow default action", func() {
		spec := minimalSpec()
		spec.DefaultAction = &AwsWafWebAclDefaultAction{
			Type: "allow",
			CustomResponse: &AwsWafWebAclCustomResponse{
				ResponseCode: 403,
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with custom_request_headers on block default action", func() {
		spec := minimalSpec()
		spec.DefaultAction = &AwsWafWebAclDefaultAction{
			Type: "block",
			CustomRequestHeaders: []*AwsWafWebAclCustomHeader{
				{Name: "x-test", Value: "test"},
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Rule action / override_action mutual exclusivity
	// =========================================================================

	ginkgo.It("fails when managed rule group uses action instead of override_action", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-managed",
				Priority: 1,
				Statement: &AwsWafWebAclRule_ManagedRuleGroup{
					ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{
						Name:       "AWSManagedRulesCommonRuleSet",
						VendorName: "AWS",
					},
				},
				Action: "block", // Should be override_action
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when custom rule uses override_action instead of action", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-geo",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{
						CountryCodes: []string{"CN"},
					},
				},
				OverrideAction: "none", // Should be action
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when rule has both action and override_action", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "both-actions",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{
						CountryCodes: []string{"CN"},
					},
				},
				Action:         "block",
				OverrideAction: "count",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when rule has neither action nor override_action", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "no-action",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{
						CountryCodes: []string{"CN"},
					},
				},
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Rule action values
	// =========================================================================

	ginkgo.It("fails with invalid rule action value", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-action",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{CountryCodes: []string{"CN"}},
				},
				Action: "reject", // Invalid
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with invalid override_action value", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-override",
				Priority: 1,
				Statement: &AwsWafWebAclRule_ManagedRuleGroup{
					ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{
						Name:       "AWSManagedRulesCommonRuleSet",
						VendorName: "AWS",
					},
				},
				OverrideAction: "block", // Invalid for override_action
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Statement required
	// =========================================================================

	ginkgo.It("fails when rule has no statement", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "no-statement",
				Priority: 1,
				Action:   "block",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Rate-based validations
	// =========================================================================

	ginkgo.It("fails with rate-based limit below minimum (10)", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{rateBasedRule("too-low", 1, 5)}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with invalid evaluation_window_sec", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-window",
				Priority: 1,
				Statement: &AwsWafWebAclRule_RateBased{
					RateBased: &AwsWafWebAclRateBasedStatement{
						Limit:               1000,
						EvaluationWindowSec: 180, // Invalid: not in [60,120,300,600]
					},
				},
				Action: "block",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with invalid aggregate_key_type", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-key-type",
				Priority: 1,
				Statement: &AwsWafWebAclRule_RateBased{
					RateBased: &AwsWafWebAclRateBasedStatement{
						Limit:            1000,
						AggregateKeyType: "INVALID",
					},
				},
				Action: "block",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when FORWARDED_IP type has no forwarded_ip_config", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "missing-fwd-config",
				Priority: 1,
				Statement: &AwsWafWebAclRule_RateBased{
					RateBased: &AwsWafWebAclRateBasedStatement{
						Limit:            1000,
						AggregateKeyType: "FORWARDED_IP",
						// Missing forwarded_ip_config
					},
				},
				Action: "block",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Geo match validations
	// =========================================================================

	ginkgo.It("fails with empty country_codes", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "empty-countries",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{
						CountryCodes: []string{},
					},
				},
				Action: "block",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Forwarded IP config validations
	// =========================================================================

	ginkgo.It("fails with invalid fallback_behavior", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-fallback",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
						ForwardedIpConfig: &AwsWafWebAclForwardedIpConfig{
							HeaderName:       "X-Forwarded-For",
							FallbackBehavior: "INVALID",
						},
					},
				},
				Action: "allow",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with invalid forwarded IP position", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-position",
				Priority: 1,
				Statement: &AwsWafWebAclRule_IpSetReference{
					IpSetReference: &AwsWafWebAclIpSetReferenceStatement{
						Arn: "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/test/abc",
						ForwardedIpConfig: &AwsWafWebAclForwardedIpConfig{
							HeaderName:       "X-Forwarded-For",
							FallbackBehavior: "MATCH",
							Position:         "MIDDLE", // Invalid
						},
					},
				},
				Action: "block",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Custom response body validations
	// =========================================================================

	ginkgo.It("fails with invalid content_type", func() {
		spec := minimalSpec()
		spec.CustomResponseBodies = []*AwsWafWebAclCustomResponseBody{
			{
				Key:         "bad-type",
				Content:     "test",
				ContentType: "TEXT_XML", // Invalid
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with missing custom response body key", func() {
		spec := minimalSpec()
		spec.CustomResponseBodies = []*AwsWafWebAclCustomResponseBody{
			{
				Key:         "",
				Content:     "test",
				ContentType: "TEXT_PLAIN",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Custom response validations
	// =========================================================================

	ginkgo.It("fails with response_code below 200", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-response-code",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{CountryCodes: []string{"CN"}},
				},
				Action: "block",
				CustomResponse: &AwsWafWebAclCustomResponse{
					ResponseCode: 100, // Below minimum
				},
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with custom_response on non-block action", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "response-on-allow",
				Priority: 1,
				Statement: &AwsWafWebAclRule_GeoMatch{
					GeoMatch: &AwsWafWebAclGeoMatchStatement{CountryCodes: []string{"US"}},
				},
				Action: "allow",
				CustomResponse: &AwsWafWebAclCustomResponse{
					ResponseCode: 403,
				},
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — Rule action override validations
	// =========================================================================

	ginkgo.It("fails with invalid rule action override action", func() {
		spec := minimalSpec()
		spec.Rules = []*AwsWafWebAclRule{
			{
				Name:     "bad-override-action",
				Priority: 1,
				Statement: &AwsWafWebAclRule_ManagedRuleGroup{
					ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{
						Name:       "AWSManagedRulesCommonRuleSet",
						VendorName: "AWS",
						RuleActionOverrides: []*AwsWafWebAclRuleActionOverride{
							{Name: "SomeRule", Action: "reject"}, // Invalid
						},
					},
				},
				OverrideAction: "none",
			},
		}
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Failure — API envelope validations
	// =========================================================================

	ginkgo.It("fails with wrong apiVersion", func() {
		input := &AwsWafWebAcl{
			ApiVersion: "wrong/v1",
			Kind:       "AwsWafWebAcl",
			Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			Spec:       minimalSpec(),
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with wrong kind", func() {
		input := &AwsWafWebAcl{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "WrongKind",
			Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			Spec:       minimalSpec(),
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with missing metadata", func() {
		input := &AwsWafWebAcl{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsWafWebAcl",
			Spec:       minimalSpec(),
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with missing spec", func() {
		input := &AwsWafWebAcl{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsWafWebAcl",
			Metadata:   &shared.CloudResourceMetadata{Name: "test"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails with description exceeding 256 characters", func() {
		spec := minimalSpec()
		longDesc := ""
		for i := 0; i < 300; i++ {
			longDesc += "a"
		}
		spec.Description = longDesc
		input := minimalAcl(spec)
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})

// strRef creates a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}
