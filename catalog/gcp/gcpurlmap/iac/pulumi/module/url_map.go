package module

import (
	"strconv"

	"github.com/pkg/errors"
	gcpurlmapv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpurlmap/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// urlMap provisions the global Compute Engine URL map — the L7 routing brain
// of a global external Application Load Balancer. Host rules map request Host
// headers to named path matchers; path matchers evaluate route_rules (priority-
// ordered, rich matching) then path_rules (longest prefix), then their own
// default; anything unmatched falls through to the URL map's top-level default.
//
// name and project are immutable (ForceNew): changing either destroys and
// recreates the map, briefly breaking every target proxy referencing the old
// self_link. Routing tables, header actions, and tests update in place.
//
// Cross-field exclusivity (exactly one default target, path_rules XOR
// route_rules, redirect vs route_action, path_template_rewrite only in route
// rules) is enforced by the spec's CEL rules before deploy — no defensive
// logic lives here. route_action carries the full traffic-management surface
// at every site: weighted splits, rewrites, timeout/retry/mirror/CORS/
// fault-injection/stream-duration policies, and the route-scoped CDN
// cache_policy.
func urlMap(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpUrlMap.Spec

	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"urlmap-compute.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	args := &compute.URLMapArgs{
		Name: pulumi.String(locals.UrlMapName),
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	// Client-side destroy stance (DELETE/PREVENT/ABANDON) — provider-level,
	// never sent to the GCP API. Empty falls back to the provider default
	// (DELETE).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	if spec.DefaultService.GetValue() != "" {
		args.DefaultService = pulumi.String(spec.DefaultService.GetValue())
	}
	if spec.DefaultUrlRedirect != nil {
		args.DefaultUrlRedirect = topLevelUrlRedirect(spec.DefaultUrlRedirect)
	}
	if spec.DefaultRouteAction != nil {
		args.DefaultRouteAction = topLevelRouteAction(spec.DefaultRouteAction)
	}
	if spec.DefaultCustomErrorResponsePolicy != nil {
		args.DefaultCustomErrorResponsePolicy = topLevelCustomErrorPolicy(spec.DefaultCustomErrorResponsePolicy)
	}
	if spec.HeaderAction != nil {
		args.HeaderAction = topLevelHeaderAction(spec.HeaderAction)
	}
	if len(spec.HostRules) > 0 {
		args.HostRules = buildHostRules(spec.HostRules)
	}
	if len(spec.PathMatchers) > 0 {
		args.PathMatchers = buildPathMatchers(spec.PathMatchers)
	}
	if len(spec.Tests) > 0 {
		args.Tests = buildTests(spec.Tests)
	}

	createdUrlMap, err := compute.NewURLMap(ctx, "url-map", args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create url map")
	}

	ctx.Export(OpSelfLink, createdUrlMap.SelfLink)
	ctx.Export(OpUrlMapName, createdUrlMap.Name)
	ctx.Export(OpMapId, createdUrlMap.MapId.ApplyT(func(id int) string {
		return strconv.Itoa(id)
	}).(pulumi.StringOutput))
	ctx.Export(OpFingerprint, createdUrlMap.Fingerprint)

	return nil
}

func topLevelUrlRedirect(r *gcpurlmapv1alpha1.GcpUrlMapUrlRedirect) *compute.URLMapDefaultUrlRedirectArgs {
	return &compute.URLMapDefaultUrlRedirectArgs{
		HostRedirect:         emptyAsNilString(r.HostRedirect),
		HttpsRedirect:        pulumi.Bool(r.HttpsRedirect),
		PathRedirect:         emptyAsNilString(r.PathRedirect),
		PrefixRedirect:       emptyAsNilString(r.PrefixRedirect),
		RedirectResponseCode: emptyAsNilString(r.RedirectResponseCode),
		StripQuery:           pulumi.Bool(r.StripQuery),
	}
}

func topLevelRouteAction(a *gcpurlmapv1alpha1.GcpUrlMapRouteAction) *compute.URLMapDefaultRouteActionArgs {
	if a == nil {
		return nil
	}
	out := &compute.URLMapDefaultRouteActionArgs{}
	if len(a.WeightedBackendServices) > 0 {
		out.WeightedBackendServices = defaultWeightedBackends(a.WeightedBackendServices)
	}
	if a.UrlRewrite != nil {
		out.UrlRewrite = defaultUrlRewrite(a.UrlRewrite)
	}
	if a.Timeout != nil {
		out.Timeout = &compute.URLMapDefaultRouteActionTimeoutArgs{
			Seconds: durationSeconds(a.Timeout),
			Nanos:   durationNanos(a.Timeout),
		}
	}
	if a.RetryPolicy != nil {
		retry := &compute.URLMapDefaultRouteActionRetryPolicyArgs{
			RetryConditions: stringArrayOrNil(a.RetryPolicy.RetryConditions),
		}
		if a.RetryPolicy.NumRetries != 0 {
			retry.NumRetries = pulumi.Int(int(a.RetryPolicy.NumRetries))
		}
		if a.RetryPolicy.PerTryTimeout != nil {
			retry.PerTryTimeout = &compute.URLMapDefaultRouteActionRetryPolicyPerTryTimeoutArgs{
				Seconds: durationSeconds(a.RetryPolicy.PerTryTimeout),
				Nanos:   durationNanos(a.RetryPolicy.PerTryTimeout),
			}
		}
		out.RetryPolicy = retry
	}
	if a.RequestMirrorPolicy != nil {
		out.RequestMirrorPolicy = &compute.URLMapDefaultRouteActionRequestMirrorPolicyArgs{
			BackendService: pulumi.String(a.RequestMirrorPolicy.BackendService.GetValue()),
		}
	}
	if a.CorsPolicy != nil {
		cors := &compute.URLMapDefaultRouteActionCorsPolicyArgs{
			AllowCredentials:   pulumi.Bool(a.CorsPolicy.AllowCredentials),
			AllowHeaders:       stringArrayOrNil(a.CorsPolicy.AllowHeaders),
			AllowMethods:       stringArrayOrNil(a.CorsPolicy.AllowMethods),
			AllowOriginRegexes: stringArrayOrNil(a.CorsPolicy.AllowOriginRegexes),
			AllowOrigins:       stringArrayOrNil(a.CorsPolicy.AllowOrigins),
			Disabled:           pulumi.Bool(a.CorsPolicy.Disabled),
			ExposeHeaders:      stringArrayOrNil(a.CorsPolicy.ExposeHeaders),
		}
		if a.CorsPolicy.MaxAge != 0 {
			cors.MaxAge = pulumi.Int(int(a.CorsPolicy.MaxAge))
		}
		out.CorsPolicy = cors
	}
	if a.FaultInjectionPolicy != nil {
		fault := &compute.URLMapDefaultRouteActionFaultInjectionPolicyArgs{}
		if a.FaultInjectionPolicy.Abort != nil {
			abort := &compute.URLMapDefaultRouteActionFaultInjectionPolicyAbortArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Abort.Percentage),
			}
			if a.FaultInjectionPolicy.Abort.HttpStatus != 0 {
				abort.HttpStatus = pulumi.Int(int(a.FaultInjectionPolicy.Abort.HttpStatus))
			}
			fault.Abort = abort
		}
		if a.FaultInjectionPolicy.Delay != nil {
			delay := &compute.URLMapDefaultRouteActionFaultInjectionPolicyDelayArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Delay.Percentage),
			}
			if a.FaultInjectionPolicy.Delay.FixedDelay != nil {
				delay.FixedDelay = &compute.URLMapDefaultRouteActionFaultInjectionPolicyDelayFixedDelayArgs{
					Seconds: durationSeconds(a.FaultInjectionPolicy.Delay.FixedDelay),
					Nanos:   durationNanos(a.FaultInjectionPolicy.Delay.FixedDelay),
				}
			}
			fault.Delay = delay
		}
		out.FaultInjectionPolicy = fault
	}
	if a.MaxStreamDuration != nil {
		out.MaxStreamDuration = &compute.URLMapDefaultRouteActionMaxStreamDurationArgs{
			Seconds: durationSeconds(a.MaxStreamDuration),
			Nanos:   durationNanos(a.MaxStreamDuration),
		}
	}
	if a.CachePolicy != nil {
		cache := &compute.URLMapDefaultRouteActionCachePolicyArgs{
			CacheMode:                     emptyAsNilString(a.CachePolicy.CacheMode),
			CacheBypassRequestHeaderNames: stringArrayOrNil(a.CachePolicy.CacheBypassRequestHeaderNames),
			NegativeCaching:               pulumi.Bool(a.CachePolicy.NegativeCaching),
			RequestCoalescing:             pulumi.Bool(a.CachePolicy.RequestCoalescing),
		}
		if a.CachePolicy.CacheKeyPolicy != nil {
			cache.CacheKeyPolicy = &compute.URLMapDefaultRouteActionCachePolicyCacheKeyPolicyArgs{
				ExcludedQueryParameters: stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.ExcludedQueryParameters),
				IncludeHost:             pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeHost),
				IncludeProtocol:         pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeProtocol),
				IncludeQueryString:      pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeQueryString),
				IncludedCookieNames:     stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedCookieNames),
				IncludedHeaderNames:     stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedHeaderNames),
				IncludedQueryParameters: stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedQueryParameters),
			}
		}
		if a.CachePolicy.ClientTtl != nil {
			cache.ClientTtl = &compute.URLMapDefaultRouteActionCachePolicyClientTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.ClientTtl),
				Nanos:   durationNanos(a.CachePolicy.ClientTtl),
			}
		}
		if a.CachePolicy.DefaultTtl != nil {
			cache.DefaultTtl = &compute.URLMapDefaultRouteActionCachePolicyDefaultTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.DefaultTtl),
				Nanos:   durationNanos(a.CachePolicy.DefaultTtl),
			}
		}
		if a.CachePolicy.MaxTtl != nil {
			cache.MaxTtl = &compute.URLMapDefaultRouteActionCachePolicyMaxTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.MaxTtl),
				Nanos:   durationNanos(a.CachePolicy.MaxTtl),
			}
		}
		if a.CachePolicy.ServeWhileStale != nil {
			cache.ServeWhileStale = &compute.URLMapDefaultRouteActionCachePolicyServeWhileStaleArgs{
				Seconds: durationSeconds(a.CachePolicy.ServeWhileStale),
				Nanos:   durationNanos(a.CachePolicy.ServeWhileStale),
			}
		}
		if len(a.CachePolicy.NegativeCachingPolicy) > 0 {
			policies := compute.URLMapDefaultRouteActionCachePolicyNegativeCachingPolicyArray{}
			for _, p := range a.CachePolicy.NegativeCachingPolicy {
				policy := &compute.URLMapDefaultRouteActionCachePolicyNegativeCachingPolicyArgs{}
				if p.Code != 0 {
					policy.Code = pulumi.Int(int(p.Code))
				}
				if p.Ttl != nil {
					policy.Ttl = &compute.URLMapDefaultRouteActionCachePolicyNegativeCachingPolicyTtlArgs{
						Seconds: durationSeconds(p.Ttl),
						Nanos:   durationNanos(p.Ttl),
					}
				}
				policies = append(policies, policy)
			}
			cache.NegativeCachingPolicies = policies
		}
		out.CachePolicy = cache
	}
	return out
}

