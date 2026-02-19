resource "oci_load_balancer_listener" "this" {
  for_each = local.listeners_map

  load_balancer_id         = oci_load_balancer_load_balancer.this.id
  name                     = each.value.name
  port                     = each.value.port
  protocol                 = lookup(local.listener_protocol_map, each.value.protocol, upper(each.value.protocol))
  default_backend_set_name = oci_load_balancer_backend_set.this[each.value.default_backend_set_name].name

  hostname_names      = length(each.value.hostname_names) > 0 ? each.value.hostname_names : null
  rule_set_names      = length(each.value.rule_set_names) > 0 ? each.value.rule_set_names : null
  routing_policy_name = each.value.routing_policy_name != "" ? each.value.routing_policy_name : null

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
      has_session_resumption           = ssl_configuration.value.has_session_resumption
    }
  }

  dynamic "connection_configuration" {
    for_each = each.value.connection_configuration != null ? [each.value.connection_configuration] : []
    content {
      idle_timeout_in_seconds              = tostring(connection_configuration.value.idle_timeout_in_seconds)
      backend_tcp_proxy_protocol_version = connection_configuration.value.backend_tcp_proxy_protocol_version > 0 ? connection_configuration.value.backend_tcp_proxy_protocol_version : null
    }
  }

  depends_on = [
    oci_load_balancer_backend_set.this,
    oci_load_balancer_certificate.this,
    oci_load_balancer_hostname.this,
    oci_load_balancer_rule_set.this,
  ]
}
