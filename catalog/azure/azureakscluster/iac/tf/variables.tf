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
  description = "Azure AKS managed cluster specification"
  type = object({
    # The resource group and region the cluster lives in. References are
    # resolved to literals by the platform before the module runs.
    resource_group = string
    region         = string

    # The managed cluster's name (its ARM identity -- renaming replaces
    # the cluster).
    name = string

    # DNS prefix for the public API FQDN, or the private-cluster prefix.
    # At most one is set (spec-level validation); both unset derives a
    # prefix from the cluster name.
    dns_prefix                 = optional(string)
    dns_prefix_private_cluster = optional(string)

    # Control-plane Kubernetes version. Unset provisions the latest
    # AKS-recommended GA version.
    kubernetes_version = optional(string)

    # Pricing tier and support plan, as spec enum name strings.
    sku_tier     = optional(string)
    support_plan = optional(string)

    # The mandatory default (system) node pool -- always Linux, System
    # mode. Field shape converges on the standalone AzureAksNodePool kind.
    default_node_pool = object({
      name                          = string
      vm_size                       = string
      node_count                    = optional(number)
      auto_scaling_enabled          = optional(bool, false)
      min_count                     = optional(number)
      max_count                     = optional(number)
      max_pods                      = optional(number)
      zones                         = optional(list(string), [])
      vnet_subnet_id                = optional(string)
      pod_subnet_id                 = optional(string)
      os_disk_size_gb               = optional(number)
      os_disk_type                  = optional(string)
      kubelet_disk_type             = optional(string)
      os_sku                        = optional(string)
      orchestrator_version          = optional(string)
      node_labels                   = optional(map(string), {})
      only_critical_addons_enabled  = optional(bool, false)
      fips_enabled                  = optional(bool, false)
      host_encryption_enabled       = optional(bool, false)
      node_public_ip_enabled        = optional(bool, false)
      node_public_ip_prefix_id      = optional(string)
      gpu_instance                  = optional(string)
      gpu_driver                    = optional(string)
      proximity_placement_group_id  = optional(string)
      host_group_id                 = optional(string)
      capacity_reservation_group_id = optional(string)
      scale_down_mode               = optional(string)
      snapshot_id                   = optional(string)
      workload_runtime              = optional(string)
      ultra_ssd_enabled             = optional(bool, false)
      temporary_name_for_rotation   = optional(string)
      kubelet_config = optional(object({
        cpu_manager_policy        = optional(string)
        cpu_cfs_quota_enabled     = optional(bool, true)
        cpu_cfs_quota_period      = optional(string)
        image_gc_high_threshold   = optional(number)
        image_gc_low_threshold    = optional(number)
        topology_manager_policy   = optional(string)
        allowed_unsafe_sysctls    = optional(list(string), [])
        container_log_max_size_mb = optional(number)
        container_log_max_files   = optional(number)
        pod_max_pid               = optional(number)
      }))
      linux_os_config = optional(object({
        sysctl_config = optional(object({
          fs_aio_max_nr                      = optional(number)
          fs_file_max                        = optional(number)
          fs_inotify_max_user_watches        = optional(number)
          fs_nr_open                         = optional(number)
          kernel_threads_max                 = optional(number)
          net_core_netdev_max_backlog        = optional(number)
          net_core_optmem_max                = optional(number)
          net_core_rmem_default              = optional(number)
          net_core_rmem_max                  = optional(number)
          net_core_somaxconn                 = optional(number)
          net_core_wmem_default              = optional(number)
          net_core_wmem_max                  = optional(number)
          net_ipv4_ip_local_port_range_min   = optional(number)
          net_ipv4_ip_local_port_range_max   = optional(number)
          net_ipv4_neigh_default_gc_thresh1  = optional(number)
          net_ipv4_neigh_default_gc_thresh2  = optional(number)
          net_ipv4_neigh_default_gc_thresh3  = optional(number)
          net_ipv4_tcp_fin_timeout           = optional(number)
          net_ipv4_tcp_keepalive_intvl       = optional(number)
          net_ipv4_tcp_keepalive_probes      = optional(number)
          net_ipv4_tcp_keepalive_time        = optional(number)
          net_ipv4_tcp_max_syn_backlog       = optional(number)
          net_ipv4_tcp_max_tw_buckets        = optional(number)
          net_ipv4_tcp_tw_reuse              = optional(bool, false)
          net_netfilter_nf_conntrack_buckets = optional(number)
          net_netfilter_nf_conntrack_max     = optional(number)
          vm_max_map_count                   = optional(number)
          vm_swappiness                      = optional(number)
          vm_vfs_cache_pressure              = optional(number)
        }))
        transparent_huge_page        = optional(string)
        transparent_huge_page_defrag = optional(string)
        swap_file_size_mb            = optional(number)
      }))
      node_network_profile = optional(object({
        allowed_host_ports = optional(list(object({
          port_start = optional(number)
          port_end   = optional(number)
          protocol   = optional(string)
        })), [])
        application_security_group_ids = optional(list(string), [])
        node_public_ip_tags            = optional(map(string), {})
      }))
      upgrade_settings = optional(object({
        max_surge                     = string
        drain_timeout_in_minutes      = optional(number)
        node_soak_duration_in_minutes = optional(number)
        undrainable_node_behavior     = optional(string)
      }))
      tags = optional(map(string), {})
    })

    # Cluster identity (control plane) and kubelet identity (nodes).
    identity = optional(object({
      type         = optional(string)
      identity_ids = optional(list(string), [])
    }))
    kubelet_identity = optional(object({
      client_id                 = string
      object_id                 = string
      user_assigned_identity_id = string
    }))

    # Workload identity federation seam.
    oidc_issuer_enabled       = optional(bool, true)
    workload_identity_enabled = optional(bool, false)

    # Private cluster posture.
    private_cluster_enabled             = optional(bool, false)
    private_dns_zone_id                 = optional(string)
    private_cluster_public_fqdn_enabled = optional(bool, false)

    # API-server access hardening for public clusters.
    api_server_access_profile = optional(object({
      authorized_ip_ranges                = optional(list(string), [])
      virtual_network_integration_enabled = optional(bool, false)
      subnet_id                           = optional(string)
    }))

    # Authentication and authorization.
    role_based_access_control_enabled = optional(bool, true)
    local_account_disabled            = optional(bool, false)
    azure_active_directory_role_based_access_control = optional(object({
      tenant_id              = optional(string)
      azure_rbac_enabled     = optional(bool, false)
      admin_group_object_ids = optional(list(string), [])
    }))

    # The network fabric. All enum fields carry spec enum name strings.
    network_profile = optional(object({
      network_plugin      = optional(string)
      network_plugin_mode = optional(string)
      network_policy      = optional(string)
      network_data_plane  = optional(string)
      dns_service_ip      = optional(string)
      service_cidr        = optional(string)
      service_cidrs       = optional(list(string), [])
      pod_cidr            = optional(string)
      pod_cidrs           = optional(list(string), [])
      ip_versions         = optional(list(string), [])
      outbound_type       = optional(string)
      load_balancer_profile = optional(object({
        outbound_ports_allocated    = optional(number)
        idle_timeout_in_minutes     = optional(number)
        managed_outbound_ip_count   = optional(number)
        managed_outbound_ipv6_count = optional(number)
        outbound_ip_prefix_ids      = optional(list(string), [])
        outbound_ip_address_ids     = optional(list(string), [])
        backend_pool_type           = optional(string)
      }))
      nat_gateway_profile = optional(object({
        idle_timeout_in_minutes   = optional(number)
        managed_outbound_ip_count = optional(number)
      }))
      advanced_networking = optional(object({
        observability_enabled = optional(bool, false)
        security_enabled      = optional(bool, false)
      }))
    }))

    # Cluster autoscaler tuning (cluster-wide, applies to autoscaled pools).
    auto_scaler_profile = optional(object({
      balance_similar_node_groups                   = optional(bool, false)
      daemonset_eviction_for_empty_nodes_enabled    = optional(bool, false)
      daemonset_eviction_for_occupied_nodes_enabled = optional(bool, true)
      expander                                      = optional(string)
      ignore_daemonsets_utilization_enabled         = optional(bool, false)
      max_graceful_termination_sec                  = optional(number)
      max_node_provisioning_time                    = optional(string)
      max_unready_nodes                             = optional(number)
      max_unready_percentage                        = optional(number)
      new_pod_scale_up_delay                        = optional(string)
      scan_interval                                 = optional(string)
      scale_down_delay_after_add                    = optional(string)
      scale_down_delay_after_delete                 = optional(string)
      scale_down_delay_after_failure                = optional(string)
      scale_down_unneeded                           = optional(string)
      scale_down_unready                            = optional(string)
      scale_down_utilization_threshold              = optional(string)
      empty_bulk_delete_max                         = optional(number)
      skip_nodes_with_local_storage                 = optional(bool, false)
      skip_nodes_with_system_pods                   = optional(bool, true)
    }))

    # Upgrade channels and maintenance windows.
    automatic_upgrade_channel = optional(string)
    node_os_upgrade_channel   = optional(string)
    maintenance_window = optional(object({
      allowed = optional(list(object({
        day   = string
        hours = list(number)
      })), [])
      not_allowed = optional(list(object({
        start = string
        end   = string
      })), [])
    }))
    maintenance_window_auto_upgrade = optional(object({
      frequency    = string
      interval     = number
      duration     = number
      day_of_week  = optional(string)
      week_index   = optional(string)
      day_of_month = optional(number)
      start_date   = optional(string)
      start_time   = optional(string)
      utc_offset   = optional(string)
      not_allowed = optional(list(object({
        start = string
        end   = string
      })), [])
    }))
    maintenance_window_node_os = optional(object({
      frequency    = string
      interval     = number
      duration     = number
      day_of_week  = optional(string)
      week_index   = optional(string)
      day_of_month = optional(number)
      start_date   = optional(string)
      start_time   = optional(string)
      utc_offset   = optional(string)
      not_allowed = optional(list(object({
        start = string
        end   = string
      })), [])
    }))

    # Add-ons.
    oms_agent = optional(object({
      log_analytics_workspace_id      = string
      msi_auth_for_monitoring_enabled = optional(bool, false)
    }))
    key_vault_secrets_provider = optional(object({
      enabled                  = optional(bool)
      secret_rotation_enabled  = optional(bool, false)
      secret_rotation_interval = optional(string)
    }))
    azure_policy_enabled = optional(bool, false)
    microsoft_defender = optional(object({
      log_analytics_workspace_id = string
    }))
    monitor_metrics = optional(object({
      enabled             = optional(bool)
      annotations_allowed = optional(string)
      labels_allowed      = optional(string)
    }))
    ingress_application_gateway = optional(object({
      gateway_id   = optional(string)
      gateway_name = optional(string)
      subnet_cidr  = optional(string)
      subnet_id    = optional(string)
    }))
    aci_connector_linux = optional(object({
      subnet_name = string
    }))
    confidential_computing = optional(object({
      enabled                  = optional(bool)
      sgx_quote_helper_enabled = optional(bool, false)
    }))
    web_app_routing = optional(object({
      dns_zone_ids             = optional(list(string), [])
      default_nginx_controller = optional(string)
    }))

    # Platform profiles.
    service_mesh_profile = optional(object({
      mode                             = string
      revisions                        = list(string)
      internal_ingress_gateway_enabled = optional(bool, false)
      external_ingress_gateway_enabled = optional(bool, false)
      certificate_authority = optional(object({
        key_vault_id           = string
        root_cert_object_name  = string
        cert_chain_object_name = string
        cert_object_name       = string
        key_object_name        = string
      }))
    }))
    storage_profile = optional(object({
      blob_driver_enabled         = optional(bool, false)
      disk_driver_enabled         = optional(bool, true)
      file_driver_enabled         = optional(bool, true)
      snapshot_controller_enabled = optional(bool, true)
    }))
    workload_autoscaler_profile = optional(object({
      keda_enabled                    = optional(bool, false)
      vertical_pod_autoscaler_enabled = optional(bool, false)
    }))
    key_management_service = optional(object({
      key_vault_key_id         = string
      key_vault_network_access = optional(string)
    }))
    http_proxy_config = optional(object({
      http_proxy  = optional(string)
      https_proxy = optional(string)
      no_proxy    = optional(list(string), [])
      trusted_ca  = optional(string)
    }))
    linux_profile = optional(object({
      admin_username = string
      ssh_public_key = string
    }))
    windows_profile = optional(object({
      admin_username = string
      admin_password = string
      license        = optional(string)
      gmsa = optional(object({
        dns_server  = optional(string)
        root_domain = optional(string)
      }))
    }))

    # Operational hardening and platform toggles.
    image_cleaner_enabled               = optional(bool, false)
    image_cleaner_interval_hours        = optional(number)
    cost_analysis_enabled               = optional(bool, false)
    run_command_enabled                 = optional(bool, true)
    disk_encryption_set_id              = optional(string)
    edge_zone                           = optional(string)
    node_resource_group                 = optional(string)
    custom_ca_trust_certificates_base64 = optional(list(string), [])
    bootstrap_profile = optional(object({
      artifact_source       = optional(string)
      container_registry_id = optional(string)
    }))
    node_provisioning_profile = optional(object({
      mode               = optional(string)
      default_node_pools = optional(string)
    }))
    upgrade_override = optional(object({
      force_upgrade_enabled = bool
      effective_until       = optional(string)
    }))
    ai_toolchain_operator_enabled = optional(bool, false)

    # Free-form user tags, merged over the metadata-derived tags.
    tags = optional(map(string), {})
  })
}
