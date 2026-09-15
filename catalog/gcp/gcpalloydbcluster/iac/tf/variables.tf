variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
}

variable "spec" {
  description = "GcpAlloydbCluster specification"
  type = object({
    project_id   = optional(string, "")
    cluster_name = string
    location     = string
    network      = optional(string, "")
    psc_config = optional(object({
      psc_enabled = optional(bool, false)
    }), null)
    cluster_type = optional(string, "")
    secondary_config = optional(object({
      primary_cluster_name = string
    }), null)
    allocated_ip_range = optional(string, "")
    database_version   = optional(string, "")
    display_name       = optional(string, "")
    initial_user = optional(object({
      password = string
      user     = optional(string, "")
    }), null)
    automated_backup_policy = optional(object({
      enabled                        = optional(bool, true)
      backup_window                  = optional(string, "")
      location                       = optional(string, "")
      quantity_based_retention_count = optional(number, 0)
      time_based_retention_period    = optional(string, "")
      weekly_schedule = optional(object({
        days_of_week = optional(list(string), [])
        start_hour   = optional(number, 0)
      }), null)
      encryption_kms_key_name = optional(string, "")
      # Labels applied to every backup created by this policy.
      labels = optional(map(string), {})
    }), null)
    continuous_backup_config = optional(object({
      enabled                 = optional(bool, true)
      recovery_window_days    = optional(number, 0)
      encryption_kms_key_name = optional(string, "")
    }), null)
    kms_key_name = optional(string, "")
    maintenance_window = optional(object({
      day        = string
      start_hour = number
    }), null)
    annotations                      = optional(map(string), {})
    subscription_type                = optional(string, "")
    skip_await_major_version_upgrade = optional(bool, false)

    # User labels applied to the cluster and its bundled primary; merged
    # beneath the platform attribution labels.
    labels = optional(map(string), {})

    # Dataplex Universal Catalog integration (GCP enables it by default
    # when absent).
    dataplex_config = optional(object({
      # Required by the API inside the block; no literal default, so an unset
      # value can never be mistaken for an explicit false.
      enabled = optional(bool)
    }), null)

    # Restore provenance — at most one source (spec-enforced); all
    # create-time only (ForceNew).
    restore_backup_source = optional(object({
      backup_name = string
    }), null)
    restore_continuous_backup_source = optional(object({
      cluster       = string
      point_in_time = string
    }), null)
    restore_backupdr_backup_source = optional(object({
      backup = string
    }), null)
    restore_backupdr_pitr_source = optional(object({
      data_source   = string
      point_in_time = string
    }), null)

    # Client-side destroy guard (provider default TRUE). Always sent
    # explicitly so the spec is the single source of truth.
    deletion_protection = optional(bool, true)

    # Cluster destroy behavior: DEFAULT (provider default), FORCE (deletes
    # contained instances too), PREVENT, or ABANDON.
    deletion_policy = optional(string, "")

    primary_instance = object({
      instance_id       = string
      cpu_count         = optional(number, 0)
      machine_type      = optional(string, "")
      availability_type = optional(string, "")
      database_flags    = optional(map(string), {})
      display_name      = optional(string, "")
      query_insights_config = optional(object({
        query_plans_per_minute  = optional(number, 5)
        query_string_length     = optional(number, 1024)
        record_application_tags = optional(bool, true)
        record_client_address   = optional(bool, true)
      }), null)
      require_connectors = optional(bool, false)
      ssl_mode           = optional(string, "")

      # Stop/start lever (ALWAYS keeps serving; NEVER stops compute).
      activation_policy = optional(string, "")

      # Client tool metadata (annotations, not labels).
      annotations = optional(map(string), {})

      # Pin a ZONAL primary to a specific zone (ZONAL only).
      gce_zone = optional(string, "")

      # AlloyDB managed connection pooling (built-in pooler).
      connection_pool_config = optional(object({
        enabled = optional(bool, false)
        flags   = optional(map(string), {})
      }), null)

      enable_public_ip          = optional(bool, false)
      enable_outbound_public_ip = optional(bool, false)
      authorized_external_networks = optional(list(object({
        cidr_range = string
      })), [])

      # Draw the primary's private IPs from a specific PSA allocated range
      # instead of the cluster's. Immutable.
      allocated_ip_range_override = optional(string, "")

      # PSC on the bundled primary (PSC clusters only).
      psc_instance_config = optional(object({
        allowed_consumer_projects = optional(list(string), [])
        psc_auto_connections = optional(list(object({
          consumer_network = optional(string, "")
          consumer_project = optional(string, "")
        })), [])
        psc_interface_configs = optional(list(object({
          network_attachment_resource = string
        })), [])
      }), null)

      # Primary-instance destroy behavior: DELETE (provider default),
      # PREVENT, or ABANDON.
      deletion_policy = optional(string, "")
    })
  })
}
