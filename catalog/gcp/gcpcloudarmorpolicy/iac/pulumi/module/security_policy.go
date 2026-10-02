package module

import (
	"github.com/pkg/errors"
	gcpcloudarmorpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudarmorpolicy/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// securityPolicy builds the GLOBAL security policy (spec.region empty): the
// WAF in front of a global external Application Load Balancer's backend
// services and backend buckets. The regional arm lives in
// region_security_policy.go.
func securityPolicy(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, projectService pulumi.Resource) error {
	spec := locals.GcpCloudArmorPolicy.Spec

	args := &compute.SecurityPolicyArgs{
		Name:   pulumi.String(locals.PolicyName),
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}

	// What destroy does to the WAF shield: DELETE (default), PREVENT
	// (refuse), or ABANDON (drop from state, keep enforcing).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Empty project falls back to the provider's default project — the same
	// ambient contract the Terraform module honors.
	if locals.ProjectId != "" {
		args.Project = pulumi.StringPtr(locals.ProjectId)
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.Type != "" {
		args.Type = pulumi.StringPtr(spec.Type)
	}

	// Adaptive Protection (CAAP): Layer 7 DDoS detection with optional
	// per-granularity detection/auto-deploy threshold overrides.
	if spec.AdaptiveProtectionConfig != nil {
		apc := spec.AdaptiveProtectionConfig
		l7Args := &compute.SecurityPolicyAdaptiveProtectionConfigLayer7DdosDefenseConfigArgs{
			Enable: pulumi.BoolPtr(apc.EnableLayer_7DdosDefense),
		}
		if apc.RuleVisibility != "" {
			l7Args.RuleVisibility = pulumi.StringPtr(apc.RuleVisibility)
		}
		if len(apc.ThresholdConfigs) > 0 {
			l7Args.ThresholdConfigs = mapThresholdConfigs(apc.ThresholdConfigs)
		}
		args.AdaptiveProtectionConfig = &compute.SecurityPolicyAdaptiveProtectionConfigArgs{
			Layer7DdosDefenseConfig: l7Args,
		}
	}

	// Advanced options configuration.
	if spec.AdvancedOptionsConfig != nil {
		aoc := spec.AdvancedOptionsConfig
		advArgs := &compute.SecurityPolicyAdvancedOptionsConfigArgs{}
		if aoc.JsonParsing != "" {
			advArgs.JsonParsing = pulumi.StringPtr(aoc.JsonParsing)
		}
		if aoc.LogLevel != "" {
			advArgs.LogLevel = pulumi.StringPtr(aoc.LogLevel)
		}
		if len(aoc.UserIpRequestHeaders) > 0 {
			advArgs.UserIpRequestHeaders = toPulumiStringArray(aoc.UserIpRequestHeaders)
		}
		// How much of each request body the WAF inspects (8KB default).
		if aoc.RequestBodyInspectionSize != "" {
			advArgs.RequestBodyInspectionSize = pulumi.StringPtr(aoc.RequestBodyInspectionSize)
		}
		if aoc.JsonCustomConfig != nil {
			advArgs.JsonCustomConfig = &compute.SecurityPolicyAdvancedOptionsConfigJsonCustomConfigArgs{
				ContentTypes: toPulumiStringArray(aoc.JsonCustomConfig.ContentTypes),
			}
		}
		args.AdvancedOptionsConfig = advArgs
	}

	// Policy-level reCAPTCHA site key for GOOGLE_RECAPTCHA redirect rules.
	if spec.RecaptchaOptionsConfig != nil {
		args.RecaptchaOptionsConfig = &compute.SecurityPolicyRecaptchaOptionsConfigArgs{
			RedirectSiteKey: pulumi.String(spec.RecaptchaOptionsConfig.RedirectSiteKey),
		}
	}

	// Map rules.
	if len(spec.Rules) > 0 {
		args.Rules = mapRules(spec.Rules)
	}

	createdPolicy, err := compute.NewSecurityPolicy(ctx, "security-policy", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{projectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create security policy")
	}

	ctx.Export(OpPolicyId, createdPolicy.ID())
	ctx.Export(OpPolicyName, createdPolicy.Name)
	ctx.Export(OpPolicySelfLink, createdPolicy.SelfLink)
	ctx.Export(OpFingerprint, createdPolicy.Fingerprint)
	// A global policy has no region and never enrolls an edge service; the
	// outputs exist on both arms so consumers read one contract.
	ctx.Export(OpRegion, pulumi.String(""))
	ctx.Export(OpNetworkEdgeSecurityServiceSelfLink, pulumi.String(""))

	return nil
}

// mapRules converts the spec's GcpCloudArmorRule list to Pulumi SecurityPolicyRuleTypeArray.
func mapRules(rules []*gcpcloudarmorpolicyv1alpha1.GcpCloudArmorRule) compute.SecurityPolicyRuleTypeArray {
	var result compute.SecurityPolicyRuleTypeArray

	for _, rule := range rules {
		ruleArgs := &compute.SecurityPolicyRuleTypeArgs{
			Action:   pulumi.String(rule.Action),
			Priority: pulumi.Int(int(rule.GetPriority())),
			Match:    mapMatch(rule.Match),
		}

		if rule.Description != "" {
			ruleArgs.Description = pulumi.StringPtr(rule.Description)
		}

		if rule.Preview {
			ruleArgs.Preview = pulumi.BoolPtr(true)
		}

		if rule.RateLimitOptions != nil {
			ruleArgs.RateLimitOptions = mapRateLimitOptions(rule.RateLimitOptions)
		}

		if rule.RedirectOptions != nil {
			redirectArgs := &compute.SecurityPolicyRuleRedirectOptionsArgs{
				Type: pulumi.StringPtr(rule.RedirectOptions.Type),
			}
			if rule.RedirectOptions.Target != "" {
				redirectArgs.Target = pulumi.StringPtr(rule.RedirectOptions.Target)
			}
			ruleArgs.RedirectOptions = redirectArgs
		}

		if rule.HeaderAction != nil {
			ruleArgs.HeaderAction = mapHeaderAction(rule.HeaderAction)
		}

		if rule.PreconfiguredWafConfig != nil {
			ruleArgs.PreconfiguredWafConfig = mapPreconfiguredWafConfig(rule.PreconfiguredWafConfig)
		}

		result = append(result, ruleArgs)
	}

	return result
}

// mapThresholdConfigs converts adaptive-protection threshold overrides.
func mapThresholdConfigs(configs []*gcpcloudarmorpolicyv1alpha1.GcpCloudArmorThresholdConfig) compute.SecurityPolicyAdaptiveProtectionConfigLayer7DdosDefenseConfigThresholdConfigArray {
	var result compute.SecurityPolicyAdaptiveProtectionConfigLayer7DdosDefenseConfigThresholdConfigArray
	for _, config := range configs {
		configArgs := &compute.SecurityPolicyAdaptiveProtectionConfigLayer7DdosDefenseConfigThresholdConfigArgs{
			Name: pulumi.String(config.Name),
		}
		if config.AutoDeployConfidenceThreshold != nil {
			configArgs.AutoDeployConfidenceThreshold = pulumi.Float64Ptr(config.GetAutoDeployConfidenceThreshold())
		}
		if config.AutoDeployImpactedBaselineThreshold != nil {
			configArgs.AutoDeployImpactedBaselineThreshold = pulumi.Float64Ptr(config.GetAutoDeployImpactedBaselineThreshold())
		}
		if config.AutoDeployLoadThreshold != nil {
			configArgs.AutoDeployLoadThreshold = pulumi.Float64Ptr(config.GetAutoDeployLoadThreshold())
		}
		if config.AutoDeployExpirationSec != nil {
			configArgs.AutoDeployExpirationSec = pulumi.IntPtr(int(config.GetAutoDeployExpirationSec()))
		}
		if config.DetectionAbsoluteQps != nil {
			configArgs.DetectionAbsoluteQps = pulumi.Float64Ptr(config.GetDetectionAbsoluteQps())
		}
		if config.DetectionLoadThreshold != nil {
			configArgs.DetectionLoadThreshold = pulumi.Float64Ptr(config.GetDetectionLoadThreshold())
		}
		if config.DetectionRelativeToBaselineQps != nil {
			configArgs.DetectionRelativeToBaselineQps = pulumi.Float64Ptr(config.GetDetectionRelativeToBaselineQps())
		}
		if len(config.TrafficGranularityConfigs) > 0 {
			var granularities compute.SecurityPolicyAdaptiveProtectionConfigLayer7DdosDefenseConfigThresholdConfigTrafficGranularityConfigArray
			for _, granularity := range config.TrafficGranularityConfigs {
				granularityArgs := &compute.SecurityPolicyAdaptiveProtectionConfigLayer7DdosDefenseConfigThresholdConfigTrafficGranularityConfigArgs{
					Type: pulumi.String(granularity.Type),
				}
				if granularity.Value != "" {
					granularityArgs.Value = pulumi.StringPtr(granularity.Value)
				}
				if granularity.EnableEachUniqueValue {
					granularityArgs.EnableEachUniqueValue = pulumi.BoolPtr(true)
				}
				granularities = append(granularities, granularityArgs)
			}
			configArgs.TrafficGranularityConfigs = granularities
		}
		result = append(result, configArgs)
	}
	return result
}

// mapMatch reconstructs the nested match structure from flattened spec fields.
// The global arm always carries match (the spec admits network_match only
// on a regional CLOUD_ARMOR_NETWORK policy).
func mapMatch(match *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorRuleMatch) compute.SecurityPolicyRuleMatchArgs {
	args := compute.SecurityPolicyRuleMatchArgs{}

	if match.VersionedExpr != "" {
		args.VersionedExpr = pulumi.StringPtr(match.VersionedExpr)
		args.Config = &compute.SecurityPolicyRuleMatchConfigArgs{
			SrcIpRanges: toPulumiStringArray(match.SrcIpRanges),
		}
	}

	if match.Expression != "" {
		args.Expr = &compute.SecurityPolicyRuleMatchExprArgs{
			Expression: pulumi.String(match.Expression),
		}
	}

	// reCAPTCHA site-key options for expressions evaluating reCAPTCHA tokens.
	if match.ExprOptions != nil {
		recaptchaArgs := &compute.SecurityPolicyRuleMatchExprOptionsRecaptchaOptionsArgs{}
		if len(match.ExprOptions.ActionTokenSiteKeys) > 0 {
			recaptchaArgs.ActionTokenSiteKeys = toPulumiStringArray(match.ExprOptions.ActionTokenSiteKeys)
		}
		if len(match.ExprOptions.SessionTokenSiteKeys) > 0 {
			recaptchaArgs.SessionTokenSiteKeys = toPulumiStringArray(match.ExprOptions.SessionTokenSiteKeys)
		}
		args.ExprOptions = &compute.SecurityPolicyRuleMatchExprOptionsArgs{
			RecaptchaOptions: recaptchaArgs,
		}
	}

	return args
}

// mapRateLimitOptions converts the spec's rate limit options to Pulumi args.
func mapRateLimitOptions(opts *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorRateLimitOptions) *compute.SecurityPolicyRuleRateLimitOptionsArgs {
	args := &compute.SecurityPolicyRuleRateLimitOptionsArgs{
		ConformAction: pulumi.String(opts.ConformAction),
		ExceedAction:  pulumi.String(opts.ExceedAction),
		RateLimitThreshold: &compute.SecurityPolicyRuleRateLimitOptionsRateLimitThresholdArgs{
			Count:       pulumi.Int(int(opts.RateLimitThreshold.Count)),
			IntervalSec: pulumi.Int(int(opts.RateLimitThreshold.IntervalSec)),
		},
	}

	if opts.EnforceOnKey != "" {
		args.EnforceOnKey = pulumi.StringPtr(opts.EnforceOnKey)
	}

	if opts.EnforceOnKeyName != "" {
		args.EnforceOnKeyName = pulumi.StringPtr(opts.EnforceOnKeyName)
	}

	// Composite rate-limit key: component values concatenate to form the
	// key requests are counted against.
	if len(opts.EnforceOnKeyConfigs) > 0 {
		var keyConfigs compute.SecurityPolicyRuleRateLimitOptionsEnforceOnKeyConfigArray
		for _, keyConfig := range opts.EnforceOnKeyConfigs {
			keyConfigArgs := &compute.SecurityPolicyRuleRateLimitOptionsEnforceOnKeyConfigArgs{
				EnforceOnKeyType: pulumi.StringPtr(keyConfig.EnforceOnKeyType),
			}
			if keyConfig.EnforceOnKeyName != "" {
				keyConfigArgs.EnforceOnKeyName = pulumi.StringPtr(keyConfig.EnforceOnKeyName)
			}
			keyConfigs = append(keyConfigs, keyConfigArgs)
		}
		args.EnforceOnKeyConfigs = keyConfigs
	}

	if opts.BanThreshold != nil {
		args.BanThreshold = &compute.SecurityPolicyRuleRateLimitOptionsBanThresholdArgs{
			Count:       pulumi.Int(int(opts.BanThreshold.Count)),
			IntervalSec: pulumi.Int(int(opts.BanThreshold.IntervalSec)),
		}
	}

	if opts.BanDurationSec > 0 {
		args.BanDurationSec = pulumi.IntPtr(int(opts.BanDurationSec))
	}

	// Exceed redirect uses its own SDK type (not the same as rule RedirectOptions).
	if opts.ExceedRedirectOptions != nil {
		excRedirect := &compute.SecurityPolicyRuleRateLimitOptionsExceedRedirectOptionsArgs{
			Type: pulumi.StringPtr(opts.ExceedRedirectOptions.Type),
		}
		if opts.ExceedRedirectOptions.Target != "" {
			excRedirect.Target = pulumi.StringPtr(opts.ExceedRedirectOptions.Target)
		}
		args.ExceedRedirectOptions = excRedirect
	}

	return args
}

// mapHeaderAction converts the spec's header action to Pulumi args.
func mapHeaderAction(ha *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorHeaderAction) *compute.SecurityPolicyRuleHeaderActionArgs {
	var headers compute.SecurityPolicyRuleHeaderActionRequestHeadersToAddArray
	for _, h := range ha.RequestHeadersToAdds {
		headerArgs := &compute.SecurityPolicyRuleHeaderActionRequestHeadersToAddArgs{
			HeaderName: pulumi.StringPtr(h.HeaderName),
		}
		if h.HeaderValue != "" {
			headerArgs.HeaderValue = pulumi.StringPtr(h.HeaderValue)
		}
		headers = append(headers, headerArgs)
	}
	return &compute.SecurityPolicyRuleHeaderActionArgs{
		RequestHeadersToAdds: headers,
	}
}

// mapPreconfiguredWafConfig converts the spec's WAF exclusion config to Pulumi args.
func mapPreconfiguredWafConfig(wc *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorPreconfiguredWafConfig) *compute.SecurityPolicyRulePreconfiguredWafConfigArgs {
	var exclusions compute.SecurityPolicyRulePreconfiguredWafConfigExclusionArray
	for _, exc := range wc.Exclusions {
		excArgs := &compute.SecurityPolicyRulePreconfiguredWafConfigExclusionArgs{
			TargetRuleSet: pulumi.String(exc.TargetRuleSet),
		}
		if len(exc.TargetRuleIds) > 0 {
			excArgs.TargetRuleIds = toPulumiStringArray(exc.TargetRuleIds)
		}
		if len(exc.RequestHeaders) > 0 {
			excArgs.RequestHeaders = mapWafExclusionHeaders(exc.RequestHeaders)
		}
		if len(exc.RequestCookies) > 0 {
			excArgs.RequestCookies = mapWafExclusionCookies(exc.RequestCookies)
		}
		if len(exc.RequestUris) > 0 {
			excArgs.RequestUris = mapWafExclusionUris(exc.RequestUris)
		}
		if len(exc.RequestQueryParams) > 0 {
			excArgs.RequestQueryParams = mapWafExclusionQueryParams(exc.RequestQueryParams)
		}
		exclusions = append(exclusions, excArgs)
	}
	return &compute.SecurityPolicyRulePreconfiguredWafConfigArgs{
		Exclusions: exclusions,
	}
}

// The Pulumi SDK generates separate types for each WAF exclusion field
// (headers, cookies, URIs, query params) even though they all have the
// same structure (operator + value). We need per-type builder functions.

func mapWafExclusionHeaders(params []*gcpcloudarmorpolicyv1alpha1.GcpCloudArmorWafExclusionFieldParams) compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestHeaderArray {
	var result compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestHeaderArray
	for _, p := range params {
		args := &compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestHeaderArgs{
			Operator: pulumi.String(p.Operator),
		}
		if p.Value != "" {
			args.Value = pulumi.StringPtr(p.Value)
		}
		result = append(result, args)
	}
	return result
}

func mapWafExclusionCookies(params []*gcpcloudarmorpolicyv1alpha1.GcpCloudArmorWafExclusionFieldParams) compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestCookyArray {
	var result compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestCookyArray
	for _, p := range params {
		args := &compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestCookyArgs{
			Operator: pulumi.String(p.Operator),
		}
		if p.Value != "" {
			args.Value = pulumi.StringPtr(p.Value)
		}
		result = append(result, args)
	}
	return result
}

func mapWafExclusionUris(params []*gcpcloudarmorpolicyv1alpha1.GcpCloudArmorWafExclusionFieldParams) compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestUriArray {
	var result compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestUriArray
	for _, p := range params {
		args := &compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestUriArgs{
			Operator: pulumi.String(p.Operator),
		}
		if p.Value != "" {
			args.Value = pulumi.StringPtr(p.Value)
		}
		result = append(result, args)
	}
	return result
}

func mapWafExclusionQueryParams(params []*gcpcloudarmorpolicyv1alpha1.GcpCloudArmorWafExclusionFieldParams) compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestQueryParamArray {
	var result compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestQueryParamArray
	for _, p := range params {
		args := &compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestQueryParamArgs{
			Operator: pulumi.String(p.Operator),
		}
		if p.Value != "" {
			args.Value = pulumi.StringPtr(p.Value)
		}
		result = append(result, args)
	}
	return result
}

// toPulumiStringArray converts a Go string slice to a Pulumi StringArray.
func toPulumiStringArray(values []string) pulumi.StringArray {
	arr := make(pulumi.StringArray, len(values))
	for i, v := range values {
		arr[i] = pulumi.String(v)
	}
	return arr
}
