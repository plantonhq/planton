# main.tf - Main Terraform configuration for GCP Cloud CDN

# GCS Backend Bucket with CDN (80% use case)
resource "google_compute_backend_bucket" "gcs_backend" {
  count = local.is_gcs_bucket ? 1 : 0
  
  name        = local.backend_bucket_name
  project     = var.spec.gcp_project_id.value
  bucket_name = var.spec.backend.gcs_bucket.bucket_name
  enable_cdn  = true
  
  cdn_policy {
    cache_mode       = local.cache_mode_string
    default_ttl      = local.default_ttl_seconds
    max_ttl          = local.max_ttl_seconds
    client_ttl       = local.client_ttl_seconds
    negative_caching = local.negative_caching_enabled
    
    # Cache key policy (if configured)
    dynamic "cache_key_policy" {
      for_each = local.cache_key_policy != null ? [local.cache_key_policy] : []
      content {
        include_query_string   = cache_key_policy.value.include_query_string
        query_string_whitelist = cache_key_policy.value.query_string_whitelist
        include_protocol       = cache_key_policy.value.include_protocol
        include_host           = cache_key_policy.value.include_host
      }
    }
    
    # Negative caching policies
    dynamic "negative_caching_policy" {
      for_each = local.negative_caching_policies
      content {
        code = negative_caching_policy.value.code
        ttl  = negative_caching_policy.value.ttl_seconds
      }
    }
    
    # Serve-while-stale
    serve_while_stale = local.serve_while_stale_seconds
  }
}

# Health Check for Compute Engine Backend
resource "google_compute_health_check" "compute_health_check" {
  count = local.is_compute_service && var.spec.backend.compute_service.health_check != null ? 1 : 0
  
  name    = local.health_check_name
  project = var.spec.gcp_project_id.value
  
  http_health_check {
    port         = coalesce(var.spec.backend.compute_service.health_check.port, local.compute_port)
    request_path = coalesce(var.spec.backend.compute_service.health_check.path, "/")
  }
  
  check_interval_sec  = coalesce(var.spec.backend.compute_service.health_check.check_interval_seconds, 5)
  timeout_sec         = coalesce(var.spec.backend.compute_service.health_check.timeout_seconds, 5)
  healthy_threshold   = coalesce(var.spec.backend.compute_service.health_check.healthy_threshold, 2)
  unhealthy_threshold = coalesce(var.spec.backend.compute_service.health_check.unhealthy_threshold, 2)
}

# Compute Engine Backend Service with CDN
resource "google_compute_backend_service" "compute_backend" {
  count = local.is_compute_service ? 1 : 0
  
  name     = local.backend_service_name
  project  = var.spec.gcp_project_id.value
  protocol = local.compute_protocol_string
  
  enable_cdn = true
  
  cdn_policy {
    cache_mode       = local.cache_mode_string
    default_ttl      = local.default_ttl_seconds
    max_ttl          = local.max_ttl_seconds
    client_ttl       = local.client_ttl_seconds
    negative_caching = local.negative_caching_enabled
    
    # Cache key policy (if configured)
    dynamic "cache_key_policy" {
      for_each = local.cache_key_policy != null ? [local.cache_key_policy] : []
      content {
        include_query_string   = cache_key_policy.value.include_query_string
        query_string_whitelist = cache_key_policy.value.query_string_whitelist
        include_protocol       = cache_key_policy.value.include_protocol
        include_host           = cache_key_policy.value.include_host
      }
    }
    
    # Negative caching policies
    dynamic "negative_caching_policy" {
      for_each = local.negative_caching_policies
      content {
        code = negative_caching_policy.value.code
        ttl  = negative_caching_policy.value.ttl_seconds
      }
    }
    
    # Serve-while-stale
    serve_while_stale = local.serve_while_stale_seconds
  }
  
  # Health check
  health_checks = local.is_compute_service && var.spec.backend.compute_service.health_check != null ? [
    google_compute_health_check.compute_health_check[0].id
  ] : []
  
  # Note: Backends (instance groups) are not managed by this module
  # Users must attach instance groups separately after creation
}

