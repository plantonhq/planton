package module

import (
	"strconv"

	"github.com/pkg/errors"
	gcpurlmapv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpurlmap/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// regionalUrlMap builds the regional twin (spec.region set): the routing
// brain of the regional external and regional internal Application Load
// Balancers. It routes only to regional backend services in its own region
// (never to a backend bucket), and it alone honors path_template_rewrite in
// a path matcher's default route action.
//
// The builders below mirror url_map.go's global ones minus the surfaces the
// regional API lacks — Cloud CDN route caching, custom error pages,
// stream-duration limits outside a path matcher's default action, and
// header-driven routing tests — which the spec's CEL walls keep off a
// regional manifest before either engine runs. Pulumi's nested input types
// are per-resource, so the assembly is written once per arm; the two files
// are kept in the same order so a change lands in both.
func regionalUrlMap(ctx *pulumi.Context, locals *Locals, opts []pulumi.ResourceOption) error {
	spec := locals.GcpUrlMap.Spec

	args := &compute.RegionUrlMapArgs{
		Name:   pulumi.String(locals.UrlMapName),
		Region: pulumi.String(spec.Region),
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	if spec.DefaultService.GetValue() != "" {
		args.DefaultService = pulumi.String(spec.DefaultService.GetValue())
	}
	if spec.DefaultUrlRedirect != nil {
		args.DefaultUrlRedirect = regionalTopLevelUrlRedirect(spec.DefaultUrlRedirect)
	}
	if spec.DefaultRouteAction != nil {
		args.DefaultRouteAction = regionalTopLevelRouteAction(spec.DefaultRouteAction)
	}
	if spec.HeaderAction != nil {
		args.HeaderAction = regionalTopLevelHeaderAction(spec.HeaderAction)
	}
	if len(spec.HostRules) > 0 {
		args.HostRules = regionalBuildHostRules(spec.HostRules)
	}
	if len(spec.PathMatchers) > 0 {
		args.PathMatchers = regionalBuildPathMatchers(spec.PathMatchers)
	}
	if len(spec.Tests) > 0 {
		args.Tests = regionalBuildTests(spec.Tests)
	}

	createdUrlMap, err := compute.NewRegionUrlMap(ctx, "url-map", args, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create region url map")
	}

	ctx.Export(OpSelfLink, createdUrlMap.SelfLink)
	ctx.Export(OpUrlMapName, createdUrlMap.Name)
	ctx.Export(OpMapId, createdUrlMap.MapId.ApplyT(func(id int) string {
		return strconv.Itoa(id)
	}).(pulumi.StringOutput))
	ctx.Export(OpFingerprint, createdUrlMap.Fingerprint)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return nil
}

func regionalTopLevelUrlRedirect(r *gcpurlmapv1alpha1.GcpUrlMapUrlRedirect) *compute.RegionUrlMapDefaultUrlRedirectArgs {
	return &compute.RegionUrlMapDefaultUrlRedirectArgs{
		HostRedirect:         emptyAsNilString(r.HostRedirect),
		HttpsRedirect:        pulumi.Bool(r.HttpsRedirect),
		PathRedirect:         emptyAsNilString(r.PathRedirect),
		PrefixRedirect:       emptyAsNilString(r.PrefixRedirect),
		RedirectResponseCode: emptyAsNilString(r.RedirectResponseCode),
		StripQuery:           pulumi.Bool(r.StripQuery),
	}
}

func regionalTopLevelRouteAction(a *gcpurlmapv1alpha1.GcpUrlMapRouteAction) *compute.RegionUrlMapDefaultRouteActionArgs {
	if a == nil {
		return nil
	}
	out := &compute.RegionUrlMapDefaultRouteActionArgs{}
	if len(a.WeightedBackendServices) > 0 {
		out.WeightedBackendServices = regionalDefaultWeightedBackends(a.WeightedBackendServices)
	}
	if a.UrlRewrite != nil {
		out.UrlRewrite = regionalDefaultUrlRewrite(a.UrlRewrite)
	}
	if a.Timeout != nil {
		out.Timeout = &compute.RegionUrlMapDefaultRouteActionTimeoutArgs{
			Seconds: durationSeconds(a.Timeout),
			Nanos:   durationNanos(a.Timeout),
		}
	}
	if a.RetryPolicy != nil {
		retry := &compute.RegionUrlMapDefaultRouteActionRetryPolicyArgs{
			RetryConditions: stringArrayOrNil(a.RetryPolicy.RetryConditions),
		}
		if a.RetryPolicy.NumRetries != 0 {
			retry.NumRetries = pulumi.Int(int(a.RetryPolicy.NumRetries))
		}
		if a.RetryPolicy.PerTryTimeout != nil {
			retry.PerTryTimeout = &compute.RegionUrlMapDefaultRouteActionRetryPolicyPerTryTimeoutArgs{
				Seconds: durationSeconds(a.RetryPolicy.PerTryTimeout),
				Nanos:   durationNanos(a.RetryPolicy.PerTryTimeout),
			}
		}
		out.RetryPolicy = retry
	}
	if a.RequestMirrorPolicy != nil {
		out.RequestMirrorPolicy = &compute.RegionUrlMapDefaultRouteActionRequestMirrorPolicyArgs{
			BackendService: pulumi.String(a.RequestMirrorPolicy.BackendService.GetValue()),
		}
	}
	if a.CorsPolicy != nil {
		cors := &compute.RegionUrlMapDefaultRouteActionCorsPolicyArgs{
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
		fault := &compute.RegionUrlMapDefaultRouteActionFaultInjectionPolicyArgs{}
		if a.FaultInjectionPolicy.Abort != nil {
			abort := &compute.RegionUrlMapDefaultRouteActionFaultInjectionPolicyAbortArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Abort.Percentage),
			}
			if a.FaultInjectionPolicy.Abort.HttpStatus != 0 {
				abort.HttpStatus = pulumi.Int(int(a.FaultInjectionPolicy.Abort.HttpStatus))
			}
			fault.Abort = abort
		}
		if a.FaultInjectionPolicy.Delay != nil {
			delay := &compute.RegionUrlMapDefaultRouteActionFaultInjectionPolicyDelayArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Delay.Percentage),
			}
			if a.FaultInjectionPolicy.Delay.FixedDelay != nil {
				delay.FixedDelay = &compute.RegionUrlMapDefaultRouteActionFaultInjectionPolicyDelayFixedDelayArgs{
					Seconds: durationSeconds(a.FaultInjectionPolicy.Delay.FixedDelay),
					Nanos:   durationNanos(a.FaultInjectionPolicy.Delay.FixedDelay),
				}
			}
			fault.Delay = delay
		}
		out.FaultInjectionPolicy = fault
	}
	return out
}

