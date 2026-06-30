resource "alicloud_alb_load_balancer" "main" {
  load_balancer_name    = local.load_balancer_name
  vpc_id                = var.spec.vpc_id
  address_type          = var.spec.address_type
  load_balancer_edition = var.spec.load_balancer_edition
  resource_group_id     = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags                  = local.final_tags

  load_balancer_billing_config {
    pay_type = "PayAsYouGo"
  }

  dynamic "zone_mappings" {
    for_each = var.spec.zone_mappings
    content {
      zone_id    = zone_mappings.value.zone_id
      vswitch_id = zone_mappings.value.vswitch_id
    }
  }

  dynamic "access_log_config" {
    for_each = var.spec.access_log_config != null ? [var.spec.access_log_config] : []
    content {
      log_project = access_log_config.value.log_project
      log_store   = access_log_config.value.log_store
    }
  }
}