# Cloud Run Serverless Network Endpoint Group
resource "google_compute_region_network_endpoint_group" "cloud_run_neg" {
  count = local.is_cloud_run ? 1 : 0
  
  name    = "${local.backend_service_name}-neg"
  project = var.spec.gcp_project_id.value
  region  = var.spec.backend.cloud_run_service.region
  
  network_endpoint_type = "SERVERLESS"
  
  cloud_run {
    service = var.spec.backend.cloud_run_service.service_name
  }
}

# Cloud Run Backend Service with CDN
resource "google_compute_backend_service" "cloud_run_backend" {
  count = local.is_cloud_run ? 1 : 0
  
  name     = local.backend_service_name
  project  = var.spec.gcp_project_id.value
  protocol = "HTTPS"
  
  enable_cdn = true
  
  cdn_policy {
    cache_mode       = local.cache_mode_string
    default_ttl      = local.default_ttl_seconds
    max_ttl          = local.max_ttl_seconds
    client_ttl       = local.client_ttl_seconds
    negative_caching = local.negative_caching_enabled
    
    # Cache key policy (if configured)
    dynamic "cache_key_policy" {
      for_each = local.cache_key_policy != null ? [local.cache_key_policy] : []
      content {
        include_query_string   = cache_key_policy.value.include_query_string
        query_string_whitelist = cache_key_policy.value.query_string_whitelist
        include_protocol       = cache_key_policy.value.include_protocol
        include_host           = cache_key_policy.value.include_host
      }
    }
    
    # Negative caching policies
    dynamic "negative_caching_policy" {
      for_each = local.negative_caching_policies
      content {
        code = negative_caching_policy.value.code
        ttl  = negative_caching_policy.value.ttl_seconds
      }
    }
    
    # Serve-while-stale
    serve_while_stale = local.serve_while_stale_seconds
  }
  
  backend {
    group = google_compute_region_network_endpoint_group.cloud_run_neg[0].id
  }
}

# External Origin Internet Network Endpoint Group
resource "google_compute_global_network_endpoint_group" "external_neg" {
  count = local.is_external_origin ? 1 : 0
  
  name                  = "${local.backend_service_name}-neg"
  project               = var.spec.gcp_project_id.value
  network_endpoint_type = "INTERNET_FQDN_PORT"
  default_port          = local.external_port
}

# External Origin Network Endpoint
resource "google_compute_global_network_endpoint" "external_endpoint" {
  count = local.is_external_origin ? 1 : 0
  
  project                       = var.spec.gcp_project_id.value
  global_network_endpoint_group = google_compute_global_network_endpoint_group.external_neg[0].name
  fqdn                          = var.spec.backend.external_origin.hostname
  port                          = local.external_port
}

# External Origin Backend Service with CDN
resource "google_compute_backend_service" "external_backend" {
  count = local.is_external_origin ? 1 : 0
  
  name     = local.backend_service_name
  project  = var.spec.gcp_project_id.value
  protocol = local.external_protocol_string
  
  enable_cdn = true
  
  cdn_policy {
    cache_mode       = local.cache_mode_string
    default_ttl      = local.default_ttl_seconds
    max_ttl          = local.max_ttl_seconds
    client_ttl       = local.client_ttl_seconds
    negative_caching = local.negative_caching_enabled
    
    # Cache key policy (if configured)
    dynamic "cache_key_policy" {
      for_each = local.cache_key_policy != null ? [local.cache_key_policy] : []
      content {
        include_query_string   = cache_key_policy.value.include_query_string
        query_string_whitelist = cache_key_policy.value.query_string_whitelist
        include_protocol       = cache_key_policy.value.include_protocol
        include_host           = cache_key_policy.value.include_host
      }
    }
    
    # Negative caching policies
    dynamic "negative_caching_policy" {
      for_each = local.negative_caching_policies
      content {
        code = negative_caching_policy.value.code
        ttl  = negative_caching_policy.value.ttl_seconds
      }
    }
    
    # Serve-while-stale
    serve_while_stale = local.serve_while_stale_seconds
  }
  
  backend {
    group = google_compute_global_network_endpoint_group.external_neg[0].id
  }
}