func defaultWeightedBackends(backends []*gcpurlmapv1alpha1.GcpUrlMapWeightedBackendService) compute.URLMapDefaultRouteActionWeightedBackendServiceArray {
	result := compute.URLMapDefaultRouteActionWeightedBackendServiceArray{}
	for _, backend := range backends {
		args := &compute.URLMapDefaultRouteActionWeightedBackendServiceArgs{
			BackendService: pulumi.String(backend.BackendService.GetValue()),
			Weight:         pulumi.Int(int(backend.Weight)),
		}
		if backend.HeaderAction != nil {
			headerAction := &compute.URLMapDefaultRouteActionWeightedBackendServiceHeaderActionArgs{
				RequestHeadersToRemoves:  stringArrayOrNil(backend.HeaderAction.RequestHeadersToRemove),
				ResponseHeadersToRemoves: stringArrayOrNil(backend.HeaderAction.ResponseHeadersToRemove),
			}
			if len(backend.HeaderAction.RequestHeadersToAdd) > 0 {
				adds := compute.URLMapDefaultRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArray{}
				for _, h := range backend.HeaderAction.RequestHeadersToAdd {
					adds = append(adds, &compute.URLMapDefaultRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.RequestHeadersToAdds = adds
			}
			if len(backend.HeaderAction.ResponseHeadersToAdd) > 0 {
				adds := compute.URLMapDefaultRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArray{}
				for _, h := range backend.HeaderAction.ResponseHeadersToAdd {
					adds = append(adds, &compute.URLMapDefaultRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.ResponseHeadersToAdds = adds
			}
			args.HeaderAction = headerAction
		}
		result = append(result, args)
	}
	return result
}

func defaultUrlRewrite(r *gcpurlmapv1alpha1.GcpUrlMapUrlRewrite) *compute.URLMapDefaultRouteActionUrlRewriteArgs {
	return &compute.URLMapDefaultRouteActionUrlRewriteArgs{
		HostRewrite:       emptyAsNilString(r.HostRewrite),
		PathPrefixRewrite: emptyAsNilString(r.PathPrefixRewrite),
	}
}

func topLevelCustomErrorPolicy(p *gcpurlmapv1alpha1.GcpUrlMapCustomErrorResponsePolicy) *compute.URLMapDefaultCustomErrorResponsePolicyArgs {
	return &compute.URLMapDefaultCustomErrorResponsePolicyArgs{
		ErrorService:       emptyAsNilString(p.ErrorService.GetValue()),
		ErrorResponseRules: topLevelErrorRules(p.ErrorResponseRules),
	}
}

func topLevelErrorRules(rules []*gcpurlmapv1alpha1.GcpUrlMapCustomErrorResponseRule) compute.URLMapDefaultCustomErrorResponsePolicyErrorResponseRuleArray {
	result := compute.URLMapDefaultCustomErrorResponsePolicyErrorResponseRuleArray{}
	for _, rule := range rules {
		codes := pulumi.StringArray{}
		for _, code := range rule.MatchResponseCodes {
			codes = append(codes, pulumi.String(code))
		}
		args := &compute.URLMapDefaultCustomErrorResponsePolicyErrorResponseRuleArgs{
			MatchResponseCodes: codes,
			Path:               pulumi.String(rule.Path),
		}
		if rule.OverrideResponseCode != 0 {
			args.OverrideResponseCode = pulumi.Int(int(rule.OverrideResponseCode))
		}
		result = append(result, args)
	}
	return result
}

func topLevelHeaderAction(h *gcpurlmapv1alpha1.GcpUrlMapHeaderAction) *compute.URLMapHeaderActionArgs {
	return &compute.URLMapHeaderActionArgs{
		RequestHeadersToAdds:     topLevelRequestHeadersToAdd(h.RequestHeadersToAdd),
		RequestHeadersToRemoves:  stringArrayOrNil(h.RequestHeadersToRemove),
		ResponseHeadersToAdds:    topLevelResponseHeadersToAdd(h.ResponseHeadersToAdd),
		ResponseHeadersToRemoves: stringArrayOrNil(h.ResponseHeadersToRemove),
	}
}

func topLevelRequestHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.URLMapHeaderActionRequestHeadersToAddArray {
	result := compute.URLMapHeaderActionRequestHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.URLMapHeaderActionRequestHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func topLevelResponseHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.URLMapHeaderActionResponseHeadersToAddArray {
	result := compute.URLMapHeaderActionResponseHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.URLMapHeaderActionResponseHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func buildHostRules(rules []*gcpurlmapv1alpha1.GcpUrlMapHostRule) compute.URLMapHostRuleArray {
	result := compute.URLMapHostRuleArray{}
	for _, rule := range rules {
		hosts := pulumi.StringArray{}
		for _, host := range rule.Hosts {
			hosts = append(hosts, pulumi.String(host))
		}
		args := &compute.URLMapHostRuleArgs{
			Hosts:       hosts,
			PathMatcher: pulumi.String(rule.PathMatcher),
		}
		if rule.Description != "" {
			args.Description = pulumi.String(rule.Description)
		}
		result = append(result, args)
	}
	return result
}

func buildPathMatchers(matchers []*gcpurlmapv1alpha1.GcpUrlMapPathMatcher) compute.URLMapPathMatcherArray {
	result := compute.URLMapPathMatcherArray{}
	for _, matcher := range matchers {
		args := &compute.URLMapPathMatcherArgs{
			Name: pulumi.String(matcher.Name),
		}
		if matcher.Description != "" {
			args.Description = pulumi.String(matcher.Description)
		}
		if matcher.DefaultService.GetValue() != "" {
			args.DefaultService = pulumi.String(matcher.DefaultService.GetValue())
		}
		if matcher.DefaultUrlRedirect != nil {
			args.DefaultUrlRedirect = pathMatcherUrlRedirect(matcher.DefaultUrlRedirect)
		}
		if matcher.DefaultRouteAction != nil {
			args.DefaultRouteAction = pathMatcherRouteAction(matcher.DefaultRouteAction)
		}
		if matcher.DefaultCustomErrorResponsePolicy != nil {
			args.DefaultCustomErrorResponsePolicy = pathMatcherCustomErrorPolicy(matcher.DefaultCustomErrorResponsePolicy)
		}
		if matcher.HeaderAction != nil {
			args.HeaderAction = pathMatcherHeaderAction(matcher.HeaderAction)
		}
		if len(matcher.PathRules) > 0 {
			args.PathRules = buildPathRules(matcher.PathRules)
		}
		if len(matcher.RouteRules) > 0 {
			args.RouteRules = buildRouteRules(matcher.RouteRules)
		}
		result = append(result, args)
	}
	return result
}

func pathMatcherUrlRedirect(r *gcpurlmapv1alpha1.GcpUrlMapUrlRedirect) *compute.URLMapPathMatcherDefaultUrlRedirectArgs {
	return &compute.URLMapPathMatcherDefaultUrlRedirectArgs{
		HostRedirect:         emptyAsNilString(r.HostRedirect),
		HttpsRedirect:        pulumi.Bool(r.HttpsRedirect),
		PathRedirect:         emptyAsNilString(r.PathRedirect),
		PrefixRedirect:       emptyAsNilString(r.PrefixRedirect),
		RedirectResponseCode: emptyAsNilString(r.RedirectResponseCode),
		StripQuery:           pulumi.Bool(r.StripQuery),
	}
}

func pathMatcherRouteAction(a *gcpurlmapv1alpha1.GcpUrlMapRouteAction) *compute.URLMapPathMatcherDefaultRouteActionArgs {
	out := &compute.URLMapPathMatcherDefaultRouteActionArgs{}
	if len(a.WeightedBackendServices) > 0 {
		out.WeightedBackendServices = pathMatcherDefaultWeightedBackends(a.WeightedBackendServices)
	}
	if a.UrlRewrite != nil {
		out.UrlRewrite = pathMatcherDefaultUrlRewrite(a.UrlRewrite)
	}
	if a.Timeout != nil {
		out.Timeout = &compute.URLMapPathMatcherDefaultRouteActionTimeoutArgs{
			Seconds: durationSeconds(a.Timeout),
			Nanos:   durationNanos(a.Timeout),
		}
	}
	if a.RetryPolicy != nil {
		retry := &compute.URLMapPathMatcherDefaultRouteActionRetryPolicyArgs{
			RetryConditions: stringArrayOrNil(a.RetryPolicy.RetryConditions),
		}
		if a.RetryPolicy.NumRetries != 0 {
			retry.NumRetries = pulumi.Int(int(a.RetryPolicy.NumRetries))
		}
		if a.RetryPolicy.PerTryTimeout != nil {
			retry.PerTryTimeout = &compute.URLMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeoutArgs{
				Seconds: durationSeconds(a.RetryPolicy.PerTryTimeout),
				Nanos:   durationNanos(a.RetryPolicy.PerTryTimeout),
			}
		}
		out.RetryPolicy = retry
	}
	if a.RequestMirrorPolicy != nil {
		out.RequestMirrorPolicy = &compute.URLMapPathMatcherDefaultRouteActionRequestMirrorPolicyArgs{
			BackendService: pulumi.String(a.RequestMirrorPolicy.BackendService.GetValue()),
		}
	}
	if a.CorsPolicy != nil {
		cors := &compute.URLMapPathMatcherDefaultRouteActionCorsPolicyArgs{
			AllowCredentials:   pulumi.Bool(a.CorsPolicy.AllowCredentials),
			AllowHeaders:       stringArrayOrNil(a.CorsPolicy.AllowHeaders),
			AllowMethods:       stringArrayOrNil(a.CorsPolicy.AllowMethods),
			AllowOriginRegexes: stringArrayOrNil(a.CorsPolicy.AllowOriginRegexes),
			AllowOrigins:       stringArrayOrNil(a.CorsPolicy.AllowOrigins),
			Disabled:           pulumi.Bool(a.CorsPolicy.Disabled),
			ExposeHeaders:      stringArrayOrNil(a.CorsPolicy.ExposeHeaders),
		}
		if a.CorsPolicy.MaxAge != 0 {
			cors.MaxAge = pulumi.Int(int(a.CorsPolicy.MaxAge))
		}
		out.CorsPolicy = cors
	}
	if a.FaultInjectionPolicy != nil {
		fault := &compute.URLMapPathMatcherDefaultRouteActionFaultInjectionPolicyArgs{}
		if a.FaultInjectionPolicy.Abort != nil {
			abort := &compute.URLMapPathMatcherDefaultRouteActionFaultInjectionPolicyAbortArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Abort.Percentage),
			}
			if a.FaultInjectionPolicy.Abort.HttpStatus != 0 {
				abort.HttpStatus = pulumi.Int(int(a.FaultInjectionPolicy.Abort.HttpStatus))
			}
			fault.Abort = abort
		}
		if a.FaultInjectionPolicy.Delay != nil {
			delay := &compute.URLMapPathMatcherDefaultRouteActionFaultInjectionPolicyDelayArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Delay.Percentage),
			}
			if a.FaultInjectionPolicy.Delay.FixedDelay != nil {
				delay.FixedDelay = &compute.URLMapPathMatcherDefaultRouteActionFaultInjectionPolicyDelayFixedDelayArgs{
					Seconds: durationSeconds(a.FaultInjectionPolicy.Delay.FixedDelay),
					Nanos:   durationNanos(a.FaultInjectionPolicy.Delay.FixedDelay),
				}
			}
			fault.Delay = delay
		}
		out.FaultInjectionPolicy = fault
	}
	if a.MaxStreamDuration != nil {
		out.MaxStreamDuration = &compute.URLMapPathMatcherDefaultRouteActionMaxStreamDurationArgs{
			Seconds: durationSeconds(a.MaxStreamDuration),
			Nanos:   durationNanos(a.MaxStreamDuration),
		}
	}
	if a.CachePolicy != nil {
		cache := &compute.URLMapPathMatcherDefaultRouteActionCachePolicyArgs{
			CacheMode:                     emptyAsNilString(a.CachePolicy.CacheMode),
			CacheBypassRequestHeaderNames: stringArrayOrNil(a.CachePolicy.CacheBypassRequestHeaderNames),
			NegativeCaching:               pulumi.Bool(a.CachePolicy.NegativeCaching),
			RequestCoalescing:             pulumi.Bool(a.CachePolicy.RequestCoalescing),
		}
		if a.CachePolicy.CacheKeyPolicy != nil {
			cache.CacheKeyPolicy = &compute.URLMapPathMatcherDefaultRouteActionCachePolicyCacheKeyPolicyArgs{
				ExcludedQueryParameters: stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.ExcludedQueryParameters),
				IncludeHost:             pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeHost),
				IncludeProtocol:         pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeProtocol),
				IncludeQueryString:      pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeQueryString),
				IncludedCookieNames:     stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedCookieNames),
				IncludedHeaderNames:     stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedHeaderNames),
				IncludedQueryParameters: stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedQueryParameters),
			}
		}
		if a.CachePolicy.ClientTtl != nil {
			cache.ClientTtl = &compute.URLMapPathMatcherDefaultRouteActionCachePolicyClientTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.ClientTtl),
				Nanos:   durationNanos(a.CachePolicy.ClientTtl),
			}
		}
		if a.CachePolicy.DefaultTtl != nil {
			cache.DefaultTtl = &compute.URLMapPathMatcherDefaultRouteActionCachePolicyDefaultTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.DefaultTtl),
				Nanos:   durationNanos(a.CachePolicy.DefaultTtl),
			}
		}
		if a.CachePolicy.MaxTtl != nil {
			cache.MaxTtl = &compute.URLMapPathMatcherDefaultRouteActionCachePolicyMaxTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.MaxTtl),
				Nanos:   durationNanos(a.CachePolicy.MaxTtl),
			}
		}
		if a.CachePolicy.ServeWhileStale != nil {
			cache.ServeWhileStale = &compute.URLMapPathMatcherDefaultRouteActionCachePolicyServeWhileStaleArgs{
				Seconds: durationSeconds(a.CachePolicy.ServeWhileStale),
				Nanos:   durationNanos(a.CachePolicy.ServeWhileStale),
			}
		}
		if len(a.CachePolicy.NegativeCachingPolicy) > 0 {
			policies := compute.URLMapPathMatcherDefaultRouteActionCachePolicyNegativeCachingPolicyArray{}
			for _, p := range a.CachePolicy.NegativeCachingPolicy {
				policy := &compute.URLMapPathMatcherDefaultRouteActionCachePolicyNegativeCachingPolicyArgs{}
				if p.Code != 0 {
					policy.Code = pulumi.Int(int(p.Code))
				}
				if p.Ttl != nil {
					policy.Ttl = &compute.URLMapPathMatcherDefaultRouteActionCachePolicyNegativeCachingPolicyTtlArgs{
						Seconds: durationSeconds(p.Ttl),
						Nanos:   durationNanos(p.Ttl),
					}
				}
				policies = append(policies, policy)
			}
			cache.NegativeCachingPolicies = policies
		}
		out.CachePolicy = cache
	}
	return out
}

