variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareZoneSettings specification"
  type = object({
    zone_id = string
    managed_request_headers = optional(list(object({
      id      = string
      enabled = optional(bool, false)
    })), [])
    managed_response_headers = optional(list(object({
      id      = string
      enabled = optional(bool, false)
    })), [])
    url_normalization = optional(object({
      scope = string
      type  = string
    }))
    origin_cloud_regions = optional(list(object({
      origin_ip = string
      region    = string
      vendor    = string
    })), [])
    waiting_room_crawler_bypass = optional(bool)
    zero_rtt                    = optional(bool)
    advanced_ddos               = optional(bool)
    aegis = optional(object({
      enabled = optional(bool)
      pool_id = optional(string, "")
    }))
    always_online            = optional(bool)
    always_use_https         = optional(bool)
    automatic_https_rewrites = optional(bool)
    automatic_platform_optimization = optional(object({
      enabled              = optional(bool)
      cache_by_device_type = optional(bool, false)
      cf                   = optional(bool, false)
      hostnames            = list(string)
      wordpress            = optional(bool, false)
      wp_plugin            = optional(bool, false)
    }))
    brotli             = optional(bool)
    browser_cache_ttl  = optional(number)
    browser_check      = optional(bool)
    cache_level        = optional(string)
    challenge_ttl      = optional(number)
    ciphers            = optional(list(string), [])
    cname_flattening   = optional(string)
    content_converter  = optional(bool)
    development_mode   = optional(bool)
    early_hints        = optional(bool)
    edge_cache_ttl     = optional(number)
    email_obfuscation  = optional(bool)
    h2_prioritization  = optional(string)
    hotlink_protection = optional(bool)
    http2              = optional(bool)
    http3              = optional(bool)
    image_resizing     = optional(string)
    ip_geolocation     = optional(bool)
    ipv6               = optional(bool)
    max_upload         = optional(number)
    min_tls_version    = optional(string)
    mirage             = optional(bool)
    nel = optional(object({
      enabled = optional(bool)
    }))
    opportunistic_encryption    = optional(bool)
    opportunistic_onion         = optional(bool)
    orange_to_orange            = optional(bool)
    origin_error_page_pass_thru = optional(bool)
    origin_h2_max_streams       = optional(number)
    origin_max_http_version     = optional(string)
    polish                      = optional(string)
    prefetch_preload            = optional(bool)
    privacy_pass                = optional(bool)
    proxy_read_timeout          = optional(number)
    pseudo_ipv4                 = optional(string)
    redirects_for_ai_training   = optional(bool)
    replace_insecure_js         = optional(bool)
    response_buffering          = optional(bool)
    rocket_loader               = optional(bool)
    search_for_agents           = optional(bool)
    security_header = optional(object({
      enabled            = optional(bool)
      include_subdomains = optional(bool, false)
      max_age            = optional(number, 0)
      nosniff            = optional(bool, false)
      preload            = optional(bool, false)
    }))
    security_level                  = optional(string)
    server_side_exclude             = optional(bool)
    sha1_support                    = optional(bool)
    sort_query_string_for_cache     = optional(bool)
    ssl                             = optional(string)
    ssl_recommender                 = optional(bool)
    tls_1_2_only                    = optional(bool)
    tls_1_3                         = optional(string)
    tls_client_auth                 = optional(bool)
    transformations                 = optional(string)
    transformations_allowed_origins = optional(string)
    true_client_ip_header           = optional(bool)
    waf                             = optional(bool)
    webp                            = optional(bool)
    websockets                      = optional(bool)
    long_lived_grpc                 = optional(bool)
  })
}