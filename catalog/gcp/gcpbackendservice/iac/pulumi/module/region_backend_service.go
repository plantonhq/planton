package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// regionalBackendService builds the regional twin (spec.region set): the
// backend of the regional external and internal Application Load Balancers
// and of the internal and external passthrough Network Load Balancers. A
// regional ALB needs a regional health check in the same region; the
// passthrough forms take the network, failover, connection-tracking, HA,
// and zonal-affinity policies and are named directly by a regional
// forwarding rule.
//
// The assembly mirrors backend_service.go's global one minus the surfaces
// the regional API lacks — compression, custom headers, edge Cloud Armor,
// service LB policy, the migration canary, backend preference, Traffic
// Director policies, signed URL keys, and Cloud CDN's advanced knobs —
// which the spec's CEL walls keep off a regional manifest before either
// engine runs. Pulumi's nested input types are per-resource, so the
// assembly is written once per arm; the two files are kept in the same
// order so a change lands in both.
func regionalBackendService(ctx *pulumi.Context, locals *Locals, opts []pulumi.ResourceOption) error {
	spec := locals.GcpBackendService.Spec

	args := &compute.RegionBackendServiceArgs{
		Name:   pulumi.String(locals.BackendServiceName),
		Region: pulumi.String(spec.Region),
		// The regional API carries the CDN switch although regional
		// Application Load Balancers have no Cloud CDN; the spec's CEL
		// confines it to the external schemes and the GUIDE says plainly it
		// does nothing here.
		EnableCdn: pulumi.Bool(spec.EnableCdn),
	}

	// The VPC network of a passthrough NLB (required with an HA policy).
	if spec.Network.GetValue() != "" {
		args.Network = pulumi.String(spec.Network.GetValue())
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
	// Sent explicitly (see the global builder's note): the regional provider
	// default is INTERNAL, and the spec's contract is EXTERNAL on both
	// scopes, so an empty spec value becomes EXTERNAL here too.
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
	// With an HA policy the leader IS the routing decision, and the provider
	// rejects session_affinity beside it even at the spec's NONE default (the
	// manifest loader fills NONE before the module runs), so the affinity is
	// dropped entirely on that arm. The Terraform module makes the same
	// choice.
	if spec.GetSessionAffinity() != "" && spec.HaPolicy == nil {
		args.SessionAffinity = pulumi.String(spec.GetSessionAffinity())
	}
	if spec.AffinityCookieTtlSec != 0 {
		args.AffinityCookieTtlSec = pulumi.Int(int(spec.AffinityCookieTtlSec))
	}
	if spec.LocalityLbPolicy != "" {
		args.LocalityLbPolicy = pulumi.String(spec.LocalityLbPolicy)
	}
	if spec.IpAddressSelectionPolicy != "" {
		args.IpAddressSelectionPolicy = pulumi.String(spec.IpAddressSelectionPolicy)
	}

	// At most one health check (a GCP-enforced cap — the spec models it
	// singular; the SDK flattens the one-element set to a plain string). A
	// regional ALB requires a regional one; not allowed together with
	// ha_policy (spec CEL).
	if spec.HealthCheck.GetValue() != "" {
		args.HealthChecks = pulumi.String(spec.HealthCheck.GetValue())
	}

	// A regional Cloud Armor policy in the same region, attached by reference
	// (GCP rejects a global policy here; edge policies are global-only).
	if spec.SecurityPolicy.GetValue() != "" {
		args.SecurityPolicy = pulumi.String(spec.SecurityPolicy.GetValue())
	}

	// The backends serving traffic. Instance groups and NEGs cannot be mixed
	// on one service (GCP rejects it); which max_* dial is required is
	// decided by each backend's balancing_mode and enforced by the spec
	// pre-deploy.
	if len(spec.Backends) > 0 {
		backends := compute.RegionBackendServiceBackendArray{}
		for _, backend := range spec.Backends {
			backendArgs := &compute.RegionBackendServiceBackendArgs{
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
			// The failover pool of a passthrough NLB (regional-only; the
			// API default is false, so only an explicit true is sent).
			if backend.Failover {
				backendArgs.Failover = pulumi.Bool(true)
			}
			if len(backend.CustomMetrics) > 0 {
				backendCustomMetrics := compute.RegionBackendServiceBackendCustomMetricArray{}
				for _, backendCustomMetric := range backend.CustomMetrics {
					backendCustomMetricArgs := &compute.RegionBackendServiceBackendCustomMetricArgs{
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
		cdnPolicy := &compute.RegionBackendServiceCdnPolicyArgs{}

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
		if spec.CdnPolicy.SignedUrlCacheMaxAgeSec != nil {
			cdnPolicy.SignedUrlCacheMaxAgeSec = pulumi.Int(int(spec.CdnPolicy.GetSignedUrlCacheMaxAgeSec()))
		}

		if len(spec.CdnPolicy.NegativeCachingPolicy) > 0 {
			negativeCachingPolicies := compute.RegionBackendServiceCdnPolicyNegativeCachingPolicyArray{}
			for _, negativeCachingPolicy := range spec.CdnPolicy.NegativeCachingPolicy {
				negativeCachingPolicies = append(negativeCachingPolicies,
					// The regional API carries the code alone (no per-code
					// TTL); the spec CEL keeps ttl off a regional manifest.
					&compute.RegionBackendServiceCdnPolicyNegativeCachingPolicyArgs{
						Code: pulumi.Int(int(negativeCachingPolicy.Code)),
					})
			}
			cdnPolicy.NegativeCachingPolicies = negativeCachingPolicies
		}

		// The backend-service cache key is richer than a backend bucket's:
		// host, protocol, query handling, cookies, and headers all shape it.
		if spec.CdnPolicy.CacheKeyPolicy != nil {
			cacheKeyPolicy := &compute.RegionBackendServiceCdnPolicyCacheKeyPolicyArgs{
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
			if len(spec.CdnPolicy.CacheKeyPolicy.IncludeNamedCookies) > 0 {
				includeNamedCookies := pulumi.StringArray{}
				for _, includeNamedCookie := range spec.CdnPolicy.CacheKeyPolicy.IncludeNamedCookies {
					includeNamedCookies = append(includeNamedCookies, pulumi.String(includeNamedCookie))
				}
				cacheKeyPolicy.IncludeNamedCookies = includeNamedCookies
			}
			cdnPolicy.CacheKeyPolicy = cacheKeyPolicy
		}

		args.CdnPolicy = cdnPolicy
	}

	// Identity-Aware Proxy: Google-identity authentication in front of the
	// backends. The client secret is secret material — never surfaced in
	// outputs, marked secret in the Pulumi state; GCP itself only returns
	// its SHA-256 after creation.
	if spec.Iap != nil {
		iap := &compute.RegionBackendServiceIapArgs{
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
		logConfig := &compute.RegionBackendServiceLogConfigArgs{
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
		// Each logged header is its own block in the provider; the spec
		// holds a flat name list.
		if len(spec.LogConfig.RequestHeaders) > 0 {
			requestHeaders := compute.RegionBackendServiceLogConfigRequestHeaderArray{}
			for _, headerName := range spec.LogConfig.RequestHeaders {
				requestHeaders = append(requestHeaders, &compute.RegionBackendServiceLogConfigRequestHeaderArgs{
					HeaderName: pulumi.String(headerName),
				})
			}
			logConfig.RequestHeaders = requestHeaders
		}
		if len(spec.LogConfig.ResponseHeaders) > 0 {
			responseHeaders := compute.RegionBackendServiceLogConfigResponseHeaderArray{}
			for _, headerName := range spec.LogConfig.ResponseHeaders {
				responseHeaders = append(responseHeaders, &compute.RegionBackendServiceLogConfigResponseHeaderArgs{
					HeaderName: pulumi.String(headerName),
				})
			}
			logConfig.ResponseHeaders = responseHeaders
		}
		args.LogConfig = logConfig
	}

	// GCP requires a cookie configuration with STRONG_COOKIE_AFFINITY; the
	// affinity mode is the statement and the block only customizes it, so the
	// module sends the (possibly empty, GCP-defaulted) configuration whenever
	// that mode is chosen.
	if spec.GetSessionAffinity() == "STRONG_COOKIE_AFFINITY" || spec.StrongSessionAffinityCookie != nil {
		strongSessionAffinityCookie := &compute.RegionBackendServiceStrongSessionAffinityCookieArgs{}
		if spec.StrongSessionAffinityCookie.GetName() != "" {
			strongSessionAffinityCookie.Name = pulumi.String(spec.StrongSessionAffinityCookie.GetName())
		}
		if spec.StrongSessionAffinityCookie.GetPath() != "" {
			strongSessionAffinityCookie.Path = pulumi.String(spec.StrongSessionAffinityCookie.GetPath())
		}
		if spec.StrongSessionAffinityCookie.GetTtl() != nil {
			ttl := &compute.RegionBackendServiceStrongSessionAffinityCookieTtlArgs{
				Seconds: pulumi.Int(int(spec.StrongSessionAffinityCookie.Ttl.Seconds)),
			}
			if spec.StrongSessionAffinityCookie.Ttl.Nanos != 0 {
				ttl.Nanos = pulumi.Int(int(spec.StrongSessionAffinityCookie.Ttl.Nanos))
			}
			strongSessionAffinityCookie.Ttl = ttl
		}
		args.StrongSessionAffinityCookie = strongSessionAffinityCookie
	}

	if spec.ConsistentHash != nil {
		consistentHash := &compute.RegionBackendServiceConsistentHashArgs{}
		if spec.ConsistentHash.HttpHeaderName != "" {
			consistentHash.HttpHeaderName = pulumi.String(spec.ConsistentHash.HttpHeaderName)
		}
		if spec.ConsistentHash.MinimumRingSize != nil {
			consistentHash.MinimumRingSize = pulumi.Int(int(spec.ConsistentHash.GetMinimumRingSize()))
		}
		if spec.ConsistentHash.HttpCookie != nil {
			httpCookie := &compute.RegionBackendServiceConsistentHashHttpCookieArgs{}
			if spec.ConsistentHash.HttpCookie.Name != "" {
				httpCookie.Name = pulumi.String(spec.ConsistentHash.HttpCookie.Name)
			}
			if spec.ConsistentHash.HttpCookie.Path != "" {
				httpCookie.Path = pulumi.String(spec.ConsistentHash.HttpCookie.Path)
			}
			if spec.ConsistentHash.HttpCookie.Ttl != nil {
				ttl := &compute.RegionBackendServiceConsistentHashHttpCookieTtlArgs{
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
		circuitBreakers := &compute.RegionBackendServiceCircuitBreakersArgs{}
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
		outlierDetection := &compute.RegionBackendServiceOutlierDetectionArgs{}
		if spec.OutlierDetection.BaseEjectionTime != nil {
			baseEjectionTime := &compute.RegionBackendServiceOutlierDetectionBaseEjectionTimeArgs{
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
			interval := &compute.RegionBackendServiceOutlierDetectionIntervalArgs{
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

	if spec.TlsSettings != nil {
		tlsSettings := &compute.RegionBackendServiceTlsSettingsArgs{}
		if spec.TlsSettings.AuthenticationConfig != "" {
			tlsSettings.AuthenticationConfig = pulumi.String(spec.TlsSettings.AuthenticationConfig)
		}
		if spec.TlsSettings.Sni != "" {
			tlsSettings.Sni = pulumi.String(spec.TlsSettings.Sni)
		}
		if len(spec.TlsSettings.SubjectAltNames) > 0 {
			subjectAltNames := compute.RegionBackendServiceTlsSettingsSubjectAltNameArray{}
			// Each SAN entry carries exactly one arm (proto oneof) — only the
			// set arm is sent.
			for _, subjectAltName := range spec.TlsSettings.SubjectAltNames {
				subjectAltNameArgs := &compute.RegionBackendServiceTlsSettingsSubjectAltNameArgs{}
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
		customMetrics := compute.RegionBackendServiceCustomMetricArray{}
		for _, customMetric := range spec.CustomMetrics {
			customMetrics = append(customMetrics, &compute.RegionBackendServiceCustomMetricArgs{
				Name:   pulumi.String(customMetric.Name),
				DryRun: pulumi.Bool(customMetric.DryRun),
			})
		}
		args.CustomMetrics = customMetrics
	}

	// ---- The passthrough Network Load Balancer policies (regional-only) ----
	// Every field carries presence in the spec, so nil means "not set" and
	// the API default stands.

	if spec.FailoverPolicy != nil {
		failoverPolicy := &compute.RegionBackendServiceFailoverPolicyArgs{}
		if spec.FailoverPolicy.DisableConnectionDrainOnFailover != nil {
			failoverPolicy.DisableConnectionDrainOnFailover = pulumi.Bool(spec.FailoverPolicy.GetDisableConnectionDrainOnFailover())
		}
		if spec.FailoverPolicy.DropTrafficIfUnhealthy != nil {
			failoverPolicy.DropTrafficIfUnhealthy = pulumi.Bool(spec.FailoverPolicy.GetDropTrafficIfUnhealthy())
		}
		if spec.FailoverPolicy.FailoverRatio != nil {
			failoverPolicy.FailoverRatio = pulumi.Float64(spec.FailoverPolicy.GetFailoverRatio())
		}
		args.FailoverPolicy = failoverPolicy
	}

	// The two mode strings carry spec defaults that match Google's and are
	// sent explicitly (a re-plan stays clean); idle_timeout_sec is
	// Optional+Computed on the API and sent only when set.
	if spec.ConnectionTrackingPolicy != nil {
		trackingMode := spec.ConnectionTrackingPolicy.GetTrackingMode()
		if trackingMode == "" {
			trackingMode = "PER_CONNECTION"
		}
		persistence := spec.ConnectionTrackingPolicy.GetConnectionPersistenceOnUnhealthyBackends()
		if persistence == "" {
			persistence = "DEFAULT_FOR_PROTOCOL"
		}
		connectionTracking := &compute.RegionBackendServiceConnectionTrackingPolicyArgs{
			TrackingMode:                             pulumi.String(trackingMode),
			ConnectionPersistenceOnUnhealthyBackends: pulumi.String(persistence),
		}
		if spec.ConnectionTrackingPolicy.IdleTimeoutSec != nil {
			connectionTracking.IdleTimeoutSec = pulumi.Int(int(spec.ConnectionTrackingPolicy.GetIdleTimeoutSec()))
		}
		if spec.ConnectionTrackingPolicy.EnableStrongAffinity {
			connectionTracking.EnableStrongAffinity = pulumi.Bool(true)
		}
		args.ConnectionTrackingPolicy = connectionTracking
	}

	if spec.HaPolicy != nil {
		haPolicy := &compute.RegionBackendServiceHaPolicyArgs{}
		if spec.HaPolicy.FastIpMove != "" {
			haPolicy.FastIpMove = pulumi.String(spec.HaPolicy.FastIpMove)
		}
		if spec.HaPolicy.Leader != nil {
			leader := &compute.RegionBackendServiceHaPolicyLeaderArgs{}
			if spec.HaPolicy.Leader.BackendGroup != "" {
				leader.BackendGroup = pulumi.String(spec.HaPolicy.Leader.BackendGroup)
			}
			if spec.HaPolicy.Leader.NetworkEndpoint != nil && spec.HaPolicy.Leader.NetworkEndpoint.Instance != "" {
				leader.NetworkEndpoint = &compute.RegionBackendServiceHaPolicyLeaderNetworkEndpointArgs{
					Instance: pulumi.String(spec.HaPolicy.Leader.NetworkEndpoint.Instance),
				}
			}
			haPolicy.Leader = leader
		}
		args.HaPolicy = haPolicy
	}

	// Zonal affinity: the mode carries a spec default matching Google's and
	// is sent explicitly; the ratio only when set.
	if spec.NetworkPassThroughLbTrafficPolicy != nil && spec.NetworkPassThroughLbTrafficPolicy.ZonalAffinity != nil {
		spillover := spec.NetworkPassThroughLbTrafficPolicy.ZonalAffinity.GetSpillover()
		if spillover == "" {
			spillover = "ZONAL_AFFINITY_DISABLED"
		}
		zonalAffinity := &compute.RegionBackendServiceNetworkPassThroughLbTrafficPolicyZonalAffinityArgs{
			Spillover: pulumi.String(spillover),
		}
		if spec.NetworkPassThroughLbTrafficPolicy.ZonalAffinity.SpilloverRatio != nil {
			zonalAffinity.SpilloverRatio = pulumi.Float64(spec.NetworkPassThroughLbTrafficPolicy.ZonalAffinity.GetSpilloverRatio())
		}
		args.NetworkPassThroughLbTrafficPolicy = &compute.RegionBackendServiceNetworkPassThroughLbTrafficPolicyArgs{
			ZonalAffinity: zonalAffinity,
		}
	}

	// Create-time Resource Manager tags; the provider nests them in a
	// params block that must be omitted entirely when no tags are set.
	if len(spec.ResourceManagerTags) > 0 {
		args.Params = &compute.RegionBackendServiceParamsArgs{
			ResourceManagerTags: pulumi.ToStringMap(spec.ResourceManagerTags),
		}
	}

	// What destroy does to the routing target: DELETE (default), PREVENT
	// (refuse), or ABANDON (drop from state, keep serving).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdBackendService, err := compute.NewRegionBackendService(ctx, "backend-service", args, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create region backend service")
	}

	ctx.Export(OpSelfLink, createdBackendService.SelfLink)
	ctx.Export(OpBackendServiceName, createdBackendService.Name)
	ctx.Export(OpGeneratedId, createdBackendService.GeneratedId.ApplyT(func(generatedId int) string {
		return strconv.Itoa(generatedId)
	}).(pulumi.StringOutput))
	ctx.Export(OpFingerprint, createdBackendService.Fingerprint)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return nil
}
