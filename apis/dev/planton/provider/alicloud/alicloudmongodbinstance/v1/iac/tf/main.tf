resource "alicloud_mongodb_instance" "main" {
  engine_version                 = var.spec.engine_version
  db_instance_class              = var.spec.db_instance_class
  db_instance_storage            = var.spec.db_instance_storage
  account_password               = var.spec.account_password
  vswitch_id                     = var.spec.vswitch_id
  name                           = local.instance_name
  replication_factor             = var.spec.replication_factor
  storage_engine                 = var.spec.storage_engine
  instance_charge_type           = var.spec.instance_charge_type
  zone_id                        = var.spec.zone_id != "" ? var.spec.zone_id : null
  secondary_zone_id              = var.spec.secondary_zone_id != "" ? var.spec.secondary_zone_id : null
  hidden_zone_id                 = var.spec.hidden_zone_id != "" ? var.spec.hidden_zone_id : null
  readonly_replicas              = var.spec.readonly_replicas
  storage_type                   = var.spec.storage_type != "" ? var.spec.storage_type : null
  provisioned_iops               = var.spec.provisioned_iops
  security_ip_list               = length(var.spec.security_ip_list) > 0 ? var.spec.security_ip_list : null
  security_group_id              = var.spec.security_group_id != "" ? var.spec.security_group_id : null
  resource_group_id              = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  ssl_action                     = var.spec.ssl_action != "" ? var.spec.ssl_action : null
  tde_status                     = var.spec.tde_status != "" ? var.spec.tde_status : null
  encryption_key                 = var.spec.encryption_key != "" ? var.spec.encryption_key : null
  encrypted                      = var.spec.encrypted
  cloud_disk_encryption_key      = var.spec.cloud_disk_encryption_key != "" ? var.spec.cloud_disk_encryption_key : null
  maintain_start_time            = var.spec.maintain_start_time != "" ? var.spec.maintain_start_time : null
  maintain_end_time              = var.spec.maintain_end_time != "" ? var.spec.maintain_end_time : null
  backup_time                    = var.spec.backup_time != "" ? var.spec.backup_time : null
  backup_period                  = length(var.spec.backup_period) > 0 ? var.spec.backup_period : null
  db_instance_release_protection = var.spec.db_instance_release_protection
  period                         = var.spec.period
  auto_renew                     = var.spec.auto_renew
  auto_renew_duration            = var.spec.auto_renew_duration
  tags                           = local.final_tags

  dynamic "parameters" {
    for_each = var.spec.parameters
    content {
      name  = parameters.key
      value = parameters.value
    }
  }
}
