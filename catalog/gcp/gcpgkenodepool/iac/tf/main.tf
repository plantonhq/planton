# Enable the Kubernetes Engine API so a fresh project can host the pool.
# disable_on_destroy is false: tearing down one node pool must never disable
# the API for everything else in the project (including its own cluster).
resource "google_project_service" "container_api" {
  project = local.project_id
  service = "container.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A GKE node pool — a group of identically configured Compute Engine VMs
# attached to a GcpGkeCluster. The cluster and location arrive as plain
# strings resolved from the cluster's outputs, so the pool addresses its
# parent exactly the way GKE named it — no lookup, no wildcard search.
#
# Lifecycle notes the API enforces:
#   - name, location, initial_node_count, max_pods_per_node,
#     placement_policy, queued_provisioning, and nearly all of node_config
#     (machine type, disks, image, identity, accelerators, shielded/
#     confidential settings, local SSDs) are immutable — changing them
#     replaces the pool (GKE drains and recreates the nodes).
#   - node_count, autoscaling, management, upgrade_settings, node_locations,
#     labels, taints, tags, and resource_labels update in place.
#   - For autoscaled pools the autoscaler owns node_count at runtime;
#     lifecycle.ignore_changes keeps the plan from fighting it.
resource "google_container_node_pool" "this" {
  name        = local.node_pool_name
  name_prefix = local.name_prefix
  project     = local.project_id
  cluster     = var.spec.cluster_name
  location    = var.spec.location

  node_locations = length(var.spec.node_locations) > 0 ? var.spec.node_locations : null

  # Engine-side destroy stance; empty inherits the provider default
  # (DELETE). PREVENT/ABANDON change what a destroy is allowed to do.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Quota/performance switch: skip the per-pool Instance Group Manager
  # queries that reconcile observed node counts on every read.
  ignore_node_count_changes = var.spec.ignore_node_count_changes ? true : null

  # Explicit version and auto_upgrade fight each other (the API re-upgrades
  # what the plan pins); the spec documents the trade and the field passes
  # through untouched.
  version = local.version

  max_pods_per_node  = var.spec.max_pods_per_node
  initial_node_count = var.spec.initial_node_count

  # Fixed size XOR autoscaling — the proto oneof guarantees at most one
  # arrives. Neither means GKE's default size (3 nodes), unmanaged.
  node_count = var.spec.node_count

  dynamic "autoscaling" {
    for_each = var.spec.autoscaling != null ? [var.spec.autoscaling] : []
    content {
      # Per-zone bounds XOR total bounds (spec-level CEL enforces the
      # exclusivity); nulls drop the unused arm from the API payload.
      min_node_count       = autoscaling.value.min_nodes
      max_node_count       = autoscaling.value.max_nodes
      total_min_node_count = autoscaling.value.total_min_nodes
      total_max_node_count = autoscaling.value.total_max_nodes
      location_policy      = autoscaling.value.location_policy != "" ? autoscaling.value.location_policy : null
    }
  }

  # Auto-repair/auto-upgrade both default true — GKE's own defaults — so an
  # omitted management block and GKE's behavior agree.
  management {
    auto_repair  = try(var.spec.management.auto_repair, true)
    auto_upgrade = try(var.spec.management.auto_upgrade, true)
  }

  # Emitted only when the spec configures it: GKE's default surge settings
  # (max_surge=1, max_unavailable=0) apply otherwise, and omitting the
  # block keeps the plan clean on pools that never touch it.
  dynamic "upgrade_settings" {
    for_each = var.spec.upgrade_settings != null ? [var.spec.upgrade_settings] : []
    content {
      max_surge       = upgrade_settings.value.max_surge
      max_unavailable = upgrade_settings.value.max_unavailable
      strategy        = upgrade_settings.value.strategy != "" ? upgrade_settings.value.strategy : null

      dynamic "blue_green_settings" {
        for_each = upgrade_settings.value.blue_green_settings != null ? [upgrade_settings.value.blue_green_settings] : []
        content {
          node_pool_soak_duration = blue_green_settings.value.node_pool_soak_duration != "" ? blue_green_settings.value.node_pool_soak_duration : null

          standard_rollout_policy {
            batch_percentage    = blue_green_settings.value.standard_rollout_policy.batch_percentage
            batch_node_count    = blue_green_settings.value.standard_rollout_policy.batch_node_count
            batch_soak_duration = blue_green_settings.value.standard_rollout_policy.batch_soak_duration != "" ? blue_green_settings.value.standard_rollout_policy.batch_soak_duration : null
          }
        }
      }
    }
  }

  dynamic "placement_policy" {
    for_each = var.spec.placement_policy != null ? [var.spec.placement_policy] : []
    content {
      type         = placement_policy.value.type
      policy_name  = placement_policy.value.policy_name != "" ? placement_policy.value.policy_name : null
      tpu_topology = placement_policy.value.tpu_topology != "" ? placement_policy.value.tpu_topology : null
    }
  }

  dynamic "queued_provisioning" {
    for_each = var.spec.queued_provisioning_enabled ? [1] : []
    content {
      enabled = true
    }
  }

  # Drain pacing for pool deletion/replacement — distinct from
  # upgrade_settings, which paces upgrades of a pool that keeps existing.
  dynamic "node_drain_config" {
    for_each = var.spec.node_drain_config != null ? [var.spec.node_drain_config] : []
    content {
      grace_termination_duration            = node_drain_config.value.grace_termination_duration != "" ? node_drain_config.value.grace_termination_duration : null
      pdb_timeout_duration                  = node_drain_config.value.pdb_timeout_duration != "" ? node_drain_config.value.pdb_timeout_duration : null
      respect_pdb_during_node_pool_deletion = node_drain_config.value.respect_pdb_during_node_pool_deletion
    }
  }

  # Hold the pool on its version until end of support: emitted only when
  # requested, so an unset spec leaves GKE's automatic upgrades untouched.
  # start_time and end_time are API-computed (read back, never sent).
  dynamic "maintenance_policy" {
    for_each = var.spec.exclude_upgrades_until_end_of_support ? [1] : []
    content {
      exclusion_until_end_of_support {
        enabled = true
      }
    }
  }

  dynamic "network_config" {
    for_each = var.spec.network_config != null ? [var.spec.network_config] : []
    content {
      create_pod_range            = network_config.value.create_pod_range
      pod_range                   = network_config.value.pod_range != "" ? network_config.value.pod_range : null
      pod_ipv4_cidr_block         = network_config.value.pod_ipv4_cidr_block != "" ? network_config.value.pod_ipv4_cidr_block : null
      enable_private_nodes        = network_config.value.enable_private_nodes
      subnetwork                  = network_config.value.subnetwork != "" ? network_config.value.subnetwork : null
      accelerator_network_profile = network_config.value.accelerator_network_profile != "" ? network_config.value.accelerator_network_profile : null

      dynamic "network_performance_config" {
        for_each = network_config.value.total_egress_bandwidth_tier != "" ? [1] : []
        content {
          total_egress_bandwidth_tier = network_config.value.total_egress_bandwidth_tier
        }
      }

      dynamic "pod_cidr_overprovision_config" {
        for_each = network_config.value.pod_cidr_overprovision_disabled ? [1] : []
        content {
          disabled = true
        }
      }

      dynamic "additional_node_network_configs" {
        for_each = network_config.value.additional_node_networks
        content {
          network    = additional_node_network_configs.value.network
          subnetwork = additional_node_network_configs.value.subnetwork
        }
      }

      dynamic "additional_pod_network_configs" {
        for_each = network_config.value.additional_pod_networks
        content {
          subnetwork          = additional_pod_network_configs.value.subnetwork != "" ? additional_pod_network_configs.value.subnetwork : null
          secondary_pod_range = additional_pod_network_configs.value.secondary_pod_range
          max_pods_per_node   = additional_pod_network_configs.value.max_pods_per_node
        }
      }
    }
  }

  # node_config is always emitted: the platform attribution resource_labels
  # and the disable-legacy-endpoints metadata guard apply to every pool,
  # spec block or not. Everything else inside honors "unset means GKE
  # default" by passing null.
  node_config {
    machine_type    = local.machine_type
    disk_size_gb    = try(local.nc.disk_size_gb, null)
    disk_type       = local.disk_type
    image_type      = local.image_type
    service_account = local.service_account

    # Empty applies GKE's default scopes; with Workload Identity, workload
    # permissions come from IAM on Kubernetes service accounts, not node
    # scopes, so the defaults are normally right.
    oauth_scopes = length(try(local.nc.oauth_scopes, [])) > 0 ? local.nc.oauth_scopes : null

    # Kubernetes node labels (nodeSelector targets) — distinct from the GCE
    # resource labels below, which carry the platform attribution.
    labels          = try(local.nc.labels, {})
    resource_labels = local.final_resource_labels
    tags            = length(try(local.nc.tags, [])) > 0 ? local.nc.tags : null
    metadata        = local.node_metadata

    dynamic "taint" {
      for_each = try(local.nc.taints, [])
      content {
        key    = taint.value.key
        value  = taint.value.value
        effect = taint.value.effect
      }
    }

    # spot supersedes preemptible (no 24h lifetime); spec-level CEL rejects
    # both together, so each passes through independently.
    spot        = try(local.nc.spot, false)
    preemptible = try(local.nc.preemptible, false)

    dynamic "guest_accelerator" {
      for_each = try(local.nc.guest_accelerators, [])
      content {
        type               = guest_accelerator.value.type
        count              = guest_accelerator.value.count
        gpu_partition_size = guest_accelerator.value.gpu_partition_size != "" ? guest_accelerator.value.gpu_partition_size : null

        dynamic "gpu_driver_installation_config" {
          for_each = guest_accelerator.value.gpu_driver_version != "" ? [1] : []
          content {
            gpu_driver_version = guest_accelerator.value.gpu_driver_version
          }
        }

        dynamic "gpu_sharing_config" {
          for_each = guest_accelerator.value.gpu_sharing_config != null ? [guest_accelerator.value.gpu_sharing_config] : []
          content {
            gpu_sharing_strategy       = gpu_sharing_config.value.gpu_sharing_strategy
            max_shared_clients_per_gpu = gpu_sharing_config.value.max_shared_clients_per_gpu
          }
        }
      }
    }

    dynamic "shielded_instance_config" {
      for_each = try(local.nc.shielded_instance_config, null) != null ? [local.nc.shielded_instance_config] : []
      content {
        enable_secure_boot          = shielded_instance_config.value.enable_secure_boot
        enable_integrity_monitoring = shielded_instance_config.value.enable_integrity_monitoring
      }
    }

    dynamic "confidential_nodes" {
      for_each = try(local.nc.confidential_nodes, null) != null ? [local.nc.confidential_nodes] : []
      content {
        enabled                    = confidential_nodes.value.enabled
        confidential_instance_type = confidential_nodes.value.confidential_instance_type != "" ? confidential_nodes.value.confidential_instance_type : null
      }
    }

    min_cpu_platform = local.min_cpu_platform
    local_ssd_count  = try(local.nc.local_ssd_count, null)

    dynamic "ephemeral_storage_local_ssd_config" {
      for_each = try(local.nc.ephemeral_storage_local_ssd, null) != null ? [local.nc.ephemeral_storage_local_ssd] : []
      content {
        local_ssd_count  = ephemeral_storage_local_ssd_config.value.local_ssd_count
        data_cache_count = ephemeral_storage_local_ssd_config.value.data_cache_count
      }
    }

    dynamic "local_nvme_ssd_block_config" {
      for_each = try(local.nc.local_nvme_ssd_block, null) != null ? [local.nc.local_nvme_ssd_block] : []
      content {
        local_ssd_count = local_nvme_ssd_block_config.value.local_ssd_count
      }
    }

    dynamic "gcfs_config" {
      for_each = try(local.nc.gcfs_enabled, false) ? [1] : []
      content {
        enabled = true
      }
    }

    dynamic "gvnic" {
      for_each = try(local.nc.gvnic_enabled, false) ? [1] : []
      content {
        enabled = true
      }
    }

    dynamic "fast_socket" {
      for_each = try(local.nc.fast_socket_enabled, false) ? [1] : []
      content {
        enabled = true
      }
    }

    boot_disk_kms_key = local.boot_disk_kms_key

    # Storage/security/tenancy pass-throughs — null when unset so GKE
    # defaults stay in charge and re-plans stay clean.
    enable_confidential_storage = try(local.nc.enable_confidential_storage, false) ? true : null
    local_ssd_encryption_mode   = try(local.nc.local_ssd_encryption_mode, "") != "" ? local.nc.local_ssd_encryption_mode : null
    gpudirect_strategy          = try(local.nc.gpudirect_strategy, "") != "" ? local.nc.gpudirect_strategy : null
    node_group                  = try(local.nc.node_group, "") != "" ? local.nc.node_group : null
    storage_pools               = length(try(local.nc.storage_pools, [])) > 0 ? local.nc.storage_pools : null
    resource_manager_tags       = length(try(local.nc.resource_manager_tags, {})) > 0 ? local.nc.resource_manager_tags : null

    dynamic "advanced_machine_features" {
      for_each = try(local.nc.advanced_machine_features, null) != null ? [local.nc.advanced_machine_features] : []
      content {
        threads_per_core             = advanced_machine_features.value.threads_per_core
        enable_nested_virtualization = advanced_machine_features.value.enable_nested_virtualization
        performance_monitoring_unit  = advanced_machine_features.value.performance_monitoring_unit != "" ? advanced_machine_features.value.performance_monitoring_unit : null
      }
    }

    dynamic "boot_disk" {
      for_each = try(local.nc.boot_disk, null) != null ? [local.nc.boot_disk] : []
      content {
        disk_type              = boot_disk.value.disk_type != "" ? boot_disk.value.disk_type : null
        size_gb                = boot_disk.value.size_gb
        provisioned_iops       = boot_disk.value.provisioned_iops
        provisioned_throughput = boot_disk.value.provisioned_throughput
      }
    }

    dynamic "node_image_config" {
      for_each = try(local.nc.node_image, null) != null ? [local.nc.node_image] : []
      content {
        image         = node_image_config.value.image
        image_project = node_image_config.value.image_project != "" ? node_image_config.value.image_project : null
      }
    }

    dynamic "sole_tenant_config" {
      for_each = try(local.nc.sole_tenant_config, null) != null ? [local.nc.sole_tenant_config] : []
      content {
        min_node_cpus = sole_tenant_config.value.min_node_cpus

        dynamic "node_affinity" {
          for_each = sole_tenant_config.value.node_affinities
          content {
            key      = node_affinity.value.key
            operator = node_affinity.value.operator
            values   = node_affinity.value.values
          }
        }
      }
    }

    dynamic "sandbox_config" {
      for_each = try(local.nc.sandbox_type, "") != "" ? [1] : []
      content {
        type = local.nc.sandbox_type
      }
    }

    dynamic "windows_node_config" {
      for_each = try(local.nc.windows_os_version, "") != "" ? [1] : []
      content {
        osversion = local.nc.windows_os_version
      }
    }

    dynamic "host_maintenance_policy" {
      for_each = try(local.nc.host_maintenance_interval, "") != "" ? [1] : []
      content {
        maintenance_interval = local.nc.host_maintenance_interval
      }
    }

    dynamic "taint_config" {
      for_each = try(local.nc.architecture_taint_behavior, "") != "" ? [1] : []
      content {
        architecture_taint_behavior = local.nc.architecture_taint_behavior
      }
    }

    dynamic "containerd_config" {
      for_each = try(local.nc.containerd_config, null) != null ? [local.nc.containerd_config] : []
      content {
        dynamic "private_registry_access_config" {
          for_each = containerd_config.value.private_registry_access != null ? [containerd_config.value.private_registry_access] : []
          content {
            enabled = private_registry_access_config.value.enabled

            dynamic "certificate_authority_domain_config" {
              for_each = private_registry_access_config.value.certificate_authority_domains
              content {
                fqdns = certificate_authority_domain_config.value.fqdns

                gcp_secret_manager_certificate_config {
                  secret_uri = certificate_authority_domain_config.value.gcp_secret_manager_certificate_uri
                }
              }
            }
          }
        }

        dynamic "registry_hosts" {
          for_each = containerd_config.value.registry_hosts
          content {
            server = registry_hosts.value.server

            dynamic "hosts" {
              for_each = registry_hosts.value.hosts
              content {
                host          = hosts.value.host
                capabilities  = length(hosts.value.capabilities) > 0 ? hosts.value.capabilities : null
                dial_timeout  = hosts.value.dial_timeout != "" ? hosts.value.dial_timeout : null
                override_path = hosts.value.override_path

                dynamic "ca" {
                  for_each = hosts.value.ca_secret_uri != "" ? [1] : []
                  content {
                    gcp_secret_manager_secret_uri = hosts.value.ca_secret_uri
                  }
                }

                dynamic "client" {
                  for_each = (hosts.value.client_cert_secret_uri != "" || hosts.value.client_key_secret_uri != "") ? [1] : []
                  content {
                    dynamic "cert" {
                      for_each = hosts.value.client_cert_secret_uri != "" ? [1] : []
                      content {
                        gcp_secret_manager_secret_uri = hosts.value.client_cert_secret_uri
                      }
                    }
                    dynamic "key" {
                      for_each = hosts.value.client_key_secret_uri != "" ? [1] : []
                      content {
                        gcp_secret_manager_secret_uri = hosts.value.client_key_secret_uri
                      }
                    }
                  }
                }

                dynamic "header" {
                  for_each = hosts.value.headers
                  content {
                    key   = header.key
                    value = header.value
                  }
                }
              }
            }
          }
        }

        dynamic "writable_cgroups" {
          for_each = containerd_config.value.writable_cgroups_enabled != null ? [1] : []
          content {
            enabled = containerd_config.value.writable_cgroups_enabled
          }
        }
      }
    }

    dynamic "workload_metadata_config" {
      for_each = try(local.nc.workload_metadata_mode, "") != "" ? [1] : []
      content {
        mode = local.nc.workload_metadata_mode
      }
    }

    dynamic "reservation_affinity" {
      for_each = try(local.nc.reservation_affinity, null) != null ? [local.nc.reservation_affinity] : []
      content {
        consume_reservation_type = reservation_affinity.value.consume_reservation_type
        key                      = reservation_affinity.value.key != "" ? reservation_affinity.value.key : null
        values                   = length(reservation_affinity.value.values) > 0 ? reservation_affinity.value.values : null
      }
    }

    dynamic "secondary_boot_disks" {
      for_each = try(local.nc.secondary_boot_disks, [])
      content {
        disk_image = secondary_boot_disks.value.disk_image
        mode       = secondary_boot_disks.value.mode != "" ? secondary_boot_disks.value.mode : null
      }
    }

    dynamic "kubelet_config" {
      for_each = try(local.nc.kubelet_config, null) != null ? [local.nc.kubelet_config] : []
      content {
        cpu_manager_policy                     = kubelet_config.value.cpu_manager_policy != "" ? kubelet_config.value.cpu_manager_policy : null
        cpu_cfs_quota                          = kubelet_config.value.cpu_cfs_quota
        cpu_cfs_quota_period                   = kubelet_config.value.cpu_cfs_quota_period != "" ? kubelet_config.value.cpu_cfs_quota_period : null
        pod_pids_limit                         = kubelet_config.value.pod_pids_limit
        insecure_kubelet_readonly_port_enabled = kubelet_config.value.insecure_kubelet_readonly_port_enabled != "" ? kubelet_config.value.insecure_kubelet_readonly_port_enabled : null
        max_parallel_image_pulls               = kubelet_config.value.max_parallel_image_pulls
        container_log_max_size                 = kubelet_config.value.container_log_max_size != "" ? kubelet_config.value.container_log_max_size : null
        container_log_max_files                = kubelet_config.value.container_log_max_files
        image_gc_low_threshold_percent         = kubelet_config.value.image_gc_low_threshold_percent
        image_gc_high_threshold_percent        = kubelet_config.value.image_gc_high_threshold_percent
        image_minimum_gc_age                   = kubelet_config.value.image_minimum_gc_age != "" ? kubelet_config.value.image_minimum_gc_age : null
        image_maximum_gc_age                   = kubelet_config.value.image_maximum_gc_age != "" ? kubelet_config.value.image_maximum_gc_age : null
        allowed_unsafe_sysctls                 = length(kubelet_config.value.allowed_unsafe_sysctls) > 0 ? kubelet_config.value.allowed_unsafe_sysctls : null
        eviction_max_pod_grace_period_seconds  = kubelet_config.value.eviction_max_pod_grace_period_seconds
        single_process_oom_kill                = kubelet_config.value.single_process_oom_kill

        # Graceful node shutdown (Spot/preemptible pools). API-computed, so
        # sent only when the spec sets them (null otherwise).
        shutdown_grace_period_seconds               = kubelet_config.value.shutdown_grace_period_seconds
        shutdown_grace_period_critical_pods_seconds = kubelet_config.value.shutdown_grace_period_critical_pods_seconds

        dynamic "eviction_soft" {
          for_each = kubelet_config.value.eviction_soft != null ? [kubelet_config.value.eviction_soft] : []
          content {
            memory_available    = eviction_soft.value.memory_available != "" ? eviction_soft.value.memory_available : null
            nodefs_available    = eviction_soft.value.nodefs_available != "" ? eviction_soft.value.nodefs_available : null
            nodefs_inodes_free  = eviction_soft.value.nodefs_inodes_free != "" ? eviction_soft.value.nodefs_inodes_free : null
            imagefs_available   = eviction_soft.value.imagefs_available != "" ? eviction_soft.value.imagefs_available : null
            imagefs_inodes_free = eviction_soft.value.imagefs_inodes_free != "" ? eviction_soft.value.imagefs_inodes_free : null
            pid_available       = eviction_soft.value.pid_available != "" ? eviction_soft.value.pid_available : null
          }
        }

        dynamic "eviction_soft_grace_period" {
          for_each = kubelet_config.value.eviction_soft_grace_period != null ? [kubelet_config.value.eviction_soft_grace_period] : []
          content {
            memory_available    = eviction_soft_grace_period.value.memory_available != "" ? eviction_soft_grace_period.value.memory_available : null
            nodefs_available    = eviction_soft_grace_period.value.nodefs_available != "" ? eviction_soft_grace_period.value.nodefs_available : null
            nodefs_inodes_free  = eviction_soft_grace_period.value.nodefs_inodes_free != "" ? eviction_soft_grace_period.value.nodefs_inodes_free : null
            imagefs_available   = eviction_soft_grace_period.value.imagefs_available != "" ? eviction_soft_grace_period.value.imagefs_available : null
            imagefs_inodes_free = eviction_soft_grace_period.value.imagefs_inodes_free != "" ? eviction_soft_grace_period.value.imagefs_inodes_free : null
            pid_available       = eviction_soft_grace_period.value.pid_available != "" ? eviction_soft_grace_period.value.pid_available : null
          }
        }

        dynamic "eviction_minimum_reclaim" {
          for_each = kubelet_config.value.eviction_minimum_reclaim != null ? [kubelet_config.value.eviction_minimum_reclaim] : []
          content {
            memory_available    = eviction_minimum_reclaim.value.memory_available != "" ? eviction_minimum_reclaim.value.memory_available : null
            nodefs_available    = eviction_minimum_reclaim.value.nodefs_available != "" ? eviction_minimum_reclaim.value.nodefs_available : null
            nodefs_inodes_free  = eviction_minimum_reclaim.value.nodefs_inodes_free != "" ? eviction_minimum_reclaim.value.nodefs_inodes_free : null
            imagefs_available   = eviction_minimum_reclaim.value.imagefs_available != "" ? eviction_minimum_reclaim.value.imagefs_available : null
            imagefs_inodes_free = eviction_minimum_reclaim.value.imagefs_inodes_free != "" ? eviction_minimum_reclaim.value.imagefs_inodes_free : null
            pid_available       = eviction_minimum_reclaim.value.pid_available != "" ? eviction_minimum_reclaim.value.pid_available : null
          }
        }

        dynamic "crash_loop_back_off" {
          for_each = kubelet_config.value.crash_loop_back_off != null ? [kubelet_config.value.crash_loop_back_off] : []
          content {
            max_container_restart_period = crash_loop_back_off.value.max_container_restart_period != "" ? crash_loop_back_off.value.max_container_restart_period : null
          }
        }

        dynamic "memory_manager" {
          for_each = kubelet_config.value.memory_manager != null ? [kubelet_config.value.memory_manager] : []
          content {
            policy = memory_manager.value.policy != "" ? memory_manager.value.policy : null
          }
        }

        dynamic "topology_manager" {
          for_each = kubelet_config.value.topology_manager != null ? [kubelet_config.value.topology_manager] : []
          content {
            policy = topology_manager.value.policy != "" ? topology_manager.value.policy : null
            scope  = topology_manager.value.scope != "" ? topology_manager.value.scope : null
          }
        }
      }
    }

    dynamic "linux_node_config" {
      for_each = try(local.nc.linux_node_config, null) != null ? [local.nc.linux_node_config] : []
      content {
        sysctls                      = length(linux_node_config.value.sysctls) > 0 ? linux_node_config.value.sysctls : null
        cgroup_mode                  = linux_node_config.value.cgroup_mode != "" ? linux_node_config.value.cgroup_mode : null
        transparent_hugepage_enabled = linux_node_config.value.transparent_hugepage_enabled != "" ? linux_node_config.value.transparent_hugepage_enabled : null
        transparent_hugepage_defrag  = linux_node_config.value.transparent_hugepage_defrag != "" ? linux_node_config.value.transparent_hugepage_defrag : null

        dynamic "hugepages_config" {
          for_each = linux_node_config.value.hugepages_config != null ? [linux_node_config.value.hugepages_config] : []
          content {
            hugepage_size_2m = hugepages_config.value.hugepage_size_2m
            hugepage_size_1g = hugepages_config.value.hugepage_size_1g
          }
        }

        dynamic "node_kernel_module_loading" {
          for_each = linux_node_config.value.node_kernel_module_loading_policy != "" ? [1] : []
          content {
            policy = linux_node_config.value.node_kernel_module_loading_policy
          }
        }

        dynamic "accurate_time_config" {
          for_each = linux_node_config.value.enable_ptp_kvm_time_sync != null ? [1] : []
          content {
            enable_ptp_kvm_time_sync = linux_node_config.value.enable_ptp_kvm_time_sync
          }
        }

        dynamic "swap_config" {
          for_each = linux_node_config.value.swap_config != null ? [linux_node_config.value.swap_config] : []
          content {
            enabled = swap_config.value.enabled

            dynamic "boot_disk_profile" {
              for_each = swap_config.value.boot_disk_profile != null ? [swap_config.value.boot_disk_profile] : []
              content {
                swap_size_gib     = boot_disk_profile.value.swap_size_gib
                swap_size_percent = boot_disk_profile.value.swap_size_percent
              }
            }

            dynamic "dedicated_local_ssd_profile" {
              for_each = swap_config.value.dedicated_local_ssd_profile != null ? [swap_config.value.dedicated_local_ssd_profile] : []
              content {
                disk_count = dedicated_local_ssd_profile.value.disk_count
              }
            }

            dynamic "ephemeral_local_ssd_profile" {
              for_each = swap_config.value.ephemeral_local_ssd_profile != null ? [swap_config.value.ephemeral_local_ssd_profile] : []
              content {
                swap_size_gib     = ephemeral_local_ssd_profile.value.swap_size_gib
                swap_size_percent = ephemeral_local_ssd_profile.value.swap_size_percent
              }
            }

            dynamic "encryption_config" {
              for_each = swap_config.value.encryption_config != null ? [swap_config.value.encryption_config] : []
              content {
                disabled = encryption_config.value.disabled
              }
            }
          }
        }

        # Boot-time init script from Cloud Storage or Secret Manager
        # (exactly one source, spec-enforced).
        dynamic "custom_node_init" {
          for_each = linux_node_config.value.custom_node_init != null ? [linux_node_config.value.custom_node_init] : []
          content {
            init_script {
              gcs_uri                       = custom_node_init.value.gcs_uri != "" ? custom_node_init.value.gcs_uri : null
              gcs_generation                = custom_node_init.value.gcs_generation
              gcp_secret_manager_secret_uri = custom_node_init.value.secret_manager_secret_uri != "" ? custom_node_init.value.secret_manager_secret_uri : null
            }
          }
        }
      }
    }

    logging_variant  = local.logging_variant
    flex_start       = try(local.nc.flex_start, false) ? true : null
    max_run_duration = local.max_run_duration
  }

  # For autoscaled pools the cluster autoscaler owns the live node count;
  # without this, every plan after a scale event would try to reset it.
  lifecycle {
    ignore_changes = [
      node_count
    ]
  }

  depends_on = [google_project_service.container_api]
}
