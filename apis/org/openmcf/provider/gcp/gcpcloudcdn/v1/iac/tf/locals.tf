# locals.tf - Local value transformations and computed values for GCP Cloud CDN

locals {
  # GCP labels from metadata
  gcp_labels = merge(
    {
      "planton-resource-type" = "gcp-cloud-cdn"
    },
    var.metadata.labels
  )
  
  # Resource naming based on metadata
  cdn_name                = var.metadata.name
  backend_bucket_name     = "${var.metadata.name}-backend-bucket"
  backend_service_name    = "${var.metadata.name}-backend-service"
  health_check_name       = "${var.metadata.name}-health-check"
  url_map_name            = "${var.metadata.name}-url-map"
  https_proxy_name        = "${var.metadata.name}-https-proxy"
  http_proxy_name         = "${var.metadata.name}-http-proxy"
  global_address_name     = "${var.metadata.name}-global-ip"
  ssl_cert_name           = "${var.metadata.name}-ssl-cert"
  
  # Backend type detection
  is_gcs_bucket      = var.spec.backend.gcs_bucket != null
  is_compute_service = var.spec.backend.compute_service != null
  is_cloud_run       = var.spec.backend.cloud_run_service != null
  is_external_origin = var.spec.backend.external_origin != null
  
  # Backend type string
  backend_type = (
    local.is_gcs_bucket ? "GCS_BUCKET" :
    local.is_compute_service ? "COMPUTE_SERVICE" :
    local.is_cloud_run ? "CLOUD_RUN" :
    local.is_external_origin ? "EXTERNAL" :
    "UNKNOWN"
  )
  
  # Cache configuration with defaults
  cache_mode = coalesce(
    var.spec.cache_mode,
    1  # CACHE_ALL_STATIC (enum value 1)
  )
  
  cache_mode_string = (
    local.cache_mode == 1 ? "CACHE_ALL_STATIC" :
    local.cache_mode == 2 ? "USE_ORIGIN_HEADERS" :
    local.cache_mode == 3 ? "FORCE_CACHE_ALL" :
    "CACHE_ALL_STATIC"  # Default fallback
  )
  
  default_ttl_seconds = coalesce(var.spec.default_ttl_seconds, 3600)      # 1 hour default
  max_ttl_seconds     = coalesce(var.spec.max_ttl_seconds, 86400)         # 1 day default
  client_ttl_seconds  = coalesce(var.spec.client_ttl_seconds, local.max_ttl_seconds)
  
  negative_caching_enabled = coalesce(var.spec.enable_negative_caching, false)
  
  # Frontend configuration flags
  has_frontend_config = var.spec.frontend_config != null
  has_custom_domains  = local.has_frontend_config && length(var.spec.frontend_config.custom_domains) > 0
  has_ssl_config      = local.has_frontend_config && var.spec.frontend_config.ssl_certificate != null
  enable_https_redirect = local.has_frontend_config ? coalesce(
    var.spec.frontend_config.enable_https_redirect,
    true  # Default: enable HTTPS redirect
  ) : false
  
  # Advanced configuration
  has_advanced_config    = var.spec.advanced_config != null
  has_cache_key_policy   = local.has_advanced_config && var.spec.advanced_config.cache_key_policy != null
  has_signed_url_config  = local.has_advanced_config && var.spec.advanced_config.signed_url_config != null
  
  # CDN policy for backend bucket
  backend_bucket_cdn_policy = {
    cache_mode              = local.cache_mode_string
    default_ttl             = local.default_ttl_seconds
    max_ttl                 = local.max_ttl_seconds
    client_ttl              = local.client_ttl_seconds
    negative_caching        = local.negative_caching_enabled
  }
  
  # CDN policy for backend service
  backend_service_cdn_policy = {
    cache_mode       = local.cache_mode_string
    default_ttl      = local.default_ttl_seconds
    max_ttl          = local.max_ttl_seconds
    client_ttl       = local.client_ttl_seconds
    negative_caching = local.negative_caching_enabled
  }
  
  # Cache key policy (if configured)
  cache_key_policy = local.has_cache_key_policy ? {
    include_query_string = coalesce(
      var.spec.advanced_config.cache_key_policy.include_query_string,
      true
    )
    query_string_whitelist = coalesce(
      var.spec.advanced_config.cache_key_policy.query_string_whitelist,
      []
    )
    include_protocol = coalesce(
      var.spec.advanced_config.cache_key_policy.include_protocol,
      true
    )
    include_host = coalesce(
      var.spec.advanced_config.cache_key_policy.include_host,
      true
    )
  } : null
  
  # Negative caching policies (if configured)
  negative_caching_policies = local.has_advanced_config ? coalesce(
    var.spec.advanced_config.negative_caching_policies,
    []
  ) : []
  
  # Serve-while-stale configuration
  serve_while_stale_seconds = local.has_advanced_config ? coalesce(
    var.spec.advanced_config.serve_while_stale_seconds,
    0
  ) : 0
  
  # Compute Engine backend configuration
  compute_protocol = local.is_compute_service ? coalesce(
    var.spec.backend.compute_service.protocol,
    1  # HTTP (enum value 1)
  ) : 1
  
  compute_protocol_string = local.compute_protocol == 2 ? "HTTPS" : "HTTP"
  
  compute_port = local.is_compute_service ? coalesce(
    var.spec.backend.compute_service.port,
    local.compute_protocol == 2 ? 443 : 80
  ) : 80
  
  # External origin configuration
  external_protocol = local.is_external_origin ? coalesce(
    var.spec.backend.external_origin.protocol,
    2  # HTTPS (enum value 2, recommended default)
  ) : 2
  
  external_protocol_string = local.external_protocol == 1 ? "HTTP" : "HTTPS"
  
  external_port = local.is_external_origin ? coalesce(
    var.spec.backend.external_origin.port,
    local.external_protocol == 1 ? 80 : 443
  ) : 443
}