func pathMatcherDefaultWeightedBackends(backends []*gcpurlmapv1alpha1.GcpUrlMapWeightedBackendService) compute.URLMapPathMatcherDefaultRouteActionWeightedBackendServiceArray {
	result := compute.URLMapPathMatcherDefaultRouteActionWeightedBackendServiceArray{}
	for _, backend := range backends {
		args := &compute.URLMapPathMatcherDefaultRouteActionWeightedBackendServiceArgs{
			BackendService: pulumi.String(backend.BackendService.GetValue()),
			Weight:         pulumi.Int(int(backend.Weight)),
		}
		if backend.HeaderAction != nil {
			headerAction := &compute.URLMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionArgs{
				RequestHeadersToRemoves:  stringArrayOrNil(backend.HeaderAction.RequestHeadersToRemove),
				ResponseHeadersToRemoves: stringArrayOrNil(backend.HeaderAction.ResponseHeadersToRemove),
			}
			if len(backend.HeaderAction.RequestHeadersToAdd) > 0 {
				adds := compute.URLMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArray{}
				for _, h := range backend.HeaderAction.RequestHeadersToAdd {
					adds = append(adds, &compute.URLMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.RequestHeadersToAdds = adds
			}
			if len(backend.HeaderAction.ResponseHeadersToAdd) > 0 {
				adds := compute.URLMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArray{}
				for _, h := range backend.HeaderAction.ResponseHeadersToAdd {
					adds = append(adds, &compute.URLMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.ResponseHeadersToAdds = adds
			}
			args.HeaderAction = headerAction
		}
		result = append(result, args)
	}
	return result
}

func pathMatcherDefaultUrlRewrite(r *gcpurlmapv1alpha1.GcpUrlMapUrlRewrite) *compute.URLMapPathMatcherDefaultRouteActionUrlRewriteArgs {
	return &compute.URLMapPathMatcherDefaultRouteActionUrlRewriteArgs{
		HostRewrite:       emptyAsNilString(r.HostRewrite),
		PathPrefixRewrite: emptyAsNilString(r.PathPrefixRewrite),
	}
}

func pathMatcherCustomErrorPolicy(p *gcpurlmapv1alpha1.GcpUrlMapCustomErrorResponsePolicy) *compute.URLMapPathMatcherDefaultCustomErrorResponsePolicyArgs {
	return &compute.URLMapPathMatcherDefaultCustomErrorResponsePolicyArgs{
		ErrorService:       emptyAsNilString(p.ErrorService.GetValue()),
		ErrorResponseRules: pathMatcherErrorRules(p.ErrorResponseRules),
	}
}

func pathMatcherErrorRules(rules []*gcpurlmapv1alpha1.GcpUrlMapCustomErrorResponseRule) compute.URLMapPathMatcherDefaultCustomErrorResponsePolicyErrorResponseRuleArray {
	result := compute.URLMapPathMatcherDefaultCustomErrorResponsePolicyErrorResponseRuleArray{}
	for _, rule := range rules {
		codes := pulumi.StringArray{}
		for _, code := range rule.MatchResponseCodes {
			codes = append(codes, pulumi.String(code))
		}
		args := &compute.URLMapPathMatcherDefaultCustomErrorResponsePolicyErrorResponseRuleArgs{
			MatchResponseCodes: codes,
			Path:               pulumi.String(rule.Path),
		}
		if rule.OverrideResponseCode != 0 {
			args.OverrideResponseCode = pulumi.Int(int(rule.OverrideResponseCode))
		}
		result = append(result, args)
	}
	return result
}

func pathMatcherHeaderAction(h *gcpurlmapv1alpha1.GcpUrlMapHeaderAction) *compute.URLMapPathMatcherHeaderActionArgs {
	return &compute.URLMapPathMatcherHeaderActionArgs{
		RequestHeadersToAdds:     pathMatcherRequestHeadersToAdd(h.RequestHeadersToAdd),
		RequestHeadersToRemoves:  stringArrayOrNil(h.RequestHeadersToRemove),
		ResponseHeadersToAdds:    pathMatcherResponseHeadersToAdd(h.ResponseHeadersToAdd),
		ResponseHeadersToRemoves: stringArrayOrNil(h.ResponseHeadersToRemove),
	}
}

func pathMatcherRequestHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.URLMapPathMatcherHeaderActionRequestHeadersToAddArray {
	result := compute.URLMapPathMatcherHeaderActionRequestHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.URLMapPathMatcherHeaderActionRequestHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func pathMatcherResponseHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.URLMapPathMatcherHeaderActionResponseHeadersToAddArray {
	result := compute.URLMapPathMatcherHeaderActionResponseHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.URLMapPathMatcherHeaderActionResponseHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func buildPathRules(rules []*gcpurlmapv1alpha1.GcpUrlMapPathRule) compute.URLMapPathMatcherPathRuleArray {
	result := compute.URLMapPathMatcherPathRuleArray{}
	for _, rule := range rules {
		paths := pulumi.StringArray{}
		for _, path := range rule.Paths {
			paths = append(paths, pulumi.String(path))
		}
		args := &compute.URLMapPathMatcherPathRuleArgs{
			Paths: paths,
		}
		if rule.Service.GetValue() != "" {
			args.Service = pulumi.String(rule.Service.GetValue())
		}
		if rule.UrlRedirect != nil {
			args.UrlRedirect = pathRuleUrlRedirect(rule.UrlRedirect)
		}
		if rule.RouteAction != nil {
			args.RouteAction = pathRuleRouteAction(rule.RouteAction)
		}
		if rule.CustomErrorResponsePolicy != nil {
			args.CustomErrorResponsePolicy = pathRuleCustomErrorPolicy(rule.CustomErrorResponsePolicy)
		}
		result = append(result, args)
	}
	return result
}

func pathRuleUrlRedirect(r *gcpurlmapv1alpha1.GcpUrlMapUrlRedirect) *compute.URLMapPathMatcherPathRuleUrlRedirectArgs {
	return &compute.URLMapPathMatcherPathRuleUrlRedirectArgs{
		HostRedirect:         emptyAsNilString(r.HostRedirect),
		HttpsRedirect:        pulumi.Bool(r.HttpsRedirect),
		PathRedirect:         emptyAsNilString(r.PathRedirect),
		PrefixRedirect:       emptyAsNilString(r.PrefixRedirect),
		RedirectResponseCode: emptyAsNilString(r.RedirectResponseCode),
		StripQuery:           pulumi.Bool(r.StripQuery),
	}
}

func pathRuleRouteAction(a *gcpurlmapv1alpha1.GcpUrlMapRouteAction) *compute.URLMapPathMatcherPathRuleRouteActionArgs {
	out := &compute.URLMapPathMatcherPathRuleRouteActionArgs{}
	if len(a.WeightedBackendServices) > 0 {
		out.WeightedBackendServices = pathRuleWeightedBackends(a.WeightedBackendServices)
	}
	if a.UrlRewrite != nil {
		out.UrlRewrite = pathRuleUrlRewrite(a.UrlRewrite)
	}
	if a.Timeout != nil {
		out.Timeout = &compute.URLMapPathMatcherPathRuleRouteActionTimeoutArgs{
			Seconds: durationSeconds(a.Timeout),
			Nanos:   durationNanos(a.Timeout),
		}
	}
	if a.RetryPolicy != nil {
		retry := &compute.URLMapPathMatcherPathRuleRouteActionRetryPolicyArgs{
			RetryConditions: stringArrayOrNil(a.RetryPolicy.RetryConditions),
		}
		if a.RetryPolicy.NumRetries != 0 {
			retry.NumRetries = pulumi.Int(int(a.RetryPolicy.NumRetries))
		}
		if a.RetryPolicy.PerTryTimeout != nil {
			retry.PerTryTimeout = &compute.URLMapPathMatcherPathRuleRouteActionRetryPolicyPerTryTimeoutArgs{
				Seconds: durationSeconds(a.RetryPolicy.PerTryTimeout),
				Nanos:   durationNanos(a.RetryPolicy.PerTryTimeout),
			}
		}
		out.RetryPolicy = retry
	}
	if a.RequestMirrorPolicy != nil {
		out.RequestMirrorPolicy = &compute.URLMapPathMatcherPathRuleRouteActionRequestMirrorPolicyArgs{
			BackendService: pulumi.String(a.RequestMirrorPolicy.BackendService.GetValue()),
		}
	}
	if a.CorsPolicy != nil {
		cors := &compute.URLMapPathMatcherPathRuleRouteActionCorsPolicyArgs{
			AllowCredentials:   pulumi.Bool(a.CorsPolicy.AllowCredentials),
			AllowHeaders:       stringArrayOrNil(a.CorsPolicy.AllowHeaders),
			AllowMethods:       stringArrayOrNil(a.CorsPolicy.AllowMethods),
			AllowOriginRegexes: stringArrayOrNil(a.CorsPolicy.AllowOriginRegexes),
			AllowOrigins:       stringArrayOrNil(a.CorsPolicy.AllowOrigins),
			Disabled:           pulumi.Bool(a.CorsPolicy.Disabled),
			ExposeHeaders:      stringArrayOrNil(a.CorsPolicy.ExposeHeaders),
		}
		if a.CorsPolicy.MaxAge != 0 {
			cors.MaxAge = pulumi.Int(int(a.CorsPolicy.MaxAge))
		}
		out.CorsPolicy = cors
	}
	if a.FaultInjectionPolicy != nil {
		fault := &compute.URLMapPathMatcherPathRuleRouteActionFaultInjectionPolicyArgs{}
		if a.FaultInjectionPolicy.Abort != nil {
			abort := &compute.URLMapPathMatcherPathRuleRouteActionFaultInjectionPolicyAbortArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Abort.Percentage),
			}
			if a.FaultInjectionPolicy.Abort.HttpStatus != 0 {
				abort.HttpStatus = pulumi.Int(int(a.FaultInjectionPolicy.Abort.HttpStatus))
			}
			fault.Abort = abort
		}
		if a.FaultInjectionPolicy.Delay != nil {
			delay := &compute.URLMapPathMatcherPathRuleRouteActionFaultInjectionPolicyDelayArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Delay.Percentage),
			}
			if a.FaultInjectionPolicy.Delay.FixedDelay != nil {
				delay.FixedDelay = &compute.URLMapPathMatcherPathRuleRouteActionFaultInjectionPolicyDelayFixedDelayArgs{
					Seconds: durationSeconds(a.FaultInjectionPolicy.Delay.FixedDelay),
					Nanos:   durationNanos(a.FaultInjectionPolicy.Delay.FixedDelay),
				}
			}
			fault.Delay = delay
		}
		out.FaultInjectionPolicy = fault
	}
	if a.MaxStreamDuration != nil {
		out.MaxStreamDuration = &compute.URLMapPathMatcherPathRuleRouteActionMaxStreamDurationArgs{
			Seconds: durationSeconds(a.MaxStreamDuration),
			Nanos:   durationNanos(a.MaxStreamDuration),
		}
	}
	if a.CachePolicy != nil {
		cache := &compute.URLMapPathMatcherPathRuleRouteActionCachePolicyArgs{
			CacheMode:                     emptyAsNilString(a.CachePolicy.CacheMode),
			CacheBypassRequestHeaderNames: stringArrayOrNil(a.CachePolicy.CacheBypassRequestHeaderNames),
			NegativeCaching:               pulumi.Bool(a.CachePolicy.NegativeCaching),
			RequestCoalescing:             pulumi.Bool(a.CachePolicy.RequestCoalescing),
		}
		if a.CachePolicy.CacheKeyPolicy != nil {
			cache.CacheKeyPolicy = &compute.URLMapPathMatcherPathRuleRouteActionCachePolicyCacheKeyPolicyArgs{
				ExcludedQueryParameters: stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.ExcludedQueryParameters),
				IncludeHost:             pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeHost),
				IncludeProtocol:         pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeProtocol),
				IncludeQueryString:      pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeQueryString),
				IncludedCookieNames:     stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedCookieNames),
				IncludedHeaderNames:     stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedHeaderNames),
				IncludedQueryParameters: stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedQueryParameters),
			}
		}
		if a.CachePolicy.ClientTtl != nil {
			cache.ClientTtl = &compute.URLMapPathMatcherPathRuleRouteActionCachePolicyClientTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.ClientTtl),
				Nanos:   durationNanos(a.CachePolicy.ClientTtl),
			}
		}
		if a.CachePolicy.DefaultTtl != nil {
			cache.DefaultTtl = &compute.URLMapPathMatcherPathRuleRouteActionCachePolicyDefaultTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.DefaultTtl),
				Nanos:   durationNanos(a.CachePolicy.DefaultTtl),
			}
		}
		if a.CachePolicy.MaxTtl != nil {
			cache.MaxTtl = &compute.URLMapPathMatcherPathRuleRouteActionCachePolicyMaxTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.MaxTtl),
				Nanos:   durationNanos(a.CachePolicy.MaxTtl),
			}
		}
		if a.CachePolicy.ServeWhileStale != nil {
			cache.ServeWhileStale = &compute.URLMapPathMatcherPathRuleRouteActionCachePolicyServeWhileStaleArgs{
				Seconds: durationSeconds(a.CachePolicy.ServeWhileStale),
				Nanos:   durationNanos(a.CachePolicy.ServeWhileStale),
			}
		}
		if len(a.CachePolicy.NegativeCachingPolicy) > 0 {
			policies := compute.URLMapPathMatcherPathRuleRouteActionCachePolicyNegativeCachingPolicyArray{}
			for _, p := range a.CachePolicy.NegativeCachingPolicy {
				policy := &compute.URLMapPathMatcherPathRuleRouteActionCachePolicyNegativeCachingPolicyArgs{}
				if p.Code != 0 {
					policy.Code = pulumi.Int(int(p.Code))
				}
				if p.Ttl != nil {
					policy.Ttl = &compute.URLMapPathMatcherPathRuleRouteActionCachePolicyNegativeCachingPolicyTtlArgs{
						Seconds: durationSeconds(p.Ttl),
						Nanos:   durationNanos(p.Ttl),
					}
				}
				policies = append(policies, policy)
			}
			cache.NegativeCachingPolicies = policies
		}
		out.CachePolicy = cache
	}
	return out
}

