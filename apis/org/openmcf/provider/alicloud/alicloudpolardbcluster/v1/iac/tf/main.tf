resource "alicloud_polardb_cluster" "main" {
  db_type       = var.spec.db_type
  db_version    = var.spec.db_version
  db_node_class = var.spec.db_node_class
  vswitch_id    = var.spec.vswitch_id
  db_node_count = var.spec.db_node_count
  description   = local.cluster_description
  pay_type      = var.spec.pay_type
  tags          = local.final_tags

  period            = var.spec.period
  renewal_status    = var.spec.renewal_status != "" ? var.spec.renewal_status : null
  auto_renew_period = var.spec.auto_renew_period

  zone_id            = var.spec.zone_id != "" ? var.spec.zone_id : null
  security_ips       = length(var.spec.security_ips) > 0 ? toset(var.spec.security_ips) : null
  security_group_ids = length(var.spec.security_group_ids) > 0 ? toset(var.spec.security_group_ids) : null
  maintain_time      = var.spec.maintain_time != "" ? var.spec.maintain_time : null
  resource_group_id  = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null

  creation_category = var.spec.creation_category != "" ? var.spec.creation_category : null
  sub_category      = var.spec.sub_category != "" ? var.spec.sub_category : null
  storage_type      = var.spec.storage_type != "" ? var.spec.storage_type : null
  storage_space     = var.spec.storage_space

  tde_status     = var.spec.tde_status != "" ? var.spec.tde_status : null
  encryption_key = var.spec.encryption_key != "" ? var.spec.encryption_key : null
  deletion_lock  = var.spec.deletion_lock

  collector_status                            = var.spec.collector_status != "" ? var.spec.collector_status : null
  backup_retention_policy_on_cluster_deletion = var.spec.backup_retention_policy_on_cluster_deletion != "" ? var.spec.backup_retention_policy_on_cluster_deletion : null

  dynamic "parameters" {
    for_each = var.spec.parameters
    content {
      name  = parameters.value.name
      value = parameters.value.value
    }
  }
}
