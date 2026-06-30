###############################################################################
# Cloud Composer Environment (Managed Apache Airflow)
###############################################################################

resource "google_composer_environment" "environment" {
  name    = local.environment_name
  region  = var.spec.region
  project = var.spec.project_id.value
  labels  = local.labels

  config {

    # -- Node config (networking) ------------------------------------------

    dynamic "node_config" {
      for_each = var.spec.node_config != null ? [var.spec.node_config] : []
      content {
        network        = node_config.value.network != null ? node_config.value.network.value : null
        subnetwork     = node_config.value.subnetwork != null ? node_config.value.subnetwork.value : null
        service_account = node_config.value.service_account != null ? node_config.value.service_account.value : null
        tags           = length(node_config.value.tags) > 0 ? node_config.value.tags : null

        composer_network_attachment        = node_config.value.composer_network_attachment != "" ? node_config.value.composer_network_attachment : null
        composer_internal_ipv4_cidr_block = node_config.value.composer_internal_ipv4_cidr_block != "" ? node_config.value.composer_internal_ipv4_cidr_block : null
      }
    }

    # -- Software config ---------------------------------------------------

    dynamic "software_config" {
      for_each = var.spec.software_config != null ? [var.spec.software_config] : []
      content {
        image_version            = software_config.value.image_version != "" ? software_config.value.image_version : null
        airflow_config_overrides = length(software_config.value.airflow_config_overrides) > 0 ? software_config.value.airflow_config_overrides : null
        pypi_packages            = length(software_config.value.pypi_packages) > 0 ? software_config.value.pypi_packages : null
        env_variables            = length(software_config.value.env_variables) > 0 ? software_config.value.env_variables : null
        web_server_plugins_mode  = software_config.value.web_server_plugins_mode != "" ? software_config.value.web_server_plugins_mode : null
      }
    }

    # -- Private environment config (Composer 2.x) -------------------------

    dynamic "private_environment_config" {
      for_each = var.spec.private_environment_config != null ? [var.spec.private_environment_config] : []
      content {
        enable_private_endpoint                = private_environment_config.value.enable_private_endpoint
        connection_type                        = private_environment_config.value.connection_type != "" ? private_environment_config.value.connection_type : null
        master_ipv4_cidr_block                 = private_environment_config.value.master_ipv4_cidr_block != "" ? private_environment_config.value.master_ipv4_cidr_block : null
        cloud_sql_ipv4_cidr_block              = private_environment_config.value.cloud_sql_ipv4_cidr_block != "" ? private_environment_config.value.cloud_sql_ipv4_cidr_block : null
        cloud_composer_network_ipv4_cidr_block = private_environment_config.value.cloud_composer_network_ipv4_cidr_block != "" ? private_environment_config.value.cloud_composer_network_ipv4_cidr_block : null
        cloud_composer_connection_subnetwork   = private_environment_config.value.cloud_composer_connection_subnetwork != "" ? private_environment_config.value.cloud_composer_connection_subnetwork : null
        enable_privately_used_public_ips       = private_environment_config.value.enable_privately_used_public_ips
      }
    }

    # -- Workloads config --------------------------------------------------

    dynamic "workloads_config" {
      for_each = var.spec.workloads_config != null ? [var.spec.workloads_config] : []
      content {
        dynamic "scheduler" {
          for_each = workloads_config.value.scheduler != null ? [workloads_config.value.scheduler] : []
          content {
            cpu        = scheduler.value.cpu > 0 ? scheduler.value.cpu : null
            memory_gb  = scheduler.value.memory_gb > 0 ? scheduler.value.memory_gb : null
            storage_gb = scheduler.value.storage_gb > 0 ? scheduler.value.storage_gb : null
            count      = scheduler.value.count > 0 ? scheduler.value.count : null
          }
        }

        dynamic "web_server" {
          for_each = workloads_config.value.web_server != null ? [workloads_config.value.web_server] : []
          content {
            cpu        = web_server.value.cpu > 0 ? web_server.value.cpu : null
            memory_gb  = web_server.value.memory_gb > 0 ? web_server.value.memory_gb : null
            storage_gb = web_server.value.storage_gb > 0 ? web_server.value.storage_gb : null
          }
        }

        dynamic "worker" {
          for_each = workloads_config.value.worker != null ? [workloads_config.value.worker] : []
          content {
            cpu        = worker.value.cpu > 0 ? worker.value.cpu : null
            memory_gb  = worker.value.memory_gb > 0 ? worker.value.memory_gb : null
            storage_gb = worker.value.storage_gb > 0 ? worker.value.storage_gb : null
            min_count  = worker.value.min_count > 0 ? worker.value.min_count : null
            max_count  = worker.value.max_count > 0 ? worker.value.max_count : null
          }
        }

        dynamic "triggerer" {
          for_each = workloads_config.value.triggerer != null ? [workloads_config.value.triggerer] : []
          content {
            cpu       = triggerer.value.cpu
            memory_gb = triggerer.value.memory_gb
            count     = triggerer.value.count
          }
        }

        dynamic "dag_processor" {
          for_each = workloads_config.value.dag_processor != null ? [workloads_config.value.dag_processor] : []
          content {
            cpu        = dag_processor.value.cpu > 0 ? dag_processor.value.cpu : null
            memory_gb  = dag_processor.value.memory_gb > 0 ? dag_processor.value.memory_gb : null
            storage_gb = dag_processor.value.storage_gb > 0 ? dag_processor.value.storage_gb : null
            count      = dag_processor.value.count > 0 ? dag_processor.value.count : null
          }
        }
      }
    }

    # -- Environment size --------------------------------------------------

    environment_size = var.spec.environment_size != "" ? var.spec.environment_size : null

    # -- Resilience mode ---------------------------------------------------

    resilience_mode = var.spec.resilience_mode != "" ? var.spec.resilience_mode : null

    # -- Encryption config (CMEK) ------------------------------------------

    dynamic "encryption_config" {
      for_each = var.spec.kms_key_name != null ? [var.spec.kms_key_name] : []
      content {
        kms_key_name = encryption_config.value.value
      }
    }

    # -- Maintenance window ------------------------------------------------

    dynamic "maintenance_window" {
      for_each = var.spec.maintenance_window != null ? [var.spec.maintenance_window] : []
      content {
        start_time = maintenance_window.value.start_time
        end_time   = maintenance_window.value.end_time
        recurrence = maintenance_window.value.recurrence
      }
    }

    # -- Recovery config ---------------------------------------------------

    dynamic "recovery_config" {
      for_each = var.spec.recovery_config != null ? [var.spec.recovery_config] : []
      content {
        scheduled_snapshots_config {
          enabled                    = recovery_config.value.enabled
          snapshot_location          = recovery_config.value.snapshot_location != "" ? recovery_config.value.snapshot_location : null
          snapshot_creation_schedule = recovery_config.value.snapshot_creation_schedule != "" ? recovery_config.value.snapshot_creation_schedule : null
          time_zone                  = recovery_config.value.time_zone != "" ? recovery_config.value.time_zone : null
        }
      }
    }

    # -- Web server network access control ---------------------------------

    dynamic "web_server_network_access_control" {
      for_each = var.spec.web_server_network_access_control != null ? [var.spec.web_server_network_access_control] : []
      content {
        dynamic "allowed_ip_range" {
          for_each = web_server_network_access_control.value.allowed_ip_ranges
          content {
            value       = allowed_ip_range.value.value
            description = allowed_ip_range.value.description != "" ? allowed_ip_range.value.description : null
          }
        }
      }
    }

    # -- Composer 3 private environment flags ------------------------------

    enable_private_environment = var.spec.enable_private_environment ? true : null
    enable_private_builds_only = var.spec.enable_private_builds_only ? true : null
  }
}
