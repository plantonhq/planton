resource "oci_psql_db_system" "this" {
  compartment_id = var.spec.compartment_id.value
  db_version     = var.spec.db_version
  display_name   = local.display_name
  shape          = var.spec.shape
  freeform_tags  = local.freeform_tags

  network_details {
    subnet_id                      = var.spec.network_details.subnet_id.value
    nsg_ids                        = length(local.nsg_ids) > 0 ? local.nsg_ids : null
    is_reader_endpoint_enabled     = var.spec.network_details.is_reader_endpoint_enabled
    primary_db_endpoint_private_ip = var.spec.network_details.primary_db_endpoint_private_ip != "" ? var.spec.network_details.primary_db_endpoint_private_ip : null
  }

  storage_details {
    is_regionally_durable = var.spec.storage_details.is_regionally_durable
    system_type           = "OCI_OPTIMIZED_STORAGE"
    availability_domain   = var.spec.storage_details.availability_domain != "" ? var.spec.storage_details.availability_domain : null
    iops                  = var.spec.storage_details.iops > 0 ? tostring(var.spec.storage_details.iops) : null
  }

  instance_ocpu_count         = var.spec.instance_ocpu_count > 0 ? var.spec.instance_ocpu_count : null
  instance_memory_size_in_gbs = var.spec.instance_memory_size_in_gbs > 0 ? var.spec.instance_memory_size_in_gbs : null
  instance_count              = var.spec.instance_count > 0 ? var.spec.instance_count : null
  description                 = var.spec.description != "" ? var.spec.description : null

  dynamic "credentials" {
    for_each = var.spec.credentials != null ? [var.spec.credentials] : []
    content {
      username = credentials.value.username

      password_details {
        password_type  = credentials.value.password_details.password_type != "" ? lookup(local.password_type_map, credentials.value.password_details.password_type, null) : null
        password       = credentials.value.password_details.password != "" ? credentials.value.password_details.password : null
        secret_id      = credentials.value.password_details.secret_id != null ? credentials.value.password_details.secret_id.value : null
        secret_version = credentials.value.password_details.secret_version != "" ? credentials.value.password_details.secret_version : null
      }
    }
  }

  config_id = var.spec.config_id != null ? var.spec.config_id.value : null

  dynamic "management_policy" {
    for_each = var.spec.management_policy != null ? [var.spec.management_policy] : []
    content {
      dynamic "backup_policy" {
        for_each = management_policy.value.backup_policy != null ? [management_policy.value.backup_policy] : []
        content {
          kind              = backup_policy.value.kind != "" ? lookup(local.backup_kind_map, backup_policy.value.kind, null) : null
          backup_start      = backup_policy.value.backup_start != "" ? backup_policy.value.backup_start : null
          retention_days    = backup_policy.value.retention_days > 0 ? backup_policy.value.retention_days : null
          days_of_the_month = length(backup_policy.value.days_of_the_month) > 0 ? backup_policy.value.days_of_the_month : null
          days_of_the_week  = length(backup_policy.value.days_of_the_week) > 0 ? backup_policy.value.days_of_the_week : null
        }
      }

      maintenance_window_start = management_policy.value.maintenance_window_start != "" ? management_policy.value.maintenance_window_start : null
    }
  }

  dynamic "instances_details" {
    for_each = var.spec.instances_details
    content {
      display_name = instances_details.value.display_name != "" ? instances_details.value.display_name : null
      description  = instances_details.value.description != "" ? instances_details.value.description : null
      private_ip   = instances_details.value.private_ip != "" ? instances_details.value.private_ip : null
    }
  }

  lifecycle {
    ignore_changes = [credentials]
  }
}
