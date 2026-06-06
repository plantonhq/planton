variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for Kubernetes Rook Ceph Cluster deployment"
  type = object({
    # Kubernetes namespace where Ceph cluster will be deployed
    namespace = optional(string, "rook-ceph")

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # Namespace where the Rook Ceph Operator is installed
    operator_namespace = optional(string, "rook-ceph")

    # The version of the Rook Ceph Cluster Helm chart to deploy
    helm_chart_version = optional(string, "v1.16.6")

    # Ceph container image configuration
    ceph_image = optional(object({
      repository        = optional(string, "quay.io/ceph/ceph")
      tag               = optional(string, "v19.2.3")
      allow_unsupported = optional(bool, false)
    }))

    # Core Ceph cluster configuration
    cluster = optional(object({
      data_dir_host_path = optional(string, "/var/lib/rook")

      mon = optional(object({
        count                   = optional(number, 3)
        allow_multiple_per_node = optional(bool, false)
      }))

      mgr = optional(object({
        count                   = optional(number, 2)
        allow_multiple_per_node = optional(bool, false)
      }))

      storage = optional(object({
        use_all_nodes   = optional(bool, true)
        use_all_devices = optional(bool, true)
        device_filter   = optional(string)
        nodes = optional(list(object({
          name          = string
          devices       = optional(list(string))
          device_filter = optional(string)
        })))
      }))

      network = optional(object({
        enable_encryption  = optional(bool, false)
        enable_compression = optional(bool, false)
        require_msgr2      = optional(bool, false)
      }))

      resources = optional(object({
        mon = optional(object({
          limits = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
          requests = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
        }))
        mgr = optional(object({
          limits = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
          requests = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
        }))
        osd = optional(object({
          limits = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
          requests = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
        }))
      }))
    }))

    # Block storage pool configuration
    block_pools = optional(list(object({
      name            = string
      failure_domain  = optional(string, "host")
      replicated_size = optional(number, 3)
      storage_class = optional(object({
        enabled                = optional(bool, true)
        name                   = string
        is_default             = optional(bool, false)
        reclaim_policy         = optional(string, "Delete")
        allow_volume_expansion = optional(bool, true)
        volume_binding_mode    = optional(string, "Immediate")
      }))
    })))

    # Filesystem configuration for CephFS
    filesystems = optional(list(object({
      name                          = string
      metadata_pool_replicated_size = optional(number, 3)
      data_pool_replicated_size     = optional(number, 3)
      failure_domain                = optional(string, "host")
      active_mds_count              = optional(number, 1)
      active_standby                = optional(bool, true)
      mds_resources = optional(object({
        limits = optional(object({
          cpu    = optional(string)
          memory = optional(string)
        }))
        requests = optional(object({
          cpu    = optional(string)
          memory = optional(string)
        }))
      }))
      storage_class = optional(object({
        enabled                = optional(bool, true)
        name                   = string
        is_default             = optional(bool, false)
        reclaim_policy         = optional(string, "Delete")
        allow_volume_expansion = optional(bool, true)
        volume_binding_mode    = optional(string, "Immediate")
      }))
    })))

    # Object store configuration for S3-compatible storage
    object_stores = optional(list(object({
      name                            = string
      metadata_pool_replicated_size   = optional(number, 3)
      data_pool_erasure_data_chunks   = optional(number, 2)
      data_pool_erasure_coding_chunks = optional(number, 1)
      failure_domain                  = optional(string, "host")
      preserve_pools_on_delete        = optional(bool, true)
      gateway_port                    = optional(number, 80)
      gateway_instances               = optional(number, 1)
      gateway_resources = optional(object({
        limits = optional(object({
          cpu    = optional(string)
          memory = optional(string)
        }))
        requests = optional(object({
          cpu    = optional(string)
          memory = optional(string)
        }))
      }))
      storage_class = optional(object({
        enabled                = optional(bool, true)
        name                   = string
        is_default             = optional(bool, false)
        reclaim_policy         = optional(string, "Delete")
        allow_volume_expansion = optional(bool, true)
        volume_binding_mode    = optional(string, "Immediate")
      }))
    })))

    # Enable the Ceph toolbox deployment for debugging
    enable_toolbox = optional(bool, false)

    # Enable Prometheus monitoring integration
    enable_monitoring = optional(bool, false)

    # Enable Ceph dashboard for web-based cluster management
    enable_dashboard = optional(bool, true)
  })
}
