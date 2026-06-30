resource "alicloud_db_instance" "main" {
  engine                   = var.spec.engine
  engine_version           = var.spec.engine_version
  instance_type            = var.spec.instance_type
  instance_storage         = var.spec.instance_storage
  vswitch_id               = var.spec.vswitch_id
  instance_name            = local.instance_name
  instance_charge_type     = var.spec.instance_charge_type
  category                 = var.spec.category
  db_instance_storage_type = var.spec.db_instance_storage_type != "" ? var.spec.db_instance_storage_type : null
  zone_id                  = var.spec.zone_id != "" ? var.spec.zone_id : null
  zone_id_slave_a          = var.spec.zone_id_slave_a != "" ? var.spec.zone_id_slave_a : null
  security_ips             = length(var.spec.security_ips) > 0 ? var.spec.security_ips : null
  security_group_ids       = length(var.spec.security_group_ids) > 0 ? var.spec.security_group_ids : null
  monitoring_period        = var.spec.monitoring_period
  maintain_time            = var.spec.maintain_time != "" ? var.spec.maintain_time : null
  deletion_protection      = var.spec.deletion_protection
  ssl_action               = var.spec.ssl_action != "" ? var.spec.ssl_action : null
  tde_status               = var.spec.tde_status != "" ? var.spec.tde_status : null
  encryption_key           = var.spec.encryption_key != "" ? var.spec.encryption_key : null
  auto_renew               = var.spec.auto_renew
  auto_renew_period        = var.spec.auto_renew_period
  period                   = var.spec.period
  resource_group_id        = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags                     = local.final_tags

  dynamic "parameters" {
    for_each = var.spec.parameters
    content {
      name  = parameters.value.name
      value = parameters.value.value
    }
  }
}
