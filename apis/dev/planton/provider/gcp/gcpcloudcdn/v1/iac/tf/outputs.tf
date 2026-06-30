# outputs.tf - Terraform outputs matching GcpCloudCdnStackOutputs proto schema

output "cdn_url" {
  description = "URL of the Cloud CDN endpoint (load balancer frontend)"
  value       = "https://${google_compute_global_address.cdn_ip.address}"
}

output "global_ip_address" {
  description = "Global IP address assigned to the load balancer"
  value       = google_compute_global_address.cdn_ip.address
}

output "backend_name" {
  description = "Name of the backend resource (BackendBucket or BackendService)"
  value = local.is_gcs_bucket ? google_compute_backend_bucket.gcs_backend[0].name : (
    local.is_compute_service ? google_compute_backend_service.compute_backend[0].name : (
      local.is_cloud_run ? google_compute_backend_service.cloud_run_backend[0].name : (
        local.is_external_origin ? google_compute_backend_service.external_backend[0].name : null
      )
    )
  )
}

output "backend_id" {
  description = "Full resource ID of the backend"
  value = local.is_gcs_bucket ? google_compute_backend_bucket.gcs_backend[0].id : (
    local.is_compute_service ? google_compute_backend_service.compute_backend[0].id : (
      local.is_cloud_run ? google_compute_backend_service.cloud_run_backend[0].id : (
        local.is_external_origin ? google_compute_backend_service.external_backend[0].id : null
      )
    )
  )
}

output "cdn_enabled" {
  description = "Whether Cloud CDN is enabled on the backend"
  value       = true
}

output "cache_mode" {
  description = "Cache mode configured for this CDN"
  value       = local.cache_mode_string
}

output "url_map_name" {
  description = "URL map name for load balancer routing configuration"
  value       = google_compute_url_map.cdn_url_map.name
}

output "https_proxy_name" {
  description = "Target HTTPS proxy name (if HTTPS is configured)"
  value       = local.has_frontend_config ? google_compute_target_https_proxy.cdn_https_proxy[0].name : null
}

output "ssl_certificate_name" {
  description = "SSL certificate name or ID (if configured)"
  value       = local.has_ssl_config && var.spec.frontend_config.ssl_certificate.google_managed != null ? google_compute_managed_ssl_certificate.cdn_cert[0].name : null
}

output "cloud_armor_policy_name" {
  description = "Cloud Armor security policy name (if Cloud Armor is enabled)"
  value       = local.has_frontend_config && var.spec.frontend_config.cloud_armor != null && var.spec.frontend_config.cloud_armor.enabled ? var.spec.frontend_config.cloud_armor.security_policy_name : null
}

output "backend_type" {
  description = "Backend type (GCS_BUCKET, COMPUTE_SERVICE, CLOUD_RUN, EXTERNAL)"
  value       = local.backend_type
}

output "gcs_bucket_name" {
  description = "GCS bucket name (only populated if backend_type is GCS_BUCKET)"
  value       = local.is_gcs_bucket ? var.spec.backend.gcs_bucket.bucket_name : null
}

output "instance_group_name" {
  description = "Compute Engine instance group name (only populated if backend_type is COMPUTE_SERVICE)"
  value       = local.is_compute_service ? var.spec.backend.compute_service.instance_group_name : null
}

output "cloud_run_service_name" {
  description = "Cloud Run service name (only populated if backend_type is CLOUD_RUN)"
  value       = local.is_cloud_run ? var.spec.backend.cloud_run_service.service_name : null
}

output "cloud_run_region" {
  description = "Cloud Run service region (only populated if backend_type is CLOUD_RUN)"
  value       = local.is_cloud_run ? var.spec.backend.cloud_run_service.region : null
}

output "external_hostname" {
  description = "External origin hostname (only populated if backend_type is EXTERNAL)"
  value       = local.is_external_origin ? var.spec.backend.external_origin.hostname : null
}

output "custom_domains" {
  description = "Custom domains configured for this CDN"
  value       = local.has_custom_domains ? var.spec.frontend_config.custom_domains : []
}

output "health_check_url" {
  description = "Health check URL (if health check is configured)"
  value = local.is_compute_service && var.spec.backend.compute_service.health_check != null ? format(
    "http://%s:%d%s",
    "backend-instances",
    coalesce(var.spec.backend.compute_service.health_check.port, local.compute_port),
    coalesce(var.spec.backend.compute_service.health_check.path, "/")
  ) : null
}

output "monitoring_dashboard_url" {
  description = "Cloud Console link to view CDN metrics"
  value = format(
    "https://console.cloud.google.com/net-services/loadbalancing/details/http/%s?project=%s",
    google_compute_url_map.cdn_url_map.name,
    var.spec.gcp_project_id
  )
}

