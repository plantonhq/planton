package module

import (
	"github.com/pkg/errors"
	gcpcloudarmorpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudarmorpolicy/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// regionSecurityPolicy builds the REGIONAL security policy (spec.region
// set): the WAF a regional backend service attaches (regional external and
// internal Application Load Balancers) or, as a CLOUD_ARMOR_NETWORK policy,
// the packet filter in front of the region's passthrough Network Load
// Balancers, protocol forwarding, and public-IP VMs. The regional
// collection carries no labels, Adaptive Protection, reCAPTCHA options,
// request-body inspection size, redirect action, or header injection -- the
// spec rejects each when region is set, so none is wired here.
//
// PARITY: Pulumi's nested Args types are per resource, so the rule and
// rate-limit mappings below are the regional twins of the ones in
// security_policy.go; the Terraform module carries the same duplication
// as two resource blocks.
func regionSecurityPolicy(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, projectService pulumi.Resource) error {
	spec := locals.GcpCloudArmorPolicy.Spec

	args := &compute.RegionSecurityPolicyArgs{
		Name:   pulumi.String(locals.PolicyName),
		Region: pulumi.String(spec.Region),
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

	// Advanced options shared with the global arm: JSON body parsing (with
	// custom content types), logging verbosity, and true-client-IP headers.
	if spec.AdvancedOptionsConfig != nil {
		aoc := spec.AdvancedOptionsConfig
		advArgs := &compute.RegionSecurityPolicyAdvancedOptionsConfigArgs{}
		if aoc.JsonParsing != "" {
			advArgs.JsonParsing = pulumi.StringPtr(aoc.JsonParsing)
		}
		if aoc.LogLevel != "" {
			advArgs.LogLevel = pulumi.StringPtr(aoc.LogLevel)
		}
		if len(aoc.UserIpRequestHeaders) > 0 {
			advArgs.UserIpRequestHeaders = toPulumiStringArray(aoc.UserIpRequestHeaders)
		}
		if aoc.JsonCustomConfig != nil {
			advArgs.JsonCustomConfig = &compute.RegionSecurityPolicyAdvancedOptionsConfigJsonCustomConfigArgs{
				ContentTypes: toPulumiStringArray(aoc.JsonCustomConfig.ContentTypes),
			}
		}
		args.AdvancedOptionsConfig = advArgs
	}

	// Network DDoS protection level (CLOUD_ARMOR_NETWORK policies).
	if spec.DdosProtectionConfig != nil {
		args.DdosProtectionConfig = &compute.RegionSecurityPolicyDdosProtectionConfigArgs{
			DdosProtection: pulumi.String(spec.DdosProtectionConfig.DdosProtection),
		}
	}

	// Custom packet fields the network rules match on. offset and size are
	// tri-state in the spec: unset is left nil so Google applies its
	// default, while an explicit 0 offset is a real byte position.
	if len(spec.UserDefinedFields) > 0 {
		var fields compute.RegionSecurityPolicyUserDefinedFieldArray
		for _, f := range spec.UserDefinedFields {
			fieldArgs := &compute.RegionSecurityPolicyUserDefinedFieldArgs{
				Base: pulumi.String(f.Base),
			}
			if f.Name != "" {
				fieldArgs.Name = pulumi.StringPtr(f.Name)
			}
			if f.Offset != nil {
				fieldArgs.Offset = pulumi.IntPtr(int(f.GetOffset()))
			}
			if f.Size != nil {
				fieldArgs.Size = pulumi.IntPtr(int(f.GetSize()))
			}
			if f.Mask != "" {
				fieldArgs.Mask = pulumi.StringPtr(f.Mask)
			}
			fields = append(fields, fieldArgs)
		}
		args.UserDefinedFields = fields
	}

	if len(spec.Rules) > 0 {
		args.Rules = mapRegionRules(spec.Rules)
	}

	createdPolicy, err := compute.NewRegionSecurityPolicy(ctx, "security-policy", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{projectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create regional security policy")
	}

	ctx.Export(OpPolicyId, createdPolicy.ID())
	ctx.Export(OpPolicyName, createdPolicy.Name)
	ctx.Export(OpPolicySelfLink, createdPolicy.SelfLink)
	ctx.Export(OpFingerprint, createdPolicy.Fingerprint)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	// The region's enrollment in advanced network DDoS protection: Google's
	// per-region, per-project network edge security service with this
	// policy attached. Created only when the spec declares the block; shares
	// the policy's deletion_policy so the enrollment and the policy live and
	// die together.
	if spec.NetworkEdgeSecurityService == nil {
		ctx.Export(OpNetworkEdgeSecurityServiceSelfLink, pulumi.String(""))
		return nil
	}

	edgeName := spec.NetworkEdgeSecurityService.Name
	if edgeName == "" {
		edgeName = locals.PolicyName
	}
	edgeArgs := &compute.NetworkEdgeSecurityServiceArgs{
		Name:           pulumi.String(edgeName),
		Region:         pulumi.String(spec.Region),
		SecurityPolicy: createdPolicy.SelfLink,
	}
	if locals.ProjectId != "" {
		edgeArgs.Project = pulumi.StringPtr(locals.ProjectId)
	}
	if spec.NetworkEdgeSecurityService.Description != "" {
		edgeArgs.Description = pulumi.StringPtr(spec.NetworkEdgeSecurityService.Description)
	}
	if spec.DeletionPolicy != "" {
		edgeArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdEdgeService, err := compute.NewNetworkEdgeSecurityService(ctx, "network-edge-security-service", edgeArgs,
		pulumi.Provider(gcpProvider),
		pulumi.Parent(createdPolicy))
	if err != nil {
		return errors.Wrap(err, "failed to create network edge security service")
	}
	ctx.Export(OpNetworkEdgeSecurityServiceSelfLink, createdEdgeService.SelfLink)

	return nil
}

// mapRegionRules converts the spec's rules to the regional resource's rule
// type. Each rule matches through exactly one arm: match (HTTP attributes)
// or network_match (packet headers).
func mapRegionRules(rules []*gcpcloudarmorpolicyv1alpha1.GcpCloudArmorRule) compute.RegionSecurityPolicyRuleTypeArray {
	var result compute.RegionSecurityPolicyRuleTypeArray

	for _, rule := range rules {
		ruleArgs := &compute.RegionSecurityPolicyRuleTypeArgs{
			Action:   pulumi.String(rule.Action),
			Priority: pulumi.Int(int(rule.GetPriority())),
		}

		if rule.Description != "" {
			ruleArgs.Description = pulumi.StringPtr(rule.Description)
		}

		if rule.Preview {
			ruleArgs.Preview = pulumi.BoolPtr(true)
		}

		if rule.Match != nil {
			ruleArgs.Match = mapRegionMatch(rule.Match)
		}

		if rule.NetworkMatch != nil {
			ruleArgs.NetworkMatch = mapNetworkMatch(rule.NetworkMatch)
		}

		if rule.RateLimitOptions != nil {
			ruleArgs.RateLimitOptions = mapRegionRateLimitOptions(rule.RateLimitOptions)
		}

		if rule.PreconfiguredWafConfig != nil {
			ruleArgs.PreconfiguredWafConfig = mapRegionPreconfiguredWafConfig(rule.PreconfiguredWafConfig)
		}

		result = append(result, ruleArgs)
	}

	return result
}

// mapRegionMatch reconstructs the nested HTTP match from the flattened spec
// fields. expr_options is a global-only lever the spec rejects regionally.
func mapRegionMatch(match *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorRuleMatch) *compute.RegionSecurityPolicyRuleMatchArgs {
	args := &compute.RegionSecurityPolicyRuleMatchArgs{}

	if match.VersionedExpr != "" {
		args.VersionedExpr = pulumi.StringPtr(match.VersionedExpr)
		args.Config = &compute.RegionSecurityPolicyRuleMatchConfigArgs{
			SrcIpRanges: toPulumiStringArray(match.SrcIpRanges),
		}
	}

	if match.Expression != "" {
		args.Expr = &compute.RegionSecurityPolicyRuleMatchExprArgs{
			Expression: pulumi.String(match.Expression),
		}
	}

	return args
}

// mapNetworkMatch converts the packet-level match. Every listed field must
// match; a field left empty is omitted so Google treats it as
// unconstrained.
func mapNetworkMatch(nm *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorNetworkMatch) *compute.RegionSecurityPolicyRuleNetworkMatchArgs {
	args := &compute.RegionSecurityPolicyRuleNetworkMatchArgs{}

	if len(nm.SrcIpRanges) > 0 {
		args.SrcIpRanges = toPulumiStringArray(nm.SrcIpRanges)
	}
	if len(nm.DestIpRanges) > 0 {
		args.DestIpRanges = toPulumiStringArray(nm.DestIpRanges)
	}
	if len(nm.IpProtocols) > 0 {
		args.IpProtocols = toPulumiStringArray(nm.IpProtocols)
	}
	if len(nm.SrcPorts) > 0 {
		args.SrcPorts = toPulumiStringArray(nm.SrcPorts)
	}
	if len(nm.DestPorts) > 0 {
		args.DestPorts = toPulumiStringArray(nm.DestPorts)
	}
	if len(nm.SrcRegionCodes) > 0 {
		args.SrcRegionCodes = toPulumiStringArray(nm.SrcRegionCodes)
	}
	if len(nm.SrcAsns) > 0 {
		asns := make(pulumi.IntArray, len(nm.SrcAsns))
		for i, asn := range nm.SrcAsns {
			asns[i] = pulumi.Int(int(asn))
		}
		args.SrcAsns = asns
	}
	if len(nm.UserDefinedFields) > 0 {
		var fields compute.RegionSecurityPolicyRuleNetworkMatchUserDefinedFieldArray
		for _, f := range nm.UserDefinedFields {
			fields = append(fields, &compute.RegionSecurityPolicyRuleNetworkMatchUserDefinedFieldArgs{
				Name:   pulumi.StringPtr(f.Name),
				Values: toPulumiStringArray(f.Values),
			})
		}
		args.UserDefinedFields = fields
	}

	return args
}

// mapRegionRateLimitOptions converts the rate-limit options for the regional
// rule type. The regional collection exceeds to deny(STATUS) only -- the
// spec rejects an exceed redirect when region is set.
func mapRegionRateLimitOptions(opts *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorRateLimitOptions) *compute.RegionSecurityPolicyRuleRateLimitOptionsArgs {
	args := &compute.RegionSecurityPolicyRuleRateLimitOptionsArgs{
		ConformAction: pulumi.StringPtr(opts.ConformAction),
		ExceedAction:  pulumi.StringPtr(opts.ExceedAction),
		RateLimitThreshold: &compute.RegionSecurityPolicyRuleRateLimitOptionsRateLimitThresholdArgs{
			Count:       pulumi.IntPtr(int(opts.RateLimitThreshold.Count)),
			IntervalSec: pulumi.IntPtr(int(opts.RateLimitThreshold.IntervalSec)),
		},
	}

	if opts.EnforceOnKey != "" {
		args.EnforceOnKey = pulumi.StringPtr(opts.EnforceOnKey)
	}

	if opts.EnforceOnKeyName != "" {
		args.EnforceOnKeyName = pulumi.StringPtr(opts.EnforceOnKeyName)
	}

	if len(opts.EnforceOnKeyConfigs) > 0 {
		var keyConfigs compute.RegionSecurityPolicyRuleRateLimitOptionsEnforceOnKeyConfigArray
		for _, keyConfig := range opts.EnforceOnKeyConfigs {
			keyConfigArgs := &compute.RegionSecurityPolicyRuleRateLimitOptionsEnforceOnKeyConfigArgs{
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
		args.BanThreshold = &compute.RegionSecurityPolicyRuleRateLimitOptionsBanThresholdArgs{
			Count:       pulumi.IntPtr(int(opts.BanThreshold.Count)),
			IntervalSec: pulumi.IntPtr(int(opts.BanThreshold.IntervalSec)),
		}
	}

	if opts.BanDurationSec > 0 {
		args.BanDurationSec = pulumi.IntPtr(int(opts.BanDurationSec))
	}

	return args
}

// mapRegionPreconfiguredWafConfig converts the WAF exclusion config for the
// regional rule type. The SDK generates separate types per field kind
// (headers, cookies, URIs, query params) though they share one shape.
func mapRegionPreconfiguredWafConfig(wc *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorPreconfiguredWafConfig) *compute.RegionSecurityPolicyRulePreconfiguredWafConfigArgs {
	var exclusions compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionArray
	for _, exc := range wc.Exclusions {
		excArgs := &compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionArgs{
			TargetRuleSet: pulumi.String(exc.TargetRuleSet),
		}
		if len(exc.TargetRuleIds) > 0 {
			excArgs.TargetRuleIds = toPulumiStringArray(exc.TargetRuleIds)
		}
		if len(exc.RequestHeaders) > 0 {
			var headers compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionRequestHeaderArray
			for _, p := range exc.RequestHeaders {
				a := &compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionRequestHeaderArgs{Operator: pulumi.String(p.Operator)}
				if p.Value != "" {
					a.Value = pulumi.StringPtr(p.Value)
				}
				headers = append(headers, a)
			}
			excArgs.RequestHeaders = headers
		}
		if len(exc.RequestCookies) > 0 {
			var cookies compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionRequestCookyArray
			for _, p := range exc.RequestCookies {
				a := &compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionRequestCookyArgs{Operator: pulumi.String(p.Operator)}
				if p.Value != "" {
					a.Value = pulumi.StringPtr(p.Value)
				}
				cookies = append(cookies, a)
			}
			excArgs.RequestCookies = cookies
		}
		if len(exc.RequestUris) > 0 {
			var uris compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionRequestUriArray
			for _, p := range exc.RequestUris {
				a := &compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionRequestUriArgs{Operator: pulumi.String(p.Operator)}
				if p.Value != "" {
					a.Value = pulumi.StringPtr(p.Value)
				}
				uris = append(uris, a)
			}
			excArgs.RequestUris = uris
		}
		if len(exc.RequestQueryParams) > 0 {
			var params compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionRequestQueryParamArray
			for _, p := range exc.RequestQueryParams {
				a := &compute.RegionSecurityPolicyRulePreconfiguredWafConfigExclusionRequestQueryParamArgs{Operator: pulumi.String(p.Operator)}
				if p.Value != "" {
					a.Value = pulumi.StringPtr(p.Value)
				}
				params = append(params, a)
			}
			excArgs.RequestQueryParams = params
		}
		exclusions = append(exclusions, excArgs)
	}
	return &compute.RegionSecurityPolicyRulePreconfiguredWafConfigArgs{
		Exclusions: exclusions,
	}
}
