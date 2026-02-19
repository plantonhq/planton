variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "OciAutonomousDatabase specification"
  type = object({
    compartment_id = object({
      value = string
    })

    db_name      = string
    display_name = optional(string, "")

    db_workload      = optional(string, "")
    db_version       = optional(string, "")
    database_edition = optional(string, "")
    license_model    = optional(string, "")
    character_set    = optional(string, "")
    ncharacter_set   = optional(string, "")

    compute_model = optional(string, "")
    compute_count = optional(number, null)

    data_storage_size_in_tbs = optional(number, 0)
    data_storage_size_in_gb  = optional(number, 0)

    is_auto_scaling_enabled             = optional(bool, null)
    is_auto_scaling_for_storage_enabled = optional(bool, null)

    admin_password        = optional(string, "")
    secret_id = optional(object({
      value = string
    }), null)
    secret_version_number = optional(number, 0)

    subnet_id = optional(object({
      value = string
    }), null)
    nsg_ids = optional(list(object({
      value = string
    })), [])
    private_endpoint_label = optional(string, "")
    private_endpoint_ip    = optional(string, "")
    whitelisted_ips        = optional(list(string), [])

    is_mtls_connection_required = optional(bool, null)
    is_access_control_enabled   = optional(bool, null)

    kms_key_id = optional(object({
      value = string
    }), null)
    vault_id = optional(object({
      value = string
    }), null)

    is_dedicated = optional(bool, null)
    is_free_tier = optional(bool, null)
    is_dev_tier  = optional(bool, null)

    autonomous_container_database_id = optional(object({
      value = string
    }), null)

    backup_retention_period_in_days = optional(number, 0)

    is_local_data_guard_enabled = optional(bool, null)

    autonomous_maintenance_schedule_type = optional(string, "")

    customer_contacts = optional(list(object({
      email = string
    })), [])
  })
}
