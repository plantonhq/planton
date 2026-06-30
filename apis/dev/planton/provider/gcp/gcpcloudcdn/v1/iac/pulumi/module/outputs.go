package module

// Output constant names for GCP Cloud CDN stack outputs
const (
	// CdnUrlOutput is the URL of the Cloud CDN endpoint (load balancer frontend)
	CdnUrlOutput = "cdn_url"

	// GlobalIpAddressOutput is the global IP address assigned to the load balancer
	GlobalIpAddressOutput = "global_ip_address"

	// BackendNameOutput is the name of the backend resource (BackendBucket or BackendService)
	BackendNameOutput = "backend_name"

	// BackendIdOutput is the full resource ID of the backend
	BackendIdOutput = "backend_id"

	// CdnEnabledOutput indicates whether Cloud CDN is enabled
	CdnEnabledOutput = "cdn_enabled"

	// CacheModeOutput is the cache mode configured for this CDN
	CacheModeOutput = "cache_mode"

	// UrlMapNameOutput is the URL map name for load balancer routing
	UrlMapNameOutput = "url_map_name"

	// HttpsProxyNameOutput is the target HTTPS proxy name
	HttpsProxyNameOutput = "https_proxy_name"

	// SslCertificateNameOutput is the SSL certificate name or ID
	SslCertificateNameOutput = "ssl_certificate_name"

	// CloudArmorPolicyNameOutput is the Cloud Armor security policy name (if enabled)
	CloudArmorPolicyNameOutput = "cloud_armor_policy_name"

	// BackendTypeOutput indicates the backend type (GCS_BUCKET, COMPUTE_SERVICE, CLOUD_RUN, EXTERNAL)
	BackendTypeOutput = "backend_type"

	// GcsBucketNameOutput is the GCS bucket name (only for GCS backend)
	GcsBucketNameOutput = "gcs_bucket_name"

	// InstanceGroupNameOutput is the Compute Engine instance group name (only for Compute backend)
	InstanceGroupNameOutput = "instance_group_name"

	// CloudRunServiceNameOutput is the Cloud Run service name (only for Cloud Run backend)
	CloudRunServiceNameOutput = "cloud_run_service_name"

	// CloudRunRegionOutput is the Cloud Run service region (only for Cloud Run backend)
	CloudRunRegionOutput = "cloud_run_region"

	// ExternalHostnameOutput is the external origin hostname (only for External backend)
	ExternalHostnameOutput = "external_hostname"

	// CustomDomainsOutput is the list of custom domains configured for this CDN
	CustomDomainsOutput = "custom_domains"

	// HealthCheckUrlOutput is the health check URL (if configured)
	HealthCheckUrlOutput = "health_check_url"

	// MonitoringDashboardUrlOutput is the Cloud Console link to view CDN metrics
	MonitoringDashboardUrlOutput = "monitoring_dashboard_url"
)
