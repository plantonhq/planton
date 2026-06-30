package module

import (
	"github.com/pkg/errors"
	gcpcloudarmorpolicyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudarmorpolicy/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func securityPolicy(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudArmorPolicy.Spec

	args := &compute.SecurityPolicyArgs{
		Name:    pulumi.String(locals.PolicyName),
		Project: pulumi.StringPtr(spec.ProjectId.GetValue()),
		Labels:  pulumi.ToStringMap(locals.GcpLabels),
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.Type != "" {
		args.Type = pulumi.StringPtr(spec.Type)
	}

	// Adaptive Protection configuration.
	if spec.AdaptiveProtectionConfig != nil {
		apc := spec.AdaptiveProtectionConfig
		l7Args := &compute.SecurityPolicyAdaptiveProtectionConfigLayer7DdosDefenseConfigArgs{
			Enable: pulumi.BoolPtr(apc.EnableLayer_7DdosDefense),
		}
		if apc.RuleVisibility != "" {
			l7Args.RuleVisibility = pulumi.StringPtr(apc.RuleVisibility)
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
		if aoc.RequestBodyInspectionSize != "" {
			advArgs.RequestBodyInspectionSize = pulumi.StringPtr(aoc.RequestBodyInspectionSize)
		}
		args.AdvancedOptionsConfig = advArgs
	}

	// Map rules.
	if len(spec.Rules) > 0 {
		args.Rules = mapRules(spec.Rules)
	}

	createdPolicy, err := compute.NewSecurityPolicy(ctx, "security-policy", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create security policy")
	}

	ctx.Export(OpPolicyId, createdPolicy.ID())
	ctx.Export(OpPolicyName, createdPolicy.Name)
	ctx.Export(OpPolicySelfLink, createdPolicy.SelfLink)
	ctx.Export(OpFingerprint, createdPolicy.Fingerprint)

	return nil
}

// mapRules converts the spec's GcpCloudArmorRule list to Pulumi SecurityPolicyRuleTypeArray.
func mapRules(rules []*gcpcloudarmorpolicyv1.GcpCloudArmorRule) compute.SecurityPolicyRuleTypeArray {
	var result compute.SecurityPolicyRuleTypeArray

	for _, rule := range rules {
		ruleArgs := &compute.SecurityPolicyRuleTypeArgs{
			Action:   pulumi.String(rule.Action),
			Priority: pulumi.Int(int(rule.Priority)),
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

// mapMatch reconstructs the nested match structure from flattened spec fields.
func mapMatch(match *gcpcloudarmorpolicyv1.GcpCloudArmorRuleMatch) compute.SecurityPolicyRuleMatchArgs {
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

	return args
}

// mapRateLimitOptions converts the spec's rate limit options to Pulumi args.
func mapRateLimitOptions(opts *gcpcloudarmorpolicyv1.GcpCloudArmorRateLimitOptions) *compute.SecurityPolicyRuleRateLimitOptionsArgs {
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
func mapHeaderAction(ha *gcpcloudarmorpolicyv1.GcpCloudArmorHeaderAction) *compute.SecurityPolicyRuleHeaderActionArgs {
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
func mapPreconfiguredWafConfig(wc *gcpcloudarmorpolicyv1.GcpCloudArmorPreconfiguredWafConfig) *compute.SecurityPolicyRulePreconfiguredWafConfigArgs {
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

func mapWafExclusionHeaders(params []*gcpcloudarmorpolicyv1.GcpCloudArmorWafExclusionFieldParams) compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestHeaderArray {
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

func mapWafExclusionCookies(params []*gcpcloudarmorpolicyv1.GcpCloudArmorWafExclusionFieldParams) compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestCookyArray {
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

func mapWafExclusionUris(params []*gcpcloudarmorpolicyv1.GcpCloudArmorWafExclusionFieldParams) compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestUriArray {
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

func mapWafExclusionQueryParams(params []*gcpcloudarmorpolicyv1.GcpCloudArmorWafExclusionFieldParams) compute.SecurityPolicyRulePreconfiguredWafConfigExclusionRequestQueryParamArray {
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
