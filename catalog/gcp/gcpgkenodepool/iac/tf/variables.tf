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
  description = "Specification for the GCP GKE node pool"
  type = object({
    # The GCP project the pool is created in (must be the cluster's
    # project). The CLI's tfvars converter resolves StringValueOrRef fields
    # to their literal string before the module runs, so this arrives as a
    # plain string. Empty falls back to the provider's default project.
    project_id = optional(string, "")

    # Parent cluster name (plain string after ref resolution). Immutable.
    cluster_name = string

    # Parent cluster location — region or zone (plain string after ref
    # resolution). Immutable.
    location = string

    # Pool name in GKE. Empty means "use metadata.name". Immutable.
    node_pool_name = optional(string, "")

    # Prefix for a GKE-generated pool name (mutually exclusive with
    # node_pool_name). Immutable.
    name_prefix = optional(string, "")

    # Zones the pool's nodes run in; empty inherits the cluster's
    # node_locations.
    node_locations = optional(list(string), [])

    # Explicit node Kubernetes version; empty lets GKE/auto-upgrade drive.
    version = optional(string, "")

    # Pods-per-node override (8-256); null inherits the cluster default.
    max_pods_per_node = optional(number, null)

    # Starting size for autoscaled pools (per zone). Immutable.
    initial_node_count = optional(number, null)

    # Fixed pool size (per zone). Mutually exclusive with autoscaling —
    # the proto oneof guarantees at most one arrives.
    node_count = optional(number, null)

    # Cluster-autoscaler bounds: per-zone (min/max) XOR total limits.
    autoscaling = optional(object({
      min_nodes       = optional(number, null)
      max_nodes       = optional(number, null)
      total_min_nodes = optional(number, null)
      total_max_nodes = optional(number, null)
      location_policy = optional(string, "")
    }), null)

    # Auto-repair / auto-upgrade; both default true (GKE's own defaults).
    management = optional(object({
      auto_repair  = optional(bool, true)
      auto_upgrade = optional(bool, true)
    }), null)

    # Surge or blue-green upgrade rollout controls.
    upgrade_settings = optional(object({
      max_surge       = optional(number, null)
      max_unavailable = optional(number, null)
      strategy        = optional(string, "")
      blue_green_settings = optional(object({
        standard_rollout_policy = object({
          batch_percentage    = optional(number, null)
          batch_node_count    = optional(number, null)
          batch_soak_duration = optional(string, "")
        })
        node_pool_soak_duration = optional(string, "")
      }), null)
    }), null)

    # Compact placement / TPU topology. Immutable.
    placement_policy = optional(object({
      type         = string
      policy_name  = optional(string, "")
      tpu_topology = optional(string, "")
    }), null)

    # Dynamic Workload Scheduler queued provisioning. Immutable.
    queued_provisioning_enabled = optional(bool, false)

    # Engine-side destroy stance: DELETE (default), PREVENT, or ABANDON.
    deletion_policy = optional(string, "")

    # Skip node-count drift queries against the Instance Group Managers.
    ignore_node_count_changes = optional(bool, false)

    # Hold the pool on its current version until end of support.
    exclude_upgrades_until_end_of_support = optional(bool, false)

    # Drain behavior when the pool itself is deleted or replaced.
    # Allowlist-gated: GCP support must enable customized node drain on the
    # project or the API rejects the create.
    node_drain_config = optional(object({
      grace_termination_duration            = optional(string, "")
      pdb_timeout_duration                  = optional(string, "")
      respect_pdb_during_node_pool_deletion = optional(bool, null)
    }), null)

    # Pool-level networking overrides.
    network_config = optional(object({
      create_pod_range                = optional(bool, false)
      pod_range                       = optional(string, "")
      pod_ipv4_cidr_block             = optional(string, "")
      enable_private_nodes            = optional(bool, null)
      total_egress_bandwidth_tier     = optional(string, "")
      pod_cidr_overprovision_disabled = optional(bool, false)

      # Pool-specific subnetwork (plain string after ref resolution).
      subnetwork = optional(string, "")

      # RDMA-capable accelerator network profile.
      accelerator_network_profile = optional(string, "")

      # Multi-networking: extra node interfaces and pod ranges.
      additional_node_networks = optional(list(object({
        network    = string
        subnetwork = string
      })), [])
      additional_pod_networks = optional(list(object({
        subnetwork          = optional(string, "")
        secondary_pod_range = string
        max_pods_per_node   = optional(number, null)
      })), [])
    }), null)

    # Node VM configuration shared by every node in the pool.
    node_config = optional(object({
      machine_type = optional(string, "e2-medium")
      disk_size_gb = optional(number, null)
      disk_type    = optional(string, "")
      image_type   = optional(string, "COS_CONTAINERD")

      # Service account email (plain string after ref resolution).
      service_account = optional(string, "")
      oauth_scopes    = optional(list(string), [])

      labels          = optional(map(string), {})
      resource_labels = optional(map(string), {})
      tags            = optional(list(string), [])
      metadata        = optional(map(string), {})

      taints = optional(list(object({
        key    = string
        value  = string
        effect = string
      })), [])

      spot        = optional(bool, false)
      preemptible = optional(bool, false)

      guest_accelerators = optional(list(object({
        type               = string
        count              = number
        gpu_partition_size = optional(string, "")
        gpu_driver_version = optional(string, "")
        gpu_sharing_config = optional(object({
          gpu_sharing_strategy       = string
          max_shared_clients_per_gpu = number
        }), null)
      })), [])

      shielded_instance_config = optional(object({
        enable_secure_boot          = optional(bool, false)
        enable_integrity_monitoring = optional(bool, true)
      }), null)

      confidential_nodes = optional(object({
        enabled                    = optional(bool, false)
        confidential_instance_type = optional(string, "")
      }), null)

      min_cpu_platform = optional(string, "")
      local_ssd_count  = optional(number, null)

      ephemeral_storage_local_ssd = optional(object({
        local_ssd_count  = number
        data_cache_count = optional(number, null)
      }), null)

      local_nvme_ssd_block = optional(object({
        local_ssd_count = number
      }), null)

      gcfs_enabled        = optional(bool, false)
      gvnic_enabled       = optional(bool, false)
      fast_socket_enabled = optional(bool, false)

      # KMS key path (plain string after ref resolution).
      boot_disk_kms_key = optional(string, "")

      workload_metadata_mode = optional(string, "")

      reservation_affinity = optional(object({
        consume_reservation_type = string
        key                      = optional(string, "")
        values                   = optional(list(string), [])
      }), null)

      secondary_boot_disks = optional(list(object({
        disk_image = string
        mode       = optional(string, "")
      })), [])

      kubelet_config = optional(object({
        cpu_manager_policy                     = optional(string, "")
        cpu_cfs_quota                          = optional(bool, null)
        cpu_cfs_quota_period                   = optional(string, "")
        pod_pids_limit                         = optional(number, null)
        insecure_kubelet_readonly_port_enabled = optional(string, "")
        max_parallel_image_pulls               = optional(number, null)
        container_log_max_size                 = optional(string, "")
        container_log_max_files                = optional(number, null)
        image_gc_low_threshold_percent         = optional(number, null)
        image_gc_high_threshold_percent        = optional(number, null)
        image_minimum_gc_age                   = optional(string, "")
        image_maximum_gc_age                   = optional(string, "")
        allowed_unsafe_sysctls                 = optional(list(string), [])
        eviction_max_pod_grace_period_seconds  = optional(number, null)
        single_process_oom_kill                = optional(bool, null)

        eviction_soft = optional(object({
          memory_available    = optional(string, "")
          nodefs_available    = optional(string, "")
          nodefs_inodes_free  = optional(string, "")
          imagefs_available   = optional(string, "")
          imagefs_inodes_free = optional(string, "")
          pid_available       = optional(string, "")
        }), null)
        eviction_soft_grace_period = optional(object({
          memory_available    = optional(string, "")
          nodefs_available    = optional(string, "")
          nodefs_inodes_free  = optional(string, "")
          imagefs_available   = optional(string, "")
          imagefs_inodes_free = optional(string, "")
          pid_available       = optional(string, "")
        }), null)
        # Percentage-only values ("10%") — GKE rejects absolute quantities
        # for minimum reclaim (the spec CEL enforces this upstream).
        eviction_minimum_reclaim = optional(object({
          memory_available    = optional(string, "")
          nodefs_available    = optional(string, "")
          nodefs_inodes_free  = optional(string, "")
          imagefs_available   = optional(string, "")
          imagefs_inodes_free = optional(string, "")
          pid_available       = optional(string, "")
        }), null)

        crash_loop_back_off = optional(object({
          max_container_restart_period = optional(string, "")
        }), null)
        memory_manager = optional(object({
          policy = optional(string, "")
        }), null)
        topology_manager = optional(object({
          policy = optional(string, "")
          scope  = optional(string, "")
        }), null)
        # Graceful node shutdown (Spot/preemptible pools), seconds.
        shutdown_grace_period_seconds               = optional(number, null)
        shutdown_grace_period_critical_pods_seconds = optional(number, null)
      }), null)

      linux_node_config = optional(object({
        sysctls     = optional(map(string), {})
        cgroup_mode = optional(string, "")
        hugepages_config = optional(object({
          hugepage_size_2m = optional(number, null)
          hugepage_size_1g = optional(number, null)
        }), null)
        transparent_hugepage_enabled      = optional(string, "")
        transparent_hugepage_defrag       = optional(string, "")
        node_kernel_module_loading_policy = optional(string, "")
        enable_ptp_kvm_time_sync          = optional(bool, null)
        swap_config = optional(object({
          enabled = optional(bool, null)
          boot_disk_profile = optional(object({
            swap_size_gib     = optional(number, null)
            swap_size_percent = optional(number, null)
          }), null)
          dedicated_local_ssd_profile = optional(object({
            disk_count = optional(number, null)
          }), null)
          ephemeral_local_ssd_profile = optional(object({
            swap_size_gib     = optional(number, null)
            swap_size_percent = optional(number, null)
          }), null)
          encryption_config = optional(object({
            disabled = optional(bool, null)
          }), null)
        }), null)
        # Boot-time init script: one of gcs_uri (optionally pinned to a
        # generation) or secret_manager_secret_uri.
        custom_node_init = optional(object({
          gcs_uri                   = optional(string, "")
          gcs_generation            = optional(number, null)
          secret_manager_secret_uri = optional(string, "")
        }), null)
      }), null)

      logging_variant  = optional(string, "")
      flex_start       = optional(bool, false)
      max_run_duration = optional(string, "")

      # Storage/security/tenancy additions.
      enable_confidential_storage = optional(bool, false)
      local_ssd_encryption_mode   = optional(string, "")
      gpudirect_strategy          = optional(string, "")
      node_group                  = optional(string, "")
      storage_pools               = optional(list(string), [])
      resource_manager_tags       = optional(map(string), {})

      advanced_machine_features = optional(object({
        threads_per_core             = number
        enable_nested_virtualization = optional(bool, null)
        performance_monitoring_unit  = optional(string, "")
      }), null)

      boot_disk = optional(object({
        disk_type              = optional(string, "")
        size_gb                = optional(number, null)
        provisioned_iops       = optional(number, null)
        provisioned_throughput = optional(number, null)
      }), null)

      node_image = optional(object({
        image         = string
        image_project = optional(string, "")
      }), null)

      sole_tenant_config = optional(object({
        node_affinities = list(object({
          key      = string
          operator = string
          values   = list(string)
        }))
        min_node_cpus = optional(number, null)
      }), null)

      sandbox_type                = optional(string, "")
      windows_os_version          = optional(string, "")
      host_maintenance_interval   = optional(string, "")
      architecture_taint_behavior = optional(string, "")

      containerd_config = optional(object({
        private_registry_access = optional(object({
          enabled = optional(bool, false)
          certificate_authority_domains = optional(list(object({
            fqdns                              = list(string)
            gcp_secret_manager_certificate_uri = string
          })), [])
        }), null)
        registry_hosts = optional(list(object({
          server = string
          hosts = optional(list(object({
            host                   = string
            capabilities           = optional(list(string), [])
            dial_timeout           = optional(string, "")
            override_path          = optional(bool, null)
            ca_secret_uri          = optional(string, "")
            client_cert_secret_uri = optional(string, "")
            client_key_secret_uri  = optional(string, "")
            headers                = optional(map(string), {})
          })), [])
        })), [])
        writable_cgroups_enabled = optional(bool, null)
      }), null)
    }), null)
  })

  validation {
    condition     = try(var.spec.node_config.kubelet_config.shutdown_grace_period_seconds, null) == null || (var.spec.node_config.kubelet_config.shutdown_grace_period_seconds >= 10 && var.spec.node_config.kubelet_config.shutdown_grace_period_seconds <= 10000)
    error_message = "node_config.kubelet_config.shutdown_grace_period_seconds must be between 10 and 10000."
  }

  validation {
    condition     = try(var.spec.node_config.kubelet_config.shutdown_grace_period_critical_pods_seconds, null) == null || try(var.spec.node_config.kubelet_config.shutdown_grace_period_seconds, null) == null || var.spec.node_config.kubelet_config.shutdown_grace_period_critical_pods_seconds <= var.spec.node_config.kubelet_config.shutdown_grace_period_seconds
    error_message = "node_config.kubelet_config.shutdown_grace_period_critical_pods_seconds cannot exceed shutdown_grace_period_seconds."
  }

  validation {
    condition     = try(var.spec.node_config.linux_node_config.custom_node_init, null) == null || ((var.spec.node_config.linux_node_config.custom_node_init.gcs_uri != "") != (var.spec.node_config.linux_node_config.custom_node_init.secret_manager_secret_uri != ""))
    error_message = "node_config.linux_node_config.custom_node_init must set exactly one of gcs_uri or secret_manager_secret_uri."
  }
}
