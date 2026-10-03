package module

import (
	"github.com/pkg/errors"
	azurefrontdoorfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorfirewallpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorFirewallPolicy.Spec

	policyArgs := &cdn.FrontdoorFirewallPolicyArgs{
		Name:              pulumi.String(spec.PolicyName),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// The policy is a global resource -- ARM fixes its location to
		// Global; no region is sent. sku_name and mode are the only other
		// required inputs.
		SkuName: pulumi.String(locals.SkuName),
		Mode:    pulumi.String(modeStrings[spec.Mode]),
		Tags:    pulumi.ToStringMap(locals.AzureTags),
	}

	// enabled / request_body_check_enabled default true on the provider;
	// send them only when the spec carries an explicit choice (stack
	// inputs never materialize proto defaults, so absence means "take
	// Azure's default").
	if spec.Enabled != nil {
		policyArgs.Enabled = pulumi.Bool(spec.GetEnabled())
	}
	if spec.RequestBodyCheckEnabled != nil {
		policyArgs.RequestBodyCheckEnabled = pulumi.Bool(spec.GetRequestBodyCheckEnabled())
	}

	if spec.RedirectUrl != "" {
		policyArgs.RedirectUrl = pulumi.String(spec.RedirectUrl)
	}
	if spec.CustomBlockResponseStatusCode != nil {
		policyArgs.CustomBlockResponseStatusCode = pulumi.Int(int(spec.GetCustomBlockResponseStatusCode()))
	}
	if spec.CustomBlockResponseBody != "" {
		policyArgs.CustomBlockResponseBody = pulumi.String(spec.CustomBlockResponseBody)
	}

	// The JS-challenge and CAPTCHA lifetimes exist only on Premium:
	// Azure ALWAYS enables both policies there (rejecting them on
	// Standard), so on Premium the modules pin the documented default of
	// 30 minutes when the spec is silent -- sending nothing would leave
	// the value to drift with Azure's server-side default instead of the
	// declared contract. The spec's Premium-only CELs keep the Standard
	// path from ever reaching here with a value.
	if locals.IsPremium {
		jsChallengeMinutes := 30
		if spec.JsChallengeCookieExpirationInMinutes != nil {
			jsChallengeMinutes = int(spec.GetJsChallengeCookieExpirationInMinutes())
		}
		policyArgs.JsChallengeCookieExpirationInMinutes = pulumi.Int(jsChallengeMinutes)

		captchaMinutes := 30
		if spec.CaptchaCookieExpirationInMinutes != nil {
			captchaMinutes = int(spec.GetCaptchaCookieExpirationInMinutes())
		}
		policyArgs.CaptchaCookieExpirationInMinutes = pulumi.Int(captchaMinutes)
	}

	if len(spec.CustomRules) > 0 {
		policyArgs.CustomRules = buildCustomRules(spec.CustomRules)
	}
	if len(spec.ManagedRules) > 0 {
		policyArgs.ManagedRules = buildManagedRules(spec.ManagedRules)
	}
	if spec.LogScrubbing != nil {
		policyArgs.LogScrubbing = buildLogScrubbing(spec.LogScrubbing)
	}

	createdPolicy, err := cdn.NewFrontdoorFirewallPolicy(ctx,
		spec.PolicyName,
		policyArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door firewall policy %s", spec.PolicyName)
	}

	// Export outputs. firewall_policy_id is what
	// AzureFrontDoorSecurityPolicy's firewall_policy_id references --
	// the policy enforces nothing until a security policy associates it.
	ctx.Export(OpFirewallPolicyId, createdPolicy.ID())
	ctx.Export(OpFirewallPolicyName, createdPolicy.Name)

	return nil
}