# Global Static IP Address
resource "google_compute_global_address" "cdn_ip" {
  name    = local.global_address_name
  project = var.spec.gcp_project_id.value
}

# URL Map (routing configuration)
resource "google_compute_url_map" "cdn_url_map" {
  name    = local.url_map_name
  project = var.spec.gcp_project_id.value
  
  default_service = local.is_gcs_bucket ? google_compute_backend_bucket.gcs_backend[0].id : (
    local.is_compute_service ? google_compute_backend_service.compute_backend[0].id : (
      local.is_cloud_run ? google_compute_backend_service.cloud_run_backend[0].id : (
        local.is_external_origin ? google_compute_backend_service.external_backend[0].id : null
      )
    )
  )
}

# Google-Managed SSL Certificate
resource "google_compute_managed_ssl_certificate" "cdn_cert" {
  count = local.has_ssl_config && var.spec.frontend_config.ssl_certificate.google_managed != null ? 1 : 0
  
  name    = local.ssl_cert_name
  project = var.spec.gcp_project_id.value
  
  managed {
    domains = var.spec.frontend_config.ssl_certificate.google_managed.domains
  }
}

# HTTPS Target Proxy
resource "google_compute_target_https_proxy" "cdn_https_proxy" {
  count = local.has_frontend_config ? 1 : 0
  
  name    = local.https_proxy_name
  project = var.spec.gcp_project_id.value
  url_map = google_compute_url_map.cdn_url_map.id
  
  ssl_certificates = local.has_ssl_config ? [
    google_compute_managed_ssl_certificate.cdn_cert[0].id
  ] : []
}

# HTTPS Global Forwarding Rule
resource "google_compute_global_forwarding_rule" "cdn_https_rule" {
  count = local.has_frontend_config ? 1 : 0
  
  name       = "${local.https_proxy_name}-rule"
  project    = var.spec.gcp_project_id.value
  target     = google_compute_target_https_proxy.cdn_https_proxy[0].id
  port_range = "443"
  ip_address = google_compute_global_address.cdn_ip.address
}

# HTTP to HTTPS Redirect URL Map
resource "google_compute_url_map" "http_redirect" {
  count = local.enable_https_redirect ? 1 : 0
  
  name    = "${local.url_map_name}-redirect"
  project = var.spec.gcp_project_id.value
  
  default_url_redirect {
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
  }
}

# HTTP Target Proxy for Redirect
resource "google_compute_target_http_proxy" "http_redirect_proxy" {
  count = local.enable_https_redirect ? 1 : 0
  
  name    = "${local.http_proxy_name}-redirect"
  project = var.spec.gcp_project_id.value
  url_map = google_compute_url_map.http_redirect[0].id
}

# HTTP Global Forwarding Rule for Redirect
resource "google_compute_global_forwarding_rule" "http_redirect_rule" {
  count = local.enable_https_redirect ? 1 : 0
  
  name       = "${local.http_proxy_name}-redirect-rule"
  project    = var.spec.gcp_project_id.value
  target     = google_compute_target_http_proxy.http_redirect_proxy[0].id
  port_range = "80"
  ip_address = google_compute_global_address.cdn_ip.address
}

# HTTP Target Proxy (if no HTTPS configured)
resource "google_compute_target_http_proxy" "cdn_http_proxy" {
  count = !local.has_frontend_config ? 1 : 0
  
  name    = local.http_proxy_name
  project = var.spec.gcp_project_id.value
  url_map = google_compute_url_map.cdn_url_map.id
}

# HTTP Global Forwarding Rule (if no HTTPS configured)
resource "google_compute_global_forwarding_rule" "cdn_http_rule" {
  count = !local.has_frontend_config ? 1 : 0
  
  name       = "${local.http_proxy_name}-rule"
  project    = var.spec.gcp_project_id.value
  target     = google_compute_target_http_proxy.cdn_http_proxy[0].id
  port_range = "80"
  ip_address = google_compute_global_address.cdn_ip.address
}

