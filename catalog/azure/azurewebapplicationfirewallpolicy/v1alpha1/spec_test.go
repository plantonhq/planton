package azurewebapplicationfirewallpolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureWebApplicationFirewallPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureWebApplicationFirewallPolicySpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimal valid spec: an OWASP 3.2 policy with no custom rules.
func minimalSpec() *AzureWebApplicationFirewallPolicy {
	return &AzureWebApplicationFirewallPolicy{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureWebApplicationFirewallPolicy",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-waf",
		},
		Spec: &AzureWebApplicationFirewallPolicySpec{
			Region:        "eastus",
			ResourceGroup: literal("my-rg"),
			PolicyName:    "test-waf-policy",
			ManagedRules: &AzureWebApplicationFirewallPolicyManagedRules{
				ManagedRuleSets: []*AzureWebApplicationFirewallPolicyManagedRuleSet{
					{Version: "3.2"},
				},
			},
		},
	}
}

func matchRule(priority int32) *AzureWebApplicationFirewallPolicyCustomRule {
	return &AzureWebApplicationFirewallPolicyCustomRule{
		Name:     "allowOffice",
		Priority: priority,
		RuleType: AzureWebApplicationFirewallPolicyCustomRuleType_MATCH_RULE,
		Action:   AzureWebApplicationFirewallPolicyCustomRuleAction_ALLOW,
		MatchConditions: []*AzureWebApplicationFirewallPolicyMatchCondition{
			{
				MatchVariables: []*AzureWebApplicationFirewallPolicyMatchVariable{
					{VariableName: AzureWebApplicationFirewallPolicyMatchVariableName_REMOTE_ADDR},
				},
				Operator:    AzureWebApplicationFirewallPolicyMatchOperator_IP_MATCH,
				MatchValues: []string{"203.0.113.0/24"},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureWebApplicationFirewallPolicySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal OWASP policy", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept every managed rule-set version", func() {
			for _, v := range []string{"0.1", "1.0", "1.1", "2.1", "2.2", "2.2.9", "3.0", "3.1", "3.2"} {
				input := minimalSpec()
				input.Spec.ManagedRules.ManagedRuleSets[0].Version = v
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "version %q must be accepted", v)
			}
		})

		ginkgo.It("should accept the bot-manager set alongside OWASP", func() {
			input := minimalSpec()
			input.Spec.ManagedRules.ManagedRuleSets = append(input.Spec.ManagedRules.ManagedRuleSets,
				&AzureWebApplicationFirewallPolicyManagedRuleSet{
					Type:    AzureWebApplicationFirewallPolicyManagedRuleSetType_MICROSOFT_BOT_MANAGER_RULE_SET,
					Version: "1.1",
				})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an IP-allowlist match rule", func() {
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{matchRule(10)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a geo rate-limit rule with grouping", func() {
			threshold := int32(300)
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{
				{
					Name:               "rateLimitPerClient",
					Priority:           20,
					RuleType:           AzureWebApplicationFirewallPolicyCustomRuleType_RATE_LIMIT_RULE,
					Action:             AzureWebApplicationFirewallPolicyCustomRuleAction_BLOCK,
					RateLimitDuration:  AzureWebApplicationFirewallPolicyRateLimitDuration_ONE_MIN,
					RateLimitThreshold: &threshold,
					GroupRateLimitBy:   AzureWebApplicationFirewallPolicyGroupRateLimitBy_CLIENT_ADDR,
					MatchConditions: []*AzureWebApplicationFirewallPolicyMatchCondition{
						{
							MatchVariables: []*AzureWebApplicationFirewallPolicyMatchVariable{
								{VariableName: AzureWebApplicationFirewallPolicyMatchVariableName_REQUEST_URI},
							},
							Operator:    AzureWebApplicationFirewallPolicyMatchOperator_BEGINS_WITH,
							MatchValues: []string{"/api/"},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an ANY-operator condition without match values", func() {
			input := minimalSpec()
			rule := matchRule(10)
			rule.MatchConditions[0].Operator = AzureWebApplicationFirewallPolicyMatchOperator_ANY
			rule.MatchConditions[0].MatchValues = nil
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{rule}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept transforms and negation on a header condition", func() {
			input := minimalSpec()
			rule := matchRule(10)
			rule.MatchConditions[0] = &AzureWebApplicationFirewallPolicyMatchCondition{
				MatchVariables: []*AzureWebApplicationFirewallPolicyMatchVariable{
					{
						VariableName: AzureWebApplicationFirewallPolicyMatchVariableName_REQUEST_HEADERS,
						Selector:     "User-Agent",
					},
				},
				Operator:          AzureWebApplicationFirewallPolicyMatchOperator_CONTAINS,
				MatchValues:       []string{"badbot"},
				NegationCondition: true,
				Transforms: []AzureWebApplicationFirewallPolicyTransform{
					AzureWebApplicationFirewallPolicyTransform_LOWERCASE,
					AzureWebApplicationFirewallPolicyTransform_TRIM,
				},
			}
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{rule}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept rule-group overrides", func() {
			enabled := false
			input := minimalSpec()
			input.Spec.ManagedRules.ManagedRuleSets[0].RuleGroupOverrides = []*AzureWebApplicationFirewallPolicyRuleGroupOverride{
				{
					RuleGroupName: "REQUEST-942-APPLICATION-ATTACK-SQLI",
					Rules: []*AzureWebApplicationFirewallPolicyRuleOverride{
						{Id: "942100", Enabled: &enabled},
						{Id: "942200", Action: AzureWebApplicationFirewallPolicyRuleOverrideAction_OVERRIDE_LOG},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a scoped exclusion", func() {
			input := minimalSpec()
			input.Spec.ManagedRules.Exclusions = []*AzureWebApplicationFirewallPolicyManagedRulesExclusion{
				{
					MatchVariable:         AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_COOKIE_NAMES,
					SelectorMatchOperator: AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS,
					Selector:              "session-token",
					ExcludedRuleSet: &AzureWebApplicationFirewallPolicyExcludedRuleSet{
						RuleGroups: []*AzureWebApplicationFirewallPolicyExcludedRuleGroup{
							{
								RuleGroupName: "REQUEST-942-APPLICATION-ATTACK-SQLI",
								ExcludedRules: []string{"942440"},
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept detection-mode policy settings with log scrubbing", func() {
			inspect := int32(0)
			input := minimalSpec()
			input.Spec.PolicySettings = &AzureWebApplicationFirewallPolicySettings{
				Mode:                        AzureWebApplicationFirewallPolicyMode_DETECTION,
				RequestBodyInspectLimitInKb: &inspect,
				LogScrubbing: &AzureWebApplicationFirewallPolicyLogScrubbing{
					Rules: []*AzureWebApplicationFirewallPolicyLogScrubbingRule{
						{
							MatchVariable:         AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES,
							SelectorMatchOperator: AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS,
							Selector:              "Authorization",
						},
						{
							MatchVariable:         AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_IP_ADDRESS,
							SelectorMatchOperator: AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS,
						},
						{
							MatchVariable:         AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_COOKIE_NAMES,
							SelectorMatchOperator: AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS_ANY,
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing managed_rules block", func() {
			input := minimalSpec()
			input.Spec.ManagedRules = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an empty managed rule-set list", func() {
			input := minimalSpec()
			input.Spec.ManagedRules.ManagedRuleSets = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown rule-set version", func() {
			input := minimalSpec()
			input.Spec.ManagedRules.ManagedRuleSets[0].Version = "4.0"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a custom rule without match conditions", func() {
			input := minimalSpec()
			rule := matchRule(10)
			rule.MatchConditions = nil
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{rule}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a custom rule priority outside 1-100", func() {
			for _, priority := range []int32{0, 101} {
				input := minimalSpec()
				input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{matchRule(priority)}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("should reject a custom rule name with a hyphen", func() {
			input := minimalSpec()
			rule := matchRule(10)
			rule.Name = "allow-office"
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{rule}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a rate-limit rule without duration and threshold", func() {
			input := minimalSpec()
			rule := matchRule(10)
			rule.RuleType = AzureWebApplicationFirewallPolicyCustomRuleType_RATE_LIMIT_RULE
			rule.Action = AzureWebApplicationFirewallPolicyCustomRuleAction_BLOCK
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{rule}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject rate-limit dials on a match rule", func() {
			threshold := int32(100)
			input := minimalSpec()
			rule := matchRule(10)
			rule.RateLimitThreshold = &threshold
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{rule}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject ALLOW on a rate-limit rule", func() {
			threshold := int32(100)
			input := minimalSpec()
			rule := matchRule(10)
			rule.RuleType = AzureWebApplicationFirewallPolicyCustomRuleType_RATE_LIMIT_RULE
			rule.RateLimitDuration = AzureWebApplicationFirewallPolicyRateLimitDuration_ONE_MIN
			rule.RateLimitThreshold = &threshold
			rule.Action = AzureWebApplicationFirewallPolicyCustomRuleAction_ALLOW
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{rule}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject match values on an ANY-operator condition", func() {
			input := minimalSpec()
			rule := matchRule(10)
			rule.MatchConditions[0].Operator = AzureWebApplicationFirewallPolicyMatchOperator_ANY
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{rule}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a non-ANY condition without match values", func() {
			input := minimalSpec()
			rule := matchRule(10)
			rule.MatchConditions[0].MatchValues = nil
			input.Spec.CustomRules = []*AzureWebApplicationFirewallPolicyCustomRule{rule}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an exclusion without a selector", func() {
			input := minimalSpec()
			input.Spec.ManagedRules.Exclusions = []*AzureWebApplicationFirewallPolicyManagedRulesExclusion{
				{
					MatchVariable:         AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_HEADER_NAMES,
					SelectorMatchOperator: AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an excluded-rule-set version outside its narrower allowed set", func() {
			badVersion := "3.0"
			input := minimalSpec()
			input.Spec.ManagedRules.Exclusions = []*AzureWebApplicationFirewallPolicyManagedRulesExclusion{
				{
					MatchVariable:         AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_HEADER_NAMES,
					SelectorMatchOperator: AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS,
					Selector:              "x-debug",
					ExcludedRuleSet: &AzureWebApplicationFirewallPolicyExcludedRuleSet{
						Version: &badVersion,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a rule-group override without rules", func() {
			input := minimalSpec()
			input.Spec.ManagedRules.ManagedRuleSets[0].RuleGroupOverrides = []*AzureWebApplicationFirewallPolicyRuleGroupOverride{
				{RuleGroupName: "SQLI"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject max_request_body_size_in_kb outside 8-2000", func() {
			for _, size := range []int32{4, 2500} {
				s := size
				input := minimalSpec()
				input.Spec.PolicySettings = &AzureWebApplicationFirewallPolicySettings{
					MaxRequestBodySizeInKb: &s,
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("should reject a JS-challenge cookie lifetime outside 5-1440", func() {
			minutes := int32(2)
			input := minimalSpec()
			input.Spec.PolicySettings = &AzureWebApplicationFirewallPolicySettings{
				JsChallengeCookieExpirationInMinutes: &minutes,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a scrubbing rule with a CONTAINS operator", func() {
			input := minimalSpec()
			input.Spec.PolicySettings = &AzureWebApplicationFirewallPolicySettings{
				LogScrubbing: &AzureWebApplicationFirewallPolicyLogScrubbing{
					Rules: []*AzureWebApplicationFirewallPolicyLogScrubbingRule{
						{
							MatchVariable:         AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES,
							SelectorMatchOperator: AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_CONTAINS,
							Selector:              "auth",
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a SELECTOR_EQUALS scrubbing rule without a selector", func() {
			input := minimalSpec()
			input.Spec.PolicySettings = &AzureWebApplicationFirewallPolicySettings{
				LogScrubbing: &AzureWebApplicationFirewallPolicyLogScrubbing{
					Rules: []*AzureWebApplicationFirewallPolicyLogScrubbingRule{
						{
							MatchVariable:         AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES,
							SelectorMatchOperator: AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS,
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a selector on a SELECTOR_EQUALS_ANY scrubbing rule", func() {
			input := minimalSpec()
			input.Spec.PolicySettings = &AzureWebApplicationFirewallPolicySettings{
				LogScrubbing: &AzureWebApplicationFirewallPolicyLogScrubbing{
					Rules: []*AzureWebApplicationFirewallPolicyLogScrubbingRule{
						{
							MatchVariable:         AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_COOKIE_NAMES,
							SelectorMatchOperator: AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS_ANY,
							Selector:              "session",
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an empty log-scrubbing block", func() {
			input := minimalSpec()
			input.Spec.PolicySettings = &AzureWebApplicationFirewallPolicySettings{
				LogScrubbing: &AzureWebApplicationFirewallPolicyLogScrubbing{},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing policy name", func() {
			input := minimalSpec()
			input.Spec.PolicyName = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
