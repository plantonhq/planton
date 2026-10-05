package azurefrontdoorfirewallpolicyv1alpha1

import (
	"fmt"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureFrontDoorFirewallPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorFirewallPolicySpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimal valid spec: a STANDARD detection-mode policy with no rules --
// the smallest thing Azure accepts.
func minimalSpec() *AzureFrontDoorFirewallPolicy {
	return &AzureFrontDoorFirewallPolicy{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorFirewallPolicy",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-front-door-firewall-policy",
		},
		Spec: &AzureFrontDoorFirewallPolicySpec{
			ResourceGroup: literal("test-rg"),
			PolicyName:    "edgewaf",
			Mode:          AzureFrontDoorFirewallPolicyMode_PREVENTION,
		},
	}
}

func premiumSpec() *AzureFrontDoorFirewallPolicy {
	input := minimalSpec()
	input.Spec.Sku = AzureFrontDoorFirewallPolicySku_PREMIUM
	return input
}

// blockRule is the simplest valid custom rule: block one CIDR.
func blockRule(name string) *AzureFrontDoorFirewallPolicyCustomRule {
	return &AzureFrontDoorFirewallPolicyCustomRule{
		Name:     name,
		RuleType: AzureFrontDoorFirewallPolicyCustomRuleType_MATCH_RULE,
		Action:   AzureFrontDoorFirewallPolicyCustomRuleAction_BLOCK,
		MatchConditions: []*AzureFrontDoorFirewallPolicyMatchCondition{{
			MatchVariable: AzureFrontDoorFirewallPolicyMatchVariable_REMOTE_ADDR,
			Operator:      AzureFrontDoorFirewallPolicyOperator_IP_MATCH,
			MatchValues:   []string{"203.0.113.0/24"},
		}},
	}
}

// managedRuleSet is a valid Microsoft_DefaultRuleSet 2.1 entry.
func managedRuleSet() *AzureFrontDoorFirewallPolicyManagedRuleSet {
	return &AzureFrontDoorFirewallPolicyManagedRuleSet{
		Type:    "Microsoft_DefaultRuleSet",
		Version: "2.1",
		Action:  AzureFrontDoorFirewallPolicyManagedRuleSetAction_RULE_SET_BLOCK,
	}
}

