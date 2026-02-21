resource "alicloud_kvstore_instance" "main" {
  instance_class              = var.spec.instance_class
  password                    = var.spec.password
  engine_version              = var.spec.engine_version
  instance_type               = var.spec.instance_type
  db_instance_name            = local.instance_name
  payment_type                = var.spec.payment_type
  vswitch_id                  = var.spec.vswitch_id
  vpc_auth_mode               = var.spec.vpc_auth_mode
  zone_id                     = var.spec.zone_id != "" ? var.spec.zone_id : null
  secondary_zone_id           = var.spec.secondary_zone_id != "" ? var.spec.secondary_zone_id : null
  security_ips                = length(var.spec.security_ips) > 0 ? var.spec.security_ips : null
  security_group_id           = var.spec.security_group_id != "" ? var.spec.security_group_id : null
  resource_group_id           = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  shard_count                 = var.spec.shard_count
  read_only_count             = var.spec.read_only_count
  ssl_enable                  = var.spec.ssl_enable != "" ? var.spec.ssl_enable : null
  tde_status                  = var.spec.tde_status != "" ? var.spec.tde_status : null
  encryption_key              = var.spec.encryption_key != "" ? var.spec.encryption_key : null
  config                      = length(var.spec.config) > 0 ? var.spec.config : null
  instance_release_protection = var.spec.instance_release_protection
  maintain_start_time         = var.spec.maintain_start_time != "" ? var.spec.maintain_start_time : null
  maintain_end_time           = var.spec.maintain_end_time != "" ? var.spec.maintain_end_time : null
  backup_period               = length(var.spec.backup_period) > 0 ? var.spec.backup_period : null
  backup_time                 = var.spec.backup_time != "" ? var.spec.backup_time : null
  private_connection_prefix   = var.spec.private_connection_prefix != "" ? var.spec.private_connection_prefix : null
  auto_renew                  = var.spec.auto_renew
  auto_renew_period           = var.spec.auto_renew_period
  period                      = var.spec.period != "" ? var.spec.period : null
  tags                        = local.final_tags
}
