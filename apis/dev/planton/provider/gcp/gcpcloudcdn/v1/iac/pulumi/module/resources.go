package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpcloudcdnv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudcdn/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput) error {
	// Create GCP provider using the credentials from the input
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	// Initialize locals (labels, resource names, configuration)
	locals := initializeLocals(stackInput)

	// Create resources based on backend type
	var backendId pulumi.StringOutput
	var backendName pulumi.StringOutput
	var backendType string

	if locals.IsGcsBucket {
		backendId, backendName, err = createGcsBackendBucket(ctx, stackInput, locals, gcpProvider)
		backendType = "GCS_BUCKET"
	} else if locals.IsComputeService {
		backendId, backendName, err = createComputeBackendService(ctx, stackInput, locals, gcpProvider)
		backendType = "COMPUTE_SERVICE"
	} else if locals.IsCloudRun {
		backendId, backendName, err = createCloudRunBackendService(ctx, stackInput, locals, gcpProvider)
		backendType = "CLOUD_RUN"
	} else if locals.IsExternalOrigin {
		backendId, backendName, err = createExternalBackendService(ctx, stackInput, locals, gcpProvider)
		backendType = "EXTERNAL"
	} else {
		return errors.New("no backend type specified - must specify one of: gcs_bucket, compute_service, cloud_run_service, external_origin")
	}

	if err != nil {
		return errors.Wrap(err, "failed to create backend")
	}

	// Create global IP address for load balancer
	globalIp, err := compute.NewGlobalAddress(ctx, locals.GlobalAddressName, &compute.GlobalAddressArgs{
		Name:    pulumi.String(locals.GlobalAddressName),
		Project: pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
	}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create global IP address")
	}

	// Create URL map (routing configuration)
	urlMap, err := compute.NewURLMap(ctx, locals.UrlMapName, &compute.URLMapArgs{
		Name:           pulumi.String(locals.UrlMapName),
		Project:        pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		DefaultService: backendId,
	}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create URL map")
	}

	// Create SSL certificate and HTTPS proxy if frontend config is specified
	var httpsProxy *compute.TargetHttpsProxy
	var sslCertName pulumi.StringOutput

	if stackInput.Target.Spec.FrontendConfig != nil {
		sslCert, err := createSslCertificate(ctx, stackInput, locals, gcpProvider)
		if err != nil {
			return errors.Wrap(err, "failed to create SSL certificate")
		}
		sslCertName = sslCert.Name

		httpsProxy, err = compute.NewTargetHttpsProxy(ctx, locals.HttpsProxyName, &compute.TargetHttpsProxyArgs{
			Name:            pulumi.String(locals.HttpsProxyName),
			Project:         pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
			UrlMap:          urlMap.ID(),
			SslCertificates: pulumi.StringArray{sslCert.ID()},
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create HTTPS proxy")
		}

		// Create HTTPS forwarding rule
		_, err = compute.NewGlobalForwardingRule(ctx, locals.HttpsProxyName+"-rule", &compute.GlobalForwardingRuleArgs{
			Name:      pulumi.String(locals.HttpsProxyName + "-rule"),
			Project:   pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
			Target:    httpsProxy.ID(),
			PortRange: pulumi.String("443"),
			IpAddress: globalIp.Address,
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create HTTPS forwarding rule")
		}

		// Create HTTP to HTTPS redirect if enabled (default: true)
		if stackInput.Target.Spec.FrontendConfig.EnableHttpsRedirect == nil || *stackInput.Target.Spec.FrontendConfig.EnableHttpsRedirect {
			if err := createHttpRedirect(ctx, stackInput, locals, globalIp, gcpProvider); err != nil {
				return errors.Wrap(err, "failed to create HTTP to HTTPS redirect")
			}
		}
	} else {
		// No frontend config - create basic HTTP forwarding rule
		httpProxy, err := compute.NewTargetHttpProxy(ctx, locals.HttpProxyName, &compute.TargetHttpProxyArgs{
			Name:    pulumi.String(locals.HttpProxyName),
			Project: pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
			UrlMap:  urlMap.ID(),
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create HTTP proxy")
		}

		_, err = compute.NewGlobalForwardingRule(ctx, locals.HttpProxyName+"-rule", &compute.GlobalForwardingRuleArgs{
			Name:      pulumi.String(locals.HttpProxyName + "-rule"),
			Project:   pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
			Target:    httpProxy.ID(),
			PortRange: pulumi.String("80"),
			IpAddress: globalIp.Address,
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create HTTP forwarding rule")
		}
	}

	// Export outputs
	cdnUrl := globalIp.Address.ApplyT(func(addr string) string {
		return fmt.Sprintf("https://%s", addr)
	}).(pulumi.StringOutput)

	ctx.Export("cdn_url", cdnUrl)
	ctx.Export("global_ip_address", globalIp.Address)
	ctx.Export("backend_name", backendName)
	ctx.Export("backend_id", backendId)
	ctx.Export("cdn_enabled", pulumi.Bool(true))
	ctx.Export("cache_mode", pulumi.String(locals.CacheMode))
	ctx.Export("url_map_name", urlMap.Name)
	ctx.Export("backend_type", pulumi.String(backendType))

	if httpsProxy != nil {
		ctx.Export("https_proxy_name", httpsProxy.Name)
		ctx.Export("ssl_certificate_name", sslCertName)
	}

	// Export backend-specific information
	if locals.IsGcsBucket {
		gcsBucket := stackInput.Target.Spec.Backend.GetGcsBucket()
		if gcsBucket != nil {
			ctx.Export("gcs_bucket_name", pulumi.String(gcsBucket.BucketName))
		}
	} else if locals.IsCloudRun {
		cloudRun := stackInput.Target.Spec.Backend.GetCloudRunService()
		if cloudRun != nil {
			ctx.Export("cloud_run_service_name", pulumi.String(cloudRun.ServiceName))
			ctx.Export("cloud_run_region", pulumi.String(cloudRun.Region))
		}
	} else if locals.IsComputeService {
		compute := stackInput.Target.Spec.Backend.GetComputeService()
		if compute != nil {
			ctx.Export("instance_group_name", pulumi.String(compute.InstanceGroupName))
		}
	} else if locals.IsExternalOrigin {
		external := stackInput.Target.Spec.Backend.GetExternalOrigin()
		if external != nil {
			ctx.Export("external_hostname", pulumi.String(external.Hostname))
		}
	}

	// Export custom domains if configured
	if stackInput.Target.Spec.FrontendConfig != nil && len(stackInput.Target.Spec.FrontendConfig.CustomDomains) > 0 {
		domains := make([]string, len(stackInput.Target.Spec.FrontendConfig.CustomDomains))
		copy(domains, stackInput.Target.Spec.FrontendConfig.CustomDomains)
		ctx.Export("custom_domains", pulumi.ToStringArray(domains))
	}

	// Export monitoring dashboard URL
	monitoringUrl := pulumi.Sprintf(
		"https://console.cloud.google.com/net-services/loadbalancing/details/http/%s?project=%s",
		urlMap.Name,
		stackInput.Target.Spec.GcpProjectId.GetValue(),
	)
	ctx.Export("monitoring_dashboard_url", monitoringUrl)

	return nil
}

// createGcsBackendBucket creates a Backend Bucket for GCS origin with CDN enabled
func createGcsBackendBucket(ctx *pulumi.Context, stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput,
	locals *Locals, gcpProvider pulumi.ProviderResource) (pulumi.StringOutput, pulumi.StringOutput, error) {

	gcsBucket := stackInput.Target.Spec.Backend.GetGcsBucket()
	if gcsBucket == nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, errors.New("gcs_bucket configuration is nil")
	}

	// Build CDN policy
	cdnPolicy := &compute.BackendBucketCdnPolicyArgs{
		CacheMode:       pulumi.String(locals.CacheMode),
		DefaultTtl:      pulumi.Int(locals.DefaultTtl),
		MaxTtl:          pulumi.Int(locals.MaxTtl),
		ClientTtl:       pulumi.Int(locals.ClientTtl),
		NegativeCaching: pulumi.Bool(locals.NegativeCachingEnabled),
	}

	// Add advanced cache key policy if specified
	if stackInput.Target.Spec.AdvancedConfig != nil && stackInput.Target.Spec.AdvancedConfig.CacheKeyPolicy != nil {
		cdnPolicy.CacheKeyPolicy = buildBackendBucketCacheKeyPolicy(stackInput.Target.Spec.AdvancedConfig.CacheKeyPolicy)
	}

	// Add negative caching policies if specified
	if stackInput.Target.Spec.AdvancedConfig != nil && len(stackInput.Target.Spec.AdvancedConfig.NegativeCachingPolicies) > 0 {
		negativeCachingPolicies := make(compute.BackendBucketCdnPolicyNegativeCachingPolicyArray,
			len(stackInput.Target.Spec.AdvancedConfig.NegativeCachingPolicies))
		for i, policy := range stackInput.Target.Spec.AdvancedConfig.NegativeCachingPolicies {
			negativeCachingPolicies[i] = &compute.BackendBucketCdnPolicyNegativeCachingPolicyArgs{
				Code: pulumi.Int(int(policy.Code)),
				Ttl:  pulumi.Int(int(policy.TtlSeconds)),
			}
		}
		cdnPolicy.NegativeCachingPolicies = negativeCachingPolicies
	}

	// Add serve-while-stale if specified
	if stackInput.Target.Spec.AdvancedConfig != nil && stackInput.Target.Spec.AdvancedConfig.ServeWhileStaleSeconds != nil {
		cdnPolicy.ServeWhileStale = pulumi.Int(int(*stackInput.Target.Spec.AdvancedConfig.ServeWhileStaleSeconds))
	}

	// Create Backend Bucket with CDN
	backendBucket, err := compute.NewBackendBucket(ctx, locals.BackendBucketName, &compute.BackendBucketArgs{
		Name:       pulumi.String(locals.BackendBucketName),
		Project:    pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		BucketName: pulumi.String(gcsBucket.BucketName),
		EnableCdn:  pulumi.Bool(true),
		CdnPolicy:  cdnPolicy,
	}, pulumi.Provider(gcpProvider), pulumi.DeleteBeforeReplace(true))

	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	return backendBucket.ID().ToStringOutput(), backendBucket.Name, nil
}

// createComputeBackendService creates a Backend Service for Compute Engine with CDN enabled
func createComputeBackendService(ctx *pulumi.Context, stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput,
	locals *Locals, gcpProvider pulumi.ProviderResource) (pulumi.StringOutput, pulumi.StringOutput, error) {

	computeConfig := stackInput.Target.Spec.Backend.GetComputeService()
	if computeConfig == nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, errors.New("compute_service configuration is nil")
	}

	// Create health check if configured
	var healthCheckId pulumi.StringOutput
	hasHealthCheck := false
	if computeConfig.HealthCheck != nil {
		healthCheck, err := createHealthCheck(ctx, stackInput, locals, computeConfig.HealthCheck, gcpProvider)
		if err != nil {
			return pulumi.StringOutput{}, pulumi.StringOutput{}, errors.Wrap(err, "failed to create health check")
		}
		healthCheckId = healthCheck.ID().ToStringOutput()
		hasHealthCheck = true
	}

	// Determine protocol
	protocol := "HTTP"
	if computeConfig.Protocol != nil && *computeConfig.Protocol == gcpcloudcdnv1.BackendProtocol_HTTPS {
		protocol = "HTTPS"
	}

	// Build CDN policy
	cdnPolicy := buildBackendServiceCdnPolicy(stackInput, locals)

	// Create Backend Service
	backendServiceArgs := &compute.BackendServiceArgs{
		Name:      pulumi.String(locals.BackendServiceName),
		Project:   pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		Protocol:  pulumi.String(protocol),
		PortName:  pulumi.String("http"),
		EnableCdn: pulumi.Bool(true),
		CdnPolicy: cdnPolicy,
	}

	if hasHealthCheck {
		backendServiceArgs.HealthChecks = healthCheckId
	}

	backendService, err := compute.NewBackendService(ctx, locals.BackendServiceName, backendServiceArgs,
		pulumi.Provider(gcpProvider))

	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	return backendService.ID().ToStringOutput(), backendService.Name, nil
}

// createCloudRunBackendService creates a Backend Service for Cloud Run with CDN enabled
func createCloudRunBackendService(ctx *pulumi.Context, stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput,
	locals *Locals, gcpProvider pulumi.ProviderResource) (pulumi.StringOutput, pulumi.StringOutput, error) {

	cloudRunConfig := stackInput.Target.Spec.Backend.GetCloudRunService()
	if cloudRunConfig == nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, errors.New("cloud_run_service configuration is nil")
	}

	// Create Serverless Network Endpoint Group for Cloud Run
	neg, err := compute.NewRegionNetworkEndpointGroup(ctx, locals.BackendServiceName+"-neg", &compute.RegionNetworkEndpointGroupArgs{
		Name:                pulumi.String(locals.BackendServiceName + "-neg"),
		Project:             pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		Region:              pulumi.String(cloudRunConfig.Region),
		NetworkEndpointType: pulumi.String("SERVERLESS"),
		CloudRun: &compute.RegionNetworkEndpointGroupCloudRunArgs{
			Service: pulumi.String(cloudRunConfig.ServiceName),
		},
	}, pulumi.Provider(gcpProvider))

	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, errors.Wrap(err, "failed to create Cloud Run NEG")
	}

	// Build CDN policy
	cdnPolicy := buildBackendServiceCdnPolicy(stackInput, locals)

	// Create Backend Service with Cloud Run NEG
	backendService, err := compute.NewBackendService(ctx, locals.BackendServiceName, &compute.BackendServiceArgs{
		Name:      pulumi.String(locals.BackendServiceName),
		Project:   pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		Protocol:  pulumi.String("HTTPS"),
		EnableCdn: pulumi.Bool(true),
		CdnPolicy: cdnPolicy,
		Backends: compute.BackendServiceBackendArray{
			&compute.BackendServiceBackendArgs{
				Group: neg.ID(),
			},
		},
	}, pulumi.Provider(gcpProvider))

	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	return backendService.ID().ToStringOutput(), backendService.Name, nil
}