var _ = ginkgo.Describe("AzureFrontDoorFirewallPolicySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for the minimal spec", func() {
			err := protovalidate.Validate(minimalSpec())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a STANDARD policy with custom rules and policy settings", func() {
			input := minimalSpec()
			input.Spec.Sku = AzureFrontDoorFirewallPolicySku_STANDARD
			input.Spec.Enabled = proto.Bool(true)
			input.Spec.RequestBodyCheckEnabled = proto.Bool(false)
			input.Spec.RedirectUrl = "https://example.com/blocked"
			input.Spec.CustomBlockResponseStatusCode = proto.Int32(429)
			input.Spec.CustomBlockResponseBody = "PGgxPmJsb2NrZWQ8L2gxPg=="
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{blockRule("blockbadcidr")}
			input.Spec.Tags = map[string]string{"team": "edge"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a rate-limit rule with window, threshold, and priority", func() {
			rule := blockRule("ratelimit")
			rule.RuleType = AzureFrontDoorFirewallPolicyCustomRuleType_RATE_LIMIT_RULE
			rule.Priority = proto.Int32(50)
			rule.RateLimitDurationInMinutes = proto.Int32(5)
			rule.RateLimitThreshold = proto.Int32(100)
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a condition with selector, negation, and transforms", func() {
			rule := blockRule("headercheck")
			rule.MatchConditions = []*AzureFrontDoorFirewallPolicyMatchCondition{{
				MatchVariable:   AzureFrontDoorFirewallPolicyMatchVariable_REQUEST_HEADER,
				Selector:        "User-Agent",
				Operator:        AzureFrontDoorFirewallPolicyOperator_CONTAINS,
				MatchValues:     []string{"badbot"},
				NegateCondition: true,
				Transforms: []AzureFrontDoorFirewallPolicyTransform{
					AzureFrontDoorFirewallPolicyTransform_LOWERCASE,
					AzureFrontDoorFirewallPolicyTransform_TRIM,
				},
			}}
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept PREMIUM with managed rules, challenge expirations, and challenge actions", func() {
			input := premiumSpec()
			input.Spec.JsChallengeCookieExpirationInMinutes = proto.Int32(45)
			input.Spec.CaptchaCookieExpirationInMinutes = proto.Int32(60)
			challengeRule := blockRule("botgate")
			challengeRule.Action = AzureFrontDoorFirewallPolicyCustomRuleAction_JS_CHALLENGE
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{challengeRule}
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{managedRuleSet()}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a CAPTCHA custom-rule action on PREMIUM", func() {
			rule := blockRule("humangate")
			rule.Action = AzureFrontDoorFirewallPolicyCustomRuleAction_CAPTCHA
			input := premiumSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept the legacy DefaultRuleSet at version preview-0.1", func() {
			set := &AzureFrontDoorFirewallPolicyManagedRuleSet{
				Type:    "DefaultRuleSet",
				Version: "preview-0.1",
				Action:  AzureFrontDoorFirewallPolicyManagedRuleSetAction_RULE_SET_LOG,
			}
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept managed-rule exclusions and anomaly-scoring overrides on a 2.x set", func() {
			set := managedRuleSet()
			set.Exclusions = []*AzureFrontDoorFirewallPolicyManagedRuleExclusion{{
				MatchVariable: AzureFrontDoorFirewallPolicyExclusionMatchVariable_EXCLUDE_REQUEST_COOKIE_NAMES,
				Operator:      AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS,
				Selector:      "session-token",
			}}
			// Both actions the 2.x gate allows: ANOMALY_SCORING and LOG.
			set.Overrides = []*AzureFrontDoorFirewallPolicyManagedRuleGroupOverride{{
				RuleGroupName: "SQLI",
				Rules: []*AzureFrontDoorFirewallPolicyManagedRuleOverride{
					{
						RuleId:  "942100",
						Enabled: proto.Bool(true),
						Action:  AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_ANOMALY_SCORING,
					},
					{
						RuleId:  "942200",
						Enabled: proto.Bool(true),
						Action:  AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_LOG,
					},
				},
			}}
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept the legacy DefaultRuleSet at version 1.0 with a block override", func() {
			set := &AzureFrontDoorFirewallPolicyManagedRuleSet{
				Type:    "DefaultRuleSet",
				Version: "1.0",
				Action:  AzureFrontDoorFirewallPolicyManagedRuleSetAction_RULE_SET_LOG,
				Overrides: []*AzureFrontDoorFirewallPolicyManagedRuleGroupOverride{{
					RuleGroupName: "SQLI",
					Rules: []*AzureFrontDoorFirewallPolicyManagedRuleOverride{{
						RuleId: "942100",
						Action: AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_BLOCK,
					}},
				}},
			}
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a JS-challenge override on the bot manager set", func() {
			set := &AzureFrontDoorFirewallPolicyManagedRuleSet{
				Type:    "Microsoft_BotManagerRuleSet",
				Version: "1.1",
				Action:  AzureFrontDoorFirewallPolicyManagedRuleSetAction_RULE_SET_BLOCK,
				Overrides: []*AzureFrontDoorFirewallPolicyManagedRuleGroupOverride{{
					RuleGroupName: "UnknownBots",
					Rules: []*AzureFrontDoorFirewallPolicyManagedRuleOverride{{
						RuleId: "Bot300700",
						Action: AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_JS_CHALLENGE,
					}},
				}},
			}
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept log scrubbing with keyed and equals-any rules", func() {
			input := minimalSpec()
			input.Spec.LogScrubbing = &AzureFrontDoorFirewallPolicyLogScrubbing{
				ScrubbingRules: []*AzureFrontDoorFirewallPolicyScrubbingRule{
					{
						MatchVariable: AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES,
						Operator:      AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS,
						Selector:      "Authorization",
					},
					{
						MatchVariable: AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_IP_ADDRESS,
						Operator:      AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS_ANY,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("policy_name format", func() {

		ginkgo.It("should accept a name at the 128-character maximum", func() {
			input := minimalSpec()
			input.Spec.PolicyName = "a" + strings.Repeat("b", 127)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a name with hyphens", func() {
			input := minimalSpec()
			input.Spec.PolicyName = "edge-waf"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a name starting with a digit", func() {
			input := minimalSpec()
			input.Spec.PolicyName = "1edgewaf"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a name longer than 128 characters", func() {
			name := "a"
			for i := 0; i < 128; i++ {
				name += "a"
			}
			input := minimalSpec()
			input.Spec.PolicyName = name
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("required scalar fields", func() {

		ginkgo.It("should reject a missing mode", func() {
			input := minimalSpec()
			input.Spec.Mode = AzureFrontDoorFirewallPolicyMode_azure_front_door_firewall_policy_mode_unspecified
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing resource group", func() {
			input := minimalSpec()
			input.Spec.ResourceGroup = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("policy settings", func() {

		ginkgo.It("should reject a redirect_url that is not http(s)", func() {
			input := minimalSpec()
			input.Spec.RedirectUrl = "ftp://example.com/blocked"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a block status code outside the allowed set", func() {
			input := minimalSpec()
			input.Spec.CustomBlockResponseStatusCode = proto.Int32(404)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept the WAF-specific 99x block status codes at both edges", func() {
			for _, code := range []int32{990, 999} {
				input := minimalSpec()
				input.Spec.CustomBlockResponseStatusCode = proto.Int32(code)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should reject 989, one below the 99x block-status range", func() {
			input := minimalSpec()
			input.Spec.CustomBlockResponseStatusCode = proto.Int32(989)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a block body that is not base64", func() {
			input := minimalSpec()
			input.Spec.CustomBlockResponseBody = "<h1>blocked</h1>"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Premium-only gates", func() {

		ginkgo.It("should reject managed rules on the STANDARD sku", func() {
			input := minimalSpec()
			input.Spec.Sku = AzureFrontDoorFirewallPolicySku_STANDARD
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{managedRuleSet()}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject managed rules when the sku is unspecified (deploys STANDARD)", func() {
			input := minimalSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{managedRuleSet()}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the JS-challenge expiration on STANDARD", func() {
			input := minimalSpec()
			input.Spec.JsChallengeCookieExpirationInMinutes = proto.Int32(30)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the CAPTCHA expiration on STANDARD", func() {
			input := minimalSpec()
			input.Spec.CaptchaCookieExpirationInMinutes = proto.Int32(30)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a JS_CHALLENGE custom-rule action on STANDARD", func() {
			rule := blockRule("botgate")
			rule.Action = AzureFrontDoorFirewallPolicyCustomRuleAction_JS_CHALLENGE
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a CAPTCHA custom-rule action on STANDARD", func() {
			rule := blockRule("humangate")
			rule.Action = AzureFrontDoorFirewallPolicyCustomRuleAction_CAPTCHA
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("challenge expiration bounds", func() {

		ginkgo.It("should accept both lifetimes at the 5 and 1440 edges", func() {
			for _, minutes := range []int32{5, 1440} {
				input := premiumSpec()
				input.Spec.JsChallengeCookieExpirationInMinutes = proto.Int32(minutes)
				input.Spec.CaptchaCookieExpirationInMinutes = proto.Int32(minutes)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should reject a JS-challenge lifetime below 5 minutes", func() {
			input := premiumSpec()
			input.Spec.JsChallengeCookieExpirationInMinutes = proto.Int32(4)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a JS-challenge lifetime above 1440 minutes", func() {
			input := premiumSpec()
			input.Spec.JsChallengeCookieExpirationInMinutes = proto.Int32(1441)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a CAPTCHA lifetime below 5 minutes", func() {
			input := premiumSpec()
			input.Spec.CaptchaCookieExpirationInMinutes = proto.Int32(4)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a CAPTCHA lifetime above 1440 minutes", func() {
			input := premiumSpec()
			input.Spec.CaptchaCookieExpirationInMinutes = proto.Int32(1441)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("custom rules", func() {

		ginkgo.It("should reject a rule without a name", func() {
			rule := blockRule("")
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule without a rule type", func() {
			rule := blockRule("norule")
			rule.RuleType = AzureFrontDoorFirewallPolicyCustomRuleType_azure_front_door_firewall_policy_custom_rule_type_unspecified
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule without an action", func() {
			rule := blockRule("noaction")
			rule.Action = AzureFrontDoorFirewallPolicyCustomRuleAction_azure_front_door_firewall_policy_custom_rule_action_unspecified
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a zero rate-limit window", func() {
			rule := blockRule("ratelimit")
			rule.RateLimitDurationInMinutes = proto.Int32(0)
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("match conditions", func() {

		ginkgo.It("should reject a condition without match values", func() {
			rule := blockRule("novalues")
			rule.MatchConditions[0].MatchValues = nil
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a match value longer than 256 characters", func() {
			long := ""
			for i := 0; i < 257; i++ {
				long += "v"
			}
			rule := blockRule("longvalue")
			rule.MatchConditions[0].MatchValues = []string{long}
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a condition without an operator", func() {
			rule := blockRule("nooperator")
			rule.MatchConditions[0].Operator = AzureFrontDoorFirewallPolicyOperator_azure_front_door_firewall_policy_operator_unspecified
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject more than 10 conditions on one rule", func() {
			rule := blockRule("toomanyconditions")
			conditions := make([]*AzureFrontDoorFirewallPolicyMatchCondition, 0, 11)
			for i := 0; i < 11; i++ {
				conditions = append(conditions, &AzureFrontDoorFirewallPolicyMatchCondition{
					MatchVariable: AzureFrontDoorFirewallPolicyMatchVariable_REMOTE_ADDR,
					Operator:      AzureFrontDoorFirewallPolicyOperator_IP_MATCH,
					MatchValues:   []string{fmt.Sprintf("203.0.113.%d/32", i)},
				})
			}
			rule.MatchConditions = conditions
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject more than 600 match values on one condition", func() {
			values := make([]string, 601)
			for i := range values {
				values[i] = fmt.Sprintf("v%d", i)
			}
			rule := blockRule("toomanyvalues")
			rule.MatchConditions[0].MatchValues = values
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject more than 5 transforms", func() {
			rule := blockRule("toomanytransforms")
			rule.MatchConditions[0].Transforms = []AzureFrontDoorFirewallPolicyTransform{
				AzureFrontDoorFirewallPolicyTransform_LOWERCASE,
				AzureFrontDoorFirewallPolicyTransform_UPPERCASE,
				AzureFrontDoorFirewallPolicyTransform_TRIM,
				AzureFrontDoorFirewallPolicyTransform_URL_DECODE,
				AzureFrontDoorFirewallPolicyTransform_URL_ENCODE,
				AzureFrontDoorFirewallPolicyTransform_REMOVE_NULLS,
			}
			input := minimalSpec()
			input.Spec.CustomRules = []*AzureFrontDoorFirewallPolicyCustomRule{rule}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("managed rule sets", func() {

		ginkgo.It("should reject the legacy DefaultRuleSet above version 1.0", func() {
			set := managedRuleSet()
			set.Type = "DefaultRuleSet"
			set.Version = "1.1"
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject Microsoft_DefaultRuleSet at version 1.0", func() {
			set := managedRuleSet()
			set.Version = "1.0"
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject Microsoft_DefaultRuleSet on a 0.x version", func() {
			set := managedRuleSet()
			set.Version = "0.9"
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a non-anomaly-scoring block override on a 2.x set", func() {
			set := managedRuleSet()
			set.Overrides = []*AzureFrontDoorFirewallPolicyManagedRuleGroupOverride{{
				RuleGroupName: "SQLI",
				Rules: []*AzureFrontDoorFirewallPolicyManagedRuleOverride{{
					RuleId: "942100",
					Action: AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_BLOCK,
				}},
			}}
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an anomaly-scoring override on a 1.x set", func() {
			set := &AzureFrontDoorFirewallPolicyManagedRuleSet{
				Type:    "Microsoft_DefaultRuleSet",
				Version: "1.1",
				Action:  AzureFrontDoorFirewallPolicyManagedRuleSetAction_RULE_SET_BLOCK,
				Overrides: []*AzureFrontDoorFirewallPolicyManagedRuleGroupOverride{{
					RuleGroupName: "SQLI",
					Rules: []*AzureFrontDoorFirewallPolicyManagedRuleOverride{{
						RuleId: "942100",
						Action: AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_ANOMALY_SCORING,
					}},
				}},
			}
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a JS-challenge override outside the bot manager set", func() {
			set := &AzureFrontDoorFirewallPolicyManagedRuleSet{
				Type:    "Microsoft_DefaultRuleSet",
				Version: "1.1",
				Action:  AzureFrontDoorFirewallPolicyManagedRuleSetAction_RULE_SET_BLOCK,
				Overrides: []*AzureFrontDoorFirewallPolicyManagedRuleGroupOverride{{
					RuleGroupName: "SQLI",
					Rules: []*AzureFrontDoorFirewallPolicyManagedRuleOverride{{
						RuleId: "942100",
						Action: AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_JS_CHALLENGE,
					}},
				}},
			}
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an exclusion without a selector", func() {
			set := managedRuleSet()
			set.Exclusions = []*AzureFrontDoorFirewallPolicyManagedRuleExclusion{{
				MatchVariable: AzureFrontDoorFirewallPolicyExclusionMatchVariable_EXCLUDE_REQUEST_HEADER_NAMES,
				Operator:      AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS,
			}}
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a set without an action", func() {
			set := managedRuleSet()
			set.Action = AzureFrontDoorFirewallPolicyManagedRuleSetAction_azure_front_door_firewall_policy_managed_rule_set_action_unspecified
			input := premiumSpec()
			input.Spec.ManagedRules = []*AzureFrontDoorFirewallPolicyManagedRuleSet{set}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("log scrubbing", func() {

		ginkgo.It("should reject an empty scrubbing-rule list", func() {
			input := minimalSpec()
			input.Spec.LogScrubbing = &AzureFrontDoorFirewallPolicyLogScrubbing{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject more than 100 scrubbing rules", func() {
			rules := make([]*AzureFrontDoorFirewallPolicyScrubbingRule, 101)
			for i := range rules {
				rules[i] = &AzureFrontDoorFirewallPolicyScrubbingRule{
					MatchVariable: AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES,
					Operator:      AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS,
					Selector:      fmt.Sprintf("header-%d", i),
				}
			}
			input := minimalSpec()
			input.Spec.LogScrubbing = &AzureFrontDoorFirewallPolicyLogScrubbing{ScrubbingRules: rules}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an exclusion-style operator on a scrubbing rule", func() {
			input := minimalSpec()
			input.Spec.LogScrubbing = &AzureFrontDoorFirewallPolicyLogScrubbing{
				ScrubbingRules: []*AzureFrontDoorFirewallPolicyScrubbingRule{{
					MatchVariable: AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES,
					Operator:      AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_CONTAINS,
					Selector:      "Auth",
				}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject SCRUB_REQUEST_IP_ADDRESS with SELECTOR_EQUALS", func() {
			input := minimalSpec()
			input.Spec.LogScrubbing = &AzureFrontDoorFirewallPolicyLogScrubbing{
				ScrubbingRules: []*AzureFrontDoorFirewallPolicyScrubbingRule{{
					MatchVariable: AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_IP_ADDRESS,
					Operator:      AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS,
					Selector:      "ip",
				}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject SELECTOR_EQUALS without a selector", func() {
			input := minimalSpec()
			input.Spec.LogScrubbing = &AzureFrontDoorFirewallPolicyLogScrubbing{
				ScrubbingRules: []*AzureFrontDoorFirewallPolicyScrubbingRule{{
					MatchVariable: AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES,
					Operator:      AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS,
				}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject SELECTOR_EQUALS_ANY with a selector", func() {
			input := minimalSpec()
			input.Spec.LogScrubbing = &AzureFrontDoorFirewallPolicyLogScrubbing{
				ScrubbingRules: []*AzureFrontDoorFirewallPolicyScrubbingRule{{
					MatchVariable: AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES,
					Operator:      AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS_ANY,
					Selector:      "Authorization",
				}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
