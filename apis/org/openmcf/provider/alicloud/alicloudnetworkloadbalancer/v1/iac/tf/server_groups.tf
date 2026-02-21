resource "alicloud_nlb_server_group" "groups" {
  for_each = local.server_groups_map

  server_group_name       = each.key
  vpc_id                  = var.spec.vpc_id
  protocol                = each.value.protocol
  scheduler               = each.value.scheduler
  connection_drain_enabled = each.value.connection_drain_enabled
  connection_drain_timeout = each.value.connection_drain_timeout
  preserve_client_ip_enabled = each.value.preserve_client_ip_enabled

  health_check {
    health_check_enabled         = each.value.health_check.health_check_enabled
    health_check_type            = each.value.health_check.health_check_type
    health_check_connect_port    = each.value.health_check.health_check_connect_port
    health_check_connect_timeout = each.value.health_check.health_check_connect_timeout
    health_check_interval        = each.value.health_check.health_check_interval
    healthy_threshold            = each.value.health_check.healthy_threshold
    unhealthy_threshold          = each.value.health_check.unhealthy_threshold
    health_check_url             = each.value.health_check.health_check_url != "" ? each.value.health_check.health_check_url : null
    health_check_domain          = each.value.health_check.health_check_domain != "" ? each.value.health_check.health_check_domain : null
    http_check_method            = each.value.health_check.health_check_type == "HTTP" ? each.value.health_check.http_check_method : null
    health_check_http_code       = length(each.value.health_check.health_check_http_codes) > 0 ? each.value.health_check.health_check_http_codes : null
  }
}
