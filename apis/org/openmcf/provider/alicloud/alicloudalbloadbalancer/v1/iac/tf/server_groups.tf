resource "alicloud_alb_server_group" "groups" {
  for_each = local.server_groups_map

  server_group_name = each.key
  vpc_id            = var.spec.vpc_id
  protocol          = each.value.protocol
  scheduler         = each.value.scheduler

  health_check_config {
    health_check_enabled      = each.value.health_check_config.health_check_enabled
    health_check_protocol     = each.value.health_check_config.health_check_protocol
    health_check_path         = each.value.health_check_config.health_check_path != "" ? each.value.health_check_config.health_check_path : null
    health_check_host         = each.value.health_check_config.health_check_host != "" ? each.value.health_check_config.health_check_host : null
    health_check_method       = each.value.health_check_config.health_check_method
    health_check_connect_port = each.value.health_check_config.health_check_connect_port
    health_check_interval     = each.value.health_check_config.health_check_interval
    health_check_timeout      = each.value.health_check_config.health_check_timeout
    healthy_threshold         = each.value.health_check_config.healthy_threshold
    unhealthy_threshold       = each.value.health_check_config.unhealthy_threshold
    health_check_codes        = length(each.value.health_check_config.health_check_codes) > 0 ? each.value.health_check_config.health_check_codes : null
  }

  dynamic "sticky_session_config" {
    for_each = each.value.sticky_session_config != null ? [each.value.sticky_session_config] : []
    content {
      sticky_session_enabled = sticky_session_config.value.sticky_session_enabled
      sticky_session_type    = sticky_session_config.value.sticky_session_type != "" ? sticky_session_config.value.sticky_session_type : null
      cookie                 = sticky_session_config.value.cookie != "" ? sticky_session_config.value.cookie : null
      cookie_timeout         = sticky_session_config.value.cookie_timeout
    }
  }
}