// createExternalBackendService creates a Backend Service for external origin with CDN enabled
func createExternalBackendService(ctx *pulumi.Context, stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput,
	locals *Locals, gcpProvider pulumi.ProviderResource) (pulumi.StringOutput, pulumi.StringOutput, error) {

	externalConfig := stackInput.Target.Spec.Backend.GetExternalOrigin()
	if externalConfig == nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, errors.New("external_origin configuration is nil")
	}

	// Determine protocol and port
	protocol := "HTTPS"
	port := 443
	if externalConfig.Protocol != nil && *externalConfig.Protocol == gcpcloudcdnv1.BackendProtocol_HTTP {
		protocol = "HTTP"
		port = 80
	}
	if externalConfig.Port != nil {
		port = int(*externalConfig.Port)
	}

	// Create Internet Network Endpoint Group
	neg, err := compute.NewNetworkEndpointGroup(ctx, locals.BackendServiceName+"-neg", &compute.NetworkEndpointGroupArgs{
		Name:                pulumi.String(locals.BackendServiceName + "-neg"),
		Project:             pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		NetworkEndpointType: pulumi.String("INTERNET_FQDN_PORT"),
		DefaultPort:         pulumi.Int(port),
	}, pulumi.Provider(gcpProvider))

	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, errors.Wrap(err, "failed to create Internet NEG")
	}

	// Note: For INTERNET_FQDN_PORT NEGs, the FQDN is configured via the NEG itself,
	// not via individual NetworkEndpoint resources. The backend service will use
	// the external hostname directly through the NEG configuration.

	// Build CDN policy
	cdnPolicy := buildBackendServiceCdnPolicy(stackInput, locals)

	// Create Backend Service with Internet NEG
	backendService, err := compute.NewBackendService(ctx, locals.BackendServiceName, &compute.BackendServiceArgs{
		Name:      pulumi.String(locals.BackendServiceName),
		Project:   pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		Protocol:  pulumi.String(protocol),
		EnableCdn: pulumi.Bool(true),
		CdnPolicy: cdnPolicy,
		Backends: compute.BackendServiceBackendArray{
			&compute.BackendServiceBackendArgs{
				Group: neg.ID(),
			},
		},
	}, pulumi.Provider(gcpProvider))

	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	return backendService.ID().ToStringOutput(), backendService.Name, nil
}

