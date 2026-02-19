resource "oci_database_autonomous_database" "this" {
  compartment_id = var.spec.compartment_id.value
  db_name        = var.spec.db_name
  display_name   = local.display_name
  freeform_tags  = local.freeform_tags

  db_workload      = var.spec.db_workload != "" ? lookup(local.db_workload_map, var.spec.db_workload, null) : null
  db_version       = var.spec.db_version != "" ? var.spec.db_version : null
  database_edition = var.spec.database_edition != "" ? lookup(local.database_edition_map, var.spec.database_edition, null) : null
  license_model    = var.spec.license_model != "" ? lookup(local.license_model_map, var.spec.license_model, null) : null
  character_set    = var.spec.character_set != "" ? var.spec.character_set : null
  ncharacter_set   = var.spec.ncharacter_set != "" ? var.spec.ncharacter_set : null

  compute_model = var.spec.compute_model != "" ? lookup(local.compute_model_map, var.spec.compute_model, null) : null
  compute_count = var.spec.compute_count

  data_storage_size_in_tbs = var.spec.data_storage_size_in_tbs > 0 ? var.spec.data_storage_size_in_tbs : null
  data_storage_size_in_gb  = var.spec.data_storage_size_in_gb > 0 ? var.spec.data_storage_size_in_gb : null

  is_auto_scaling_enabled             = var.spec.is_auto_scaling_enabled
  is_auto_scaling_for_storage_enabled = var.spec.is_auto_scaling_for_storage_enabled

  admin_password        = var.spec.admin_password != "" ? var.spec.admin_password : null
  secret_id             = var.spec.secret_id != null ? var.spec.secret_id.value : null
  secret_version_number = var.spec.secret_version_number > 0 ? var.spec.secret_version_number : null

  subnet_id              = var.spec.subnet_id != null ? var.spec.subnet_id.value : null
  nsg_ids                = length(local.nsg_ids) > 0 ? local.nsg_ids : null
  private_endpoint_label = var.spec.private_endpoint_label != "" ? var.spec.private_endpoint_label : null
  private_endpoint_ip    = var.spec.private_endpoint_ip != "" ? var.spec.private_endpoint_ip : null
  whitelisted_ips        = length(var.spec.whitelisted_ips) > 0 ? var.spec.whitelisted_ips : null

  is_mtls_connection_required = var.spec.is_mtls_connection_required
  is_access_control_enabled   = var.spec.is_access_control_enabled

  kms_key_id = var.spec.kms_key_id != null ? var.spec.kms_key_id.value : null
  vault_id   = var.spec.vault_id != null ? var.spec.vault_id.value : null

  is_dedicated = var.spec.is_dedicated
  is_free_tier = var.spec.is_free_tier
  is_dev_tier  = var.spec.is_dev_tier

  autonomous_container_database_id = var.spec.autonomous_container_database_id != null ? var.spec.autonomous_container_database_id.value : null

  backup_retention_period_in_days = var.spec.backup_retention_period_in_days > 0 ? var.spec.backup_retention_period_in_days : null

  is_local_data_guard_enabled = var.spec.is_local_data_guard_enabled

  autonomous_maintenance_schedule_type = var.spec.autonomous_maintenance_schedule_type != "" ? lookup(local.maintenance_schedule_type_map, var.spec.autonomous_maintenance_schedule_type, null) : null

  dynamic "customer_contacts" {
    for_each = var.spec.customer_contacts
    content {
      email = customer_contacts.value.email
    }
  }
}
