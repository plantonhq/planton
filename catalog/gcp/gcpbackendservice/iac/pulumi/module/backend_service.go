package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// backendService provisions the global Compute Engine backend service — the
// hub of the L7 load balancing family — plus its signed-URL keys.
//
// name and project are immutable (ForceNew in the provider): changing either
// destroys and recreates the backend service, briefly breaking every URL map
// referencing the old self_link. Everything else — backends, CDN policy,
// affinity, IAP — updates in place, which is what makes this node the
// operational lever of a running load balancer.
//
// Two provider subtleties worth knowing when comparing previews to API
// calls: the provider applies security_policy and edge_security_policy via
// dedicated setSecurityPolicy/setEdgeSecurityPolicy sub-calls (not the main
// insert/patch body), and it strips max_utilization from any NEG backend
// because the API rejects it there. Neither changes declared state — both
// engines behave identically.
//
// Cross-field applicability (CDN only on external schemes, circuit breakers /
// max_stream_duration only on INTERNAL_SELF_MANAGED, consistent_hash only
// with MAGLEV/RING_HASH, cache-mode/TTL coherence) is enforced by the spec's
// CEL rules before deploy, so no defensive logic lives here.
func backendService(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpBackendService.Spec

	// Enable the Compute Engine API so a fresh project can host the backend
	// service. disable_on_destroy stays false (the provider default): tearing
	// down one backend service must never disable the API for everything else
	// in the project. Matches the Terraform module.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"backendservice-compute.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	args := &compute.BackendServiceArgs{
		Name:      pulumi.String(locals.BackendServiceName),
		EnableCdn: pulumi.Bool(spec.EnableCdn),
	}

	// Honor the spec contract: an empty project_id falls back to the provider's
	// default project. Leaving Project unset lets the gcp provider resolve its
	// own project (configuration or the GOOGLE_PROJECT / GOOGLE_CLOUD_PROJECT
	// environment chain); an empty string would be sent verbatim and rejected.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// Omitted optionals stay unset (matching the Terraform module's null)
	// rather than being sent as empty strings the API would reject or
	// misread. Each of these has a GCP API default that matches the spec's
	// proto default (protocol HTTP, affinity NONE).
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.GetProtocol() != "" {
		args.Protocol = pulumi.String(spec.GetProtocol())
	}
	// The scheme is the ONE exception: the spec's default is EXTERNAL (the
	// classic global external ALB), but the provider's own default is
	// EXTERNAL_MANAGED, and the scheme is immutable -- letting the provider
	// decide would replace every existing classic backend service the next
	// time it was applied. The module therefore sends EXTERNAL itself when
	// the spec leaves the scheme empty (the manifest defaults applier
	// normally fills it first; this guard covers every path that bypasses
	// the applier). The Terraform module makes the same choice.
	if spec.GetLoadBalancingScheme() != "" {
		args.LoadBalancingScheme = pulumi.String(spec.GetLoadBalancingScheme())
	} else {
		args.LoadBalancingScheme = pulumi.String("EXTERNAL")
	}
	if spec.PortName != "" {
		args.PortName = pulumi.String(spec.PortName)
	}
	if spec.GetTimeoutSec() != 0 {
		args.TimeoutSec = pulumi.Int(int(spec.GetTimeoutSec()))
	}
	if spec.GetConnectionDrainingTimeoutSec() != 0 {
		args.ConnectionDrainingTimeoutSec = pulumi.Int(int(spec.GetConnectionDrainingTimeoutSec()))
	}
	if spec.GetSessionAffinity() != "" {
		args.SessionAffinity = pulumi.String(spec.GetSessionAffinity())
	}
	if spec.AffinityCookieTtlSec != 0 {
		args.AffinityCookieTtlSec = pulumi.Int(int(spec.AffinityCookieTtlSec))
	}
	if spec.LocalityLbPolicy != "" {
		args.LocalityLbPolicy = pulumi.String(spec.LocalityLbPolicy)
	}
	if spec.CompressionMode != "" {
		args.CompressionMode = pulumi.String(spec.CompressionMode)
	}
	if spec.IpAddressSelectionPolicy != "" {
		args.IpAddressSelectionPolicy = pulumi.String(spec.IpAddressSelectionPolicy)
	}
	if spec.ServiceLbPolicy != "" {
		args.ServiceLbPolicy = pulumi.String(spec.ServiceLbPolicy)
	}
	if spec.ExternalManagedMigrationState != "" {
		args.ExternalManagedMigrationState = pulumi.String(spec.ExternalManagedMigrationState)
	}
	if spec.ExternalManagedMigrationTestingPercentage != 0 {
		args.ExternalManagedMigrationTestingPercentage = pulumi.Float64(spec.ExternalManagedMigrationTestingPercentage)
	}

	// At most one health check (a GCP-enforced cap — the spec models it
	// singular; the SDK flattens the one-element set to a plain string).
	// None at all is valid only for serverless/internet NEG backends.
	if spec.HealthCheck.GetValue() != "" {
		args.HealthChecks = pulumi.String(spec.HealthCheck.GetValue())
	}

	// Attach by reference — the Cloud Armor policies are their own composable
	// nodes (GcpCloudArmorPolicy), never embedded here. security_policy
	// filters after the CDN cache; edge_security_policy before it.
	if spec.SecurityPolicy.GetValue() != "" {
		args.SecurityPolicy = pulumi.String(spec.SecurityPolicy.GetValue())
	}
	if spec.EdgeSecurityPolicy.GetValue() != "" {
		args.EdgeSecurityPolicy = pulumi.String(spec.EdgeSecurityPolicy.GetValue())
	}

	if len(spec.CustomRequestHeaders) > 0 {
		customRequestHeaders := pulumi.StringArray{}
		for _, customRequestHeader := range spec.CustomRequestHeaders {
			customRequestHeaders = append(customRequestHeaders, pulumi.String(customRequestHeader))
		}
		args.CustomRequestHeaders = customRequestHeaders
	}
	if len(spec.CustomResponseHeaders) > 0 {
		customResponseHeaders := pulumi.StringArray{}
		for _, customResponseHeader := range spec.CustomResponseHeaders {
			customResponseHeaders = append(customResponseHeaders, pulumi.String(customResponseHeader))
		}
		args.CustomResponseHeaders = customResponseHeaders
	}

	// The backends serving traffic. Instance groups and NEGs cannot be mixed
	// on one service (GCP rejects it); which max_* dial is required is
	// decided by each backend's balancing_mode and enforced by the spec
	// pre-deploy.
	if len(spec.Backends) > 0 {
		backends := compute.BackendServiceBackendArray{}
		for _, backend := range spec.Backends {
			backendArgs := &compute.BackendServiceBackendArgs{
				Group: pulumi.String(backend.Group.GetValue()),
			}
			if backend.GetBalancingMode() != "" {
				backendArgs.BalancingMode = pulumi.String(backend.GetBalancingMode())
			}
			// capacity_scaler passes through whenever present: it is
			// `optional` in the proto, so nil means unset (API default 1.0)
			// and an explicit 0 is the drain-this-backend semantics and
			// must survive.
			if backend.CapacityScaler != nil {
				backendArgs.CapacityScaler = pulumi.Float64(backend.GetCapacityScaler())
			}
			if backend.Description != "" {
				backendArgs.Description = pulumi.String(backend.Description)
			}
			if backend.MaxConnections != 0 {
				backendArgs.MaxConnections = pulumi.Int(int(backend.MaxConnections))
			}
			if backend.MaxConnectionsPerInstance != 0 {
				backendArgs.MaxConnectionsPerInstance = pulumi.Int(int(backend.MaxConnectionsPerInstance))
			}
			if backend.MaxConnectionsPerEndpoint != 0 {
				backendArgs.MaxConnectionsPerEndpoint = pulumi.Int(int(backend.MaxConnectionsPerEndpoint))
			}
			if backend.MaxRate != 0 {
				backendArgs.MaxRate = pulumi.Int(int(backend.MaxRate))
			}
			if backend.MaxRatePerInstance != 0 {
				backendArgs.MaxRatePerInstance = pulumi.Float64(backend.MaxRatePerInstance)
			}
			if backend.MaxRatePerEndpoint != 0 {
				backendArgs.MaxRatePerEndpoint = pulumi.Float64(backend.MaxRatePerEndpoint)
			}
			if backend.MaxUtilization != 0 {
				backendArgs.MaxUtilization = pulumi.Float64(backend.MaxUtilization)
			}
			if backend.Preference != "" {
				backendArgs.Preference = pulumi.String(backend.Preference)
			}
			if len(backend.CustomMetrics) > 0 {
				backendCustomMetrics := compute.BackendServiceBackendCustomMetricArray{}
				for _, backendCustomMetric := range backend.CustomMetrics {
					backendCustomMetricArgs := &compute.BackendServiceBackendCustomMetricArgs{
						Name:   pulumi.String(backendCustomMetric.Name),
						DryRun: pulumi.Bool(backendCustomMetric.DryRun),
					}
					if backendCustomMetric.MaxUtilization != nil {
						backendCustomMetricArgs.MaxUtilization = pulumi.Float64(backendCustomMetric.GetMaxUtilization())
					}
					backendCustomMetrics = append(backendCustomMetrics, backendCustomMetricArgs)
				}
				backendArgs.CustomMetrics = backendCustomMetrics
			}
			backends = append(backends, backendArgs)
		}
		args.Backends = backends
	}

	if spec.CdnPolicy != nil {
		cdnPolicy := &compute.BackendServiceCdnPolicyArgs{}

		if spec.CdnPolicy.CacheMode != "" {
			cdnPolicy.CacheMode = pulumi.String(spec.CdnPolicy.CacheMode)
		}
		// The tfvars/proto contract treats 0 as unset for TTLs, letting the
		// GCP API apply its own defaults — identical to the Terraform module.
		if spec.CdnPolicy.ClientTtl != 0 {
			cdnPolicy.ClientTtl = pulumi.Int(int(spec.CdnPolicy.ClientTtl))
		}
		if spec.CdnPolicy.DefaultTtl != 0 {
			cdnPolicy.DefaultTtl = pulumi.Int(int(spec.CdnPolicy.DefaultTtl))
		}
		if spec.CdnPolicy.MaxTtl != 0 {
			cdnPolicy.MaxTtl = pulumi.Int(int(spec.CdnPolicy.MaxTtl))
		}
		if spec.CdnPolicy.NegativeCaching {
			cdnPolicy.NegativeCaching = pulumi.Bool(true)
		}
		if spec.CdnPolicy.ServeWhileStale != 0 {
			cdnPolicy.ServeWhileStale = pulumi.Int(int(spec.CdnPolicy.ServeWhileStale))
		}
		if spec.CdnPolicy.RequestCoalescing {
			cdnPolicy.RequestCoalescing = pulumi.Bool(true)
		}
		if spec.CdnPolicy.SignedUrlCacheMaxAgeSec != nil {
			cdnPolicy.SignedUrlCacheMaxAgeSec = pulumi.Int(int(spec.CdnPolicy.GetSignedUrlCacheMaxAgeSec()))
		}

		if len(spec.CdnPolicy.NegativeCachingPolicy) > 0 {
			negativeCachingPolicies := compute.BackendServiceCdnPolicyNegativeCachingPolicyArray{}
			for _, negativeCachingPolicy := range spec.CdnPolicy.NegativeCachingPolicy {
				negativeCachingPolicies = append(negativeCachingPolicies,
					&compute.BackendServiceCdnPolicyNegativeCachingPolicyArgs{
						Code: pulumi.Int(int(negativeCachingPolicy.Code)),
						// A 0 TTL means don't-cache-this-code to GCP, so it is
						// passed as-is (unlike the top-level TTLs).
						Ttl: pulumi.Int(int(negativeCachingPolicy.Ttl)),
					})
			}
			cdnPolicy.NegativeCachingPolicies = negativeCachingPolicies
		}

		// The backend-service cache key is richer than a backend bucket's:
		// host, protocol, query handling, cookies, and headers all shape it.
		if spec.CdnPolicy.CacheKeyPolicy != nil {
			cacheKeyPolicy := &compute.BackendServiceCdnPolicyCacheKeyPolicyArgs{
				IncludeHost:        pulumi.Bool(spec.CdnPolicy.CacheKeyPolicy.IncludeHost),
				IncludeProtocol:    pulumi.Bool(spec.CdnPolicy.CacheKeyPolicy.IncludeProtocol),
				IncludeQueryString: pulumi.Bool(spec.CdnPolicy.CacheKeyPolicy.IncludeQueryString),
			}
			if len(spec.CdnPolicy.CacheKeyPolicy.QueryStringWhitelist) > 0 {
				queryStringWhitelist := pulumi.StringArray{}
				for _, queryString := range spec.CdnPolicy.CacheKeyPolicy.QueryStringWhitelist {
					queryStringWhitelist = append(queryStringWhitelist, pulumi.String(queryString))
				}
				// Pulumi pluralizes this field (queryStringWhitelists) vs the
				// provider's singular query_string_whitelist — same wire field.
				cacheKeyPolicy.QueryStringWhitelists = queryStringWhitelist
			}
			if len(spec.CdnPolicy.CacheKeyPolicy.QueryStringBlacklist) > 0 {
				queryStringBlacklist := pulumi.StringArray{}
				for _, queryString := range spec.CdnPolicy.CacheKeyPolicy.QueryStringBlacklist {
					queryStringBlacklist = append(queryStringBlacklist, pulumi.String(queryString))
				}
				cacheKeyPolicy.QueryStringBlacklists = queryStringBlacklist
			}
			if len(spec.CdnPolicy.CacheKeyPolicy.IncludeHttpHeaders) > 0 {
				includeHttpHeaders := pulumi.StringArray{}
				for _, includeHttpHeader := range spec.CdnPolicy.CacheKeyPolicy.IncludeHttpHeaders {
					includeHttpHeaders = append(includeHttpHeaders, pulumi.String(includeHttpHeader))
				}
				cacheKeyPolicy.IncludeHttpHeaders = includeHttpHeaders
			}
			if len(spec.CdnPolicy.CacheKeyPolicy.IncludeNamedCookies) > 0 {
				includeNamedCookies := pulumi.StringArray{}
				for _, includeNamedCookie := range spec.CdnPolicy.CacheKeyPolicy.IncludeNamedCookies {
					includeNamedCookies = append(includeNamedCookies, pulumi.String(includeNamedCookie))
				}
				cacheKeyPolicy.IncludeNamedCookies = includeNamedCookies
			}
			cdnPolicy.CacheKeyPolicy = cacheKeyPolicy
		}

		if len(spec.CdnPolicy.BypassCacheOnRequestHeaders) > 0 {
			bypassCacheOnRequestHeaders := compute.BackendServiceCdnPolicyBypassCacheOnRequestHeaderArray{}
			for _, bypassCacheOnRequestHeader := range spec.CdnPolicy.BypassCacheOnRequestHeaders {
				bypassCacheOnRequestHeaders = append(bypassCacheOnRequestHeaders,
					&compute.BackendServiceCdnPolicyBypassCacheOnRequestHeaderArgs{
						HeaderName: pulumi.String(bypassCacheOnRequestHeader.HeaderName),
					})
			}
			cdnPolicy.BypassCacheOnRequestHeaders = bypassCacheOnRequestHeaders
		}

		args.CdnPolicy = cdnPolicy
	}

	// Identity-Aware Proxy: Google-identity authentication in front of the
	// backends. The client secret is secret material — never surfaced in
	// outputs, marked secret in the Pulumi state; GCP itself only returns
	// its SHA-256 after creation.
	if spec.Iap != nil {
		iap := &compute.BackendServiceIapArgs{
			Enabled: pulumi.Bool(spec.Iap.Enabled),
		}
		if spec.Iap.Oauth2ClientId != "" {
			iap.Oauth2ClientId = pulumi.String(spec.Iap.Oauth2ClientId)
		}
		if spec.Iap.Oauth2ClientSecret != "" {
			iap.Oauth2ClientSecret = pulumi.ToSecret(pulumi.String(spec.Iap.Oauth2ClientSecret)).(pulumi.StringOutput)
		}
		args.Iap = iap
	}

	if spec.LogConfig != nil {
		logConfig := &compute.BackendServiceLogConfigArgs{
			Enable: pulumi.Bool(spec.LogConfig.Enable),
		}
		if spec.LogConfig.SampleRate != nil {
			logConfig.SampleRate = pulumi.Float64(spec.LogConfig.GetSampleRate())
		}
		if spec.LogConfig.OptionalMode != "" {
			logConfig.OptionalMode = pulumi.String(spec.LogConfig.OptionalMode)
		}
		if len(spec.LogConfig.OptionalFields) > 0 {
			optionalFields := pulumi.StringArray{}
			for _, optionalField := range spec.LogConfig.OptionalFields {
				optionalFields = append(optionalFields, pulumi.String(optionalField))
			}
			logConfig.OptionalFields = optionalFields
		}
		args.LogConfig = logConfig
	}

	// GCP requires a cookie configuration with STRONG_COOKIE_AFFINITY; the
	// affinity mode is the statement and the block only customizes it, so the
	// module sends the (possibly empty, GCP-defaulted) configuration whenever
	// that mode is chosen.
	if spec.GetSessionAffinity() == "STRONG_COOKIE_AFFINITY" || spec.StrongSessionAffinityCookie != nil {
		strongSessionAffinityCookie := &compute.BackendServiceStrongSessionAffinityCookieArgs{}
		if spec.StrongSessionAffinityCookie.GetName() != "" {
			strongSessionAffinityCookie.Name = pulumi.String(spec.StrongSessionAffinityCookie.GetName())
		}
		if spec.StrongSessionAffinityCookie.GetPath() != "" {
			strongSessionAffinityCookie.Path = pulumi.String(spec.StrongSessionAffinityCookie.GetPath())
		}
		if spec.StrongSessionAffinityCookie.GetTtl() != nil {
			ttl := &compute.BackendServiceStrongSessionAffinityCookieTtlArgs{
				Seconds: pulumi.Int(int(spec.StrongSessionAffinityCookie.Ttl.Seconds)),
			}
			if spec.StrongSessionAffinityCookie.Ttl.Nanos != 0 {
				ttl.Nanos = pulumi.Int(int(spec.StrongSessionAffinityCookie.Ttl.Nanos))
			}
			strongSessionAffinityCookie.Ttl = ttl
		}
		args.StrongSessionAffinityCookie = strongSessionAffinityCookie
	}

	// Ordered Traffic Director policy list; each entry carries exactly one
	// of a built-in policy or a custom xDS policy (proto oneof upstream).
	if len(spec.LocalityLbPolicies) > 0 {
		localityLbPolicies := compute.BackendServiceLocalityLbPolicyArray{}
		for _, localityLbPolicyConfig := range spec.LocalityLbPolicies {
			localityLbPolicyArgs := &compute.BackendServiceLocalityLbPolicyArgs{}
			if builtinPolicy := localityLbPolicyConfig.GetPolicy(); builtinPolicy != nil {
				localityLbPolicyArgs.Policy = &compute.BackendServiceLocalityLbPolicyPolicyArgs{
					Name: pulumi.String(builtinPolicy.Name),
				}
			}
			if customPolicy := localityLbPolicyConfig.GetCustomPolicy(); customPolicy != nil {
				customPolicyArgs := &compute.BackendServiceLocalityLbPolicyCustomPolicyArgs{
					Name: pulumi.String(customPolicy.Name),
				}
				if customPolicy.Data != "" {
					customPolicyArgs.Data = pulumi.String(customPolicy.Data)
				}
				localityLbPolicyArgs.CustomPolicy = customPolicyArgs
			}
			localityLbPolicies = append(localityLbPolicies, localityLbPolicyArgs)
		}
		args.LocalityLbPolicies = localityLbPolicies
	}

	if spec.ConsistentHash != nil {
		consistentHash := &compute.BackendServiceConsistentHashArgs{}
		if spec.ConsistentHash.HttpHeaderName != "" {
			consistentHash.HttpHeaderName = pulumi.String(spec.ConsistentHash.HttpHeaderName)
		}
		if spec.ConsistentHash.MinimumRingSize != nil {
			consistentHash.MinimumRingSize = pulumi.Int(int(spec.ConsistentHash.GetMinimumRingSize()))
		}
		if spec.ConsistentHash.HttpCookie != nil {
			httpCookie := &compute.BackendServiceConsistentHashHttpCookieArgs{}
			if spec.ConsistentHash.HttpCookie.Name != "" {
				httpCookie.Name = pulumi.String(spec.ConsistentHash.HttpCookie.Name)
			}
			if spec.ConsistentHash.HttpCookie.Path != "" {
				httpCookie.Path = pulumi.String(spec.ConsistentHash.HttpCookie.Path)
			}
			if spec.ConsistentHash.HttpCookie.Ttl != nil {
				ttl := &compute.BackendServiceConsistentHashHttpCookieTtlArgs{
					Seconds: pulumi.Int(int(spec.ConsistentHash.HttpCookie.Ttl.Seconds)),
				}
				if spec.ConsistentHash.HttpCookie.Ttl.Nanos != 0 {
					ttl.Nanos = pulumi.Int(int(spec.ConsistentHash.HttpCookie.Ttl.Nanos))
				}
				httpCookie.Ttl = ttl
			}
			consistentHash.HttpCookie = httpCookie
		}
		args.ConsistentHash = consistentHash
	}

	if spec.CircuitBreakers != nil {
		circuitBreakers := &compute.BackendServiceCircuitBreakersArgs{}
		if spec.CircuitBreakers.MaxConnections != nil {
			circuitBreakers.MaxConnections = pulumi.Int(int(spec.CircuitBreakers.GetMaxConnections()))
		}
		if spec.CircuitBreakers.MaxPendingRequests != nil {
			circuitBreakers.MaxPendingRequests = pulumi.Int(int(spec.CircuitBreakers.GetMaxPendingRequests()))
		}
		if spec.CircuitBreakers.MaxRequests != nil {
			circuitBreakers.MaxRequests = pulumi.Int(int(spec.CircuitBreakers.GetMaxRequests()))
		}
		if spec.CircuitBreakers.MaxRequestsPerConnection != 0 {
			circuitBreakers.MaxRequestsPerConnection = pulumi.Int(int(spec.CircuitBreakers.MaxRequestsPerConnection))
		}
		if spec.CircuitBreakers.MaxRetries != nil {
			circuitBreakers.MaxRetries = pulumi.Int(int(spec.CircuitBreakers.GetMaxRetries()))
		}
		args.CircuitBreakers = circuitBreakers
	}

	if spec.OutlierDetection != nil {
		outlierDetection := &compute.BackendServiceOutlierDetectionArgs{}
		if spec.OutlierDetection.BaseEjectionTime != nil {
			baseEjectionTime := &compute.BackendServiceOutlierDetectionBaseEjectionTimeArgs{
				Seconds: pulumi.Int(int(spec.OutlierDetection.BaseEjectionTime.Seconds)),
			}
			if spec.OutlierDetection.BaseEjectionTime.Nanos != 0 {
				baseEjectionTime.Nanos = pulumi.Int(int(spec.OutlierDetection.BaseEjectionTime.Nanos))
			}
			outlierDetection.BaseEjectionTime = baseEjectionTime
		}
		if spec.OutlierDetection.ConsecutiveErrors != 0 {
			outlierDetection.ConsecutiveErrors = pulumi.Int(int(spec.OutlierDetection.ConsecutiveErrors))
		}
		if spec.OutlierDetection.ConsecutiveGatewayFailure != 0 {
			outlierDetection.ConsecutiveGatewayFailure = pulumi.Int(int(spec.OutlierDetection.ConsecutiveGatewayFailure))
		}
		if spec.OutlierDetection.EnforcingConsecutiveErrors != 0 {
			outlierDetection.EnforcingConsecutiveErrors = pulumi.Int(int(spec.OutlierDetection.EnforcingConsecutiveErrors))
		}
		if spec.OutlierDetection.EnforcingConsecutiveGatewayFailure != 0 {
			outlierDetection.EnforcingConsecutiveGatewayFailure = pulumi.Int(int(spec.OutlierDetection.EnforcingConsecutiveGatewayFailure))
		}
		if spec.OutlierDetection.EnforcingSuccessRate != 0 {
			outlierDetection.EnforcingSuccessRate = pulumi.Int(int(spec.OutlierDetection.EnforcingSuccessRate))
		}
		if spec.OutlierDetection.Interval != nil {
			interval := &compute.BackendServiceOutlierDetectionIntervalArgs{
				Seconds: pulumi.Int(int(spec.OutlierDetection.Interval.Seconds)),
			}
			if spec.OutlierDetection.Interval.Nanos != 0 {
				interval.Nanos = pulumi.Int(int(spec.OutlierDetection.Interval.Nanos))
			}
			outlierDetection.Interval = interval
		}
		if spec.OutlierDetection.MaxEjectionPercent != 0 {
			outlierDetection.MaxEjectionPercent = pulumi.Int(int(spec.OutlierDetection.MaxEjectionPercent))
		}
		if spec.OutlierDetection.SuccessRateMinimumHosts != 0 {
			outlierDetection.SuccessRateMinimumHosts = pulumi.Int(int(spec.OutlierDetection.SuccessRateMinimumHosts))
		}
		if spec.OutlierDetection.SuccessRateRequestVolume != 0 {
			outlierDetection.SuccessRateRequestVolume = pulumi.Int(int(spec.OutlierDetection.SuccessRateRequestVolume))
		}
		if spec.OutlierDetection.SuccessRateStdevFactor != 0 {
			outlierDetection.SuccessRateStdevFactor = pulumi.Int(int(spec.OutlierDetection.SuccessRateStdevFactor))
		}
		args.OutlierDetection = outlierDetection
	}

	// The provider models this Duration's seconds as a STRING (int64 format).
	if spec.MaxStreamDuration != nil {
		maxStreamDuration := &compute.BackendServiceMaxStreamDurationArgs{
			Seconds: pulumi.String(strconv.FormatInt(spec.MaxStreamDuration.Seconds, 10)),
		}
		if spec.MaxStreamDuration.Nanos != 0 {
			maxStreamDuration.Nanos = pulumi.Int(int(spec.MaxStreamDuration.Nanos))
		}
		args.MaxStreamDuration = maxStreamDuration
	}

	if spec.SecuritySettings != nil {
		securitySettings := &compute.BackendServiceSecuritySettingsArgs{}
		if spec.SecuritySettings.ClientTlsPolicy != "" {
			securitySettings.ClientTlsPolicy = pulumi.String(spec.SecuritySettings.ClientTlsPolicy)
		}
		if len(spec.SecuritySettings.SubjectAltNames) > 0 {
			subjectAltNames := pulumi.StringArray{}
			for _, subjectAltName := range spec.SecuritySettings.SubjectAltNames {
				subjectAltNames = append(subjectAltNames, pulumi.String(subjectAltName))
			}
			securitySettings.SubjectAltNames = subjectAltNames
		}
		// SigV4 origin signing: access_key is secret material — never
		// surfaced in outputs, marked secret in the Pulumi state; GCP never
		// returns it on reads.
		if spec.SecuritySettings.AwsV4Authentication != nil {
			awsV4Authentication := &compute.BackendServiceSecuritySettingsAwsV4AuthenticationArgs{}
			if spec.SecuritySettings.AwsV4Authentication.AccessKeyId != "" {
				awsV4Authentication.AccessKeyId = pulumi.String(spec.SecuritySettings.AwsV4Authentication.AccessKeyId)
			}
			if spec.SecuritySettings.AwsV4Authentication.AccessKey != "" {
				awsV4Authentication.AccessKey = pulumi.ToSecret(pulumi.String(spec.SecuritySettings.AwsV4Authentication.AccessKey)).(pulumi.StringOutput)
			}
			if spec.SecuritySettings.AwsV4Authentication.AccessKeyVersion != "" {
				awsV4Authentication.AccessKeyVersion = pulumi.String(spec.SecuritySettings.AwsV4Authentication.AccessKeyVersion)
			}
			if spec.SecuritySettings.AwsV4Authentication.OriginRegion != "" {
				awsV4Authentication.OriginRegion = pulumi.String(spec.SecuritySettings.AwsV4Authentication.OriginRegion)
			}
			securitySettings.AwsV4Authentication = awsV4Authentication
		}
		args.SecuritySettings = securitySettings
	}

	if spec.TlsSettings != nil {
		tlsSettings := &compute.BackendServiceTlsSettingsArgs{}
		if spec.TlsSettings.AuthenticationConfig != "" {
			tlsSettings.AuthenticationConfig = pulumi.String(spec.TlsSettings.AuthenticationConfig)
		}
		if spec.TlsSettings.Sni != "" {
			tlsSettings.Sni = pulumi.String(spec.TlsSettings.Sni)
		}
		if len(spec.TlsSettings.SubjectAltNames) > 0 {
			subjectAltNames := compute.BackendServiceTlsSettingsSubjectAltNameArray{}
			// Each SAN entry carries exactly one arm (proto oneof) — only the
			// set arm is sent.
			for _, subjectAltName := range spec.TlsSettings.SubjectAltNames {
				subjectAltNameArgs := &compute.BackendServiceTlsSettingsSubjectAltNameArgs{}
				if subjectAltName.GetDnsName() != "" {
					subjectAltNameArgs.DnsName = pulumi.String(subjectAltName.GetDnsName())
				}
				if subjectAltName.GetUniformResourceIdentifier() != "" {
					subjectAltNameArgs.UniformResourceIdentifier = pulumi.String(subjectAltName.GetUniformResourceIdentifier())
				}
				subjectAltNames = append(subjectAltNames, subjectAltNameArgs)
			}
			tlsSettings.SubjectAltNames = subjectAltNames
		}
		args.TlsSettings = tlsSettings
	}

	// Service-level ORCA metrics for WEIGHTED_ROUND_ROBIN.
	if len(spec.CustomMetrics) > 0 {
		customMetrics := compute.BackendServiceCustomMetricArray{}
		for _, customMetric := range spec.CustomMetrics {
			customMetrics = append(customMetrics, &compute.BackendServiceCustomMetricArgs{
				Name:   pulumi.String(customMetric.Name),
				DryRun: pulumi.Bool(customMetric.DryRun),
			})
		}
		args.CustomMetrics = customMetrics
	}

	// Create-time Resource Manager tags; the provider nests them in a
	// params block that must be omitted entirely when no tags are set.
	if len(spec.ResourceManagerTags) > 0 {
		args.Params = &compute.BackendServiceParamsArgs{
			ResourceManagerTags: pulumi.ToStringMap(spec.ResourceManagerTags),
		}
	}

	// What destroy does to the routing target: DELETE (default), PREVENT
	// (refuse), or ABANDON (drop from state, keep serving).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdBackendService, err := compute.NewBackendService(ctx, "backend-service", args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create backend service")
	}

	// Signed-URL keys — folded into this kind rather than modeled as a
	// separate node: keys are never referenced by other resources, GCP caps
	// them at 3 per service, and their lifecycle is the service's. Each key
	// is immutable in GCP (add/delete only), which is exactly the rotation
	// semantics signed URLs need (add new key -> re-sign -> remove old).
	for _, signedUrlKey := range spec.SignedUrlKeys {
		signedUrlKeyArgs := &compute.BackendServiceSignedUrlKeyArgs{
			Name: pulumi.String(signedUrlKey.Name),
			// Secret material; never surfaced in outputs. ToSecret marks it
			// in the Pulumi state as well.
			KeyValue:       pulumi.ToSecret(pulumi.String(signedUrlKey.KeyValue)).(pulumi.StringOutput),
			BackendService: createdBackendService.Name,
		}
		if spec.ProjectId.GetValue() != "" {
			signedUrlKeyArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		// The kind-level deletion_policy fans out to every key — the keys
		// have no life apart from the service, so one switch governs both.
		if spec.DeletionPolicy != "" {
			signedUrlKeyArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if _, err := compute.NewBackendServiceSignedUrlKey(ctx, "signed-url-key-"+signedUrlKey.Name,
			signedUrlKeyArgs, pulumi.Provider(gcpProvider)); err != nil {
			return errors.Wrap(err, "failed to create signed url key "+signedUrlKey.Name)
		}
	}

	ctx.Export(OpSelfLink, createdBackendService.SelfLink)
	ctx.Export(OpBackendServiceName, createdBackendService.Name)
	ctx.Export(OpGeneratedId, createdBackendService.GeneratedId.ApplyT(func(generatedId int) string {
		return strconv.Itoa(generatedId)
	}).(pulumi.StringOutput))
	ctx.Export(OpFingerprint, createdBackendService.Fingerprint)

	return nil
}
