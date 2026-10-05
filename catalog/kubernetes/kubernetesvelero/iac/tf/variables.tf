variable "metadata" {
  description = "Catalog object metadata"
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
  description = "KubernetesVelero specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    crds = optional(object({
      upgrade_on_install   = optional(bool)
      cleanup_on_uninstall = optional(bool, false)
    }))
    backup_storage = object({
      s3 = optional(object({
        bucket           = string
        region           = string
        s3_url           = optional(string, "")
        force_path_style = optional(bool, false)
        kms_key_id       = optional(string, "")
        ca_cert          = optional(string, "")
        irsa_role_arn    = optional(string, "")
        access_keys = optional(object({
          access_key_id     = string
          secret_access_key = string
        }))
      }))
      gcs = optional(object({
        bucket                                  = string
        workload_identity_service_account_email = optional(string, "")
        service_account_key_json                = optional(string, "")
      }))
      azure_blob = optional(object({
        storage_account             = string
        container                   = string
        resource_group              = string
        subscription_id             = string
        workload_identity_client_id = optional(string, "")
        service_principal = optional(object({
          tenant_id     = string
          client_id     = string
          client_secret = string
        }))
      }))
      prefix       = optional(string, "")
      plugin_image = optional(string, "")
    })
    volume_snapshots = optional(object({
      enabled                    = optional(bool)
      enable_csi                 = optional(bool, false)
      default_snapshot_move_data = optional(bool, false)
    }))
    fs_backup = optional(object({
      deploy_node_agent            = optional(bool, false)
      default_volumes_to_fs_backup = optional(bool, false)
      node_agent_resources = optional(object({
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      node_agent_tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
    }))
    schedules = optional(list(object({
      name                         = string
      schedule                     = string
      paused                       = optional(bool, false)
      ttl                          = optional(string, "")
      included_namespaces          = optional(list(string), [])
      excluded_namespaces          = optional(list(string), [])
      included_resources           = optional(list(string), [])
      excluded_resources           = optional(list(string), [])
      label_selector               = optional(map(string), {})
      include_cluster_resources    = optional(bool)
      snapshot_volumes             = optional(bool)
      default_volumes_to_fs_backup = optional(bool)
      storage_location             = optional(string, "")
    })), [])
    server = optional(object({
      default_backup_ttl             = optional(string, "")
      default_item_operation_timeout = optional(string, "")
      garbage_collection_frequency   = optional(string, "")
      restore_only_mode              = optional(bool, false)
      log_level                      = optional(string)
      log_format                     = optional(string)
    }))
    deployment = optional(object({
      resources = optional(object({
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      priority_class_name = optional(string, "")
      node_selector       = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
    }))
    prometheus = optional(object({
      enabled         = optional(bool)
      service_monitor = optional(bool, false)
    }))
    helm_values = optional(string, "")
  })

  # NOT marked sensitive as a whole: Terraform cannot mark individual
  # object attributes, and a blanket-sensitive spec would force every
  # derived output (namespace, release name) to be sensitive and render
  # the whole plan opaque. Instead the module isolates the credential
  # material (S3 secret access key / GCP key JSON / Azure client secret)
  # into a dedicated values document wrapped with sensitive() in
  # locals.tf — the only place those fields are ever referenced.
}
