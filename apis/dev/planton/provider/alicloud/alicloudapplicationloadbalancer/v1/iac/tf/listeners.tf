resource "alicloud_alb_listener" "listeners" {
  for_each = local.listeners_map

  load_balancer_id     = alicloud_alb_load_balancer.main.id
  listener_port        = each.value.listener_port
  listener_protocol    = each.value.listener_protocol
  listener_description = each.value.listener_description != "" ? each.value.listener_description : null
  gzip_enabled         = each.value.gzip_enabled
  http2_enabled        = each.value.http2_enabled
  idle_timeout         = each.value.idle_timeout
  request_timeout      = each.value.request_timeout
  security_policy_id   = each.value.security_policy_id != "" ? each.value.security_policy_id : null

  dynamic "certificates" {
    for_each = each.value.certificate_id != "" ? [each.value.certificate_id] : []
    content {
      certificate_id = certificates.value
    }
  }

  default_actions {
    type = "ForwardGroup"
    forward_group_config {
      server_group_tuples {
        server_group_id = alicloud_alb_server_group.groups[each.value.default_action_server_group_name].id
      }
    }
  }
}
