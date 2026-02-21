resource "oci_network_load_balancer_network_load_balancer" "this" {
  compartment_id = var.spec.compartment_id.value
  display_name   = local.display_name
  subnet_id      = var.spec.subnet_id.value
  freeform_tags  = local.freeform_tags

  is_private                     = var.spec.is_private
  is_preserve_source_destination = var.spec.is_preserve_source_destination ? true : null
  is_symmetric_hash_enabled      = var.spec.is_symmetric_hash_enabled ? true : null

  network_security_group_ids = length(var.spec.network_security_group_ids) > 0 ? [
    for n in var.spec.network_security_group_ids : n.value
  ] : null

  nlb_ip_version        = var.spec.nlb_ip_version != "" ? var.spec.nlb_ip_version : null
  assigned_ipv6         = var.spec.assigned_ipv6 != "" ? var.spec.assigned_ipv6 : null
  assigned_private_ipv4 = var.spec.assigned_private_ipv4 != "" ? var.spec.assigned_private_ipv4 : null
  subnet_ipv6cidr       = var.spec.subnet_ipv6cidr != "" ? var.spec.subnet_ipv6cidr : null

  dynamic "reserved_ips" {
    for_each = var.spec.reserved_ips
    content {
      id = reserved_ips.value.id
    }
  }
}

resource "oci_network_load_balancer_backend_set" "this" {
  for_each = local.backend_sets_map

  network_load_balancer_id = oci_network_load_balancer_network_load_balancer.this.id
  name                     = each.value.name
  policy                   = local.policy_map[each.value.policy]

  health_checker {
    protocol            = local.health_checker_protocol_map[each.value.health_checker.protocol]
    port                = each.value.health_checker.port > 0 ? each.value.health_checker.port : null
    url_path            = each.value.health_checker.url_path != "" ? each.value.health_checker.url_path : null
    return_code         = each.value.health_checker.return_code > 0 ? each.value.health_checker.return_code : null
    response_body_regex = each.value.health_checker.response_body_regex != "" ? each.value.health_checker.response_body_regex : null
    interval_in_millis  = each.value.health_checker.interval_in_millis > 0 ? each.value.health_checker.interval_in_millis : null
    timeout_in_millis   = each.value.health_checker.timeout_in_millis > 0 ? each.value.health_checker.timeout_in_millis : null
    retries             = each.value.health_checker.retries > 0 ? each.value.health_checker.retries : null
    request_data        = each.value.health_checker.request_data != "" ? each.value.health_checker.request_data : null
    response_data       = each.value.health_checker.response_data != "" ? each.value.health_checker.response_data : null

    dynamic "dns" {
      for_each = each.value.health_checker.dns_health_check != null ? [each.value.health_checker.dns_health_check] : []
      content {
        domain_name        = dns.value.domain_name
        query_class        = dns.value.query_class != "" ? dns.value.query_class : null
        query_type         = dns.value.query_type != "" ? dns.value.query_type : null
        rcodes             = length(dns.value.rcodes) > 0 ? dns.value.rcodes : null
        transport_protocol = dns.value.transport_protocol != "" ? dns.value.transport_protocol : null
      }
    }
  }

  is_preserve_source                          = each.value.is_preserve_source ? true : null
  is_fail_open                                = each.value.is_fail_open ? true : null
  is_instant_failover_enabled                 = each.value.is_instant_failover_enabled ? true : null
  is_instant_failover_tcp_reset_enabled       = each.value.is_instant_failover_tcp_reset_enabled ? true : null
  are_operationally_active_backends_preferred = each.value.are_operationally_active_backends_preferred ? true : null
  ip_version                                  = each.value.ip_version != "" ? each.value.ip_version : null
}

resource "oci_network_load_balancer_backend" "this" {
  for_each = local.backends_map

  network_load_balancer_id = oci_network_load_balancer_network_load_balancer.this.id
  backend_set_name         = oci_network_load_balancer_backend_set.this[each.value.bs_name].name
  port                     = each.value.port

  ip_address = each.value.ip_address != "" ? each.value.ip_address : null
  target_id  = each.value.target_id != "" ? each.value.target_id : null
  weight     = each.value.weight > 0 ? each.value.weight : null
  is_backup  = each.value.is_backup ? true : null
  is_drain   = each.value.is_drain ? true : null
  is_offline = each.value.is_offline ? true : null
  name       = each.value.name != "" ? each.value.name : null
}

resource "oci_network_load_balancer_listener" "this" {
  for_each = local.listeners_map

  network_load_balancer_id = oci_network_load_balancer_network_load_balancer.this.id
  name                     = each.value.name
  port                     = each.value.port
  protocol                 = local.listener_protocol_map[each.value.protocol]
  default_backend_set_name = oci_network_load_balancer_backend_set.this[each.value.default_backend_set_name].name

  ip_version       = each.value.ip_version != "" ? each.value.ip_version : null
  is_ppv2enabled   = each.value.is_ppv2_enabled ? true : null
  tcp_idle_timeout = each.value.tcp_idle_timeout > 0 ? each.value.tcp_idle_timeout : null
  udp_idle_timeout = each.value.udp_idle_timeout > 0 ? each.value.udp_idle_timeout : null
  l3ip_idle_timeout = each.value.l3ip_idle_timeout > 0 ? each.value.l3ip_idle_timeout : null
}