// buildCustomRules converts the spec's custom rules into the provider's
// custom_rule blocks.
func buildCustomRules(rules []*azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRule) cdn.FrontdoorFirewallPolicyCustomRuleArray {
	result := make(cdn.FrontdoorFirewallPolicyCustomRuleArray, 0, len(rules))

	for _, rule := range rules {
		ruleArgs := &cdn.FrontdoorFirewallPolicyCustomRuleArgs{
			Name:   pulumi.String(rule.Name),
			Type:   pulumi.String(customRuleTypeStrings[rule.RuleType]),
			Action: pulumi.String(customRuleActionStrings[rule.Action]),
		}

		// Optional dials with provider defaults (enabled true, priority
		// 1, rate-limit window 1 minute / threshold 10): sent only when
		// the spec carries an explicit choice. The rate-limit pair is
		// harmless on MatchRule rules (the provider always sends its
		// defaults; ARM ignores them there).
		if rule.Enabled != nil {
			ruleArgs.Enabled = pulumi.Bool(rule.GetEnabled())
		}
		if rule.Priority != nil {
			ruleArgs.Priority = pulumi.Int(int(rule.GetPriority()))
		}
		if rule.RateLimitDurationInMinutes != nil {
			ruleArgs.RateLimitDurationInMinutes = pulumi.Int(int(rule.GetRateLimitDurationInMinutes()))
		}
		if rule.RateLimitThreshold != nil {
			ruleArgs.RateLimitThreshold = pulumi.Int(int(rule.GetRateLimitThreshold()))
		}

		if len(rule.MatchConditions) > 0 {
			conditions := make(cdn.FrontdoorFirewallPolicyCustomRuleMatchConditionArray, 0, len(rule.MatchConditions))
			for _, condition := range rule.MatchConditions {
				conditionArgs := &cdn.FrontdoorFirewallPolicyCustomRuleMatchConditionArgs{
					MatchVariable:     pulumi.String(matchVariableStrings[condition.MatchVariable]),
					Operator:          pulumi.String(operatorStrings[condition.Operator]),
					MatchValues:       pulumi.ToStringArray(condition.MatchValues),
					NegationCondition: pulumi.Bool(condition.NegateCondition),
				}
				// Selector only for the keyed variables (Cookies,
				// PostArgs, QueryString, RequestHeader) -- ARM rejects it
				// elsewhere, so an empty spec value is simply not sent.
				if condition.Selector != "" {
					conditionArgs.Selector = pulumi.String(condition.Selector)
				}
				if len(condition.Transforms) > 0 {
					transforms := make(pulumi.StringArray, 0, len(condition.Transforms))
					for _, transform := range condition.Transforms {
						transforms = append(transforms, pulumi.String(transformStrings[transform]))
					}
					conditionArgs.Transforms = transforms
				}
				conditions = append(conditions, conditionArgs)
			}
			ruleArgs.MatchConditions = conditions
		}

		result = append(result, ruleArgs)
	}

	return result
}