// buildBackendServiceCdnPolicy creates CDN policy for Backend Service
func buildBackendServiceCdnPolicy(stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput, locals *Locals) *compute.BackendServiceCdnPolicyArgs {
	cdnPolicy := &compute.BackendServiceCdnPolicyArgs{
		CacheMode:       pulumi.String(locals.CacheMode),
		DefaultTtl:      pulumi.Int(locals.DefaultTtl),
		MaxTtl:          pulumi.Int(locals.MaxTtl),
		ClientTtl:       pulumi.Int(locals.ClientTtl),
		NegativeCaching: pulumi.Bool(locals.NegativeCachingEnabled),
	}

	// Add advanced configuration if specified
	if stackInput.Target.Spec.AdvancedConfig != nil {
		advancedConfig := stackInput.Target.Spec.AdvancedConfig

		// Cache key policy
		if advancedConfig.CacheKeyPolicy != nil {
			cdnPolicy.CacheKeyPolicy = buildBackendServiceCacheKeyPolicy(advancedConfig.CacheKeyPolicy)
		}

		// Negative caching policies
		if len(advancedConfig.NegativeCachingPolicies) > 0 {
			negativeCachingPolicies := make(compute.BackendServiceCdnPolicyNegativeCachingPolicyArray,
				len(advancedConfig.NegativeCachingPolicies))
			for i, policy := range advancedConfig.NegativeCachingPolicies {
				negativeCachingPolicies[i] = &compute.BackendServiceCdnPolicyNegativeCachingPolicyArgs{
					Code: pulumi.Int(int(policy.Code)),
					Ttl:  pulumi.Int(int(policy.TtlSeconds)),
				}
			}
			cdnPolicy.NegativeCachingPolicies = negativeCachingPolicies
		}

		// Serve-while-stale
		if advancedConfig.ServeWhileStaleSeconds != nil {
			cdnPolicy.ServeWhileStale = pulumi.Int(int(*advancedConfig.ServeWhileStaleSeconds))
		}
	}

	return cdnPolicy
}

