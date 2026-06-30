resource "alicloud_nlb_listener" "listeners" {
  for_each = local.listeners_map

  load_balancer_id     = alicloud_nlb_load_balancer.main.id
  listener_port        = each.value.listener_port
  listener_protocol    = each.value.listener_protocol
  server_group_id      = alicloud_nlb_server_group.groups[each.value.server_group_name].id
  listener_description = each.value.listener_description != "" ? each.value.listener_description : null
  idle_timeout         = each.value.idle_timeout
  proxy_protocol_enabled = each.value.proxy_protocol_enabled

  certificate_ids    = length(each.value.certificate_ids) > 0 ? each.value.certificate_ids : null
  security_policy_id = each.value.security_policy_id != "" ? each.value.security_policy_id : null
  ca_certificate_ids = length(each.value.ca_certificate_ids) > 0 ? each.value.ca_certificate_ids : null
  ca_enabled         = each.value.ca_enabled
}
