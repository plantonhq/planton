variable "spec" {
  description = "GcpCloudComposerEnvironmentSpec"
  type = object({
    project_id = object({
      value = string
    })
    region           = string
    environment_name = optional(string, "")

    node_config = optional(object({
      network = optional(object({
        value = string
      }), null)
      subnetwork = optional(object({
        value = string
      }), null)
      service_account = optional(object({
        value = string
      }), null)
      tags                              = optional(list(string), [])
      composer_network_attachment        = optional(string, "")
      composer_internal_ipv4_cidr_block = optional(string, "")
    }), null)

    software_config = optional(object({
      image_version            = optional(string, "")
      airflow_config_overrides = optional(map(string), {})
      pypi_packages            = optional(map(string), {})
      env_variables            = optional(map(string), {})
      web_server_plugins_mode  = optional(string, "")
    }), null)

    private_environment_config = optional(object({
      enable_private_endpoint                   = optional(bool, false)
      connection_type                           = optional(string, "")
      master_ipv4_cidr_block                    = optional(string, "")
      cloud_sql_ipv4_cidr_block                 = optional(string, "")
      cloud_composer_network_ipv4_cidr_block    = optional(string, "")
      cloud_composer_connection_subnetwork      = optional(string, "")
      enable_privately_used_public_ips          = optional(bool, false)
    }), null)

    workloads_config = optional(object({
      scheduler = optional(object({
        cpu        = optional(number, 0)
        memory_gb  = optional(number, 0)
        storage_gb = optional(number, 0)
        count      = optional(number, 0)
      }), null)
      web_server = optional(object({
        cpu        = optional(number, 0)
        memory_gb  = optional(number, 0)
        storage_gb = optional(number, 0)
      }), null)
      worker = optional(object({
        cpu        = optional(number, 0)
        memory_gb  = optional(number, 0)
        storage_gb = optional(number, 0)
        min_count  = optional(number, 0)
        max_count  = optional(number, 0)
      }), null)
      triggerer = optional(object({
        cpu       = optional(number, 0)
        memory_gb = optional(number, 0)
        count     = optional(number, 0)
      }), null)
      dag_processor = optional(object({
        cpu        = optional(number, 0)
        memory_gb  = optional(number, 0)
        storage_gb = optional(number, 0)
        count      = optional(number, 0)
      }), null)
    }), null)

    environment_size = optional(string, "")
    resilience_mode  = optional(string, "")

    kms_key_name = optional(object({
      value = string
    }), null)

    maintenance_window = optional(object({
      start_time = string
      end_time   = string
      recurrence = string
    }), null)

    recovery_config = optional(object({
      enabled                    = optional(bool, false)
      snapshot_location          = optional(string, "")
      snapshot_creation_schedule = optional(string, "")
      time_zone                  = optional(string, "")
    }), null)

    web_server_network_access_control = optional(object({
      allowed_ip_ranges = optional(list(object({
        value       = string
        description = optional(string, "")
      })), [])
    }), null)

    enable_private_environment = optional(bool, false)
    enable_private_builds_only = optional(bool, false)
  })
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
  default = {
    name = ""
  }
}