// buildBackendBucketCacheKeyPolicy creates cache key policy configuration for Backend Buckets
func buildBackendBucketCacheKeyPolicy(policy *gcpcloudcdnv1.CacheKeyPolicy) *compute.BackendBucketCdnPolicyCacheKeyPolicyArgs {
	cacheKeyPolicy := &compute.BackendBucketCdnPolicyCacheKeyPolicyArgs{}

	// Backend Bucket cache key policy only supports IncludeHttpHeaders and QueryStringWhitelists
	if len(policy.QueryStringWhitelist) > 0 {
		whitelist := make([]string, len(policy.QueryStringWhitelist))
		copy(whitelist, policy.QueryStringWhitelist)
		cacheKeyPolicy.QueryStringWhitelists = pulumi.ToStringArray(whitelist)
	}

	return cacheKeyPolicy
}

// buildBackendServiceCacheKeyPolicy creates cache key policy configuration for Backend Services
func buildBackendServiceCacheKeyPolicy(policy *gcpcloudcdnv1.CacheKeyPolicy) *compute.BackendServiceCdnPolicyCacheKeyPolicyArgs {
	cacheKeyPolicy := &compute.BackendServiceCdnPolicyCacheKeyPolicyArgs{}

	if policy.IncludeQueryString != nil {
		cacheKeyPolicy.IncludeQueryString = pulumi.Bool(*policy.IncludeQueryString)
	}

	if len(policy.QueryStringWhitelist) > 0 {
		whitelist := make([]string, len(policy.QueryStringWhitelist))
		copy(whitelist, policy.QueryStringWhitelist)
		cacheKeyPolicy.QueryStringWhitelists = pulumi.ToStringArray(whitelist)
	}

	if policy.IncludeProtocol != nil {
		cacheKeyPolicy.IncludeProtocol = pulumi.Bool(*policy.IncludeProtocol)
	}

	if policy.IncludeHost != nil {
		cacheKeyPolicy.IncludeHost = pulumi.Bool(*policy.IncludeHost)
	}

	return cacheKeyPolicy
}