// buildManagedRules converts the spec's managed rule sets into the
// provider's managed_rule blocks (Premium only, spec-enforced).
func buildManagedRules(sets []*azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleSet) cdn.FrontdoorFirewallPolicyManagedRuleArray {
	result := make(cdn.FrontdoorFirewallPolicyManagedRuleArray, 0, len(sets))

	for _, set := range sets {
		setArgs := &cdn.FrontdoorFirewallPolicyManagedRuleArgs{
			Type:    pulumi.String(set.Type),
			Version: pulumi.String(set.Version),
			Action:  pulumi.String(managedRuleSetActionStrings[set.Action]),
		}

		if len(set.Exclusions) > 0 {
			exclusions := make(cdn.FrontdoorFirewallPolicyManagedRuleExclusionArray, 0, len(set.Exclusions))
			for _, exclusion := range set.Exclusions {
				exclusions = append(exclusions, &cdn.FrontdoorFirewallPolicyManagedRuleExclusionArgs{
					MatchVariable: pulumi.String(exclusionMatchVariableStrings[exclusion.MatchVariable]),
					Operator:      pulumi.String(selectorOperatorStrings[exclusion.Operator]),
					Selector:      pulumi.String(exclusion.Selector),
				})
			}
			setArgs.Exclusions = exclusions
		}

		if len(set.Overrides) > 0 {
			overrides := make(cdn.FrontdoorFirewallPolicyManagedRuleOverrideArray, 0, len(set.Overrides))
			for _, override := range set.Overrides {
				overrideArgs := &cdn.FrontdoorFirewallPolicyManagedRuleOverrideArgs{
					RuleGroupName: pulumi.String(override.RuleGroupName),
				}

				if len(override.Exclusions) > 0 {
					exclusions := make(cdn.FrontdoorFirewallPolicyManagedRuleOverrideExclusionArray, 0, len(override.Exclusions))
					for _, exclusion := range override.Exclusions {
						exclusions = append(exclusions, &cdn.FrontdoorFirewallPolicyManagedRuleOverrideExclusionArgs{
							MatchVariable: pulumi.String(exclusionMatchVariableStrings[exclusion.MatchVariable]),
							Operator:      pulumi.String(selectorOperatorStrings[exclusion.Operator]),
							Selector:      pulumi.String(exclusion.Selector),
						})
					}
					overrideArgs.Exclusions = exclusions
				}

				if len(override.Rules) > 0 {
					ruleOverrides := make(cdn.FrontdoorFirewallPolicyManagedRuleOverrideRuleArray, 0, len(override.Rules))
					for _, ruleOverride := range override.Rules {
						ruleArgs := &cdn.FrontdoorFirewallPolicyManagedRuleOverrideRuleArgs{
							RuleId: pulumi.String(ruleOverride.RuleId),
							Action: pulumi.String(managedRuleOverrideActionStrings[ruleOverride.Action]),
						}
						// The provider's default here is FALSE (listing a
						// rule disables it -- the common tuning gesture);
						// sent only on an explicit choice.
						if ruleOverride.Enabled != nil {
							ruleArgs.Enabled = pulumi.Bool(ruleOverride.GetEnabled())
						}
						if len(ruleOverride.Exclusions) > 0 {
							exclusions := make(cdn.FrontdoorFirewallPolicyManagedRuleOverrideRuleExclusionArray, 0, len(ruleOverride.Exclusions))
							for _, exclusion := range ruleOverride.Exclusions {
								exclusions = append(exclusions, &cdn.FrontdoorFirewallPolicyManagedRuleOverrideRuleExclusionArgs{
									MatchVariable: pulumi.String(exclusionMatchVariableStrings[exclusion.MatchVariable]),
									Operator:      pulumi.String(selectorOperatorStrings[exclusion.Operator]),
									Selector:      pulumi.String(exclusion.Selector),
								})
							}
							ruleArgs.Exclusions = exclusions
						}
						ruleOverrides = append(ruleOverrides, ruleArgs)
					}
					overrideArgs.Rules = ruleOverrides
				}

				overrides = append(overrides, overrideArgs)
			}
			setArgs.Overrides = overrides
		}

		result = append(result, setArgs)
	}

	return result
}

// buildLogScrubbing converts the spec's log-scrubbing block into the
// provider's log_scrubbing block. The operator/selector pairing
// contracts (EqualsAny for IP/URI, selector XOR EqualsAny) are
// spec-enforced.
func buildLogScrubbing(scrubbing *azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyLogScrubbing) *cdn.FrontdoorFirewallPolicyLogScrubbingArgs {
	args := &cdn.FrontdoorFirewallPolicyLogScrubbingArgs{}

	// enabled defaults true on the provider; sent only on an explicit
	// choice.
	if scrubbing.Enabled != nil {
		args.Enabled = pulumi.Bool(scrubbing.GetEnabled())
	}

	rules := make(cdn.FrontdoorFirewallPolicyLogScrubbingScrubbingRuleArray, 0, len(scrubbing.ScrubbingRules))
	for _, rule := range scrubbing.ScrubbingRules {
		ruleArgs := &cdn.FrontdoorFirewallPolicyLogScrubbingScrubbingRuleArgs{
			MatchVariable: pulumi.String(scrubbingMatchVariableStrings[rule.MatchVariable]),
		}
		if rule.Enabled != nil {
			ruleArgs.Enabled = pulumi.Bool(rule.GetEnabled())
		}
		// Unspecified means Equals (the provider's default) -- not sent,
		// letting the provider default apply.
		if rule.Operator != azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySelectorOperator_azure_front_door_firewall_policy_selector_operator_unspecified {
			ruleArgs.Operator = pulumi.String(selectorOperatorStrings[rule.Operator])
		}
		if rule.Selector != "" {
			ruleArgs.Selector = pulumi.String(rule.Selector)
		}
		rules = append(rules, ruleArgs)
	}
	args.ScrubbingRules = rules

	return args
}
