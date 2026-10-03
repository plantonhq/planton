package module

import (
	"github.com/pkg/errors"
	azurewebapplicationfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurewebapplicationfirewallpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/waf"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureWebApplicationFirewallPolicy.Spec

	policyArgs := &waf.PolicyArgs{
		Name:              pulumi.String(spec.PolicyName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		ManagedRules:      buildManagedRules(spec.ManagedRules),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Custom rules: IP/geo allowlists, header exceptions, rate limits --
	// evaluated before the managed sets by ascending priority. The
	// rate-limit trio is only sent for RATE_LIMIT_RULE rules (spec
	// validation pairs them; ARM rejects strays).
	if len(spec.CustomRules) > 0 {
		customRules := make(waf.PolicyCustomRuleArray, 0, len(spec.CustomRules))
		for _, rule := range spec.CustomRules {
			ruleArgs := &waf.PolicyCustomRuleArgs{
				Priority:        pulumi.Int(int(rule.Priority)),
				RuleType:        pulumi.String(ruleTypeStrings[rule.RuleType]),
				Action:          pulumi.String(customRuleActionStrings[rule.Action]),
				MatchConditions: buildMatchConditions(rule.MatchConditions),
			}
			if rule.Name != "" {
				ruleArgs.Name = pulumi.String(rule.Name)
			}
			// Presence-guarded true-default: unset falls back to the spec
			// default (enabled) -- IaC inputs built from a manifest do
			// NOT materialize proto defaults.
			if rule.Enabled != nil {
				ruleArgs.Enabled = pulumi.Bool(rule.GetEnabled())
			} else {
				ruleArgs.Enabled = pulumi.Bool(true)
			}
			if rule.RateLimitDuration != azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRateLimitDuration_azure_web_application_firewall_policy_rate_limit_duration_unspecified {
				ruleArgs.RateLimitDuration = pulumi.String(rateLimitDurationStrings[rule.RateLimitDuration])
			}
			if rule.RateLimitThreshold != nil {
				ruleArgs.RateLimitThreshold = pulumi.Int(int(rule.GetRateLimitThreshold()))
			}
			if rule.GroupRateLimitBy != azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyGroupRateLimitBy_azure_web_application_firewall_policy_group_rate_limit_by_unspecified {
				ruleArgs.GroupRateLimitBy = pulumi.String(groupRateLimitByStrings[rule.GroupRateLimitBy])
			}
			customRules = append(customRules, ruleArgs)
		}
		policyArgs.CustomRules = customRules
	}

	// Enforcement mode and body-inspection dials. Omitting the block
	// applies Azure's defaults (enabled, Prevention, body check on,
	// 128 KB limits). Every optional-with-default field is
	// presence-guarded to its documented default.
	if spec.PolicySettings != nil {
		settings := spec.PolicySettings
		settingsArgs := &waf.PolicyPolicySettingsArgs{}

		if settings.Enabled != nil {
			settingsArgs.Enabled = pulumi.Bool(settings.GetEnabled())
		} else {
			settingsArgs.Enabled = pulumi.Bool(true)
		}
		// Unspecified mode means Prevention -- azurerm's own default,
		// materialized so both engines send the same payload.
		if settings.Mode != azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMode_azure_web_application_firewall_policy_mode_unspecified {
			settingsArgs.Mode = pulumi.String(modeStrings[settings.Mode])
		} else {
			settingsArgs.Mode = pulumi.String("Prevention")
		}
		if settings.RequestBodyCheck != nil {
			settingsArgs.RequestBodyCheck = pulumi.Bool(settings.GetRequestBodyCheck())
		} else {
			settingsArgs.RequestBodyCheck = pulumi.Bool(true)
		}
		if settings.RequestBodyEnforcement != nil {
			settingsArgs.RequestBodyEnforcement = pulumi.Bool(settings.GetRequestBodyEnforcement())
		} else {
			settingsArgs.RequestBodyEnforcement = pulumi.Bool(true)
		}
		if settings.RequestBodyInspectLimitInKb != nil {
			settingsArgs.RequestBodyInspectLimitInKb = pulumi.Int(int(settings.GetRequestBodyInspectLimitInKb()))
		} else {
			settingsArgs.RequestBodyInspectLimitInKb = pulumi.Int(128)
		}
		if settings.MaxRequestBodySizeInKb != nil {
			settingsArgs.MaxRequestBodySizeInKb = pulumi.Int(int(settings.GetMaxRequestBodySizeInKb()))
		} else {
			settingsArgs.MaxRequestBodySizeInKb = pulumi.Int(128)
		}
		// file_upload_enforcement is only honored with OWASP 3.2, so it is
		// forwarded solely on explicit presence -- materializing a default
		// would error on older rule sets.
		if settings.FileUploadEnforcement != nil {
			settingsArgs.FileUploadEnforcement = pulumi.Bool(settings.GetFileUploadEnforcement())
		}
		if settings.FileUploadLimitInMb != nil {
			settingsArgs.FileUploadLimitInMb = pulumi.Int(int(settings.GetFileUploadLimitInMb()))
		} else {
			settingsArgs.FileUploadLimitInMb = pulumi.Int(100)
		}
		if settings.JsChallengeCookieExpirationInMinutes != nil {
			settingsArgs.JsChallengeCookieExpirationInMinutes = pulumi.Int(int(settings.GetJsChallengeCookieExpirationInMinutes()))
		} else {
			settingsArgs.JsChallengeCookieExpirationInMinutes = pulumi.Int(30)
		}

		// Log scrubbing: redact sensitive request parts from WAF logs.
		if settings.LogScrubbing != nil {
			scrubbingArgs := &waf.PolicyPolicySettingsLogScrubbingArgs{}
			if settings.LogScrubbing.Enabled != nil {
				scrubbingArgs.Enabled = pulumi.Bool(settings.LogScrubbing.GetEnabled())
			} else {
				scrubbingArgs.Enabled = pulumi.Bool(true)
			}
			scrubbingRules := make(waf.PolicyPolicySettingsLogScrubbingRuleArray, 0, len(settings.LogScrubbing.Rules))
			for _, rule := range settings.LogScrubbing.Rules {
				ruleArgs := &waf.PolicyPolicySettingsLogScrubbingRuleArgs{
					MatchVariable: pulumi.String(scrubbingMatchVariableStrings[rule.MatchVariable]),
				}
				if rule.Enabled != nil {
					ruleArgs.Enabled = pulumi.Bool(rule.GetEnabled())
				} else {
					ruleArgs.Enabled = pulumi.Bool(true)
				}
				// Unspecified operator means Equals (azurerm's default).
				if rule.SelectorMatchOperator != azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicySelectorMatchOperator_azure_web_application_firewall_policy_selector_match_operator_unspecified {
					ruleArgs.SelectorMatchOperator = pulumi.String(selectorMatchOperatorStrings[rule.SelectorMatchOperator])
				} else {
					ruleArgs.SelectorMatchOperator = pulumi.String("Equals")
				}
				if rule.Selector != "" {
					ruleArgs.Selector = pulumi.String(rule.Selector)
				}
				scrubbingRules = append(scrubbingRules, ruleArgs)
			}
			scrubbingArgs.Rules = scrubbingRules
			settingsArgs.LogScrubbing = scrubbingArgs
		}

		policyArgs.PolicySettings = settingsArgs
	}

	createdPolicy, err := waf.NewPolicy(ctx,
		spec.PolicyName,
		policyArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create waf policy %s", spec.PolicyName)
	}

	// Export outputs from the created resource.
	ctx.Export(OpPolicyId, createdPolicy.ID())
	ctx.Export(OpPolicyName, createdPolicy.Name)

	return nil
}

// buildManagedRules translates the managed-rule configuration -- rule sets,
// per-rule overrides, and scoped exclusions.
func buildManagedRules(managedRules *azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyManagedRules) *waf.PolicyManagedRulesArgs {
	args := &waf.PolicyManagedRulesArgs{}

	if len(managedRules.Exclusions) > 0 {
		exclusions := make(waf.PolicyManagedRulesExclusionArray, 0, len(managedRules.Exclusions))
		for _, exclusion := range managedRules.Exclusions {
			exclusionArgs := &waf.PolicyManagedRulesExclusionArgs{
				MatchVariable:         pulumi.String(exclusionMatchVariableStrings[exclusion.MatchVariable]),
				SelectorMatchOperator: pulumi.String(selectorMatchOperatorStrings[exclusion.SelectorMatchOperator]),
				Selector:              pulumi.String(exclusion.Selector),
			}
			if exclusion.ExcludedRuleSet != nil {
				excludedSetArgs := &waf.PolicyManagedRulesExclusionExcludedRuleSetArgs{
					Type: pulumi.String(managedRuleSetTypeString(exclusion.ExcludedRuleSet.Type)),
				}
				// Presence-guarded: unset falls back to azurerm's default
				// (3.2).
				if exclusion.ExcludedRuleSet.Version != nil {
					excludedSetArgs.Version = pulumi.String(exclusion.ExcludedRuleSet.GetVersion())
				} else {
					excludedSetArgs.Version = pulumi.String("3.2")
				}
				if len(exclusion.ExcludedRuleSet.RuleGroups) > 0 {
					ruleGroups := make(waf.PolicyManagedRulesExclusionExcludedRuleSetRuleGroupArray, 0, len(exclusion.ExcludedRuleSet.RuleGroups))
					for _, ruleGroup := range exclusion.ExcludedRuleSet.RuleGroups {
						ruleGroups = append(ruleGroups, &waf.PolicyManagedRulesExclusionExcludedRuleSetRuleGroupArgs{
							RuleGroupName: pulumi.String(ruleGroup.RuleGroupName),
							ExcludedRules: pulumi.ToStringArray(ruleGroup.ExcludedRules),
						})
					}
					excludedSetArgs.RuleGroups = ruleGroups
				}
				exclusionArgs.ExcludedRuleSet = excludedSetArgs
			}
			exclusions = append(exclusions, exclusionArgs)
		}
		args.Exclusions = exclusions
	}

	ruleSets := make(waf.PolicyManagedRulesManagedRuleSetArray, 0, len(managedRules.ManagedRuleSets))
	for _, ruleSet := range managedRules.ManagedRuleSets {
		ruleSetArgs := &waf.PolicyManagedRulesManagedRuleSetArgs{
			Type:    pulumi.String(managedRuleSetTypeString(ruleSet.Type)),
			Version: pulumi.String(ruleSet.Version),
		}
		if len(ruleSet.RuleGroupOverrides) > 0 {
			overrides := make(waf.PolicyManagedRulesManagedRuleSetRuleGroupOverrideArray, 0, len(ruleSet.RuleGroupOverrides))
			for _, override := range ruleSet.RuleGroupOverrides {
				rules := make(waf.PolicyManagedRulesManagedRuleSetRuleGroupOverrideRuleArray, 0, len(override.Rules))
				for _, rule := range override.Rules {
					ruleArgs := &waf.PolicyManagedRulesManagedRuleSetRuleGroupOverrideRuleArgs{
						Id: pulumi.String(rule.Id),
					}
					// Presence-guarded false-default: listing a rule
					// without enabled=true disables it -- azurerm's (and
					// the spec's) documented default.
					if rule.Enabled != nil {
						ruleArgs.Enabled = pulumi.Bool(rule.GetEnabled())
					} else {
						ruleArgs.Enabled = pulumi.Bool(false)
					}
					if rule.Action != azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRuleOverrideAction_azure_web_application_firewall_policy_rule_override_action_unspecified {
						ruleArgs.Action = pulumi.String(ruleOverrideActionStrings[rule.Action])
					}
					rules = append(rules, ruleArgs)
				}
				overrides = append(overrides, &waf.PolicyManagedRulesManagedRuleSetRuleGroupOverrideArgs{
					RuleGroupName: pulumi.String(override.RuleGroupName),
					Rules:         rules,
				})
			}
			ruleSetArgs.RuleGroupOverrides = overrides
		}
		ruleSets = append(ruleSets, ruleSetArgs)
	}
	args.ManagedRuleSets = ruleSets

	return args
}

// buildMatchConditions translates a custom rule's match conditions.
func buildMatchConditions(conditions []*azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchCondition) waf.PolicyCustomRuleMatchConditionArray {
	out := make(waf.PolicyCustomRuleMatchConditionArray, 0, len(conditions))
	for _, condition := range conditions {
		variables := make(waf.PolicyCustomRuleMatchConditionMatchVariableArray, 0, len(condition.MatchVariables))
		for _, variable := range condition.MatchVariables {
			variableArgs := &waf.PolicyCustomRuleMatchConditionMatchVariableArgs{
				VariableName: pulumi.String(matchVariableNameStrings[variable.VariableName]),
			}
			if variable.Selector != "" {
				variableArgs.Selector = pulumi.String(variable.Selector)
			}
			variables = append(variables, variableArgs)
		}

		conditionArgs := &waf.PolicyCustomRuleMatchConditionArgs{
			MatchVariables:    variables,
			Operator:          pulumi.String(matchOperatorStrings[condition.Operator]),
			NegationCondition: pulumi.Bool(condition.NegationCondition),
		}
		if len(condition.MatchValues) > 0 {
			conditionArgs.MatchValues = pulumi.ToStringArray(condition.MatchValues)
		}
		if len(condition.Transforms) > 0 {
			transforms := make(pulumi.StringArray, 0, len(condition.Transforms))
			for _, transform := range condition.Transforms {
				transforms = append(transforms, pulumi.String(transformStrings[transform]))
			}
			conditionArgs.Transforms = transforms
		}
		out = append(out, conditionArgs)
	}
	return out
}

// managedRuleSetTypeString maps a rule-set type enum to ARM's value, with
// unspecified applying OWASP -- azurerm's own default, materialized so both
// engines send the same payload.
func managedRuleSetTypeString(t azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyManagedRuleSetType) string {
	if t == azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyManagedRuleSetType_azure_web_application_firewall_policy_managed_rule_set_type_unspecified {
		return "OWASP"
	}
	return managedRuleSetTypeStrings[t]
}