// createHealthCheck creates a health check for backend instances
func createHealthCheck(ctx *pulumi.Context, stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput,
	locals *Locals, healthCheckConfig *gcpcloudcdnv1.HealthCheckConfig,
	gcpProvider pulumi.ProviderResource) (*compute.HealthCheck, error) {

	// Build HTTP health check configuration
	httpHealthCheck := &compute.HealthCheckHttpHealthCheckArgs{
		Port: pulumi.Int(80),
	}

	// Set path if specified
	if healthCheckConfig.Path != nil && *healthCheckConfig.Path != "" {
		httpHealthCheck.RequestPath = pulumi.String(*healthCheckConfig.Path)
	} else {
		httpHealthCheck.RequestPath = pulumi.String("/")
	}

	// Set port if specified
	if healthCheckConfig.Port != nil {
		httpHealthCheck.Port = pulumi.Int(int(*healthCheckConfig.Port))
	}

	healthCheckArgs := &compute.HealthCheckArgs{
		Name:            pulumi.String(locals.HealthCheckName),
		Project:         pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		HttpHealthCheck: httpHealthCheck,
	}

	// Set intervals and thresholds
	if healthCheckConfig.CheckIntervalSeconds != nil {
		healthCheckArgs.CheckIntervalSec = pulumi.Int(int(*healthCheckConfig.CheckIntervalSeconds))
	}

	if healthCheckConfig.TimeoutSeconds != nil {
		healthCheckArgs.TimeoutSec = pulumi.Int(int(*healthCheckConfig.TimeoutSeconds))
	}

	if healthCheckConfig.HealthyThreshold != nil {
		healthCheckArgs.HealthyThreshold = pulumi.Int(int(*healthCheckConfig.HealthyThreshold))
	}

	if healthCheckConfig.UnhealthyThreshold != nil {
		healthCheckArgs.UnhealthyThreshold = pulumi.Int(int(*healthCheckConfig.UnhealthyThreshold))
	}

	return compute.NewHealthCheck(ctx, locals.HealthCheckName, healthCheckArgs, pulumi.Provider(gcpProvider))
}

