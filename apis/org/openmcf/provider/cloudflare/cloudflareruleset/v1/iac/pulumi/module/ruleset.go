package module

import (
	"strings"

	"github.com/pkg/errors"
	cloudflarerulesetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareruleset/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ruleset(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.Ruleset, error) {
	spec := locals.CloudflareRuleset.Spec

	rulesetArgs := &cloudflare.RulesetArgs{
		Kind:        pulumi.String(rulesetKindString(spec.GetRulesetKind())),
		Name:        pulumi.String(spec.Name),
		Phase:       pulumi.String(phaseString(spec.Phase)),
		Description: pulumi.String(spec.Description),
		Rules:       buildRules(spec.Rules),
	}

	if spec.ZoneId != nil {
		rulesetArgs.ZoneId = pulumi.String(spec.ZoneId.GetValue())
	}
	if spec.AccountId != "" {
		rulesetArgs.AccountId = pulumi.String(spec.AccountId)
	}

	created, err := cloudflare.NewRuleset(
		ctx,
		strings.ToLower(locals.CloudflareRuleset.Metadata.Name),
		rulesetArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare ruleset")
	}

	ctx.Export(OpRulesetId, created.ID())
	ctx.Export(OpVersion, created.Version)

	zoneId := ""
	if spec.ZoneId != nil {
		zoneId = spec.ZoneId.GetValue()
	}
	ctx.Export(OpZoneId, pulumi.String(zoneId))
	ctx.Export(OpPhase, pulumi.String(phaseString(spec.Phase)))

	return created, nil
}

func buildRules(protoRules []*cloudflarerulesetv1.CloudflareRulesetRule) cloudflare.RulesetRuleArray {
	rules := make(cloudflare.RulesetRuleArray, 0, len(protoRules))
	for _, pr := range protoRules {
		rule := &cloudflare.RulesetRuleArgs{
			Expression:  pulumi.String(pr.Expression),
			Action:      pulumi.String(actionString(pr.Action)),
			Description: pulumi.String(pr.Description),
			Enabled:     pulumi.Bool(pr.GetEnabled()),
		}
		if pr.Ref != "" {
			rule.Ref = pulumi.String(pr.Ref)
		}
		if pr.ActionParameters != nil {
			rule.ActionParameters = buildActionParameters(pr.ActionParameters)
		}
		rules = append(rules, rule)
	}
	return rules
}

func buildActionParameters(ap *cloudflarerulesetv1.CloudflareRulesetActionParameters) *cloudflare.RulesetRuleActionParametersArgs {
	args := &cloudflare.RulesetRuleActionParametersArgs{}

	// Origin Rules (route)
	if ap.HostHeader != "" {
		args.HostHeader = pulumi.String(ap.HostHeader)
	}
	if ap.Origin != nil {
		args.Origin = &cloudflare.RulesetRuleActionParametersOriginArgs{
			Host: pulumi.String(ap.Origin.Host),
			Port: pulumi.Int(int(ap.Origin.Port)),
		}
	}
	if ap.Sni != nil {
		args.Sni = &cloudflare.RulesetRuleActionParametersSniArgs{
			Value: pulumi.String(ap.Sni.Value),
		}
	}

	// Block
	if ap.Response != nil {
		args.Response = &cloudflare.RulesetRuleActionParametersResponseArgs{
			StatusCode:  pulumi.Int(int(ap.Response.StatusCode)),
			Content:     pulumi.String(ap.Response.Content),
			ContentType: pulumi.String(ap.Response.ContentType),
		}
	}

	// Rewrite
	if ap.Uri != nil {
		uriArgs := &cloudflare.RulesetRuleActionParametersUriArgs{}
		if ap.Uri.Path != nil {
			uriArgs.Path = &cloudflare.RulesetRuleActionParametersUriPathArgs{
				Value:      pulumi.String(ap.Uri.Path.Value),
				Expression: pulumi.String(ap.Uri.Path.Expression),
			}
		}
		if ap.Uri.Query != nil {
			uriArgs.Query = &cloudflare.RulesetRuleActionParametersUriQueryArgs{
				Value:      pulumi.String(ap.Uri.Query.Value),
				Expression: pulumi.String(ap.Uri.Query.Expression),
			}
		}
		args.Uri = uriArgs
	}
	if len(ap.Headers) > 0 {
		headerMap := cloudflare.RulesetRuleActionParametersHeadersMap{}
		for name, header := range ap.Headers {
			headerMap[name] = &cloudflare.RulesetRuleActionParametersHeadersArgs{
				Operation:  pulumi.String(header.Operation),
				Value:      pulumi.String(header.Value),
				Expression: pulumi.String(header.Expression),
			}
		}
		args.Headers = headerMap
	}

	// Redirect
	if ap.FromValue != nil {
		fv := &cloudflare.RulesetRuleActionParametersFromValueArgs{
			StatusCode:          pulumi.Int(int(ap.FromValue.StatusCode)),
			PreserveQueryString: pulumi.Bool(ap.FromValue.PreserveQueryString),
		}
		if ap.FromValue.TargetUrl != nil {
			fv.TargetUrl = &cloudflare.RulesetRuleActionParametersFromValueTargetUrlArgs{
				Value:      pulumi.String(ap.FromValue.TargetUrl.Value),
				Expression: pulumi.String(ap.FromValue.TargetUrl.Expression),
			}
		}
		args.FromValue = fv
	}

	// Skip
	if len(ap.Phases) > 0 {
		phases := pulumi.ToStringArray(ap.Phases)
		args.Phases = phases
	}
	if len(ap.Products) > 0 {
		products := pulumi.ToStringArray(ap.Products)
		args.Products = products
	}
	if ap.Ruleset != "" {
		args.Ruleset = pulumi.String(ap.Ruleset)
	}
	if len(ap.Rulesets) > 0 {
		args.Rulesets = pulumi.ToStringArray(ap.Rulesets)
	}

	// Execute
	if ap.Id != "" {
		args.Id = pulumi.String(ap.Id)
	}
	if ap.Overrides != nil {
		args.Overrides = buildOverrides(ap.Overrides)
	}

	// Cache
	if ap.Cache {
		args.Cache = pulumi.Bool(ap.Cache)
	}
	if ap.EdgeTtl != nil {
		edgeTtl := &cloudflare.RulesetRuleActionParametersEdgeTtlArgs{
			Mode:    pulumi.String(ap.EdgeTtl.Mode),
			Default: pulumi.Int(int(ap.EdgeTtl.DefaultTtl)),
		}
		if len(ap.EdgeTtl.StatusCodeTtls) > 0 {
			scTtls := cloudflare.RulesetRuleActionParametersEdgeTtlStatusCodeTtlArray{}
			for _, sct := range ap.EdgeTtl.StatusCodeTtls {
				entry := &cloudflare.RulesetRuleActionParametersEdgeTtlStatusCodeTtlArgs{
					Value: pulumi.Int(int(sct.Value)),
				}
				if sct.StatusCode > 0 {
					entry.StatusCode = pulumi.Int(int(sct.StatusCode))
				}
				if sct.StatusCodeRange != nil {
					entry.StatusCodeRange = &cloudflare.RulesetRuleActionParametersEdgeTtlStatusCodeTtlStatusCodeRangeArgs{
						From: pulumi.Int(int(sct.StatusCodeRange.From)),
						To:   pulumi.Int(int(sct.StatusCodeRange.To)),
					}
				}
				scTtls = append(scTtls, entry)
			}
			edgeTtl.StatusCodeTtls = scTtls
		}
		args.EdgeTtl = edgeTtl
	}
	if ap.BrowserTtl != nil {
		args.BrowserTtl = &cloudflare.RulesetRuleActionParametersBrowserTtlArgs{
			Mode:    pulumi.String(ap.BrowserTtl.Mode),
			Default: pulumi.Int(int(ap.BrowserTtl.DefaultTtl)),
		}
	}
	if ap.ServeStale != nil {
		args.ServeStale = &cloudflare.RulesetRuleActionParametersServeStaleArgs{
			DisableStaleWhileUpdating: pulumi.Bool(ap.ServeStale.DisableStaleWhileUpdating),
		}
	}

	return args
}

func buildOverrides(o *cloudflarerulesetv1.CloudflareRulesetOverrides) *cloudflare.RulesetRuleActionParametersOverridesArgs {
	ov := &cloudflare.RulesetRuleActionParametersOverridesArgs{}
	if o.Action != "" {
		ov.Action = pulumi.String(o.Action)
	}
	if o.Enabled {
		ov.Enabled = pulumi.Bool(o.Enabled)
	}
	if o.SensitivityLevel != "" {
		ov.SensitivityLevel = pulumi.String(o.SensitivityLevel)
	}
	if len(o.Categories) > 0 {
		cats := cloudflare.RulesetRuleActionParametersOverridesCategoryArray{}
		for _, c := range o.Categories {
			cats = append(cats, &cloudflare.RulesetRuleActionParametersOverridesCategoryArgs{
				Category:         pulumi.String(c.Category),
				Action:           pulumi.String(c.Action),
				Enabled:          pulumi.Bool(c.Enabled),
				SensitivityLevel: pulumi.String(c.SensitivityLevel),
			})
		}
		ov.Categories = cats
	}
	if len(o.Rules) > 0 {
		rules := cloudflare.RulesetRuleActionParametersOverridesRuleArray{}
		for _, r := range o.Rules {
			rules = append(rules, &cloudflare.RulesetRuleActionParametersOverridesRuleArgs{
				Id:               pulumi.String(r.Id),
				Action:           pulumi.String(r.Action),
				Enabled:          pulumi.Bool(r.Enabled),
				ScoreThreshold:   pulumi.Int(int(r.ScoreThreshold)),
				SensitivityLevel: pulumi.String(r.SensitivityLevel),
			})
		}
		ov.Rules = rules
	}
	return ov
}

func rulesetKindString(k cloudflarerulesetv1.CloudflareRulesetSpec_RulesetKind) string {
	switch k {
	case cloudflarerulesetv1.CloudflareRulesetSpec_zone:
		return "zone"
	case cloudflarerulesetv1.CloudflareRulesetSpec_custom:
		return "custom"
	case cloudflarerulesetv1.CloudflareRulesetSpec_managed:
		return "managed"
	case cloudflarerulesetv1.CloudflareRulesetSpec_root:
		return "root"
	default:
		return "zone"
	}
}

func phaseString(p cloudflarerulesetv1.CloudflareRulesetSpec_Phase) string {
	return strings.TrimPrefix(p.String(), "")
}

func actionString(a cloudflarerulesetv1.CloudflareRulesetRule_Action) string {
	return a.String()
}
