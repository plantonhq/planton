resource "oci_database_db_system" "this" {
  availability_domain = var.spec.availability_domain
  compartment_id      = var.spec.compartment_id.value
  hostname            = var.spec.hostname
  shape               = var.spec.shape
  ssh_public_keys     = var.spec.ssh_public_keys
  subnet_id           = var.spec.subnet_id.value
  display_name        = local.display_name
  freeform_tags       = local.freeform_tags

  cpu_core_count          = var.spec.cpu_core_count > 0 ? var.spec.cpu_core_count : null
  database_edition        = var.spec.database_edition != "" ? lookup(local.database_edition_map, var.spec.database_edition, null) : null
  license_model           = var.spec.license_model != "" ? lookup(local.license_model_map, var.spec.license_model, null) : null
  data_storage_size_in_gb = var.spec.data_storage_size_in_gb > 0 ? var.spec.data_storage_size_in_gb : null
  data_storage_percentage = var.spec.data_storage_percentage > 0 ? var.spec.data_storage_percentage : null
  disk_redundancy         = var.spec.disk_redundancy != "" ? lookup(local.disk_redundancy_map, var.spec.disk_redundancy, null) : null
  node_count              = var.spec.node_count > 0 ? var.spec.node_count : null
  domain                  = var.spec.domain != "" ? var.spec.domain : null
  cluster_name            = var.spec.cluster_name != "" ? var.spec.cluster_name : null
  fault_domains           = length(var.spec.fault_domains) > 0 ? var.spec.fault_domains : null

  nsg_ids              = length(local.nsg_ids) > 0 ? local.nsg_ids : null
  backup_subnet_id     = var.spec.backup_subnet_id != null ? var.spec.backup_subnet_id.value : null
  backup_network_nsg_ids = length(local.backup_network_nsg_ids) > 0 ? local.backup_network_nsg_ids : null

  kms_key_id         = var.spec.kms_key_id != null ? var.spec.kms_key_id.value : null
  kms_key_version_id = var.spec.kms_key_version_id != "" ? var.spec.kms_key_version_id : null

  time_zone        = var.spec.time_zone != "" ? var.spec.time_zone : null
  sparse_diskgroup = var.spec.sparse_diskgroup
  storage_volume_performance_mode = var.spec.storage_volume_performance_mode != "" ? lookup(local.storage_volume_performance_mode_map, var.spec.storage_volume_performance_mode, null) : null
  private_ip       = var.spec.private_ip != "" ? var.spec.private_ip : null

  db_home {
    db_version              = var.spec.db_home.db_version != "" ? var.spec.db_home.db_version : null
    display_name            = var.spec.db_home.display_name != "" ? var.spec.db_home.display_name : null
    database_software_image_id = var.spec.db_home.database_software_image_id != null ? var.spec.db_home.database_software_image_id.value : null

    database {
      admin_password   = var.spec.db_home.database.admin_password
      db_name          = var.spec.db_home.database.db_name
      character_set    = var.spec.db_home.database.character_set != "" ? var.spec.db_home.database.character_set : null
      ncharacter_set   = var.spec.db_home.database.ncharacter_set != "" ? var.spec.db_home.database.ncharacter_set : null
      pdb_name         = var.spec.db_home.database.pdb_name != "" ? var.spec.db_home.database.pdb_name : null
      db_domain        = var.spec.db_home.database.db_domain != "" ? var.spec.db_home.database.db_domain : null
      kms_key_id       = var.spec.db_home.database.kms_key_id != null ? var.spec.db_home.database.kms_key_id.value : null
      kms_key_version_id = var.spec.db_home.database.kms_key_version_id != "" ? var.spec.db_home.database.kms_key_version_id : null
      vault_id         = var.spec.db_home.database.vault_id != null ? var.spec.db_home.database.vault_id.value : null

      dynamic "db_backup_config" {
        for_each = var.spec.db_home.database.db_backup_config != null ? [var.spec.db_home.database.db_backup_config] : []
        content {
          auto_backup_enabled     = db_backup_config.value.auto_backup_enabled
          auto_backup_window      = db_backup_config.value.auto_backup_window != "" ? db_backup_config.value.auto_backup_window : null
          recovery_window_in_days = db_backup_config.value.recovery_window_in_days > 0 ? db_backup_config.value.recovery_window_in_days : null
        }
      }
    }
  }

  dynamic "data_collection_options" {
    for_each = var.spec.data_collection_options != null ? [var.spec.data_collection_options] : []
    content {
      is_diagnostics_events_enabled = data_collection_options.value.is_diagnostics_events_enabled
      is_health_monitoring_enabled  = data_collection_options.value.is_health_monitoring_enabled
      is_incident_logs_enabled      = data_collection_options.value.is_incident_logs_enabled
    }
  }

  dynamic "db_system_options" {
    for_each = var.spec.db_system_options != null ? [var.spec.db_system_options] : []
    content {
      storage_management = db_system_options.value.storage_management != "" ? lookup(local.storage_management_map, db_system_options.value.storage_management, null) : null
    }
  }

  dynamic "maintenance_window_details" {
    for_each = var.spec.maintenance_window_details != null ? [var.spec.maintenance_window_details] : []
    content {
      preference = maintenance_window_details.value.preference != "" ? lookup(local.preference_map, maintenance_window_details.value.preference, null) : null

      patching_mode      = maintenance_window_details.value.patching_mode != "" ? lookup(local.patching_mode_map, maintenance_window_details.value.patching_mode, null) : null
      lead_time_in_weeks = maintenance_window_details.value.lead_time_in_weeks > 0 ? maintenance_window_details.value.lead_time_in_weeks : null

      dynamic "months" {
        for_each = maintenance_window_details.value.months
        content {
          name = months.value
        }
      }

      weeks_of_month = length(maintenance_window_details.value.weeks_of_month) > 0 ? maintenance_window_details.value.weeks_of_month : null

      dynamic "days_of_week" {
        for_each = maintenance_window_details.value.days_of_week
        content {
          name = days_of_week.value
        }
      }

      hours_of_day = length(maintenance_window_details.value.hours_of_day) > 0 ? maintenance_window_details.value.hours_of_day : null

      custom_action_timeout_in_mins    = maintenance_window_details.value.custom_action_timeout_in_mins > 0 ? maintenance_window_details.value.custom_action_timeout_in_mins : null
      is_custom_action_timeout_enabled = maintenance_window_details.value.is_custom_action_timeout_enabled
      is_monthly_patching_enabled      = maintenance_window_details.value.is_monthly_patching_enabled
    }
  }

  lifecycle {
    ignore_changes = [db_home[0].database[0].admin_password]
  }
}