// createSslCertificate creates SSL certificate (Google-managed or self-managed)
func createSslCertificate(ctx *pulumi.Context, stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput,
	locals *Locals, gcpProvider pulumi.ProviderResource) (*compute.ManagedSslCertificate, error) {

	sslConfig := stackInput.Target.Spec.FrontendConfig.SslCertificate

	// Google-managed certificate
	if sslConfig.GetGoogleManaged() != nil {
		googleManaged := sslConfig.GetGoogleManaged()

		domains := make([]string, len(googleManaged.Domains))
		copy(domains, googleManaged.Domains)

		return compute.NewManagedSslCertificate(ctx, locals.SslCertName, &compute.ManagedSslCertificateArgs{
			Name:    pulumi.String(locals.SslCertName),
			Project: pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
			Managed: &compute.ManagedSslCertificateManagedArgs{
				Domains: pulumi.ToStringArray(domains),
			},
		}, pulumi.Provider(gcpProvider))
	}

	// Self-managed certificate (not implemented yet - would use compute.SslCertificate)
	return nil, errors.New("self-managed SSL certificates not yet implemented")
}

// createHttpRedirect creates HTTP to HTTPS redirect
func createHttpRedirect(ctx *pulumi.Context, stackInput *gcpcloudcdnv1.GcpCloudCdnStackInput,
	locals *Locals, globalIp *compute.GlobalAddress, gcpProvider pulumi.ProviderResource) error {

	// Create redirect URL map
	redirectUrlMap, err := compute.NewURLMap(ctx, locals.UrlMapName+"-redirect", &compute.URLMapArgs{
		Name:    pulumi.String(locals.UrlMapName + "-redirect"),
		Project: pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		DefaultUrlRedirect: &compute.URLMapDefaultUrlRedirectArgs{
			HttpsRedirect:        pulumi.Bool(true),
			RedirectResponseCode: pulumi.String("MOVED_PERMANENTLY_DEFAULT"),
		},
	}, pulumi.Provider(gcpProvider))

	if err != nil {
		return err
	}

	// Create HTTP proxy for redirect
	httpProxy, err := compute.NewTargetHttpProxy(ctx, locals.HttpProxyName+"-redirect", &compute.TargetHttpProxyArgs{
		Name:    pulumi.String(locals.HttpProxyName + "-redirect"),
		Project: pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		UrlMap:  redirectUrlMap.ID(),
	}, pulumi.Provider(gcpProvider))

	if err != nil {
		return err
	}

	// Create HTTP forwarding rule for redirect
	_, err = compute.NewGlobalForwardingRule(ctx, locals.HttpProxyName+"-redirect-rule", &compute.GlobalForwardingRuleArgs{
		Name:      pulumi.String(locals.HttpProxyName + "-redirect-rule"),
		Project:   pulumi.String(stackInput.Target.Spec.GcpProjectId.GetValue()),
		Target:    httpProxy.ID(),
		PortRange: pulumi.String("80"),
		IpAddress: globalIp.Address,
	}, pulumi.Provider(gcpProvider))

	return err
}
