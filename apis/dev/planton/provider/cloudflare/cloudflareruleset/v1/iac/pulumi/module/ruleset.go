package module

import (
	"strings"

	"github.com/pkg/errors"
	cloudflarerulesetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflareruleset/v1"
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
	ctx.Export(OpLastUpdated, created.LastUpdated)

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
		if pr.Ratelimit != nil {
			rule.Ratelimit = buildRatelimit(pr.Ratelimit)
		}
		if pr.Logging != nil {
			rule.Logging = &cloudflare.RulesetRuleLoggingArgs{
				Enabled: pulumi.Bool(pr.Logging.Enabled),
			}
		}
		if pr.ExposedCredentialCheck != nil {
			rule.ExposedCredentialCheck = &cloudflare.RulesetRuleExposedCredentialCheckArgs{
				UsernameExpression: pulumi.String(pr.ExposedCredentialCheck.UsernameExpression),
				PasswordExpression: pulumi.String(pr.ExposedCredentialCheck.PasswordExpression),
			}
		}
		rules = append(rules, rule)
	}
	return rules
}

func buildRatelimit(rl *cloudflarerulesetv1.CloudflareRulesetRatelimit) *cloudflare.RulesetRuleRatelimitArgs {
	args := &cloudflare.RulesetRuleRatelimitArgs{
		Characteristics:  pulumi.ToStringArray(rl.Characteristics),
		Period:           pulumi.Int(int(rl.Period)),
		RequestsToOrigin: pulumi.Bool(rl.RequestsToOrigin),
	}
	if rl.CountingExpression != "" {
		args.CountingExpression = pulumi.String(rl.CountingExpression)
	}
	if rl.MitigationTimeout > 0 {
		args.MitigationTimeout = pulumi.Int(int(rl.MitigationTimeout))
	}
	if rl.RequestsPerPeriod > 0 {
		args.RequestsPerPeriod = pulumi.Int(int(rl.RequestsPerPeriod))
	}
	if rl.ScorePerPeriod > 0 {
		args.ScorePerPeriod = pulumi.Int(int(rl.ScorePerPeriod))
	}
	if rl.ScoreResponseHeaderName != "" {
		args.ScoreResponseHeaderName = pulumi.String(rl.ScoreResponseHeaderName)
	}
	return args
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

	// Block / serve_error inline response
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
			// The provider requires exactly one of value/expression; sending the
			// unused one as an empty string fails validation, so set only the
			// populated field.
			targetUrl := &cloudflare.RulesetRuleActionParametersFromValueTargetUrlArgs{}
			if ap.FromValue.TargetUrl.Value != "" {
				targetUrl.Value = pulumi.String(ap.FromValue.TargetUrl.Value)
			}
			if ap.FromValue.TargetUrl.Expression != "" {
				targetUrl.Expression = pulumi.String(ap.FromValue.TargetUrl.Expression)
			}
			fv.TargetUrl = targetUrl
		}
		args.FromValue = fv
	}
	if ap.FromList != nil {
		fromListName := ""
		if ap.FromList.Name != nil {
			fromListName = ap.FromList.Name.GetValue()
		}
		args.FromList = &cloudflare.RulesetRuleActionParametersFromListArgs{
			Name: pulumi.String(fromListName),
			Key:  pulumi.String(ap.FromList.Key),
		}
	}

	// Skip
	if len(ap.Phases) > 0 {
		args.Phases = pulumi.ToStringArray(ap.Phases)
	}
	if len(ap.Products) > 0 {
		args.Products = pulumi.ToStringArray(ap.Products)
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
	if ap.MatchedData != nil {
		args.MatchedData = &cloudflare.RulesetRuleActionParametersMatchedDataArgs{
			PublicKey: pulumi.String(ap.MatchedData.PublicKey),
		}
	}

	// Compress response
	if len(ap.Algorithms) > 0 {
		algos := cloudflare.RulesetRuleActionParametersAlgorithmArray{}
		for _, a := range ap.Algorithms {
			algos = append(algos, &cloudflare.RulesetRuleActionParametersAlgorithmArgs{
				Name: pulumi.String(a.Name),
			})
		}
		args.Algorithms = algos
	}

	// Score
	if ap.Increment > 0 {
		args.Increment = pulumi.Int(int(ap.Increment))
	}

	// Serve error (inline)
	if ap.AssetName != "" {
		args.AssetName = pulumi.String(ap.AssetName)
	}
	if ap.Content != "" {
		args.Content = pulumi.String(ap.Content)
	}
	if ap.ContentType != "" {
		args.ContentType = pulumi.String(ap.ContentType)
	}
	if ap.StatusCode > 0 {
		args.StatusCode = pulumi.Int(int(ap.StatusCode))
	}

	// Configuration settings (set_config)
	args.AutomaticHttpsRewrites = optBool(ap.AutomaticHttpsRewrites)
	if ap.Autominify != nil {
		args.Autominify = &cloudflare.RulesetRuleActionParametersAutominifyArgs{
			Css:  optBool(ap.Autominify.Css),
			Html: optBool(ap.Autominify.Html),
			Js:   optBool(ap.Autominify.Js),
		}
	}
	args.Bic = optBool(ap.Bic)
	args.ContentConverter = optBool(ap.ContentConverter)
	args.DisableApps = optBool(ap.DisableApps)
	args.DisableRum = optBool(ap.DisableRum)
	args.DisableZaraz = optBool(ap.DisableZaraz)
	args.EmailObfuscation = optBool(ap.EmailObfuscation)
	args.Fonts = optBool(ap.Fonts)
	args.HotlinkProtection = optBool(ap.HotlinkProtection)
	args.Mirage = optBool(ap.Mirage)
	args.OpportunisticEncryption = optBool(ap.OpportunisticEncryption)
	if ap.Polish != "" {
		args.Polish = pulumi.String(ap.Polish)
	}
	args.RedirectsForAiTraining = optBool(ap.RedirectsForAiTraining)
	if ap.RequestBodyBuffering != "" {
		args.RequestBodyBuffering = pulumi.String(ap.RequestBodyBuffering)
	}
	if ap.ResponseBodyBuffering != "" {
		args.ResponseBodyBuffering = pulumi.String(ap.ResponseBodyBuffering)
	}
	args.RocketLoader = optBool(ap.RocketLoader)
	if ap.SecurityLevel != "" {
		args.SecurityLevel = pulumi.String(ap.SecurityLevel)
	}
	args.ServerSideExcludes = optBool(ap.ServerSideExcludes)
	if ap.Ssl != "" {
		args.Ssl = pulumi.String(ap.Ssl)
	}
	args.Sxg = optBool(ap.Sxg)

	// Cache (set_cache_settings)
	if ap.Cache {
		args.Cache = pulumi.Bool(ap.Cache)
	}
	if len(ap.AdditionalCacheablePorts) > 0 {
		ports := make(pulumi.IntArray, 0, len(ap.AdditionalCacheablePorts))
		for _, p := range ap.AdditionalCacheablePorts {
			ports = append(ports, pulumi.Int(int(p)))
		}
		args.AdditionalCacheablePorts = ports
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
	if ap.CacheKey != nil {
		args.CacheKey = buildCacheKey(ap.CacheKey)
	}
	if ap.CacheReserve != nil {
		cr := &cloudflare.RulesetRuleActionParametersCacheReserveArgs{
			Eligible: pulumi.Bool(ap.CacheReserve.Eligible),
		}
		if ap.CacheReserve.MinimumFileSize > 0 {
			cr.MinimumFileSize = pulumi.Int(int(ap.CacheReserve.MinimumFileSize))
		}
		args.CacheReserve = cr
	}
	args.OriginCacheControl = optBool(ap.OriginCacheControl)
	args.OriginErrorPagePassthru = optBool(ap.OriginErrorPagePassthru)
	if ap.ReadTimeout > 0 {
		args.ReadTimeout = pulumi.Int(int(ap.ReadTimeout))
	}
	args.RespectStrongEtags = optBool(ap.RespectStrongEtags)
	args.StripEtags = optBool(ap.StripEtags)
	args.StripLastModified = optBool(ap.StripLastModified)
	args.StripSetCookie = optBool(ap.StripSetCookie)
	// NOTE: `vary` is intentionally not mapped here — the pulumi-cloudflare SDK
	// (v6.17.0) does not expose it on ruleset action parameters, while the Terraform
	// provider (v5.21.1) does. The proto models it as the future-proof contract; the
	// Terraform module provisions it today and Pulumi will once the SDK catches up.
	// See pkg/iac/MODULE_PARITY.md.

	// Log custom fields
	if len(ap.CookieFields) > 0 {
		args.CookieFields = logFieldArray(ap.CookieFields)
	}
	if len(ap.RequestFields) > 0 {
		fields := cloudflare.RulesetRuleActionParametersRequestFieldArray{}
		for _, f := range ap.RequestFields {
			fields = append(fields, &cloudflare.RulesetRuleActionParametersRequestFieldArgs{Name: pulumi.String(f.Name)})
		}
		args.RequestFields = fields
	}
	if len(ap.TransformedRequestFields) > 0 {
		fields := cloudflare.RulesetRuleActionParametersTransformedRequestFieldArray{}
		for _, f := range ap.TransformedRequestFields {
			fields = append(fields, &cloudflare.RulesetRuleActionParametersTransformedRequestFieldArgs{Name: pulumi.String(f.Name)})
		}
		args.TransformedRequestFields = fields
	}
	if len(ap.ResponseFields) > 0 {
		fields := cloudflare.RulesetRuleActionParametersResponseFieldArray{}
		for _, f := range ap.ResponseFields {
			fields = append(fields, &cloudflare.RulesetRuleActionParametersResponseFieldArgs{
				Name:               pulumi.String(f.Name),
				PreserveDuplicates: pulumi.Bool(f.PreserveDuplicates),
			})
		}
		args.ResponseFields = fields
	}
	if len(ap.RawResponseFields) > 0 {
		fields := cloudflare.RulesetRuleActionParametersRawResponseFieldArray{}
		for _, f := range ap.RawResponseFields {
			fields = append(fields, &cloudflare.RulesetRuleActionParametersRawResponseFieldArgs{
				Name:               pulumi.String(f.Name),
				PreserveDuplicates: pulumi.Bool(f.PreserveDuplicates),
			})
		}
		args.RawResponseFields = fields
	}

	// Set Cache-Control directives
	if ap.MaxAge != nil {
		args.MaxAge = &cloudflare.RulesetRuleActionParametersMaxAgeArgs{
			Operation: pulumi.String(ap.MaxAge.Operation), Value: optInt(ap.MaxAge.Value), CloudflareOnly: pulumi.Bool(ap.MaxAge.CloudflareOnly),
		}
	}
	if ap.SMaxage != nil {
		args.SMaxage = &cloudflare.RulesetRuleActionParametersSMaxageArgs{
			Operation: pulumi.String(ap.SMaxage.Operation), Value: optInt(ap.SMaxage.Value), CloudflareOnly: pulumi.Bool(ap.SMaxage.CloudflareOnly),
		}
	}
	if ap.StaleWhileRevalidate != nil {
		args.StaleWhileRevalidate = &cloudflare.RulesetRuleActionParametersStaleWhileRevalidateArgs{
			Operation: pulumi.String(ap.StaleWhileRevalidate.Operation), Value: optInt(ap.StaleWhileRevalidate.Value), CloudflareOnly: pulumi.Bool(ap.StaleWhileRevalidate.CloudflareOnly),
		}
	}
	if ap.StaleIfError != nil {
		args.StaleIfError = &cloudflare.RulesetRuleActionParametersStaleIfErrorArgs{
			Operation: pulumi.String(ap.StaleIfError.Operation), Value: optInt(ap.StaleIfError.Value), CloudflareOnly: pulumi.Bool(ap.StaleIfError.CloudflareOnly),
		}
	}
	if ap.Private != nil {
		args.Private = &cloudflare.RulesetRuleActionParametersPrivateArgs{
			Operation: pulumi.String(ap.Private.Operation), Qualifiers: pulumi.ToStringArray(ap.Private.Qualifiers), CloudflareOnly: pulumi.Bool(ap.Private.CloudflareOnly),
		}
	}
	if ap.NoCache != nil {
		args.NoCache = &cloudflare.RulesetRuleActionParametersNoCacheArgs{
			Operation: pulumi.String(ap.NoCache.Operation), Qualifiers: pulumi.ToStringArray(ap.NoCache.Qualifiers), CloudflareOnly: pulumi.Bool(ap.NoCache.CloudflareOnly),
		}
	}
	if ap.MustRevalidate != nil {
		args.MustRevalidate = &cloudflare.RulesetRuleActionParametersMustRevalidateArgs{Operation: pulumi.String(ap.MustRevalidate.Operation), CloudflareOnly: pulumi.Bool(ap.MustRevalidate.CloudflareOnly)}
	}
	if ap.ProxyRevalidate != nil {
		args.ProxyRevalidate = &cloudflare.RulesetRuleActionParametersProxyRevalidateArgs{Operation: pulumi.String(ap.ProxyRevalidate.Operation), CloudflareOnly: pulumi.Bool(ap.ProxyRevalidate.CloudflareOnly)}
	}
	if ap.MustUnderstand != nil {
		args.MustUnderstand = &cloudflare.RulesetRuleActionParametersMustUnderstandArgs{Operation: pulumi.String(ap.MustUnderstand.Operation), CloudflareOnly: pulumi.Bool(ap.MustUnderstand.CloudflareOnly)}
	}
	if ap.NoTransform != nil {
		args.NoTransform = &cloudflare.RulesetRuleActionParametersNoTransformArgs{Operation: pulumi.String(ap.NoTransform.Operation), CloudflareOnly: pulumi.Bool(ap.NoTransform.CloudflareOnly)}
	}
	if ap.Immutable != nil {
		args.Immutable = &cloudflare.RulesetRuleActionParametersImmutableArgs{Operation: pulumi.String(ap.Immutable.Operation), CloudflareOnly: pulumi.Bool(ap.Immutable.CloudflareOnly)}
	}
	if ap.NoStore != nil {
		args.NoStore = &cloudflare.RulesetRuleActionParametersNoStoreArgs{Operation: pulumi.String(ap.NoStore.Operation), CloudflareOnly: pulumi.Bool(ap.NoStore.CloudflareOnly)}
	}
	if ap.Public != nil {
		args.Public = &cloudflare.RulesetRuleActionParametersPublicArgs{Operation: pulumi.String(ap.Public.Operation), CloudflareOnly: pulumi.Bool(ap.Public.CloudflareOnly)}
	}

	// Set cache tags
	if ap.Operation != "" {
		args.Operation = pulumi.String(ap.Operation)
	}
	if len(ap.Values) > 0 {
		args.Values = pulumi.ToStringArray(ap.Values)
	}
	if ap.Expression != "" {
		args.Expression = pulumi.String(ap.Expression)
	}

	return args
}

func buildCacheKey(ck *cloudflarerulesetv1.CloudflareRulesetCacheKey) *cloudflare.RulesetRuleActionParametersCacheKeyArgs {
	args := &cloudflare.RulesetRuleActionParametersCacheKeyArgs{
		CacheByDeviceType:       optBool(ck.CacheByDeviceType),
		CacheDeceptionArmor:     optBool(ck.CacheDeceptionArmor),
		IgnoreQueryStringsOrder: optBool(ck.IgnoreQueryStringsOrder),
	}
	if ck.CustomKey != nil {
		custom := &cloudflare.RulesetRuleActionParametersCacheKeyCustomKeyArgs{}
		if ck.CustomKey.Cookie != nil {
			custom.Cookie = &cloudflare.RulesetRuleActionParametersCacheKeyCustomKeyCookieArgs{
				CheckPresences: pulumi.ToStringArray(ck.CustomKey.Cookie.CheckPresence),
				Includes:       pulumi.ToStringArray(ck.CustomKey.Cookie.Include),
			}
		}
		if ck.CustomKey.Header != nil {
			header := &cloudflare.RulesetRuleActionParametersCacheKeyCustomKeyHeaderArgs{
				CheckPresences: pulumi.ToStringArray(ck.CustomKey.Header.CheckPresence),
				ExcludeOrigin:  optBool(ck.CustomKey.Header.ExcludeOrigin),
				Includes:       pulumi.ToStringArray(ck.CustomKey.Header.Include),
			}
			if len(ck.CustomKey.Header.Contains) > 0 {
				m := pulumi.StringArrayMap{}
				for k, v := range ck.CustomKey.Header.Contains {
					m[k] = pulumi.ToStringArray(v.Values)
				}
				header.Contains = m
			}
			custom.Header = header
		}
		if ck.CustomKey.Host != nil {
			custom.Host = &cloudflare.RulesetRuleActionParametersCacheKeyCustomKeyHostArgs{
				Resolved: optBool(ck.CustomKey.Host.Resolved),
			}
		}
		if ck.CustomKey.QueryString != nil {
			qs := &cloudflare.RulesetRuleActionParametersCacheKeyCustomKeyQueryStringArgs{}
			if ck.CustomKey.QueryString.Include != nil {
				qs.Include = &cloudflare.RulesetRuleActionParametersCacheKeyCustomKeyQueryStringIncludeArgs{
					Lists: pulumi.ToStringArray(ck.CustomKey.QueryString.Include.List),
					All:   optBool(ck.CustomKey.QueryString.Include.All),
				}
			}
			if ck.CustomKey.QueryString.Exclude != nil {
				qs.Exclude = &cloudflare.RulesetRuleActionParametersCacheKeyCustomKeyQueryStringExcludeArgs{
					Lists: pulumi.ToStringArray(ck.CustomKey.QueryString.Exclude.List),
					All:   optBool(ck.CustomKey.QueryString.Exclude.All),
				}
			}
			custom.QueryString = qs
		}
		if ck.CustomKey.User != nil {
			custom.User = &cloudflare.RulesetRuleActionParametersCacheKeyCustomKeyUserArgs{
				DeviceType: optBool(ck.CustomKey.User.DeviceType),
				Geo:        optBool(ck.CustomKey.User.Geo),
				Lang:       optBool(ck.CustomKey.User.Lang),
			}
		}
		args.CustomKey = custom
	}
	return args
}

func logFieldArray(fields []*cloudflarerulesetv1.CloudflareRulesetLogField) cloudflare.RulesetRuleActionParametersCookieFieldArray {
	out := cloudflare.RulesetRuleActionParametersCookieFieldArray{}
	for _, f := range fields {
		out = append(out, &cloudflare.RulesetRuleActionParametersCookieFieldArgs{Name: pulumi.String(f.Name)})
	}
	return out
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

// optBool returns a pulumi bool input for a proto3 optional bool, or nil when unset
// so the provider's computed default applies.
func optBool(b *bool) pulumi.BoolPtrInput {
	if b == nil {
		return nil
	}
	return pulumi.Bool(*b)
}

// optInt returns a pulumi int input for a value, or nil when it is zero (unset).
func optInt(v int64) pulumi.IntPtrInput {
	if v == 0 {
		return nil
	}
	return pulumi.Int(int(v))
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
	return p.String()
}

func actionString(a cloudflarerulesetv1.CloudflareRulesetRule_Action) string {
	return a.String()
}