func regionalDefaultWeightedBackends(backends []*gcpurlmapv1alpha1.GcpUrlMapWeightedBackendService) compute.RegionUrlMapDefaultRouteActionWeightedBackendServiceArray {
	result := compute.RegionUrlMapDefaultRouteActionWeightedBackendServiceArray{}
	for _, backend := range backends {
		args := &compute.RegionUrlMapDefaultRouteActionWeightedBackendServiceArgs{
			BackendService: pulumi.String(backend.BackendService.GetValue()),
			Weight:         pulumi.Int(int(backend.Weight)),
		}
		if backend.HeaderAction != nil {
			headerAction := &compute.RegionUrlMapDefaultRouteActionWeightedBackendServiceHeaderActionArgs{
				RequestHeadersToRemoves:  stringArrayOrNil(backend.HeaderAction.RequestHeadersToRemove),
				ResponseHeadersToRemoves: stringArrayOrNil(backend.HeaderAction.ResponseHeadersToRemove),
			}
			if len(backend.HeaderAction.RequestHeadersToAdd) > 0 {
				adds := compute.RegionUrlMapDefaultRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArray{}
				for _, h := range backend.HeaderAction.RequestHeadersToAdd {
					adds = append(adds, &compute.RegionUrlMapDefaultRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.RequestHeadersToAdds = adds
			}
			if len(backend.HeaderAction.ResponseHeadersToAdd) > 0 {
				adds := compute.RegionUrlMapDefaultRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArray{}
				for _, h := range backend.HeaderAction.ResponseHeadersToAdd {
					adds = append(adds, &compute.RegionUrlMapDefaultRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArgs{
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

func regionalDefaultUrlRewrite(r *gcpurlmapv1alpha1.GcpUrlMapUrlRewrite) *compute.RegionUrlMapDefaultRouteActionUrlRewriteArgs {
	return &compute.RegionUrlMapDefaultRouteActionUrlRewriteArgs{
		HostRewrite:       emptyAsNilString(r.HostRewrite),
		PathPrefixRewrite: emptyAsNilString(r.PathPrefixRewrite),
	}
}

func regionalTopLevelHeaderAction(h *gcpurlmapv1alpha1.GcpUrlMapHeaderAction) *compute.RegionUrlMapHeaderActionArgs {
	return &compute.RegionUrlMapHeaderActionArgs{
		RequestHeadersToAdds:     regionalTopLevelRequestHeadersToAdd(h.RequestHeadersToAdd),
		RequestHeadersToRemoves:  stringArrayOrNil(h.RequestHeadersToRemove),
		ResponseHeadersToAdds:    regionalTopLevelResponseHeadersToAdd(h.ResponseHeadersToAdd),
		ResponseHeadersToRemoves: stringArrayOrNil(h.ResponseHeadersToRemove),
	}
}

func regionalTopLevelRequestHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.RegionUrlMapHeaderActionRequestHeadersToAddArray {
	result := compute.RegionUrlMapHeaderActionRequestHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.RegionUrlMapHeaderActionRequestHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func regionalTopLevelResponseHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.RegionUrlMapHeaderActionResponseHeadersToAddArray {
	result := compute.RegionUrlMapHeaderActionResponseHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.RegionUrlMapHeaderActionResponseHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func regionalBuildHostRules(rules []*gcpurlmapv1alpha1.GcpUrlMapHostRule) compute.RegionUrlMapHostRuleArray {
	result := compute.RegionUrlMapHostRuleArray{}
	for _, rule := range rules {
		hosts := pulumi.StringArray{}
		for _, host := range rule.Hosts {
			hosts = append(hosts, pulumi.String(host))
		}
		args := &compute.RegionUrlMapHostRuleArgs{
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

func regionalBuildPathMatchers(matchers []*gcpurlmapv1alpha1.GcpUrlMapPathMatcher) compute.RegionUrlMapPathMatcherArray {
	result := compute.RegionUrlMapPathMatcherArray{}
	for _, matcher := range matchers {
		args := &compute.RegionUrlMapPathMatcherArgs{
			Name: pulumi.String(matcher.Name),
		}
		if matcher.Description != "" {
			args.Description = pulumi.String(matcher.Description)
		}
		if matcher.DefaultService.GetValue() != "" {
			args.DefaultService = pulumi.String(matcher.DefaultService.GetValue())
		}
		if matcher.DefaultUrlRedirect != nil {
			args.DefaultUrlRedirect = regionalPathMatcherUrlRedirect(matcher.DefaultUrlRedirect)
		}
		if matcher.DefaultRouteAction != nil {
			args.DefaultRouteAction = regionalPathMatcherRouteAction(matcher.DefaultRouteAction)
		}
		if matcher.HeaderAction != nil {
			args.HeaderAction = regionalPathMatcherHeaderAction(matcher.HeaderAction)
		}
		if len(matcher.PathRules) > 0 {
			args.PathRules = regionalBuildPathRules(matcher.PathRules)
		}
		if len(matcher.RouteRules) > 0 {
			args.RouteRules = regionalBuildRouteRules(matcher.RouteRules)
		}
		result = append(result, args)
	}
	return result
}

func regionalPathMatcherUrlRedirect(r *gcpurlmapv1alpha1.GcpUrlMapUrlRedirect) *compute.RegionUrlMapPathMatcherDefaultUrlRedirectArgs {
	return &compute.RegionUrlMapPathMatcherDefaultUrlRedirectArgs{
		HostRedirect:         emptyAsNilString(r.HostRedirect),
		HttpsRedirect:        pulumi.Bool(r.HttpsRedirect),
		PathRedirect:         emptyAsNilString(r.PathRedirect),
		PrefixRedirect:       emptyAsNilString(r.PrefixRedirect),
		RedirectResponseCode: emptyAsNilString(r.RedirectResponseCode),
		StripQuery:           pulumi.Bool(r.StripQuery),
	}
}

func regionalPathMatcherRouteAction(a *gcpurlmapv1alpha1.GcpUrlMapRouteAction) *compute.RegionUrlMapPathMatcherDefaultRouteActionArgs {
	out := &compute.RegionUrlMapPathMatcherDefaultRouteActionArgs{}
	if len(a.WeightedBackendServices) > 0 {
		out.WeightedBackendServices = regionalPathMatcherDefaultWeightedBackends(a.WeightedBackendServices)
	}
	if a.UrlRewrite != nil {
		out.UrlRewrite = regionalPathMatcherDefaultUrlRewrite(a.UrlRewrite)
	}
	if a.Timeout != nil {
		out.Timeout = &compute.RegionUrlMapPathMatcherDefaultRouteActionTimeoutArgs{
			Seconds: durationSeconds(a.Timeout),
			Nanos:   durationNanos(a.Timeout),
		}
	}
	if a.RetryPolicy != nil {
		retry := &compute.RegionUrlMapPathMatcherDefaultRouteActionRetryPolicyArgs{
			RetryConditions: stringArrayOrNil(a.RetryPolicy.RetryConditions),
		}
		if a.RetryPolicy.NumRetries != 0 {
			retry.NumRetries = pulumi.Int(int(a.RetryPolicy.NumRetries))
		}
		if a.RetryPolicy.PerTryTimeout != nil {
			retry.PerTryTimeout = &compute.RegionUrlMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeoutArgs{
				Seconds: durationSeconds(a.RetryPolicy.PerTryTimeout),
				Nanos:   durationNanos(a.RetryPolicy.PerTryTimeout),
			}
		}
		out.RetryPolicy = retry
	}
	if a.RequestMirrorPolicy != nil {
		out.RequestMirrorPolicy = &compute.RegionUrlMapPathMatcherDefaultRouteActionRequestMirrorPolicyArgs{
			BackendService: pulumi.String(a.RequestMirrorPolicy.BackendService.GetValue()),
		}
	}
	if a.CorsPolicy != nil {
		cors := &compute.RegionUrlMapPathMatcherDefaultRouteActionCorsPolicyArgs{
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
		fault := &compute.RegionUrlMapPathMatcherDefaultRouteActionFaultInjectionPolicyArgs{}
		if a.FaultInjectionPolicy.Abort != nil {
			abort := &compute.RegionUrlMapPathMatcherDefaultRouteActionFaultInjectionPolicyAbortArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Abort.Percentage),
			}
			if a.FaultInjectionPolicy.Abort.HttpStatus != 0 {
				abort.HttpStatus = pulumi.Int(int(a.FaultInjectionPolicy.Abort.HttpStatus))
			}
			fault.Abort = abort
		}
		if a.FaultInjectionPolicy.Delay != nil {
			delay := &compute.RegionUrlMapPathMatcherDefaultRouteActionFaultInjectionPolicyDelayArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Delay.Percentage),
			}
			if a.FaultInjectionPolicy.Delay.FixedDelay != nil {
				delay.FixedDelay = &compute.RegionUrlMapPathMatcherDefaultRouteActionFaultInjectionPolicyDelayFixedDelayArgs{
					Seconds: durationSeconds(a.FaultInjectionPolicy.Delay.FixedDelay),
					Nanos:   durationNanos(a.FaultInjectionPolicy.Delay.FixedDelay),
				}
			}
			fault.Delay = delay
		}
		out.FaultInjectionPolicy = fault
	}
	if a.MaxStreamDuration != nil {
		out.MaxStreamDuration = &compute.RegionUrlMapPathMatcherDefaultRouteActionMaxStreamDurationArgs{
			Seconds: durationSeconds(a.MaxStreamDuration),
			Nanos:   durationNanos(a.MaxStreamDuration),
		}
	}
	return out
}

func regionalPathMatcherDefaultWeightedBackends(backends []*gcpurlmapv1alpha1.GcpUrlMapWeightedBackendService) compute.RegionUrlMapPathMatcherDefaultRouteActionWeightedBackendServiceArray {
	result := compute.RegionUrlMapPathMatcherDefaultRouteActionWeightedBackendServiceArray{}
	for _, backend := range backends {
		args := &compute.RegionUrlMapPathMatcherDefaultRouteActionWeightedBackendServiceArgs{
			BackendService: pulumi.String(backend.BackendService.GetValue()),
			Weight:         pulumi.Int(int(backend.Weight)),
		}
		if backend.HeaderAction != nil {
			headerAction := &compute.RegionUrlMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionArgs{
				RequestHeadersToRemoves:  stringArrayOrNil(backend.HeaderAction.RequestHeadersToRemove),
				ResponseHeadersToRemoves: stringArrayOrNil(backend.HeaderAction.ResponseHeadersToRemove),
			}
			if len(backend.HeaderAction.RequestHeadersToAdd) > 0 {
				adds := compute.RegionUrlMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArray{}
				for _, h := range backend.HeaderAction.RequestHeadersToAdd {
					adds = append(adds, &compute.RegionUrlMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.RequestHeadersToAdds = adds
			}
			if len(backend.HeaderAction.ResponseHeadersToAdd) > 0 {
				adds := compute.RegionUrlMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArray{}
				for _, h := range backend.HeaderAction.ResponseHeadersToAdd {
					adds = append(adds, &compute.RegionUrlMapPathMatcherDefaultRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArgs{
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

// The regional map alone honors path_template_rewrite in a path matcher's
// default route action (the spec CEL rejects it on a global map).
func regionalPathMatcherDefaultUrlRewrite(r *gcpurlmapv1alpha1.GcpUrlMapUrlRewrite) *compute.RegionUrlMapPathMatcherDefaultRouteActionUrlRewriteArgs {
	return &compute.RegionUrlMapPathMatcherDefaultRouteActionUrlRewriteArgs{
		HostRewrite:         emptyAsNilString(r.HostRewrite),
		PathPrefixRewrite:   emptyAsNilString(r.PathPrefixRewrite),
		PathTemplateRewrite: emptyAsNilString(r.PathTemplateRewrite),
	}
}

func regionalPathMatcherHeaderAction(h *gcpurlmapv1alpha1.GcpUrlMapHeaderAction) *compute.RegionUrlMapPathMatcherHeaderActionArgs {
	return &compute.RegionUrlMapPathMatcherHeaderActionArgs{
		RequestHeadersToAdds:     regionalPathMatcherRequestHeadersToAdd(h.RequestHeadersToAdd),
		RequestHeadersToRemoves:  stringArrayOrNil(h.RequestHeadersToRemove),
		ResponseHeadersToAdds:    regionalPathMatcherResponseHeadersToAdd(h.ResponseHeadersToAdd),
		ResponseHeadersToRemoves: stringArrayOrNil(h.ResponseHeadersToRemove),
	}
}

func regionalPathMatcherRequestHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.RegionUrlMapPathMatcherHeaderActionRequestHeadersToAddArray {
	result := compute.RegionUrlMapPathMatcherHeaderActionRequestHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.RegionUrlMapPathMatcherHeaderActionRequestHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func regionalPathMatcherResponseHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.RegionUrlMapPathMatcherHeaderActionResponseHeadersToAddArray {
	result := compute.RegionUrlMapPathMatcherHeaderActionResponseHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.RegionUrlMapPathMatcherHeaderActionResponseHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func regionalBuildPathRules(rules []*gcpurlmapv1alpha1.GcpUrlMapPathRule) compute.RegionUrlMapPathMatcherPathRuleArray {
	result := compute.RegionUrlMapPathMatcherPathRuleArray{}
	for _, rule := range rules {
		paths := pulumi.StringArray{}
		for _, path := range rule.Paths {
			paths = append(paths, pulumi.String(path))
		}
		args := &compute.RegionUrlMapPathMatcherPathRuleArgs{
			Paths: paths,
		}
		if rule.Service.GetValue() != "" {
			args.Service = pulumi.String(rule.Service.GetValue())
		}
		if rule.UrlRedirect != nil {
			args.UrlRedirect = regionalPathRuleUrlRedirect(rule.UrlRedirect)
		}
		if rule.RouteAction != nil {
			args.RouteAction = regionalPathRuleRouteAction(rule.RouteAction)
		}
		result = append(result, args)
	}
	return result
}

func regionalPathRuleUrlRedirect(r *gcpurlmapv1alpha1.GcpUrlMapUrlRedirect) *compute.RegionUrlMapPathMatcherPathRuleUrlRedirectArgs {
	return &compute.RegionUrlMapPathMatcherPathRuleUrlRedirectArgs{
		HostRedirect:         emptyAsNilString(r.HostRedirect),
		HttpsRedirect:        pulumi.Bool(r.HttpsRedirect),
		PathRedirect:         emptyAsNilString(r.PathRedirect),
		PrefixRedirect:       emptyAsNilString(r.PrefixRedirect),
		RedirectResponseCode: emptyAsNilString(r.RedirectResponseCode),
		StripQuery:           pulumi.Bool(r.StripQuery),
	}
}

func regionalPathRuleRouteAction(a *gcpurlmapv1alpha1.GcpUrlMapRouteAction) *compute.RegionUrlMapPathMatcherPathRuleRouteActionArgs {
	out := &compute.RegionUrlMapPathMatcherPathRuleRouteActionArgs{}
	if len(a.WeightedBackendServices) > 0 {
		out.WeightedBackendServices = regionalPathRuleWeightedBackends(a.WeightedBackendServices)
	}
	if a.UrlRewrite != nil {
		out.UrlRewrite = regionalPathRuleUrlRewrite(a.UrlRewrite)
	}
	if a.Timeout != nil {
		out.Timeout = &compute.RegionUrlMapPathMatcherPathRuleRouteActionTimeoutArgs{
			Seconds: durationSeconds(a.Timeout),
			Nanos:   durationNanos(a.Timeout),
		}
	}
	if a.RetryPolicy != nil {
		retry := &compute.RegionUrlMapPathMatcherPathRuleRouteActionRetryPolicyArgs{
			RetryConditions: stringArrayOrNil(a.RetryPolicy.RetryConditions),
		}
		if a.RetryPolicy.NumRetries != 0 {
			retry.NumRetries = pulumi.Int(int(a.RetryPolicy.NumRetries))
		}
		if a.RetryPolicy.PerTryTimeout != nil {
			retry.PerTryTimeout = &compute.RegionUrlMapPathMatcherPathRuleRouteActionRetryPolicyPerTryTimeoutArgs{
				Seconds: durationSeconds(a.RetryPolicy.PerTryTimeout),
				Nanos:   durationNanos(a.RetryPolicy.PerTryTimeout),
			}
		}
		out.RetryPolicy = retry
	}
	if a.RequestMirrorPolicy != nil {
		out.RequestMirrorPolicy = &compute.RegionUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicyArgs{
			BackendService: pulumi.String(a.RequestMirrorPolicy.BackendService.GetValue()),
		}
	}
	if a.CorsPolicy != nil {
		cors := &compute.RegionUrlMapPathMatcherPathRuleRouteActionCorsPolicyArgs{
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
		fault := &compute.RegionUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicyArgs{}
		if a.FaultInjectionPolicy.Abort != nil {
			abort := &compute.RegionUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicyAbortArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Abort.Percentage),
			}
			if a.FaultInjectionPolicy.Abort.HttpStatus != 0 {
				abort.HttpStatus = pulumi.Int(int(a.FaultInjectionPolicy.Abort.HttpStatus))
			}
			fault.Abort = abort
		}
		if a.FaultInjectionPolicy.Delay != nil {
			delay := &compute.RegionUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicyDelayArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Delay.Percentage),
			}
			if a.FaultInjectionPolicy.Delay.FixedDelay != nil {
				delay.FixedDelay = &compute.RegionUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicyDelayFixedDelayArgs{
					Seconds: durationSeconds(a.FaultInjectionPolicy.Delay.FixedDelay),
					Nanos:   durationNanos(a.FaultInjectionPolicy.Delay.FixedDelay),
				}
			}
			fault.Delay = delay
		}
		out.FaultInjectionPolicy = fault
	}
	return out
}

func regionalPathRuleWeightedBackends(backends []*gcpurlmapv1alpha1.GcpUrlMapWeightedBackendService) compute.RegionUrlMapPathMatcherPathRuleRouteActionWeightedBackendServiceArray {
	result := compute.RegionUrlMapPathMatcherPathRuleRouteActionWeightedBackendServiceArray{}
	for _, backend := range backends {
		args := &compute.RegionUrlMapPathMatcherPathRuleRouteActionWeightedBackendServiceArgs{
			BackendService: pulumi.String(backend.BackendService.GetValue()),
			Weight:         pulumi.Int(int(backend.Weight)),
		}
		if backend.HeaderAction != nil {
			headerAction := &compute.RegionUrlMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionArgs{
				RequestHeadersToRemoves:  stringArrayOrNil(backend.HeaderAction.RequestHeadersToRemove),
				ResponseHeadersToRemoves: stringArrayOrNil(backend.HeaderAction.ResponseHeadersToRemove),
			}
			if len(backend.HeaderAction.RequestHeadersToAdd) > 0 {
				adds := compute.RegionUrlMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArray{}
				for _, h := range backend.HeaderAction.RequestHeadersToAdd {
					adds = append(adds, &compute.RegionUrlMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.RequestHeadersToAdds = adds
			}
			if len(backend.HeaderAction.ResponseHeadersToAdd) > 0 {
				adds := compute.RegionUrlMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArray{}
				for _, h := range backend.HeaderAction.ResponseHeadersToAdd {
					adds = append(adds, &compute.RegionUrlMapPathMatcherPathRuleRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArgs{
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

func regionalPathRuleUrlRewrite(r *gcpurlmapv1alpha1.GcpUrlMapUrlRewrite) *compute.RegionUrlMapPathMatcherPathRuleRouteActionUrlRewriteArgs {
	return &compute.RegionUrlMapPathMatcherPathRuleRouteActionUrlRewriteArgs{
		HostRewrite:       emptyAsNilString(r.HostRewrite),
		PathPrefixRewrite: emptyAsNilString(r.PathPrefixRewrite),
	}
}

func regionalBuildRouteRules(rules []*gcpurlmapv1alpha1.GcpUrlMapRouteRule) compute.RegionUrlMapPathMatcherRouteRuleArray {
	result := compute.RegionUrlMapPathMatcherRouteRuleArray{}
	for _, rule := range rules {
		args := &compute.RegionUrlMapPathMatcherRouteRuleArgs{
			Priority:   pulumi.Int(int(rule.Priority)),
			MatchRules: regionalBuildMatchRules(rule.MatchRules),
		}
		if rule.Service.GetValue() != "" {
			args.Service = pulumi.String(rule.Service.GetValue())
		}
		if rule.UrlRedirect != nil {
			args.UrlRedirect = regionalRouteRuleUrlRedirect(rule.UrlRedirect)
		}
		if rule.RouteAction != nil {
			args.RouteAction = regionalRouteRuleRouteAction(rule.RouteAction)
		}
		if rule.HeaderAction != nil {
			args.HeaderAction = regionalRouteRuleHeaderAction(rule.HeaderAction)
		}
		result = append(result, args)
	}
	return result
}

func regionalRouteRuleUrlRedirect(r *gcpurlmapv1alpha1.GcpUrlMapUrlRedirect) *compute.RegionUrlMapPathMatcherRouteRuleUrlRedirectArgs {
	return &compute.RegionUrlMapPathMatcherRouteRuleUrlRedirectArgs{
		HostRedirect:         emptyAsNilString(r.HostRedirect),
		HttpsRedirect:        pulumi.Bool(r.HttpsRedirect),
		PathRedirect:         emptyAsNilString(r.PathRedirect),
		PrefixRedirect:       emptyAsNilString(r.PrefixRedirect),
		RedirectResponseCode: emptyAsNilString(r.RedirectResponseCode),
		StripQuery:           pulumi.Bool(r.StripQuery),
	}
}

func regionalRouteRuleRouteAction(a *gcpurlmapv1alpha1.GcpUrlMapRouteAction) *compute.RegionUrlMapPathMatcherRouteRuleRouteActionArgs {
	out := &compute.RegionUrlMapPathMatcherRouteRuleRouteActionArgs{}
	if len(a.WeightedBackendServices) > 0 {
		out.WeightedBackendServices = regionalRouteRuleWeightedBackends(a.WeightedBackendServices)
	}
	if a.UrlRewrite != nil {
		out.UrlRewrite = regionalRouteRuleUrlRewrite(a.UrlRewrite)
	}
	if a.Timeout != nil {
		out.Timeout = &compute.RegionUrlMapPathMatcherRouteRuleRouteActionTimeoutArgs{
			Seconds: durationSeconds(a.Timeout),
			Nanos:   durationNanos(a.Timeout),
		}
	}
	if a.RetryPolicy != nil {
		retry := &compute.RegionUrlMapPathMatcherRouteRuleRouteActionRetryPolicyArgs{
			RetryConditions: stringArrayOrNil(a.RetryPolicy.RetryConditions),
		}
		if a.RetryPolicy.NumRetries != 0 {
			retry.NumRetries = pulumi.Int(int(a.RetryPolicy.NumRetries))
		}
		if a.RetryPolicy.PerTryTimeout != nil {
			retry.PerTryTimeout = &compute.RegionUrlMapPathMatcherRouteRuleRouteActionRetryPolicyPerTryTimeoutArgs{
				Seconds: durationSeconds(a.RetryPolicy.PerTryTimeout),
				Nanos:   durationNanos(a.RetryPolicy.PerTryTimeout),
			}
		}
		out.RetryPolicy = retry
	}
	if a.RequestMirrorPolicy != nil {
		out.RequestMirrorPolicy = &compute.RegionUrlMapPathMatcherRouteRuleRouteActionRequestMirrorPolicyArgs{
			BackendService: pulumi.String(a.RequestMirrorPolicy.BackendService.GetValue()),
		}
	}
	if a.CorsPolicy != nil {
		cors := &compute.RegionUrlMapPathMatcherRouteRuleRouteActionCorsPolicyArgs{
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
		fault := &compute.RegionUrlMapPathMatcherRouteRuleRouteActionFaultInjectionPolicyArgs{}
		if a.FaultInjectionPolicy.Abort != nil {
			abort := &compute.RegionUrlMapPathMatcherRouteRuleRouteActionFaultInjectionPolicyAbortArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Abort.Percentage),
			}
			if a.FaultInjectionPolicy.Abort.HttpStatus != 0 {
				abort.HttpStatus = pulumi.Int(int(a.FaultInjectionPolicy.Abort.HttpStatus))
			}
			fault.Abort = abort
		}
		if a.FaultInjectionPolicy.Delay != nil {
			delay := &compute.RegionUrlMapPathMatcherRouteRuleRouteActionFaultInjectionPolicyDelayArgs{
				Percentage: pulumi.Float64(a.FaultInjectionPolicy.Delay.Percentage),
			}
			if a.FaultInjectionPolicy.Delay.FixedDelay != nil {
				delay.FixedDelay = &compute.RegionUrlMapPathMatcherRouteRuleRouteActionFaultInjectionPolicyDelayFixedDelayArgs{
					Seconds: durationSeconds(a.FaultInjectionPolicy.Delay.FixedDelay),
					Nanos:   durationNanos(a.FaultInjectionPolicy.Delay.FixedDelay),
				}
			}
			fault.Delay = delay
		}
		out.FaultInjectionPolicy = fault
	}
	return out
}

func regionalRouteRuleWeightedBackends(backends []*gcpurlmapv1alpha1.GcpUrlMapWeightedBackendService) compute.RegionUrlMapPathMatcherRouteRuleRouteActionWeightedBackendServiceArray {
	result := compute.RegionUrlMapPathMatcherRouteRuleRouteActionWeightedBackendServiceArray{}
	for _, backend := range backends {
		args := &compute.RegionUrlMapPathMatcherRouteRuleRouteActionWeightedBackendServiceArgs{
			BackendService: pulumi.String(backend.BackendService.GetValue()),
			Weight:         pulumi.Int(int(backend.Weight)),
		}
		if backend.HeaderAction != nil {
			headerAction := &compute.RegionUrlMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionArgs{
				RequestHeadersToRemoves:  stringArrayOrNil(backend.HeaderAction.RequestHeadersToRemove),
				ResponseHeadersToRemoves: stringArrayOrNil(backend.HeaderAction.ResponseHeadersToRemove),
			}
			if len(backend.HeaderAction.RequestHeadersToAdd) > 0 {
				adds := compute.RegionUrlMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArray{}
				for _, h := range backend.HeaderAction.RequestHeadersToAdd {
					adds = append(adds, &compute.RegionUrlMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionRequestHeadersToAddArgs{
						HeaderName:  pulumi.String(h.HeaderName),
						HeaderValue: pulumi.String(h.HeaderValue),
						Replace:     pulumi.Bool(h.Replace),
					})
				}
				headerAction.RequestHeadersToAdds = adds
			}
			if len(backend.HeaderAction.ResponseHeadersToAdd) > 0 {
				adds := compute.RegionUrlMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArray{}
				for _, h := range backend.HeaderAction.ResponseHeadersToAdd {
					adds = append(adds, &compute.RegionUrlMapPathMatcherRouteRuleRouteActionWeightedBackendServiceHeaderActionResponseHeadersToAddArgs{
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

func regionalRouteRuleUrlRewrite(r *gcpurlmapv1alpha1.GcpUrlMapUrlRewrite) *compute.RegionUrlMapPathMatcherRouteRuleRouteActionUrlRewriteArgs {
	return &compute.RegionUrlMapPathMatcherRouteRuleRouteActionUrlRewriteArgs{
		HostRewrite:         emptyAsNilString(r.HostRewrite),
		PathPrefixRewrite:   emptyAsNilString(r.PathPrefixRewrite),
		PathTemplateRewrite: emptyAsNilString(r.PathTemplateRewrite),
	}
}

func regionalRouteRuleHeaderAction(h *gcpurlmapv1alpha1.GcpUrlMapHeaderAction) *compute.RegionUrlMapPathMatcherRouteRuleHeaderActionArgs {
	return &compute.RegionUrlMapPathMatcherRouteRuleHeaderActionArgs{
		RequestHeadersToAdds:     regionalRouteRuleRequestHeadersToAdd(h.RequestHeadersToAdd),
		RequestHeadersToRemoves:  stringArrayOrNil(h.RequestHeadersToRemove),
		ResponseHeadersToAdds:    regionalRouteRuleResponseHeadersToAdd(h.ResponseHeadersToAdd),
		ResponseHeadersToRemoves: stringArrayOrNil(h.ResponseHeadersToRemove),
	}
}

func regionalRouteRuleRequestHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.RegionUrlMapPathMatcherRouteRuleHeaderActionRequestHeadersToAddArray {
	result := compute.RegionUrlMapPathMatcherRouteRuleHeaderActionRequestHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.RegionUrlMapPathMatcherRouteRuleHeaderActionRequestHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func regionalRouteRuleResponseHeadersToAdd(headers []*gcpurlmapv1alpha1.GcpUrlMapHeaderValue) compute.RegionUrlMapPathMatcherRouteRuleHeaderActionResponseHeadersToAddArray {
	result := compute.RegionUrlMapPathMatcherRouteRuleHeaderActionResponseHeadersToAddArray{}
	for _, header := range headers {
		result = append(result, &compute.RegionUrlMapPathMatcherRouteRuleHeaderActionResponseHeadersToAddArgs{
			HeaderName:  pulumi.String(header.HeaderName),
			HeaderValue: pulumi.String(header.HeaderValue),
			Replace:     pulumi.Bool(header.Replace),
		})
	}
	return result
}

func regionalBuildMatchRules(rules []*gcpurlmapv1alpha1.GcpUrlMapRouteRuleMatchRule) compute.RegionUrlMapPathMatcherRouteRuleMatchRuleArray {
	result := compute.RegionUrlMapPathMatcherRouteRuleMatchRuleArray{}
	for _, rule := range rules {
		args := &compute.RegionUrlMapPathMatcherRouteRuleMatchRuleArgs{
			PrefixMatch:           emptyAsNilString(rule.PrefixMatch),
			FullPathMatch:         emptyAsNilString(rule.FullPathMatch),
			RegexMatch:            emptyAsNilString(rule.RegexMatch),
			PathTemplateMatch:     emptyAsNilString(rule.PathTemplateMatch),
			IgnoreCase:            pulumi.Bool(rule.IgnoreCase),
			HeaderMatches:         regionalBuildHeaderMatches(rule.HeaderMatches),
			QueryParameterMatches: regionalBuildQueryParameterMatches(rule.QueryParameterMatches),
			MetadataFilters:       regionalBuildMetadataFilters(rule.MetadataFilters),
		}
		result = append(result, args)
	}
	return result
}

func regionalBuildHeaderMatches(matches []*gcpurlmapv1alpha1.GcpUrlMapHeaderMatch) compute.RegionUrlMapPathMatcherRouteRuleMatchRuleHeaderMatchArray {
	result := compute.RegionUrlMapPathMatcherRouteRuleMatchRuleHeaderMatchArray{}
	for _, match := range matches {
		args := &compute.RegionUrlMapPathMatcherRouteRuleMatchRuleHeaderMatchArgs{
			HeaderName:   pulumi.String(match.HeaderName),
			ExactMatch:   emptyAsNilString(match.ExactMatch),
			PrefixMatch:  emptyAsNilString(match.PrefixMatch),
			SuffixMatch:  emptyAsNilString(match.SuffixMatch),
			RegexMatch:   emptyAsNilString(match.RegexMatch),
			PresentMatch: pulumi.Bool(match.PresentMatch),
			InvertMatch:  pulumi.Bool(match.InvertMatch),
		}
		if match.RangeMatch != nil {
			args.RangeMatch = &compute.RegionUrlMapPathMatcherRouteRuleMatchRuleHeaderMatchRangeMatchArgs{
				RangeStart: pulumi.Int(int(match.RangeMatch.RangeStart)),
				RangeEnd:   pulumi.Int(int(match.RangeMatch.RangeEnd)),
			}
		}
		result = append(result, args)
	}
	return result
}

func regionalBuildQueryParameterMatches(matches []*gcpurlmapv1alpha1.GcpUrlMapQueryParameterMatch) compute.RegionUrlMapPathMatcherRouteRuleMatchRuleQueryParameterMatchArray {
	result := compute.RegionUrlMapPathMatcherRouteRuleMatchRuleQueryParameterMatchArray{}
	for _, match := range matches {
		result = append(result, &compute.RegionUrlMapPathMatcherRouteRuleMatchRuleQueryParameterMatchArgs{
			Name:         pulumi.String(match.Name),
			ExactMatch:   emptyAsNilString(match.ExactMatch),
			PresentMatch: pulumi.Bool(match.PresentMatch),
			RegexMatch:   emptyAsNilString(match.RegexMatch),
		})
	}
	return result
}

func regionalBuildMetadataFilters(filters []*gcpurlmapv1alpha1.GcpUrlMapMetadataFilter) compute.RegionUrlMapPathMatcherRouteRuleMatchRuleMetadataFilterArray {
	result := compute.RegionUrlMapPathMatcherRouteRuleMatchRuleMetadataFilterArray{}
	for _, filter := range filters {
		labels := compute.RegionUrlMapPathMatcherRouteRuleMatchRuleMetadataFilterFilterLabelArray{}
		for _, label := range filter.FilterLabels {
			labels = append(labels, &compute.RegionUrlMapPathMatcherRouteRuleMatchRuleMetadataFilterFilterLabelArgs{
				Name:  pulumi.String(label.Name),
				Value: pulumi.String(label.Value),
			})
		}
		result = append(result, &compute.RegionUrlMapPathMatcherRouteRuleMatchRuleMetadataFilterArgs{
			FilterMatchCriteria: pulumi.String(filter.FilterMatchCriteria),
			FilterLabels:        labels,
		})
	}
	return result
}

func regionalBuildTests(tests []*gcpurlmapv1alpha1.GcpUrlMapTest) compute.RegionUrlMapTestArray {
	result := compute.RegionUrlMapTestArray{}
	for _, test := range tests {
		args := &compute.RegionUrlMapTestArgs{
			Host: pulumi.String(test.Host),
			Path: pulumi.String(test.Path),
		}
		if test.Service.GetValue() != "" {
			args.Service = pulumi.String(test.Service.GetValue())
		}
		if test.Description != "" {
			args.Description = pulumi.String(test.Description)
		}
		result = append(result, args)
	}
	return result
}
