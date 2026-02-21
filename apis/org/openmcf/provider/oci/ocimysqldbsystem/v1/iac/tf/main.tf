resource "oci_mysql_mysql_db_system" "this" {
  compartment_id      = var.spec.compartment_id.value
  availability_domain = var.spec.availability_domain
  shape_name          = var.spec.shape_name
  subnet_id           = var.spec.subnet_id.value
  display_name        = local.display_name
  freeform_tags       = local.freeform_tags

  admin_username = var.spec.admin_username != "" ? var.spec.admin_username : null
  admin_password = var.spec.admin_password != "" ? var.spec.admin_password : null
  mysql_version  = var.spec.mysql_version != "" ? var.spec.mysql_version : null

  configuration_id = var.spec.configuration_id != null ? var.spec.configuration_id.value : null

  is_highly_available = var.spec.is_highly_available
  hostname_label      = var.spec.hostname_label != "" ? var.spec.hostname_label : null
  ip_address          = var.spec.ip_address != "" ? var.spec.ip_address : null
  fault_domain        = var.spec.fault_domain != "" ? var.spec.fault_domain : null
  port                = var.spec.port > 0 ? var.spec.port : null
  port_x              = var.spec.port_x > 0 ? var.spec.port_x : null
  description         = var.spec.description != "" ? var.spec.description : null
  crash_recovery      = var.spec.crash_recovery != "" ? var.spec.crash_recovery : null
  database_management = var.spec.database_management != "" ? var.spec.database_management : null

  nsg_ids = length(local.nsg_ids) > 0 ? local.nsg_ids : null

  dynamic "data_storage" {
    for_each = var.spec.data_storage != null ? [var.spec.data_storage] : []
    content {
      data_storage_size_in_gb        = data_storage.value.data_storage_size_in_gb > 0 ? data_storage.value.data_storage_size_in_gb : null
      is_auto_expand_storage_enabled = data_storage.value.is_auto_expand_storage_enabled
      max_storage_size_in_gbs        = data_storage.value.max_storage_size_in_gbs > 0 ? data_storage.value.max_storage_size_in_gbs : null
    }
  }

  dynamic "backup_policy" {
    for_each = var.spec.backup_policy != null ? [var.spec.backup_policy] : []
    content {
      is_enabled        = backup_policy.value.is_enabled
      retention_in_days = backup_policy.value.retention_in_days > 0 ? backup_policy.value.retention_in_days : null
      window_start_time = backup_policy.value.window_start_time != "" ? backup_policy.value.window_start_time : null

      dynamic "pitr_policy" {
        for_each = backup_policy.value.pitr_policy != null ? [backup_policy.value.pitr_policy] : []
        content {
          is_enabled = pitr_policy.value.is_enabled
        }
      }
    }
  }

  dynamic "maintenance" {
    for_each = var.spec.maintenance != null ? [var.spec.maintenance] : []
    content {
      window_start_time        = maintenance.value.window_start_time
      maintenance_schedule_type = maintenance.value.maintenance_schedule_type != "" ? lookup(local.maintenance_schedule_type_map, maintenance.value.maintenance_schedule_type, null) : null
      version_preference        = maintenance.value.version_preference != "" ? lookup(local.version_preference_map, maintenance.value.version_preference, null) : null
      version_track_preference  = maintenance.value.version_track_preference != "" ? lookup(local.version_track_preference_map, maintenance.value.version_track_preference, null) : null
    }
  }

  dynamic "deletion_policy" {
    for_each = var.spec.deletion_policy != null ? [var.spec.deletion_policy] : []
    content {
      automatic_backup_retention = deletion_policy.value.automatic_backup_retention != "" ? deletion_policy.value.automatic_backup_retention : null
      final_backup               = deletion_policy.value.final_backup != "" ? deletion_policy.value.final_backup : null
      is_delete_protected        = deletion_policy.value.is_delete_protected
    }
  }

  dynamic "encrypt_data" {
    for_each = var.spec.encrypt_data != null ? [var.spec.encrypt_data] : []
    content {
      key_generation_type = encrypt_data.value.key_generation_type != "" ? lookup(local.key_generation_type_map, encrypt_data.value.key_generation_type, null) : null
      key_id              = encrypt_data.value.key_id != null ? encrypt_data.value.key_id.value : null
    }
  }

  dynamic "secure_connections" {
    for_each = var.spec.secure_connections != null ? [var.spec.secure_connections] : []
    content {
      certificate_generation_type = secure_connections.value.certificate_generation_type != "" ? lookup(local.certificate_generation_type_map, secure_connections.value.certificate_generation_type, null) : null
      certificate_id              = secure_connections.value.certificate_id != null ? secure_connections.value.certificate_id.value : null
    }
  }

  dynamic "customer_contacts" {
    for_each = var.spec.customer_contacts
    content {
      email = customer_contacts.value.email
    }
  }

  dynamic "read_endpoint" {
    for_each = var.spec.read_endpoint != null ? [var.spec.read_endpoint] : []
    content {
      is_enabled                   = read_endpoint.value.is_enabled
      exclude_ips                  = length(read_endpoint.value.exclude_ips) > 0 ? read_endpoint.value.exclude_ips : null
      read_endpoint_hostname_label = read_endpoint.value.read_endpoint_hostname_label != "" ? read_endpoint.value.read_endpoint_hostname_label : null
      read_endpoint_ip_address     = read_endpoint.value.read_endpoint_ip_address != "" ? read_endpoint.value.read_endpoint_ip_address : null
    }
  }

  dynamic "database_console" {
    for_each = var.spec.database_console != null ? [var.spec.database_console] : []
    content {
      status = database_console.value.status != "" ? lookup(local.database_console_status_map, database_console.value.status, null) : null
      port   = database_console.value.port > 0 ? database_console.value.port : null
    }
  }

  dynamic "rest" {
    for_each = var.spec.rest != null ? [var.spec.rest] : []
    content {
      configuration = rest.value.configuration != "" ? rest.value.configuration : null
      port          = rest.value.port > 0 ? rest.value.port : null
    }
  }

  lifecycle {
    ignore_changes = [admin_password]
  }
}
