locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-tunnel")

  # Ingress is only managed for a remotely-configured (cloudflare) tunnel with rules.
  is_remote     = var.spec.config_src == "cloudflare"
  manage_config = local.is_remote && length(var.spec.ingress) > 0

  # Per-ingress origin settings, mapped to the provider's field names (note the
  # provider's match_sn_ito_host spelling). Unset numbers/strings/bools become null so
  # the two engines stay byte-for-byte equivalent (false is omitted, like Pulumi).
  ingress = [for r in var.spec.ingress : {
    hostname = try(r.hostname, "") != "" ? r.hostname : null
    service  = r.service
    path     = try(r.path, "") != "" ? r.path : null
    origin_request = try(r.origin_request, null) == null ? null : {
      access = try(r.origin_request.access, null) == null ? null : {
        aud_tag   = r.origin_request.access.aud_tag
        team_name = r.origin_request.access.team_name
        required  = r.origin_request.access.required
      }
      ca_pool                  = try(r.origin_request.ca_pool, "") != "" ? r.origin_request.ca_pool : null
      connect_timeout          = try(r.origin_request.connect_timeout, 0) > 0 ? r.origin_request.connect_timeout : null
      disable_chunked_encoding = try(r.origin_request.disable_chunked_encoding, false) ? true : null
      http2_origin             = try(r.origin_request.http2_origin, false) ? true : null
      http_host_header         = try(r.origin_request.http_host_header, "") != "" ? r.origin_request.http_host_header : null
      keep_alive_connections   = try(r.origin_request.keep_alive_connections, 0) > 0 ? r.origin_request.keep_alive_connections : null
      keep_alive_timeout       = try(r.origin_request.keep_alive_timeout, 0) > 0 ? r.origin_request.keep_alive_timeout : null
      match_sn_ito_host        = try(r.origin_request.match_sni_to_host, false) ? true : null
      no_happy_eyeballs        = try(r.origin_request.no_happy_eyeballs, false) ? true : null
      no_tls_verify            = try(r.origin_request.no_tls_verify, false) ? true : null
      origin_server_name       = try(r.origin_request.origin_server_name, "") != "" ? r.origin_request.origin_server_name : null
      proxy_type               = try(r.origin_request.proxy_type, "") != "" ? r.origin_request.proxy_type : null
      tcp_keep_alive           = try(r.origin_request.tcp_keep_alive, 0) > 0 ? r.origin_request.tcp_keep_alive : null
      tls_timeout              = try(r.origin_request.tls_timeout, 0) > 0 ? r.origin_request.tls_timeout : null
    }
  }]

  cfg_or = var.spec.origin_request
  origin_request = local.cfg_or == null ? null : {
    access = try(local.cfg_or.access, null) == null ? null : {
      aud_tag   = local.cfg_or.access.aud_tag
      team_name = local.cfg_or.access.team_name
      required  = local.cfg_or.access.required
    }
    ca_pool                  = try(local.cfg_or.ca_pool, "") != "" ? local.cfg_or.ca_pool : null
    connect_timeout          = try(local.cfg_or.connect_timeout, 0) > 0 ? local.cfg_or.connect_timeout : null
    disable_chunked_encoding = try(local.cfg_or.disable_chunked_encoding, false) ? true : null
    http2_origin             = try(local.cfg_or.http2_origin, false) ? true : null
    http_host_header         = try(local.cfg_or.http_host_header, "") != "" ? local.cfg_or.http_host_header : null
    keep_alive_connections   = try(local.cfg_or.keep_alive_connections, 0) > 0 ? local.cfg_or.keep_alive_connections : null
    keep_alive_timeout       = try(local.cfg_or.keep_alive_timeout, 0) > 0 ? local.cfg_or.keep_alive_timeout : null
    match_sn_ito_host        = try(local.cfg_or.match_sni_to_host, false) ? true : null
    no_happy_eyeballs        = try(local.cfg_or.no_happy_eyeballs, false) ? true : null
    no_tls_verify            = try(local.cfg_or.no_tls_verify, false) ? true : null
    origin_server_name       = try(local.cfg_or.origin_server_name, "") != "" ? local.cfg_or.origin_server_name : null
    proxy_type               = try(local.cfg_or.proxy_type, "") != "" ? local.cfg_or.proxy_type : null
    tcp_keep_alive           = try(local.cfg_or.tcp_keep_alive, 0) > 0 ? local.cfg_or.tcp_keep_alive : null
    tls_timeout              = try(local.cfg_or.tls_timeout, 0) > 0 ? local.cfg_or.tls_timeout : null
  }
}
