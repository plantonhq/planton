locals {
  zone_id = var.spec.zone_id

  # On/off toggles: null = not managed (never sent); true/false = "on"/"off".
  # Keys are the Cloudflare setting_ids (the spec's zero_rtt maps to "0rtt" --
  # proto identifiers cannot start with a digit).
  bool_settings = {
    "0rtt"                        = var.spec.zero_rtt
    "advanced_ddos"               = var.spec.advanced_ddos
    "always_online"               = var.spec.always_online
    "always_use_https"            = var.spec.always_use_https
    "automatic_https_rewrites"    = var.spec.automatic_https_rewrites
    "brotli"                      = var.spec.brotli
    "browser_check"               = var.spec.browser_check
    "content_converter"           = var.spec.content_converter
    "development_mode"            = var.spec.development_mode
    "early_hints"                 = var.spec.early_hints
    "email_obfuscation"           = var.spec.email_obfuscation
    "hotlink_protection"          = var.spec.hotlink_protection
    "http2"                       = var.spec.http2
    "http3"                       = var.spec.http3
    "ip_geolocation"              = var.spec.ip_geolocation
    "ipv6"                        = var.spec.ipv6
    "long_lived_grpc"             = var.spec.long_lived_grpc
    "mirage"                      = var.spec.mirage
    "opportunistic_encryption"    = var.spec.opportunistic_encryption
    "opportunistic_onion"         = var.spec.opportunistic_onion
    "orange_to_orange"            = var.spec.orange_to_orange
    "origin_error_page_pass_thru" = var.spec.origin_error_page_pass_thru
    "prefetch_preload"            = var.spec.prefetch_preload
    "privacy_pass"                = var.spec.privacy_pass
    "redirects_for_ai_training"   = var.spec.redirects_for_ai_training
    "replace_insecure_js"         = var.spec.replace_insecure_js
    "response_buffering"          = var.spec.response_buffering
    "rocket_loader"               = var.spec.rocket_loader
    "search_for_agents"           = var.spec.search_for_agents
    "server_side_exclude"         = var.spec.server_side_exclude
    "sha1_support"                = var.spec.sha1_support
    "sort_query_string_for_cache" = var.spec.sort_query_string_for_cache
    # ssl_recommender: the provider's schema requires the on/off value form on
    # writes (its documented enabled-attribute form fails validation at v5.23.0).
    "ssl_recommender"       = var.spec.ssl_recommender
    "tls_1_2_only"          = var.spec.tls_1_2_only
    "tls_client_auth"       = var.spec.tls_client_auth
    "true_client_ip_header" = var.spec.true_client_ip_header
    "waf"                   = var.spec.waf
    "webp"                  = var.spec.webp
    "websockets"            = var.spec.websockets
  }

  # Enum and free-string settings: null = not managed.
  string_settings = {
    "cache_level"                     = var.spec.cache_level
    "cname_flattening"                = var.spec.cname_flattening
    "h2_prioritization"               = var.spec.h2_prioritization
    "image_resizing"                  = var.spec.image_resizing
    "min_tls_version"                 = var.spec.min_tls_version
    "origin_max_http_version"         = var.spec.origin_max_http_version
    "polish"                          = var.spec.polish
    "pseudo_ipv4"                     = var.spec.pseudo_ipv4
    "security_level"                  = var.spec.security_level
    "ssl"                             = var.spec.ssl
    "tls_1_3"                         = var.spec.tls_1_3
    "transformations"                 = var.spec.transformations
    "transformations_allowed_origins" = var.spec.transformations_allowed_origins
  }

  # Numeric settings: null = not managed (value sets validated by the API).
  number_settings = {
    "browser_cache_ttl"     = var.spec.browser_cache_ttl
    "challenge_ttl"         = var.spec.challenge_ttl
    "edge_cache_ttl"        = var.spec.edge_cache_ttl
    "max_upload"            = var.spec.max_upload
    "origin_h2_max_streams" = var.spec.origin_h2_max_streams
    "proxy_read_timeout"    = var.spec.proxy_read_timeout
  }

  # The complete managed-settings map: one entry per setting the manifest
  # actually manages. The zone_setting value attribute is dynamic, so entries
  # of different value types coexist in one fan-out.
  settings = merge(
    { for id, v in local.bool_settings : id => (v ? "on" : "off") if v != null },
    { for id, v in local.string_settings : id => v if v != null },
    { for id, v in local.number_settings : id => v if v != null },
    length(var.spec.ciphers) > 0 ? { "ciphers" = var.spec.ciphers } : {},
    var.spec.security_header != null ? {
      # The API nests the HSTS fields under strict_transport_security.
      "security_header" = {
        strict_transport_security = {
          enabled            = var.spec.security_header.enabled
          include_subdomains = var.spec.security_header.include_subdomains
          max_age            = var.spec.security_header.max_age
          nosniff            = var.spec.security_header.nosniff
          preload            = var.spec.security_header.preload
        }
      }
    } : {},
    var.spec.nel != null ? {
      "nel" = { enabled = var.spec.nel.enabled }
    } : {},
    var.spec.aegis != null ? {
      "aegis" = merge(
        var.spec.aegis.enabled != null ? { enabled = var.spec.aegis.enabled } : {},
        var.spec.aegis.pool_id != "" ? { pool_id = var.spec.aegis.pool_id } : {},
      )
    } : {},
    var.spec.automatic_platform_optimization != null ? {
      # The APO API requires every member of the value object on writes.
      "automatic_platform_optimization" = {
        enabled              = var.spec.automatic_platform_optimization.enabled
        cache_by_device_type = var.spec.automatic_platform_optimization.cache_by_device_type
        cf                   = var.spec.automatic_platform_optimization.cf
        hostnames            = var.spec.automatic_platform_optimization.hostnames
        wordpress            = var.spec.automatic_platform_optimization.wordpress
        wp_plugin            = var.spec.automatic_platform_optimization.wp_plugin
      }
    } : {},
  )
}
