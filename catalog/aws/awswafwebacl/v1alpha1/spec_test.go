package awswafwebaclv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsWafWebAclSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsWafWebAclSpec Validation Suite")
}

// helper to create a minimal valid AwsWafWebAcl wrapper.
func minimalAcl(spec *AwsWafWebAclSpec) *AwsWafWebAcl {
	return &AwsWafWebAcl{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsWafWebAcl",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-acl"},
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

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// helper wrapping a statement arm into the Statement node.
func stmt(arm isAwsWafWebAclStatement_Statement) *AwsWafWebAclStatement {
	return &AwsWafWebAclStatement{Statement: arm}
}

// helper to create a managed rule group rule.
func managedRuleGroupRule(name string, priority int32, groupName string) *AwsWafWebAclRule {
	return &AwsWafWebAclRule{
		Name:     name,
		Priority: proto.Int32(priority),
		Statement: stmt(&AwsWafWebAclStatement_ManagedRuleGroup{
			ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{
				Name:       groupName,
				VendorName: "AWS",
			},
		}),
		OverrideAction: "none",
	}
}

// helper to create a rate-based rule.
func rateBasedRule(name string, priority int32, limit int32) *AwsWafWebAclRule {
	return &AwsWafWebAclRule{
		Name:     name,
		Priority: proto.Int32(priority),
		Statement: stmt(&AwsWafWebAclStatement_RateBased{
			RateBased: &AwsWafWebAclRateBasedStatement{
				Limit: limit,
			},
		}),
		Action: "block",
	}
}

// helper to create a geo match rule.
func geoMatchRule(name string, priority int32, countryCodes []string) *AwsWafWebAclRule {
	return &AwsWafWebAclRule{
		Name:     name,
		Priority: proto.Int32(priority),
		Statement: stmt(&AwsWafWebAclStatement_GeoMatch{
			GeoMatch: &AwsWafWebAclGeoMatchStatement{
				CountryCodes: countryCodes,
			},
		}),
		Action: "block",
	}
}

// helper to create an IP set reference rule.
func ipSetRefRule(name string, priority int32, arn string) *AwsWafWebAclRule {
	return &AwsWafWebAclRule{
		Name:     name,
		Priority: proto.Int32(priority),
		Statement: stmt(&AwsWafWebAclStatement_IpSetReference{
			IpSetReference: &AwsWafWebAclIpSetReferenceStatement{
				Arn: strRef(arn),
			},
		}),
		Action: "block",
	}
}

// helper for a byte-match statement inspecting the URI path.
func byteMatchStatement(search string) *AwsWafWebAclStatement {
	return stmt(&AwsWafWebAclStatement_ByteMatch{
		ByteMatch: &AwsWafWebAclByteMatchStatement{
			SearchString:         search,
			PositionalConstraint: "STARTS_WITH",
			FieldToMatch: &AwsWafWebAclFieldToMatch{
				Field: &AwsWafWebAclFieldToMatch_UriPath{UriPath: true},
			},
			TextTransformations: []*AwsWafWebAclTextTransformation{{Priority: 0, Type: "NONE"}},
		},
	})
}

// helper wrapping a statement into a match rule.
func matchRule(name string, priority int32, statement *AwsWafWebAclStatement) *AwsWafWebAclRule {
	return &AwsWafWebAclRule{
		Name:      name,
		Priority:  proto.Int32(priority),
		Statement: statement,
		Action:    "block",
	}
}

// helper wrapping rules into a valid spec + acl.
func aclWithRules(rules ...*AwsWafWebAclRule) *AwsWafWebAcl {
	spec := minimalSpec()
	spec.Rules = rules
	return minimalAcl(spec)
}

var _ = ginkgo.Describe("AwsWafWebAclSpec validations", func() {

	// =========================================================================
	// Happy path — Spec level
	// =========================================================================

	ginkgo.It("accepts a minimal spec with REGIONAL scope and allow default action", func() {
		gomega.Expect(protovalidate.Validate(minimalAcl(minimalSpec()))).To(gomega.BeNil())
	})

	ginkgo.It("accepts CLOUDFRONT scope in us-east-1", func() {
		spec := minimalSpec()
		spec.Scope = "CLOUDFRONT"
		spec.Region = "us-east-1"
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).To(gomega.BeNil())
	})

	ginkgo.It("rejects CLOUDFRONT scope outside us-east-1", func() {
		spec := minimalSpec()
		spec.Scope = "CLOUDFRONT"
		spec.Region = "us-west-2"
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid scope", func() {
		spec := minimalSpec()
		spec.Scope = "GLOBAL"
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a missing region", func() {
		spec := minimalSpec()
		spec.Region = ""
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a description over 256 characters", func() {
		spec := minimalSpec()
		for i := 0; i < 26; i++ {
			spec.Description += "0123456789"
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Default action
	// =========================================================================

	ginkgo.It("rejects an invalid default action type", func() {
		spec := minimalSpec()
		spec.DefaultAction = &AwsWafWebAclDefaultAction{Type: "count"}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a block default action with custom response", func() {
		spec := minimalSpec()
		spec.DefaultAction = &AwsWafWebAclDefaultAction{
			Type:           "block",
			CustomResponse: &AwsWafWebAclCustomResponse{ResponseCode: 403},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).To(gomega.BeNil())
	})

	ginkgo.It("rejects custom_response on an allow default action", func() {
		spec := minimalSpec()
		spec.DefaultAction = &AwsWafWebAclDefaultAction{
			Type:           "allow",
			CustomResponse: &AwsWafWebAclCustomResponse{ResponseCode: 403},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects custom_request_headers on a block default action", func() {
		spec := minimalSpec()
		spec.DefaultAction = &AwsWafWebAclDefaultAction{
			Type:                 "block",
			CustomRequestHeaders: []*AwsWafWebAclCustomHeader{{Name: "x-a", Value: "b"}},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Rule action model
	// =========================================================================

	ginkgo.It("accepts a managed rule group rule with override_action", func() {
		gomega.Expect(protovalidate.Validate(aclWithRules(
			managedRuleGroupRule("common", 1, "AWSManagedRulesCommonRuleSet"),
		))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a managed rule group rule with action instead of override_action", func() {
		rule := managedRuleGroupRule("common", 1, "AWSManagedRulesCommonRuleSet")
		rule.OverrideAction = ""
		rule.Action = "block"
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a rule group reference rule with action instead of override_action", func() {
		rule := &AwsWafWebAclRule{
			Name:     "custom-group",
			Priority: proto.Int32(1),
			Statement: stmt(&AwsWafWebAclStatement_RuleGroupReference{
				RuleGroupReference: &AwsWafWebAclRuleGroupReferenceStatement{
					Arn: "arn:aws:wafv2:us-west-2:111122223333:regional/rulegroup/x/1",
				},
			}),
			Action: "block",
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a rule group reference rule with override_action", func() {
		rule := &AwsWafWebAclRule{
			Name:     "custom-group",
			Priority: proto.Int32(1),
			Statement: stmt(&AwsWafWebAclStatement_RuleGroupReference{
				RuleGroupReference: &AwsWafWebAclRuleGroupReferenceStatement{
					Arn: "arn:aws:wafv2:us-west-2:111122223333:regional/rulegroup/x/1",
				},
			}),
			OverrideAction: "none",
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a match rule with override_action instead of action", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Action = ""
		rule.OverrideAction = "none"
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid rule action", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Action = "drop"
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid override_action", func() {
		rule := managedRuleGroupRule("common", 1, "AWSManagedRulesCommonRuleSet")
		rule.OverrideAction = "block"
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a rule without a statement", func() {
		rule := &AwsWafWebAclRule{Name: "no-stmt", Priority: proto.Int32(1), Action: "block"}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a statement with no arm set", func() {
		rule := &AwsWafWebAclRule{
			Name:      "empty-stmt",
			Priority:  proto.Int32(1),
			Statement: &AwsWafWebAclStatement{},
			Action:    "block",
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects custom_response on a non-block rule action", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Action = "count"
		rule.CustomResponse = &AwsWafWebAclCustomResponse{ResponseCode: 429}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a rule with per-rule captcha and challenge immunity", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Action = "captcha"
		rule.CaptchaConfig = &AwsWafWebAclImmunityTimeConfig{ImmunityTimeSec: 120}
		rule.ChallengeConfig = &AwsWafWebAclImmunityTimeConfig{ImmunityTimeSec: 600}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	// =========================================================================
	// Rate-based statements
	// =========================================================================

	ginkgo.It("rejects a rate limit below 10", func() {
		gomega.Expect(protovalidate.Validate(aclWithRules(rateBasedRule("rate", 1, 5)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid evaluation window", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Statement.GetRateBased().EvaluationWindowSec = 90
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects FORWARDED_IP aggregation without forwarded_ip_config", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Statement.GetRateBased().AggregateKeyType = "FORWARDED_IP"
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts FORWARDED_IP aggregation with forwarded_ip_config", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rb := rule.Statement.GetRateBased()
		rb.AggregateKeyType = "FORWARDED_IP"
		rb.ForwardedIpConfig = &AwsWafWebAclForwardedIpConfig{
			HeaderName:       "X-Forwarded-For",
			FallbackBehavior: "MATCH",
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects custom_keys without CUSTOM_KEYS aggregation", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Statement.GetRateBased().CustomKeys = []*AwsWafWebAclRateBasedCustomKey{
			{Key: &AwsWafWebAclRateBasedCustomKey_Ip{Ip: true}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects CUSTOM_KEYS aggregation without custom_keys", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Statement.GetRateBased().AggregateKeyType = "CUSTOM_KEYS"
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts CUSTOM_KEYS aggregation with header + uri_path keys", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rb := rule.Statement.GetRateBased()
		rb.AggregateKeyType = "CUSTOM_KEYS"
		rb.CustomKeys = []*AwsWafWebAclRateBasedCustomKey{
			{Key: &AwsWafWebAclRateBasedCustomKey_Header{Header: &AwsWafWebAclKeyWithTransformations{
				Name:                "x-api-key",
				TextTransformations: []*AwsWafWebAclTextTransformation{{Priority: 0, Type: "NONE"}},
			}}},
			{Key: &AwsWafWebAclRateBasedCustomKey_UriPath{UriPath: &AwsWafWebAclTransformationsOnlyKey{
				TextTransformations: []*AwsWafWebAclTextTransformation{{Priority: 0, Type: "LOWERCASE"}},
			}}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects more than five custom keys", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rb := rule.Statement.GetRateBased()
		rb.AggregateKeyType = "CUSTOM_KEYS"
		for i := 0; i < 6; i++ {
			rb.CustomKeys = append(rb.CustomKeys, &AwsWafWebAclRateBasedCustomKey{
				Key: &AwsWafWebAclRateBasedCustomKey_QueryString{QueryString: &AwsWafWebAclTransformationsOnlyKey{
					TextTransformations: []*AwsWafWebAclTextTransformation{{Priority: int32(i), Type: "NONE"}},
				}},
			})
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a custom key with no arm set", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rb := rule.Statement.GetRateBased()
		rb.AggregateKeyType = "CUSTOM_KEYS"
		rb.CustomKeys = []*AwsWafWebAclRateBasedCustomKey{{}}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a false bool custom-key arm", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rb := rule.Statement.GetRateBased()
		rb.AggregateKeyType = "CUSTOM_KEYS"
		rb.CustomKeys = []*AwsWafWebAclRateBasedCustomKey{
			{Key: &AwsWafWebAclRateBasedCustomKey_Ip{Ip: false}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Scope-down statements
	// =========================================================================

	ginkgo.It("accepts a geo scope-down on a managed rule group", func() {
		rule := managedRuleGroupRule("common", 1, "AWSManagedRulesCommonRuleSet")
		rule.Statement.GetManagedRuleGroup().ScopeDownStatement = stmt(&AwsWafWebAclStatement_GeoMatch{
			GeoMatch: &AwsWafWebAclGeoMatchStatement{CountryCodes: []string{"US"}},
		})
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a managed rule group inside a scope-down", func() {
		rule := managedRuleGroupRule("common", 1, "AWSManagedRulesCommonRuleSet")
		rule.Statement.GetManagedRuleGroup().ScopeDownStatement = stmt(&AwsWafWebAclStatement_ManagedRuleGroup{
			ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{Name: "x", VendorName: "AWS"},
		})
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a rate-based statement inside a rate-based scope-down", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Statement.GetRateBased().ScopeDownStatement = stmt(&AwsWafWebAclStatement_RateBased{
			RateBased: &AwsWafWebAclRateBasedStatement{Limit: 100},
		})
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a byte-match scope-down on a rate-based statement", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.Statement.GetRateBased().ScopeDownStatement = byteMatchStatement("/login")
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	// =========================================================================
	// Logical composition (and/or/not)
	// =========================================================================

	ginkgo.It("accepts an AND of geo and byte-match", func() {
		rule := matchRule("and-rule", 1, stmt(&AwsWafWebAclStatement_AndStatement{
			AndStatement: &AwsWafWebAclAndStatement{Statements: []*AwsWafWebAclStatement{
				stmt(&AwsWafWebAclStatement_GeoMatch{GeoMatch: &AwsWafWebAclGeoMatchStatement{CountryCodes: []string{"US"}}}),
				byteMatchStatement("/admin"),
			}},
		}))
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an AND with a single statement", func() {
		rule := matchRule("and-rule", 1, stmt(&AwsWafWebAclStatement_AndStatement{
			AndStatement: &AwsWafWebAclAndStatement{Statements: []*AwsWafWebAclStatement{
				byteMatchStatement("/admin"),
			}},
		}))
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a NOT of an OR nested three levels", func() {
		rule := matchRule("nested", 1, stmt(&AwsWafWebAclStatement_NotStatement{
			NotStatement: &AwsWafWebAclNotStatement{Statement: stmt(&AwsWafWebAclStatement_OrStatement{
				OrStatement: &AwsWafWebAclOrStatement{Statements: []*AwsWafWebAclStatement{
					byteMatchStatement("/health"),
					stmt(&AwsWafWebAclStatement_LabelMatch{LabelMatch: &AwsWafWebAclLabelMatchStatement{
						Scope: "LABEL",
						Key:   "awswaf:managed:aws:bot-control:bot:category:monitoring",
					}}),
				}},
			})},
		}))
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a NOT without a statement", func() {
		rule := matchRule("not-empty", 1, stmt(&AwsWafWebAclStatement_NotStatement{
			NotStatement: &AwsWafWebAclNotStatement{},
		}))
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Match statements + field_to_match + text transformations
	// =========================================================================

	ginkgo.It("accepts geo, ip-set, and regex-set reference rules", func() {
		regexRefRule := matchRule("regex-set", 3, stmt(&AwsWafWebAclStatement_RegexPatternSetReference{
			RegexPatternSetReference: &AwsWafWebAclRegexPatternSetReferenceStatement{
				Arn: strRef("arn:aws:wafv2:us-west-2:111122223333:regional/regexpatternset/x/1"),
				FieldToMatch: &AwsWafWebAclFieldToMatch{
					Field: &AwsWafWebAclFieldToMatch_UriPath{UriPath: true},
				},
				TextTransformations: []*AwsWafWebAclTextTransformation{{Priority: 0, Type: "LOWERCASE"}},
			},
		}))
		gomega.Expect(protovalidate.Validate(aclWithRules(
			geoMatchRule("geo", 1, []string{"US", "CA"}),
			ipSetRefRule("ips", 2, "arn:aws:wafv2:us-west-2:111122223333:regional/ipset/x/1"),
			regexRefRule,
		))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a geo match without country codes", func() {
		gomega.Expect(protovalidate.Validate(aclWithRules(geoMatchRule("geo", 1, nil)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an ip-set reference without an arn", func() {
		rule := &AwsWafWebAclRule{
			Name:     "ips",
			Priority: proto.Int32(1),
			Statement: stmt(&AwsWafWebAclStatement_IpSetReference{
				IpSetReference: &AwsWafWebAclIpSetReferenceStatement{},
			}),
			Action: "block",
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a byte match without text transformations", func() {
		statement := byteMatchStatement("/admin")
		statement.GetByteMatch().TextTransformations = nil
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("bm", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid positional constraint", func() {
		statement := byteMatchStatement("/admin")
		statement.GetByteMatch().PositionalConstraint = "NEAR"
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("bm", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid text transformation type", func() {
		statement := byteMatchStatement("/admin")
		statement.GetByteMatch().TextTransformations = []*AwsWafWebAclTextTransformation{{Priority: 0, Type: "ROT13"}}
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("bm", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a field_to_match with no component set", func() {
		statement := byteMatchStatement("/admin")
		statement.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{}
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("bm", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a false bool field_to_match arm", func() {
		statement := byteMatchStatement("/admin")
		statement.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_UriPath{UriPath: false},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("bm", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a headers inspection with an include list", func() {
		statement := byteMatchStatement("bot")
		statement.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_Headers{Headers: &AwsWafWebAclHeadersMatch{
				MatchPattern:     &AwsWafWebAclNamePattern{IncludedNames: []string{"user-agent"}},
				MatchScope:       "VALUE",
				OversizeHandling: "CONTINUE",
			}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("hdr", 1, statement)))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a name pattern selecting both all and an include list", func() {
		statement := byteMatchStatement("bot")
		statement.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_Headers{Headers: &AwsWafWebAclHeadersMatch{
				MatchPattern:     &AwsWafWebAclNamePattern{All: true, IncludedNames: []string{"user-agent"}},
				MatchScope:       "VALUE",
				OversizeHandling: "CONTINUE",
			}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("hdr", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a cookies inspection without oversize handling", func() {
		statement := byteMatchStatement("session")
		statement.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_Cookies{Cookies: &AwsWafWebAclCookiesMatch{
				MatchPattern: &AwsWafWebAclNamePattern{All: true},
				MatchScope:   "ALL",
			}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("ck", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a json body inspection", func() {
		statement := byteMatchStatement("admin")
		statement.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_JsonBody{JsonBody: &AwsWafWebAclJsonBodyMatch{
				MatchScope:       "VALUE",
				IncludedPaths:    []string{"/role"},
				OversizeHandling: "MATCH",
			}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("jb", 1, statement)))).To(gomega.BeNil())
	})

	ginkgo.It("accepts JA3/JA4 fingerprint and header-order inspections", func() {
		ja3 := byteMatchStatement("d41d8cd98f00b204e9800998ecf8427e")
		ja3.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_Ja3Fingerprint{Ja3Fingerprint: &AwsWafWebAclFingerprintMatch{FallbackBehavior: "NO_MATCH"}},
		}
		ja4 := byteMatchStatement("t13d1516h2_8daaf6152771_e5627efa2ab1")
		ja4.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_Ja4Fingerprint{Ja4Fingerprint: &AwsWafWebAclFingerprintMatch{FallbackBehavior: "MATCH"}},
		}
		order := byteMatchStatement("host,user-agent,accept")
		order.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_HeaderOrder{HeaderOrder: &AwsWafWebAclHeaderOrderMatch{OversizeHandling: "CONTINUE"}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(
			matchRule("ja3", 1, ja3),
			matchRule("ja4", 2, ja4),
			matchRule("order", 3, order),
		))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a JA3 fingerprint inspection without a fallback behavior", func() {
		statement := byteMatchStatement("d41d8cd98f00b204e9800998ecf8427e")
		statement.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_Ja3Fingerprint{Ja3Fingerprint: &AwsWafWebAclFingerprintMatch{}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("ja3", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a uri-fragment inspection with and without a fallback behavior", func() {
		withFallback := byteMatchStatement("section")
		withFallback.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_UriFragment{UriFragment: &AwsWafWebAclUriFragmentMatch{FallbackBehavior: "NO_MATCH"}},
		}
		withoutFallback := byteMatchStatement("section")
		withoutFallback.GetByteMatch().FieldToMatch = &AwsWafWebAclFieldToMatch{
			Field: &AwsWafWebAclFieldToMatch_UriFragment{UriFragment: &AwsWafWebAclUriFragmentMatch{}},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(
			matchRule("frag1", 1, withFallback),
			matchRule("frag2", 2, withoutFallback),
		))).To(gomega.BeNil())
	})

	// The scope-down group-statement exclusion CELs guard the IMMEDIATE child
	// only: a managed rule group nested one logical level deeper (inside an
	// AND under the scope-down) passes spec validation by design, and AWS
	// rejects it at deploy time. This case documents that acknowledged
	// boundary — if it starts failing, the spec grew a deep-walk guard and
	// the modules' comments should be updated to match.
	ginkgo.It("accepts (deferring to AWS) a group statement nested below a scope-down's immediate child", func() {
		rule := managedRuleGroupRule("common", 1, "AWSManagedRulesCommonRuleSet")
		rule.Statement.GetManagedRuleGroup().ScopeDownStatement = stmt(&AwsWafWebAclStatement_AndStatement{
			AndStatement: &AwsWafWebAclAndStatement{Statements: []*AwsWafWebAclStatement{
				byteMatchStatement("/admin"),
				stmt(&AwsWafWebAclStatement_ManagedRuleGroup{
					ManagedRuleGroup: &AwsWafWebAclManagedRuleGroupStatement{Name: "x", VendorName: "AWS"},
				}),
			}},
		})
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid sqli sensitivity level", func() {
		statement := stmt(&AwsWafWebAclStatement_SqliMatch{SqliMatch: &AwsWafWebAclSqliMatchStatement{
			FieldToMatch:        &AwsWafWebAclFieldToMatch{Field: &AwsWafWebAclFieldToMatch_Body{Body: &AwsWafWebAclBodyMatch{}}},
			TextTransformations: []*AwsWafWebAclTextTransformation{{Priority: 0, Type: "URL_DECODE"}},
			SensitivityLevel:    "MEDIUM",
		}})
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("sqli", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts an xss match on the body with oversize handling", func() {
		statement := stmt(&AwsWafWebAclStatement_XssMatch{XssMatch: &AwsWafWebAclXssMatchStatement{
			FieldToMatch:        &AwsWafWebAclFieldToMatch{Field: &AwsWafWebAclFieldToMatch_Body{Body: &AwsWafWebAclBodyMatch{OversizeHandling: "MATCH"}}},
			TextTransformations: []*AwsWafWebAclTextTransformation{{Priority: 0, Type: "HTML_ENTITY_DECODE"}},
		}})
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("xss", 1, statement)))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid size constraint operator", func() {
		statement := stmt(&AwsWafWebAclStatement_SizeConstraint{SizeConstraint: &AwsWafWebAclSizeConstraintStatement{
			ComparisonOperator:  "GTE",
			Size:                8192,
			FieldToMatch:        &AwsWafWebAclFieldToMatch{Field: &AwsWafWebAclFieldToMatch_Body{Body: &AwsWafWebAclBodyMatch{}}},
			TextTransformations: []*AwsWafWebAclTextTransformation{{Priority: 0, Type: "NONE"}},
		}})
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("size", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a regex match over 512 characters", func() {
		long := ""
		for i := 0; i < 52; i++ {
			long += "0123456789"
		}
		statement := stmt(&AwsWafWebAclStatement_RegexMatch{RegexMatch: &AwsWafWebAclRegexMatchStatement{
			RegexString:         long,
			FieldToMatch:        &AwsWafWebAclFieldToMatch{Field: &AwsWafWebAclFieldToMatch_UriPath{UriPath: true}},
			TextTransformations: []*AwsWafWebAclTextTransformation{{Priority: 0, Type: "NONE"}},
		}})
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("re", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid label match scope", func() {
		statement := stmt(&AwsWafWebAclStatement_LabelMatch{LabelMatch: &AwsWafWebAclLabelMatchStatement{
			Scope: "PREFIX",
			Key:   "ns:label",
		}})
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("lbl", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an asn match with more than 100 entries", func() {
		asns := make([]uint32, 101)
		for i := range asns {
			asns[i] = uint32(64500 + i)
		}
		statement := stmt(&AwsWafWebAclStatement_AsnMatch{AsnMatch: &AwsWafWebAclAsnMatchStatement{AsnList: asns}})
		gomega.Expect(protovalidate.Validate(aclWithRules(matchRule("asn", 1, statement)))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a custom statement escape hatch", func() {
		payload, err := structpb.NewStruct(map[string]interface{}{
			"SqliMatchStatement": map[string]interface{}{
				"FieldToMatch":        map[string]interface{}{"Body": map[string]interface{}{}},
				"TextTransformations": []interface{}{map[string]interface{}{"Priority": 0, "Type": "URL_DECODE"}},
			},
		})
		gomega.Expect(err).To(gomega.BeNil())
		rule := matchRule("custom", 1, stmt(&AwsWafWebAclStatement_CustomStatement{CustomStatement: payload}))
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	// =========================================================================
	// Managed rule group configs
	// =========================================================================

	ginkgo.It("accepts a bot control config", func() {
		rule := managedRuleGroupRule("bots", 1, "AWSManagedRulesBotControlRuleSet")
		rule.Statement.GetManagedRuleGroup().ManagedRuleGroupConfigs = &AwsWafWebAclManagedRuleGroupConfigs{
			BotControl: &AwsWafWebAclBotControlConfig{InspectionLevel: "TARGETED"},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid bot control inspection level", func() {
		rule := managedRuleGroupRule("bots", 1, "AWSManagedRulesBotControlRuleSet")
		rule.Statement.GetManagedRuleGroup().ManagedRuleGroupConfigs = &AwsWafWebAclManagedRuleGroupConfigs{
			BotControl: &AwsWafWebAclBotControlConfig{InspectionLevel: "DEEP"},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts an ATP config with request and response inspection", func() {
		rule := managedRuleGroupRule("atp", 1, "AWSManagedRulesATPRuleSet")
		rule.Statement.GetManagedRuleGroup().ManagedRuleGroupConfigs = &AwsWafWebAclManagedRuleGroupConfigs{
			AccountTakeoverPrevention: &AwsWafWebAclAtpConfig{
				LoginPath: "/api/login",
				RequestInspection: &AwsWafWebAclAtpRequestInspection{
					PayloadType:   "JSON",
					UsernameField: &AwsWafWebAclFieldIdentifier{Identifier: "/email"},
					PasswordField: &AwsWafWebAclFieldIdentifier{Identifier: "/password"},
				},
				ResponseInspection: &AwsWafWebAclResponseInspection{
					StatusCode: &AwsWafWebAclResponseInspectionStatusCode{
						SuccessCodes: []int32{200},
						FailureCodes: []int32{401, 403},
					},
				},
			},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an ATP config without a login path", func() {
		rule := managedRuleGroupRule("atp", 1, "AWSManagedRulesATPRuleSet")
		rule.Statement.GetManagedRuleGroup().ManagedRuleGroupConfigs = &AwsWafWebAclManagedRuleGroupConfigs{
			AccountTakeoverPrevention: &AwsWafWebAclAtpConfig{},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an ATP request inspection with an invalid payload type", func() {
		rule := managedRuleGroupRule("atp", 1, "AWSManagedRulesATPRuleSet")
		rule.Statement.GetManagedRuleGroup().ManagedRuleGroupConfigs = &AwsWafWebAclManagedRuleGroupConfigs{
			AccountTakeoverPrevention: &AwsWafWebAclAtpConfig{
				LoginPath: "/api/login",
				RequestInspection: &AwsWafWebAclAtpRequestInspection{
					PayloadType:   "XML",
					UsernameField: &AwsWafWebAclFieldIdentifier{Identifier: "/email"},
					PasswordField: &AwsWafWebAclFieldIdentifier{Identifier: "/password"},
				},
			},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts an ACFP config", func() {
		rule := managedRuleGroupRule("acfp", 1, "AWSManagedRulesACFPRuleSet")
		rule.Statement.GetManagedRuleGroup().ManagedRuleGroupConfigs = &AwsWafWebAclManagedRuleGroupConfigs{
			AccountCreationFraudPrevention: &AwsWafWebAclAcfpConfig{
				CreationPath:         "/api/signup",
				RegistrationPagePath: "/register",
				RequestInspection: &AwsWafWebAclAcfpRequestInspection{
					PayloadType: "JSON",
					EmailField:  &AwsWafWebAclFieldIdentifier{Identifier: "/email"},
					PhoneNumberFields: &AwsWafWebAclFieldIdentifiers{
						Identifiers: []string{"/phone"},
					},
				},
			},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an ACFP config without request inspection", func() {
		rule := managedRuleGroupRule("acfp", 1, "AWSManagedRulesACFPRuleSet")
		rule.Statement.GetManagedRuleGroup().ManagedRuleGroupConfigs = &AwsWafWebAclManagedRuleGroupConfigs{
			AccountCreationFraudPrevention: &AwsWafWebAclAcfpConfig{
				CreationPath:         "/api/signup",
				RegistrationPagePath: "/register",
			},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts an anti-DDoS config", func() {
		rule := managedRuleGroupRule("ddos", 1, "AWSManagedRulesAntiDDoSRuleSet")
		rule.Statement.GetManagedRuleGroup().ManagedRuleGroupConfigs = &AwsWafWebAclManagedRuleGroupConfigs{
			AntiDdos: &AwsWafWebAclAntiDdosConfig{
				ClientSideAction: &AwsWafWebAclAntiDdosClientSideAction{
					UsageOfAction:               "ENABLED",
					Sensitivity:                 "HIGH",
					ExemptUriRegularExpressions: []string{"^/api/webhooks/.*"},
				},
				SensitivityToBlock: "LOW",
			},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an anti-DDoS config without a client side action", func() {
		rule := managedRuleGroupRule("ddos", 1, "AWSManagedRulesAntiDDoSRuleSet")
		rule.Statement.GetManagedRuleGroup().ManagedRuleGroupConfigs = &AwsWafWebAclManagedRuleGroupConfigs{
			AntiDdos: &AwsWafWebAclAntiDdosConfig{},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid rule action override", func() {
		rule := managedRuleGroupRule("common", 1, "AWSManagedRulesCommonRuleSet")
		rule.Statement.GetManagedRuleGroup().RuleActionOverrides = []*AwsWafWebAclRuleActionOverride{
			{Name: "SizeRestrictions_BODY", Action: "drop"},
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Top-level configs
	// =========================================================================

	ginkgo.It("accepts web-ACL-wide captcha and challenge immunity windows", func() {
		spec := minimalSpec()
		spec.CaptchaConfig = &AwsWafWebAclImmunityTimeConfig{ImmunityTimeSec: 300}
		spec.ChallengeConfig = &AwsWafWebAclImmunityTimeConfig{ImmunityTimeSec: 86400}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an immunity window below 60 seconds", func() {
		spec := minimalSpec()
		spec.CaptchaConfig = &AwsWafWebAclImmunityTimeConfig{ImmunityTimeSec: 30}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts association config body limits", func() {
		spec := minimalSpec()
		spec.AssociationConfig = &AwsWafWebAclAssociationConfig{
			ApiGatewayRequestBodyLimit:      "KB_48",
			CognitoUserPoolRequestBodyLimit: "KB_32",
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid body inspection limit", func() {
		spec := minimalSpec()
		spec.AssociationConfig = &AwsWafWebAclAssociationConfig{
			CloudfrontRequestBodyLimit: "KB_128",
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a data protection config", func() {
		spec := minimalSpec()
		spec.DataProtectionConfig = &AwsWafWebAclDataProtectionConfig{
			DataProtections: []*AwsWafWebAclDataProtection{
				{FieldType: "SINGLE_HEADER", FieldKeys: []string{"authorization"}, Action: "HASH"},
				{FieldType: "BODY", Action: "SUBSTITUTION"},
			},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).To(gomega.BeNil())
	})

	ginkgo.It("accepts a SINGLE_HEADER data protection without field keys (protects all headers)", func() {
		// AWS contract: omitting field_keys for a SINGLE_* type protects
		// ALL keys of that type — this shape is legal, not an error.
		spec := minimalSpec()
		spec.DataProtectionConfig = &AwsWafWebAclDataProtectionConfig{
			DataProtections: []*AwsWafWebAclDataProtection{
				{FieldType: "SINGLE_HEADER", Action: "HASH"},
			},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a BODY data protection with field keys", func() {
		spec := minimalSpec()
		spec.DataProtectionConfig = &AwsWafWebAclDataProtectionConfig{
			DataProtections: []*AwsWafWebAclDataProtection{
				{FieldType: "BODY", FieldKeys: []string{"x"}, Action: "HASH"},
			},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid data protection action", func() {
		spec := minimalSpec()
		spec.DataProtectionConfig = &AwsWafWebAclDataProtectionConfig{
			DataProtections: []*AwsWafWebAclDataProtection{
				{FieldType: "QUERY_STRING", Action: "DELETE"},
			},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Custom response bodies + logging
	// =========================================================================

	ginkgo.It("accepts custom response bodies with valid content types", func() {
		spec := minimalSpec()
		spec.CustomResponseBodies = []*AwsWafWebAclCustomResponseBody{
			{Key: "blocked", Content: "{\"error\": \"blocked\"}", ContentType: "APPLICATION_JSON"},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).To(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid custom response body content type", func() {
		spec := minimalSpec()
		spec.CustomResponseBodies = []*AwsWafWebAclCustomResponseBody{
			{Key: "blocked", Content: "<xml/>", ContentType: "APPLICATION_XML"},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a custom response code outside 200-600", func() {
		rule := rateBasedRule("rate", 1, 1000)
		rule.CustomResponse = &AwsWafWebAclCustomResponse{ResponseCode: 199}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a logging config with a destination reference", func() {
		spec := minimalSpec()
		spec.Logging = &AwsWafWebAclLoggingConfig{
			DestinationArn:      strRef("arn:aws:logs:us-west-2:111122223333:log-group:aws-waf-logs-test"),
			RedactedHeaderNames: []string{"authorization"},
			RedactUriPath:       true,
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a logging config without a destination", func() {
		spec := minimalSpec()
		spec.Logging = &AwsWafWebAclLoggingConfig{}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Forwarded IP config
	// =========================================================================

	ginkgo.It("rejects an invalid forwarded-ip fallback behavior", func() {
		rule := geoMatchRule("geo", 1, []string{"US"})
		rule.Statement.GetGeoMatch().ForwardedIpConfig = &AwsWafWebAclForwardedIpConfig{
			HeaderName:       "X-Forwarded-For",
			FallbackBehavior: "IGNORE",
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid forwarded-ip position", func() {
		rule := ipSetRefRule("ips", 1, "arn:aws:wafv2:us-west-2:111122223333:regional/ipset/x/1")
		rule.Statement.GetIpSetReference().ForwardedIpConfig = &AwsWafWebAclForwardedIpConfig{
			HeaderName:       "X-Forwarded-For",
			FallbackBehavior: "MATCH",
			Position:         "MIDDLE",
		}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Rule priority presence (0 is AWS-legal)
	// =========================================================================

	ginkgo.It("accepts a rule with priority 0", func() {
		gomega.Expect(protovalidate.Validate(aclWithRules(
			matchRule("first", 0, byteMatchStatement("/admin")),
		))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a rule without a priority", func() {
		rule := matchRule("no-priority", 0, byteMatchStatement("/admin"))
		rule.Priority = nil
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a negative rule priority", func() {
		gomega.Expect(protovalidate.Validate(aclWithRules(
			matchRule("negative", -1, byteMatchStatement("/admin")),
		))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Logging: full redaction surface + log filtering
	// =========================================================================

	// helper for a logging config with a valid destination.
	loggingWith := func(mutate func(*AwsWafWebAclLoggingConfig)) *AwsWafWebAcl {
		spec := minimalSpec()
		spec.Logging = &AwsWafWebAclLoggingConfig{
			DestinationArn: strRef("arn:aws:logs:us-west-2:111122223333:log-group:aws-waf-logs-test"),
		}
		mutate(spec.Logging)
		return minimalAcl(spec)
	}

	ginkgo.It("accepts method redaction", func() {
		gomega.Expect(protovalidate.Validate(loggingWith(func(l *AwsWafWebAclLoggingConfig) {
			l.RedactMethod = true
		}))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a redacted header name over 64 characters", func() {
		gomega.Expect(protovalidate.Validate(loggingWith(func(l *AwsWafWebAclLoggingConfig) {
			l.RedactedHeaderNames = []string{string(make([]byte, 65))}
		}))).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a logging filter keeping only blocked requests", func() {
		gomega.Expect(protovalidate.Validate(loggingWith(func(l *AwsWafWebAclLoggingConfig) {
			l.Filter = &AwsWafWebAclLoggingFilterConfig{
				DefaultBehavior: "DROP",
				Filters: []*AwsWafWebAclLoggingFilter{{
					Behavior:    "KEEP",
					Requirement: "MEETS_ANY",
					Conditions: []*AwsWafWebAclLoggingFilterCondition{
						{Action: "BLOCK"},
						{Action: "EXCLUDED_AS_COUNT"},
					},
				}},
			}
		}))).To(gomega.BeNil())
	})

	ginkgo.It("accepts a logging filter condition on a label name", func() {
		gomega.Expect(protovalidate.Validate(loggingWith(func(l *AwsWafWebAclLoggingConfig) {
			l.Filter = &AwsWafWebAclLoggingFilterConfig{
				DefaultBehavior: "KEEP",
				Filters: []*AwsWafWebAclLoggingFilter{{
					Behavior:    "DROP",
					Requirement: "MEETS_ALL",
					Conditions: []*AwsWafWebAclLoggingFilterCondition{
						{LabelName: "awswaf:managed:aws:bot-control:bot:category:monitoring"},
					},
				}},
			}
		}))).To(gomega.BeNil())
	})

	ginkgo.It("rejects a logging filter without filters", func() {
		gomega.Expect(protovalidate.Validate(loggingWith(func(l *AwsWafWebAclLoggingConfig) {
			l.Filter = &AwsWafWebAclLoggingFilterConfig{DefaultBehavior: "KEEP"}
		}))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a logging filter with an invalid default behavior", func() {
		gomega.Expect(protovalidate.Validate(loggingWith(func(l *AwsWafWebAclLoggingConfig) {
			l.Filter = &AwsWafWebAclLoggingFilterConfig{
				DefaultBehavior: "IGNORE",
				Filters: []*AwsWafWebAclLoggingFilter{{
					Behavior:    "KEEP",
					Requirement: "MEETS_ANY",
					Conditions:  []*AwsWafWebAclLoggingFilterCondition{{Action: "BLOCK"}},
				}},
			}
		}))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a logging filter condition with an invalid action", func() {
		gomega.Expect(protovalidate.Validate(loggingWith(func(l *AwsWafWebAclLoggingConfig) {
			l.Filter = &AwsWafWebAclLoggingFilterConfig{
				DefaultBehavior: "KEEP",
				Filters: []*AwsWafWebAclLoggingFilter{{
					Behavior:    "DROP",
					Requirement: "MEETS_ANY",
					Conditions:  []*AwsWafWebAclLoggingFilterCondition{{Action: "DENY"}},
				}},
			}
		}))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a logging filter condition carrying both action and label", func() {
		gomega.Expect(protovalidate.Validate(loggingWith(func(l *AwsWafWebAclLoggingConfig) {
			l.Filter = &AwsWafWebAclLoggingFilterConfig{
				DefaultBehavior: "KEEP",
				Filters: []*AwsWafWebAclLoggingFilter{{
					Behavior:    "DROP",
					Requirement: "MEETS_ANY",
					Conditions: []*AwsWafWebAclLoggingFilterCondition{
						{Action: "BLOCK", LabelName: "awswaf:x:y"},
					},
				}},
			}
		}))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an empty logging filter condition", func() {
		gomega.Expect(protovalidate.Validate(loggingWith(func(l *AwsWafWebAclLoggingConfig) {
			l.Filter = &AwsWafWebAclLoggingFilterConfig{
				DefaultBehavior: "KEEP",
				Filters: []*AwsWafWebAclLoggingFilter{{
					Behavior:    "DROP",
					Requirement: "MEETS_ANY",
					Conditions:  []*AwsWafWebAclLoggingFilterCondition{{}},
				}},
			}
		}))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Data protection: field_keys scope (omitting keys protects all keys)
	// =========================================================================

	ginkgo.It("rejects a QUERY_STRING data protection carrying field keys", func() {
		spec := minimalSpec()
		spec.DataProtectionConfig = &AwsWafWebAclDataProtectionConfig{
			DataProtections: []*AwsWafWebAclDataProtection{{
				FieldType: "QUERY_STRING",
				FieldKeys: []string{"q"},
				Action:    "SUBSTITUTION",
			}},
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	// =========================================================================
	// Value-domain repairs: token domains, metric names, country codes,
	// custom response body keys, custom header bounds, rule labels
	// =========================================================================

	ginkgo.It("rejects a token domain with illegal characters", func() {
		spec := minimalSpec()
		spec.TokenDomains = []string{"bad domain.com"}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a reserved CloudWatch metric name", func() {
		spec := minimalSpec()
		spec.VisibilityConfig = &AwsWafWebAclVisibilityConfig{
			CloudwatchMetricsEnabled: true,
			SampledRequestsEnabled:   true,
			MetricName:               "Default_Action",
		}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a lowercase country code", func() {
		gomega.Expect(protovalidate.Validate(aclWithRules(
			geoMatchRule("geo", 1, []string{"us"}),
		))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a custom response body key with illegal characters", func() {
		spec := minimalSpec()
		spec.CustomResponseBodies = []*AwsWafWebAclCustomResponseBody{{
			Key:         "bad key!",
			Content:     "denied",
			ContentType: "TEXT_PLAIN",
		}}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a custom header value over 255 characters", func() {
		spec := minimalSpec()
		spec.DefaultAction.CustomRequestHeaders = []*AwsWafWebAclCustomHeader{{
			Name:  "x-origin-check",
			Value: string(make([]byte, 256)),
		}}
		gomega.Expect(protovalidate.Validate(minimalAcl(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a rule label with illegal characters", func() {
		rule := matchRule("labeled", 1, byteMatchStatement("/admin"))
		rule.RuleLabels = []string{"has space:label"}
		gomega.Expect(protovalidate.Validate(aclWithRules(rule))).NotTo(gomega.BeNil())
	})
})