func pathRuleWeightedBackends(backends []*gcpurlmapv1alpha1.GcpUrlMapWeightedBackendService) compute.URLMapPathMatcherPathRuleRouteActionWeightedBackendServiceArray {
	result := compute.URLMapPathMatcherPathRuleRouteActionWeightedBackendServiceArray{}
	for _, backend := range backends {
		args := &compute.URLMapPathMatcherPathRuleRouteActionWeightedBackendServiceArgs{
			BackendService: pulumi.String(backend.BackendService.GetValue()),
			Weight:         pulumi.Int(int(backend.Weight)),
		}
		if backend.HeaderAction != nil {
			headerAction := &compute.URLMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionArgs{
				RequestHeadersToRemoves:  stringArrayOrNil(backend.HeaderAction.RequestHeadersToRemove),
				ResponseHeadersToRemoves: stringArrayOrNil(backend.HeaderAction.ResponseHeadersToRemove),
			}
			if len(backend.HeaderAction.RequestHeadersToAdd) > 0 {
				adds := compute.URLMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArray{}
				for _, h := range backend.HeaderAction.RequestHeadersToAdd {
					adds = append(adds, &compute.URLMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.RequestHeadersToAdds = adds
			}
			if len(backend.HeaderAction.ResponseHeadersToAdd) > 0 {
				adds := compute.URLMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArray{}
				for _, h := range backend.HeaderAction.ResponseHeadersToAdd {
					adds = append(adds, &compute.URLMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.ResponseHeadersToAdds = adds
			}
			args.HeaderAction = headerAction
		}
		result = append(result, args)
	}
	return result
}

func pathRuleUrlRewrite(r *gcpurlmapv1alpha1.GcpUrlMapUrlRewrite) *compute.URLMapPathMatcherPathRuleRouteActionUrlRewriteArgs {
	return &compute.URLMapPathMatcherPathRuleRouteActionUrlRewriteArgs{
		HostRewrite:       emptyAsNilString(r.HostRewrite),
		PathPrefixRewrite: emptyAsNilString(r.PathPrefixRewrite),
	}
}

func pathRuleCustomErrorPolicy(p *gcpurlmapv1alpha1.GcpUrlMapCustomErrorResponsePolicy) *compute.URLMapPathMatcherPathRuleCustomErrorResponsePolicyArgs {
	return &compute.URLMapPathMatcherPathRuleCustomErrorResponsePolicyArgs{
		ErrorService:       emptyAsNilString(p.ErrorService.GetValue()),
		ErrorResponseRules: pathRuleErrorRules(p.ErrorResponseRules),
	}
}

func pathRuleErrorRules(rules []*gcpurlmapv1alpha1.GcpUrlMapCustomErrorResponseRule) compute.URLMapPathMatcherPathRuleCustomErrorResponsePolicyErrorResponseRuleArray {
	result := compute.URLMapPathMatcherPathRuleCustomErrorResponsePolicyErrorResponseRuleArray{}
	for _, rule := range rules {
		codes := pulumi.StringArray{}
		for _, code := range rule.MatchResponseCodes {
			codes = append(codes, pulumi.String(code))
		}
		args := &compute.URLMapPathMatcherPathRuleCustomErrorResponsePolicyErrorResponseRuleArgs{
			MatchResponseCodes: codes,
			Path:               pulumi.String(rule.Path),
		}
		if rule.OverrideResponseCode != 0 {
			args.OverrideResponseCode = pulumi.Int(int(rule.OverrideResponseCode))
		}
		result = append(result, args)
	}
	return result
}

func buildRouteRules(rules []*gcpurlmapv1alpha1.GcpUrlMapRouteRule) compute.URLMapPathMatcherRouteRuleArray {
	result := compute.URLMapPathMatcherRouteRuleArray{}
	for _, rule := range rules {
		args := &compute.URLMapPathMatcherRouteRuleArgs{
			Priority:   pulumi.Int(int(rule.Priority)),
			MatchRules: buildMatchRules(rule.MatchRules),
		}
		if rule.Service.GetValue() != "" {
			args.Service = pulumi.String(rule.Service.GetValue())
		}
		if rule.UrlRedirect != nil {
			args.UrlRedirect = routeRuleUrlRedirect(rule.UrlRedirect)
		}
		if rule.RouteAction != nil {
			args.RouteAction = routeRuleRouteAction(rule.RouteAction)
		}
		if rule.HeaderAction != nil {
			args.HeaderAction = routeRuleHeaderAction(rule.HeaderAction)
		}
		if rule.CustomErrorResponsePolicy != nil {
			args.CustomErrorResponsePolicy = routeRuleCustomErrorPolicy(rule.CustomErrorResponsePolicy)
		}
		result = append(result, args)
	}
	return result
}

func routeRuleUrlRedirect(r *gcpurlmapv1alpha1.GcpUrlMapUrlRedirect) *compute.URLMapPathMatcherRouteRuleUrlRedirectArgs {
	return &compute.URLMapPathMatcherRouteRuleUrlRedirectArgs{
		HostRedirect:         emptyAsNilString(r.HostRedirect),
		HttpsRedirect:        pulumi.Bool(r.HttpsRedirect),
		PathRedirect:         emptyAsNilString(r.PathRedirect),
		PrefixRedirect:       emptyAsNilString(r.PrefixRedirect),
		RedirectResponseCode: emptyAsNilString(r.RedirectResponseCode),
		StripQuery:           pulumi.Bool(r.StripQuery),
	}
}

func routeRuleRouteAction(a *gcpurlmapv1alpha1.GcpUrlMapRouteAction) *compute.URLMapPathMatcherRouteRuleRouteActionArgs {
	out := &compute.URLMapPathMatcherRouteRuleRouteActionArgs{}
	if len(a.WeightedBackendServices) > 0 {
		out.WeightedBackendServices = routeRuleWeightedBackends(a.WeightedBackendServices)
	}
	if a.UrlRewrite != nil {
		out.UrlRewrite = routeRuleUrlRewrite(a.UrlRewrite)
	}
	if a.Timeout != nil {
		out.Timeout = &compute.URLMapPathMatcherRouteRuleRouteActionTimeoutArgs{
			Seconds: durationSeconds(a.Timeout),
			Nanos:   durationNanos(a.Timeout),
		}
	}
	if a.RetryPolicy != nil {
		retry := &compute.URLMapPathMatcherRouteRuleRouteActionRetryPolicyArgs{
			RetryConditions: stringArrayOrNil(a.RetryPolicy.RetryConditions),
		}
		if a.RetryPolicy.NumRetries != 0 {
			retry.NumRetries = pulumi.Int(int(a.RetryPolicy.NumRetries))
		}
		if a.RetryPolicy.PerTryTimeout != nil {
			retry.PerTryTimeout = &compute.URLMapPathMatcherRouteRuleRouteActionRetryPolicyPerTryTimeoutArgs{
				Seconds: durationSeconds(a.RetryPolicy.PerTryTimeout),
				Nanos:   durationNanos(a.RetryPolicy.PerTryTimeout),
			}
		}
		out.RetryPolicy = retry
	}
	if a.RequestMirrorPolicy != nil {
		out.RequestMirrorPolicy = &compute.URLMapPathMatcherRouteRuleRouteActionRequestMirrorPolicyArgs{
			BackendService: pulumi.String(a.RequestMirrorPolicy.BackendService.GetValue()),
		}
	}
	if a.CorsPolicy != nil {
		cors := &compute.URLMapPathMatcherRouteRuleRouteActionCorsPolicyArgs{
			AllowCredentials:   pulumi.Bool(a.CorsPolicy.AllowCredentials),
			AllowHeaders:       stringArrayOrNil(a.CorsPolicy.AllowHeaders),
			AllowMethods:       stringArrayOrNil(a.CorsPolicy.AllowMethods),
			AllowOriginRegexes: stringArrayOrNil(a.CorsPolicy.AllowOriginRegexes),
			AllowOrigins:       stringArrayOrNil(a.CorsPolicy.AllowOrigins),
			Disabled:           pulumi.Bool(a.CorsPolicy.Disabled),
			ExposeHeaders:      stringArrayOrNil(a.CorsPolicy.ExposeHeaders),
		}
		if a.CorsPolicy.MaxAge != 0 {
			cors.MaxAge = pulumi.Int(int(a.CorsPolicy.MaxAge))
		}
		out.CorsPolicy = cors
	}
	if a.FaultInjectionPolicy != nil {
		fault := &compute.URLMapPathMatcherRouteRuleRouteActionFaultInjectionPolicyArgs{}
		if a.FaultInjectionPolicy.Abort != nil {
			abort := &compute.URLMapPathMatcherRouteRuleRouteActionFaultInjectionPolicyAbortArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Abort.Percentage),
			}
			if a.FaultInjectionPolicy.Abort.HttpStatus != 0 {
				abort.HttpStatus = pulumi.Int(int(a.FaultInjectionPolicy.Abort.HttpStatus))
			}
			fault.Abort = abort
		}
		if a.FaultInjectionPolicy.Delay != nil {
			delay := &compute.URLMapPathMatcherRouteRuleRouteActionFaultInjectionPolicyDelayArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Delay.Percentage),
			}
			if a.FaultInjectionPolicy.Delay.FixedDelay != nil {
				delay.FixedDelay = &compute.URLMapPathMatcherRouteRuleRouteActionFaultInjectionPolicyDelayFixedDelayArgs{
					Seconds: durationSeconds(a.FaultInjectionPolicy.Delay.FixedDelay),
					Nanos:   durationNanos(a.FaultInjectionPolicy.Delay.FixedDelay),
				}
			}
			fault.Delay = delay
		}
		out.FaultInjectionPolicy = fault
	}
	if a.MaxStreamDuration != nil {
		out.MaxStreamDuration = &compute.URLMapPathMatcherRouteRuleRouteActionMaxStreamDurationArgs{
			Seconds: durationSeconds(a.MaxStreamDuration),
			Nanos:   durationNanos(a.MaxStreamDuration),
		}
	}
	if a.CachePolicy != nil {
		cache := &compute.URLMapPathMatcherRouteRuleRouteActionCachePolicyArgs{
			CacheMode:                     emptyAsNilString(a.CachePolicy.CacheMode),
			CacheBypassRequestHeaderNames: stringArrayOrNil(a.CachePolicy.CacheBypassRequestHeaderNames),
			NegativeCaching:               pulumi.Bool(a.CachePolicy.NegativeCaching),
			RequestCoalescing:             pulumi.Bool(a.CachePolicy.RequestCoalescing),
		}
		if a.CachePolicy.CacheKeyPolicy != nil {
			cache.CacheKeyPolicy = &compute.URLMapPathMatcherRouteRuleRouteActionCachePolicyCacheKeyPolicyArgs{
				ExcludedQueryParameters: stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.ExcludedQueryParameters),
				IncludeHost:             pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeHost),
				IncludeProtocol:         pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeProtocol),
				IncludeQueryString:      pulumi.Bool(a.CachePolicy.CacheKeyPolicy.IncludeQueryString),
				IncludedCookieNames:     stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedCookieNames),
				IncludedHeaderNames:     stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedHeaderNames),
				IncludedQueryParameters: stringArrayOrNil(a.CachePolicy.CacheKeyPolicy.IncludedQueryParameters),
			}
		}
		if a.CachePolicy.ClientTtl != nil {
			cache.ClientTtl = &compute.URLMapPathMatcherRouteRuleRouteActionCachePolicyClientTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.ClientTtl),
				Nanos:   durationNanos(a.CachePolicy.ClientTtl),
			}
		}
		if a.CachePolicy.DefaultTtl != nil {
			cache.DefaultTtl = &compute.URLMapPathMatcherRouteRuleRouteActionCachePolicyDefaultTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.DefaultTtl),
				Nanos:   durationNanos(a.CachePolicy.DefaultTtl),
			}
		}
		if a.CachePolicy.MaxTtl != nil {
			cache.MaxTtl = &compute.URLMapPathMatcherRouteRuleRouteActionCachePolicyMaxTtlArgs{
				Seconds: durationSeconds(a.CachePolicy.MaxTtl),
				Nanos:   durationNanos(a.CachePolicy.MaxTtl),
			}
		}
		if a.CachePolicy.ServeWhileStale != nil {
			cache.ServeWhileStale = &compute.URLMapPathMatcherRouteRuleRouteActionCachePolicyServeWhileStaleArgs{
				Seconds: durationSeconds(a.CachePolicy.ServeWhileStale),
				Nanos:   durationNanos(a.CachePolicy.ServeWhileStale),
			}
		}
		if len(a.CachePolicy.NegativeCachingPolicy) > 0 {
			policies := compute.URLMapPathMatcherRouteRuleRouteActionCachePolicyNegativeCachingPolicyArray{}
			for _, p := range a.CachePolicy.NegativeCachingPolicy {
				policy := &compute.URLMapPathMatcherRouteRuleRouteActionCachePolicyNegativeCachingPolicyArgs{}
				if p.Code != 0 {
					policy.Code = pulumi.Int(int(p.Code))
				}
				if p.Ttl != nil {
					policy.Ttl = &compute.URLMapPathMatcherRouteRuleRouteActionCachePolicyNegativeCachingPolicyTtlArgs{
						Seconds: durationSeconds(p.Ttl),
						Nanos:   durationNanos(p.Ttl),
					}
				}
				policies = append(policies, policy)
			}
			cache.NegativeCachingPolicies = policies
		}
		out.CachePolicy = cache
	}
	return out
}

