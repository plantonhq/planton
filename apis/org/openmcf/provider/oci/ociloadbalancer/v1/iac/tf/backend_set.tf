resource "oci_load_balancer_backend_set" "this" {
  for_each = local.backend_sets_map

  load_balancer_id = oci_load_balancer_load_balancer.this.id
  name             = each.value.name
  policy           = lookup(local.policy_map, each.value.policy, upper(each.value.policy))

  health_checker {
    protocol            = lookup(local.health_checker_protocol_map, each.value.health_checker.protocol, upper(each.value.health_checker.protocol))
    port                = each.value.health_checker.port > 0 ? each.value.health_checker.port : null
    url_path            = each.value.health_checker.url_path != "" ? each.value.health_checker.url_path : null
    return_code         = each.value.health_checker.return_code > 0 ? each.value.health_checker.return_code : null
    response_body_regex = each.value.health_checker.response_body_regex != "" ? each.value.health_checker.response_body_regex : null
    interval_ms         = each.value.health_checker.interval_ms > 0 ? each.value.health_checker.interval_ms : null
    timeout_in_millis   = each.value.health_checker.timeout_in_millis > 0 ? each.value.health_checker.timeout_in_millis : null
    retries             = each.value.health_checker.retries > 0 ? each.value.health_checker.retries : null
    is_force_plain_text = each.value.health_checker.is_force_plain_text ? true : null
  }

  backend_max_connections = each.value.backend_max_connections > 0 ? each.value.backend_max_connections : null

  dynamic "ssl_configuration" {
    for_each = each.value.ssl_configuration != null ? [each.value.ssl_configuration] : []
    content {
      certificate_ids                  = length(ssl_configuration.value.certificate_ids) > 0 ? ssl_configuration.value.certificate_ids : null
      certificate_name                 = ssl_configuration.value.certificate_name != "" ? ssl_configuration.value.certificate_name : null
      cipher_suite_name                = ssl_configuration.value.cipher_suite_name != "" ? ssl_configuration.value.cipher_suite_name : null
      protocols                        = length(ssl_configuration.value.protocols) > 0 ? ssl_configuration.value.protocols : null
      server_order_preference          = ssl_configuration.value.server_order_preference != "" ? ssl_configuration.value.server_order_preference : null
      trusted_certificate_authority_ids = length(ssl_configuration.value.trusted_certificate_authority_ids) > 0 ? ssl_configuration.value.trusted_certificate_authority_ids : null
      verify_depth                     = ssl_configuration.value.verify_depth > 0 ? ssl_configuration.value.verify_depth : null
      verify_peer_certificate          = ssl_configuration.value.verify_peer_certificate
    }
  }

  dynamic "lb_cookie_session_persistence_configuration" {
    for_each = each.value.lb_cookie_session_persistence != null ? [each.value.lb_cookie_session_persistence] : []
    content {
      cookie_name        = lb_cookie_session_persistence_configuration.value.cookie_name != "" ? lb_cookie_session_persistence_configuration.value.cookie_name : null
      disable_fallback   = lb_cookie_session_persistence_configuration.value.disable_fallback
      domain             = lb_cookie_session_persistence_configuration.value.domain != "" ? lb_cookie_session_persistence_configuration.value.domain : null
      is_http_only       = lb_cookie_session_persistence_configuration.value.is_http_only
      is_secure          = lb_cookie_session_persistence_configuration.value.is_secure
      max_age_in_seconds = lb_cookie_session_persistence_configuration.value.max_age_in_seconds > 0 ? lb_cookie_session_persistence_configuration.value.max_age_in_seconds : null
      path               = lb_cookie_session_persistence_configuration.value.path != "" ? lb_cookie_session_persistence_configuration.value.path : null
    }
  }

  dynamic "session_persistence_configuration" {
    for_each = each.value.app_cookie_session_persistence != null ? [each.value.app_cookie_session_persistence] : []
    content {
      cookie_name      = session_persistence_configuration.value.cookie_name
      disable_fallback = session_persistence_configuration.value.disable_fallback
    }
  }
}

resource "oci_load_balancer_backend" "this" {
  for_each = local.backends_map

  load_balancer_id = oci_load_balancer_load_balancer.this.id
  backendset_name  = oci_load_balancer_backend_set.this[each.value.bs_name].name
  ip_address       = each.value.ip_address
  port             = each.value.port

  weight          = each.value.weight > 0 ? each.value.weight : null
  backup          = each.value.backup ? true : null
  drain           = each.value.drain ? true : null
  offline         = each.value.offline ? true : null
  max_connections = each.value.max_connections > 0 ? each.value.max_connections : null
}