func routeRuleWeightedBackends(backends []*gcpurlmapv1alpha1.GcpUrlMapWeightedBackendService) compute.URLMapPathMatcherRouteRuleRouteActionWeightedBackendServiceArray {
	result := compute.URLMapPathMatcherRouteRuleRouteActionWeightedBackendServiceArray{}
	for _, backend := range backends {
		args := &compute.URLMapPathMatcherRouteRuleRouteActionWeightedBackendServiceArgs{
			BackendService: pulumi.String(backend.BackendService.GetValue()),
			Weight:         pulumi.Int(int(backend.Weight)),
		}
		if backend.HeaderAction != nil {
			headerAction := &compute.URLMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionArgs{
				RequestHeadersToRemoves:  stringArrayOrNil(backend.HeaderAction.RequestHeadersToRemove),
				ResponseHeadersToRemoves: stringArrayOrNil(backend.HeaderAction.ResponseHeadersToRemove),
			}
			if len(backend.HeaderAction.RequestHeadersToAdd) > 0 {
				adds := compute.URLMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArray{}
				for _, h := range backend.HeaderAction.RequestHeadersToAdd {
					adds = append(adds, &compute.URLMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.RequestHeadersToAdds = adds
			}
			if len(backend.HeaderAction.ResponseHeadersToAdd) > 0 {
				adds := compute.URLMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArray{}
				for _, h := range backend.HeaderAction.ResponseHeadersToAdd {
					adds = append(adds, &compute.URLMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.ResponseHeadersToAdds = adds
			}
			args.HeaderAction = headerAction
		}
		result = append(result, args)
	}
	return result
}

func routeRuleUrlRewrite(r *gcpurlmapv1alpha1.GcpUrlMapUrlRewrite) *compute.URLMapPathMatcherRouteRuleRouteActionUrlRewriteArgs {
	return &compute.URLMapPathMatcherRouteRuleRouteActionUrlRewriteArgs{
		HostRewrite:         emptyAsNilString(r.HostRewrite),
		PathPrefixRewrite:   emptyAsNilString(r.PathPrefixRewrite),
		PathTemplateRewrite: emptyAsNilString(r.PathTemplateRewrite),
	}
}

func routeRuleCustomErrorPolicy(p *gcpurlmapv1alpha1.GcpUrlMapCustomErrorResponsePolicy) *compute.URLMapPathMatcherRouteRuleCustomErrorResponsePolicyArgs {
	return &compute.URLMapPathMatcherRouteRuleCustomErrorResponsePolicyArgs{
		ErrorService:       emptyAsNilString(p.ErrorService.GetValue()),
		ErrorResponseRules: routeRuleErrorRules(p.ErrorResponseRules),
	}
}

func routeRuleErrorRules(rules []*gcpurlmapv1alpha1.GcpUrlMapCustomErrorResponseRule) compute.URLMapPathMatcherRouteRuleCustomErrorResponsePolicyErrorResponseRuleArray {
	result := compute.URLMapPathMatcherRouteRuleCustomErrorResponsePolicyErrorResponseRuleArray{}
	for _, rule := range rules {
		codes := pulumi.StringArray{}
		for _, code := range rule.MatchResponseCodes {
			codes = append(codes, pulumi.String(code))
		}
		args := &compute.URLMapPathMatcherRouteRuleCustomErrorResponsePolicyErrorResponseRuleArgs{
			MatchResponseCodes: codes,
			Path:               pulumi.String(rule.Path),
		}
		if rule.OverrideResponseCode != 0 {
			args.OverrideResponseCode = pulumi.Int(int(rule.OverrideResponseCode))
		}
		result = append(result, args)
	}
	return result
}

func routeRuleHeaderAction(h *gcpurlmapv1alpha1.GcpUrlMapHeaderAction) *compute.URLMapPathMatcherRouteRuleHeaderActionArgs {
	return &compute.URLMapPathMatcherRouteRuleHeaderActionArgs{
		RequestHeadersToAdds:     routeRuleRequestHeadersToAdd(h.RequestHeadersToAdd),
		RequestHeadersToRemoves:  stringArrayOrNil(h.RequestHeadersToRemove),
		ResponseHeadersToAdds:    routeRuleResponseHeadersToAdd(h.ResponseHeadersToAdd),
		ResponseHeadersToRemoves: stringArrayOrNil(h.ResponseHeadersToRemove),
	}
}

func routeRuleRequestHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.URLMapPathMatcherRouteRuleHeaderActionRequestHeadersToAddArray {
	result := compute.URLMapPathMatcherRouteRuleHeaderActionRequestHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.URLMapPathMatcherRouteRuleHeaderActionRequestHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func routeRuleResponseHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.URLMapPathMatcherRouteRuleHeaderActionResponseHeadersToAddArray {
	result := compute.URLMapPathMatcherRouteRuleHeaderActionResponseHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.URLMapPathMatcherRouteRuleHeaderActionResponseHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func buildMatchRules(rules []*gcpurlmapv1alpha1.GcpUrlMapRouteRuleMatchRule) compute.URLMapPathMatcherRouteRuleMatchRuleArray {
	result := compute.URLMapPathMatcherRouteRuleMatchRuleArray{}
	for _, rule := range rules {
		args := &compute.URLMapPathMatcherRouteRuleMatchRuleArgs{
			PrefixMatch:           emptyAsNilString(rule.PrefixMatch),
			FullPathMatch:         emptyAsNilString(rule.FullPathMatch),
			RegexMatch:            emptyAsNilString(rule.RegexMatch),
			PathTemplateMatch:     emptyAsNilString(rule.PathTemplateMatch),
			IgnoreCase:            pulumi.Bool(rule.IgnoreCase),
			HeaderMatches:         buildHeaderMatches(rule.HeaderMatches),
			QueryParameterMatches: buildQueryParameterMatches(rule.QueryParameterMatches),
			MetadataFilters:       buildMetadataFilters(rule.MetadataFilters),
		}
		result = append(result, args)
	}
	return result
}

func buildHeaderMatches(matches []*gcpurlmapv1alpha1.GcpUrlMapHeaderMatch) compute.URLMapPathMatcherRouteRuleMatchRuleHeaderMatchArray {
	result := compute.URLMapPathMatcherRouteRuleMatchRuleHeaderMatchArray{}
	for _, match := range matches {
		args := &compute.URLMapPathMatcherRouteRuleMatchRuleHeaderMatchArgs{
			HeaderName:   pulumi.String(match.HeaderName),
			ExactMatch:   emptyAsNilString(match.ExactMatch),
			PrefixMatch:  emptyAsNilString(match.PrefixMatch),
			SuffixMatch:  emptyAsNilString(match.SuffixMatch),
			RegexMatch:   emptyAsNilString(match.RegexMatch),
			PresentMatch: pulumi.Bool(match.PresentMatch),
			InvertMatch:  pulumi.Bool(match.InvertMatch),
		}
		if match.RangeMatch != nil {
			args.RangeMatch = &compute.URLMapPathMatcherRouteRuleMatchRuleHeaderMatchRangeMatchArgs{
				RangeStart: pulumi.Int(int(match.RangeMatch.RangeStart)),
				RangeEnd:   pulumi.Int(int(match.RangeMatch.RangeEnd)),
			}
		}
		result = append(result, args)
	}
	return result
}

func buildQueryParameterMatches(matches []*gcpurlmapv1alpha1.GcpUrlMapQueryParameterMatch) compute.URLMapPathMatcherRouteRuleMatchRuleQueryParameterMatchArray {
	result := compute.URLMapPathMatcherRouteRuleMatchRuleQueryParameterMatchArray{}
	for _, match := range matches {
		result = append(result, &compute.URLMapPathMatcherRouteRuleMatchRuleQueryParameterMatchArgs{
			Name:         pulumi.String(match.Name),
			ExactMatch:   emptyAsNilString(match.ExactMatch),
			PresentMatch: pulumi.Bool(match.PresentMatch),
			RegexMatch:   emptyAsNilString(match.RegexMatch),
		})
	}
	return result
}

func buildMetadataFilters(filters []*gcpurlmapv1alpha1.GcpUrlMapMetadataFilter) compute.URLMapPathMatcherRouteRuleMatchRuleMetadataFilterArray {
	result := compute.URLMapPathMatcherRouteRuleMatchRuleMetadataFilterArray{}
	for _, filter := range filters {
		labels := compute.URLMapPathMatcherRouteRuleMatchRuleMetadataFilterFilterLabelArray{}
		for _, label := range filter.FilterLabels {
			labels = append(labels, &compute.URLMapPathMatcherRouteRuleMatchRuleMetadataFilterFilterLabelArgs{
				Name:  pulumi.String(label.Name),
				Value: pulumi.String(label.Value),
			})
		}
		result = append(result, &compute.URLMapPathMatcherRouteRuleMatchRuleMetadataFilterArgs{
			FilterMatchCriteria: pulumi.String(filter.FilterMatchCriteria),
			FilterLabels:        labels,
		})
	}
	return result
}

func buildTests(tests []*gcpurlmapv1alpha1.GcpUrlMapTest) compute.URLMapTestArray {
	result := compute.URLMapTestArray{}
	for _, test := range tests {
		args := &compute.URLMapTestArgs{
			Host: pulumi.String(test.Host),
			Path: pulumi.String(test.Path),
		}
		if test.Service.GetValue() != "" {
			args.Service = pulumi.String(test.Service.GetValue())
		}
		if test.Description != "" {
			args.Description = pulumi.String(test.Description)
		}
		if test.ExpectedOutputUrl != "" {
			args.ExpectedOutputUrl = pulumi.String(test.ExpectedOutputUrl)
		}
		if test.ExpectedRedirectResponseCode != 0 {
			args.ExpectedRedirectResponseCode = pulumi.Int(int(test.ExpectedRedirectResponseCode))
		}
		if len(test.Headers) > 0 {
			headers := compute.URLMapTestHeaderArray{}
			for _, header := range test.Headers {
				headers = append(headers, &compute.URLMapTestHeaderArgs{
					Name:  pulumi.String(header.Name),
					Value: pulumi.String(header.Value),
				})
			}
			args.Headers = headers
		}
		result = append(result, args)
	}
	return result
}

func emptyAsNilString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.String(value)
}

// durationSeconds renders a Duration's seconds for the provider, which
// models GCP Duration seconds as a string (int64 range exceeds JSON's safe
// integer range). Returns the concrete pulumi.String so it satisfies both
// the required (StringInput) and optional (StringPtrInput) SDK positions.
func durationSeconds(d *gcpurlmapv1alpha1.GcpUrlMapDuration) pulumi.String {
	return pulumi.String(strconv.FormatInt(d.GetSeconds(), 10))
}

// durationNanos renders a Duration's sub-second component, omitted when
// zero (whole-second durations).
func durationNanos(d *gcpurlmapv1alpha1.GcpUrlMapDuration) pulumi.IntPtrInput {
	if d.GetNanos() == 0 {
		return nil
	}
	return pulumi.Int(int(d.GetNanos()))
}

func stringArrayOrNil(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	result := pulumi.StringArray{}
	for _, value := range values {
		result = append(result, pulumi.String(value))
	}
	return result
}
