# Create the AKS managed cluster: the control plane, its identity and
# access model, its network fabric, the mandatory default (system) node
# pool, and the Azure-managed add-ons.
#
# Design and lifecycle notes worth knowing before operating this resource:
# - The cluster carries exactly ONE node pool -- the default pool Azure
#   requires at creation. Every additional pool is a standalone
#   AzureAksNodePool resource referencing this cluster's ID; pools have
#   independent lifecycles and coupling them here would force cluster
#   updates for pool changes.
# - The network fabric (plugin, mode, CIDRs, outbound type) mostly
#   replaces the cluster when changed -- decide the network model up
#   front. The module writes the modern AKS default explicitly (Azure CNI
#   in overlay mode) when the spec leaves it unspecified, because kubenet
#   -- azurerm's implicit fallback -- is deprecated and retires in 2028.
# - The OIDC issuer defaults ON (the spec's default, deliberately above
#   Azure's provisioning default): it is the trust anchor for workload
#   identity federation (AzureFederatedIdentityCredential consumes the
#   oidc_issuer_url output), costs nothing, and disabling it after
#   enabling forces cluster replacement.
# - Identity: the module defaults to a system-assigned managed identity.
#   A user-assigned identity is the right choice when grants must
#   pre-exist (BYO private DNS zone, BYO subnet) -- compose it with
#   AzureUserAssignedIdentity and AzureRoleAssignment.
# - The legacy service-principal auth block is deliberately not modeled:
#   managed identity is Azure's own stated direction, and a client secret
#   in cluster config is exactly the credential class the platform is
#   built to eliminate.
resource "azurerm_kubernetes_cluster" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group

  dns_prefix                 = local.dns_prefix
  dns_prefix_private_cluster = local.dns_prefix_private_cluster

  # Unset provisions the latest AKS-recommended GA version; production
  # clusters pin a version in the spec so upgrades are deliberate.
  kubernetes_version = var.spec.kubernetes_version

  sku_tier     = local.sku_tier != null ? local.sku_tier : "Free"
  support_plan = local.support_plan != null ? local.support_plan : "KubernetesOfficial"

  automatic_upgrade_channel = local.automatic_upgrade_channel
  node_os_upgrade_channel   = local.node_os_upgrade_channel != null ? local.node_os_upgrade_channel : "NodeImage"

  oidc_issuer_enabled       = var.spec.oidc_issuer_enabled
  workload_identity_enabled = var.spec.workload_identity_enabled

  private_cluster_enabled             = var.spec.private_cluster_enabled
  private_dns_zone_id                 = var.spec.private_dns_zone_id
  private_cluster_public_fqdn_enabled = var.spec.private_cluster_public_fqdn_enabled

  role_based_access_control_enabled = var.spec.role_based_access_control_enabled
  local_account_disabled            = var.spec.local_account_disabled

  azure_policy_enabled = var.spec.azure_policy_enabled

  image_cleaner_enabled        = var.spec.image_cleaner_enabled
  image_cleaner_interval_hours = var.spec.image_cleaner_interval_hours
  cost_analysis_enabled        = var.spec.cost_analysis_enabled
  run_command_enabled          = var.spec.run_command_enabled

  disk_encryption_set_id = var.spec.disk_encryption_set_id
  edge_zone              = var.spec.edge_zone
  node_resource_group    = var.spec.node_resource_group

  custom_ca_trust_certificates_base64 = length(var.spec.custom_ca_trust_certificates_base64) > 0 ? var.spec.custom_ca_trust_certificates_base64 : null

  ai_toolchain_operator_enabled = var.spec.ai_toolchain_operator_enabled

  # The default (system) pool: always Linux, always System mode -- which
  # is why it carries no os_type/mode/spot knobs. Its field shape
  # deliberately matches the standalone AzureAksNodePool kind so moving a
  # workload pool out to its own resource is a mechanical copy.
  default_node_pool {
    name    = var.spec.default_node_pool.name
    vm_size = var.spec.default_node_pool.vm_size

    # With autoscaling, ARM owns node_count after creation; without it,
    # node_count is the pool's fixed size (spec validation requires it).
    node_count           = var.spec.default_node_pool.node_count != null && var.spec.default_node_pool.node_count > 0 ? var.spec.default_node_pool.node_count : null
    auto_scaling_enabled = var.spec.default_node_pool.auto_scaling_enabled
    min_count            = var.spec.default_node_pool.auto_scaling_enabled ? var.spec.default_node_pool.min_count : null
    max_count            = var.spec.default_node_pool.auto_scaling_enabled ? var.spec.default_node_pool.max_count : null

    max_pods = var.spec.default_node_pool.max_pods
    zones    = length(var.spec.default_node_pool.zones) > 0 ? var.spec.default_node_pool.zones : null

    # Unset deploys AKS-managed networking (Azure's default); a BYO
    # subnet requires the cluster identity to hold Network Contributor.
    vnet_subnet_id = var.spec.default_node_pool.vnet_subnet_id
    pod_subnet_id  = var.spec.default_node_pool.pod_subnet_id

    os_disk_size_gb      = var.spec.default_node_pool.os_disk_size_gb
    os_disk_type         = local.pool_os_disk_type != null ? local.pool_os_disk_type : "Managed"
    kubelet_disk_type    = local.pool_kubelet_disk_type
    os_sku               = lookup(local.os_sku_map, coalesce(var.spec.default_node_pool.os_sku, "_"), null)
    orchestrator_version = var.spec.default_node_pool.orchestrator_version

    node_labels                  = length(var.spec.default_node_pool.node_labels) > 0 ? var.spec.default_node_pool.node_labels : null
    only_critical_addons_enabled = var.spec.default_node_pool.only_critical_addons_enabled

    fips_enabled            = var.spec.default_node_pool.fips_enabled
    host_encryption_enabled = var.spec.default_node_pool.host_encryption_enabled

    node_public_ip_enabled   = var.spec.default_node_pool.node_public_ip_enabled
    node_public_ip_prefix_id = var.spec.default_node_pool.node_public_ip_prefix_id

    gpu_instance = lookup(local.gpu_instance_map, coalesce(var.spec.default_node_pool.gpu_instance, "_"), null)
    gpu_driver   = local.pool_gpu_driver

    proximity_placement_group_id  = var.spec.default_node_pool.proximity_placement_group_id
    host_group_id                 = var.spec.default_node_pool.host_group_id
    capacity_reservation_group_id = var.spec.default_node_pool.capacity_reservation_group_id

    scale_down_mode  = local.pool_scale_down_mode != null ? local.pool_scale_down_mode : "Delete"
    snapshot_id      = var.spec.default_node_pool.snapshot_id
    workload_runtime = local.pool_workload_runtime

    ultra_ssd_enabled = var.spec.default_node_pool.ultra_ssd_enabled

    # A stand-in pool AKS rotates through otherwise replace-forcing
    # changes (vm_size, os_disk_type...) -- set proactively in production.
    temporary_name_for_rotation = var.spec.default_node_pool.temporary_name_for_rotation

    dynamic "kubelet_config" {
      for_each = var.spec.default_node_pool.kubelet_config != null ? [var.spec.default_node_pool.kubelet_config] : []
      content {
        cpu_manager_policy        = lookup(local.cpu_manager_policy_map, coalesce(kubelet_config.value.cpu_manager_policy, "_"), null)
        cpu_cfs_quota_enabled     = kubelet_config.value.cpu_cfs_quota_enabled
        cpu_cfs_quota_period      = kubelet_config.value.cpu_cfs_quota_period
        image_gc_high_threshold   = kubelet_config.value.image_gc_high_threshold
        image_gc_low_threshold    = kubelet_config.value.image_gc_low_threshold
        topology_manager_policy   = lookup(local.topology_manager_policy_map, coalesce(kubelet_config.value.topology_manager_policy, "_"), null)
        allowed_unsafe_sysctls    = length(kubelet_config.value.allowed_unsafe_sysctls) > 0 ? kubelet_config.value.allowed_unsafe_sysctls : null
        container_log_max_size_mb = kubelet_config.value.container_log_max_size_mb
        container_log_max_files   = kubelet_config.value.container_log_max_files
        pod_max_pid               = kubelet_config.value.pod_max_pid
      }
    }

    dynamic "linux_os_config" {
      for_each = var.spec.default_node_pool.linux_os_config != null ? [var.spec.default_node_pool.linux_os_config] : []
      content {
        transparent_huge_page        = lookup(local.transparent_huge_page_map, coalesce(linux_os_config.value.transparent_huge_page, "_"), null)
        transparent_huge_page_defrag = lookup(local.transparent_huge_page_defrag_map, coalesce(linux_os_config.value.transparent_huge_page_defrag, "_"), null)
        swap_file_size_mb            = linux_os_config.value.swap_file_size_mb

        dynamic "sysctl_config" {
          for_each = linux_os_config.value.sysctl_config != null ? [linux_os_config.value.sysctl_config] : []
          content {
            fs_aio_max_nr                      = sysctl_config.value.fs_aio_max_nr
            fs_file_max                        = sysctl_config.value.fs_file_max
            fs_inotify_max_user_watches        = sysctl_config.value.fs_inotify_max_user_watches
            fs_nr_open                         = sysctl_config.value.fs_nr_open
            kernel_threads_max                 = sysctl_config.value.kernel_threads_max
            net_core_netdev_max_backlog        = sysctl_config.value.net_core_netdev_max_backlog
            net_core_optmem_max                = sysctl_config.value.net_core_optmem_max
            net_core_rmem_default              = sysctl_config.value.net_core_rmem_default
            net_core_rmem_max                  = sysctl_config.value.net_core_rmem_max
            net_core_somaxconn                 = sysctl_config.value.net_core_somaxconn
            net_core_wmem_default              = sysctl_config.value.net_core_wmem_default
            net_core_wmem_max                  = sysctl_config.value.net_core_wmem_max
            net_ipv4_ip_local_port_range_min   = sysctl_config.value.net_ipv4_ip_local_port_range_min
            net_ipv4_ip_local_port_range_max   = sysctl_config.value.net_ipv4_ip_local_port_range_max
            net_ipv4_neigh_default_gc_thresh1  = sysctl_config.value.net_ipv4_neigh_default_gc_thresh1
            net_ipv4_neigh_default_gc_thresh2  = sysctl_config.value.net_ipv4_neigh_default_gc_thresh2
            net_ipv4_neigh_default_gc_thresh3  = sysctl_config.value.net_ipv4_neigh_default_gc_thresh3
            net_ipv4_tcp_fin_timeout           = sysctl_config.value.net_ipv4_tcp_fin_timeout
            net_ipv4_tcp_keepalive_intvl       = sysctl_config.value.net_ipv4_tcp_keepalive_intvl
            net_ipv4_tcp_keepalive_probes      = sysctl_config.value.net_ipv4_tcp_keepalive_probes
            net_ipv4_tcp_keepalive_time        = sysctl_config.value.net_ipv4_tcp_keepalive_time
            net_ipv4_tcp_max_syn_backlog       = sysctl_config.value.net_ipv4_tcp_max_syn_backlog
            net_ipv4_tcp_max_tw_buckets        = sysctl_config.value.net_ipv4_tcp_max_tw_buckets
            net_ipv4_tcp_tw_reuse              = sysctl_config.value.net_ipv4_tcp_tw_reuse
            net_netfilter_nf_conntrack_buckets = sysctl_config.value.net_netfilter_nf_conntrack_buckets
            net_netfilter_nf_conntrack_max     = sysctl_config.value.net_netfilter_nf_conntrack_max
            vm_max_map_count                   = sysctl_config.value.vm_max_map_count
            vm_swappiness                      = sysctl_config.value.vm_swappiness
            vm_vfs_cache_pressure              = sysctl_config.value.vm_vfs_cache_pressure
          }
        }
      }
    }

    dynamic "node_network_profile" {
      for_each = var.spec.default_node_pool.node_network_profile != null ? [var.spec.default_node_pool.node_network_profile] : []
      content {
        dynamic "allowed_host_ports" {
          for_each = node_network_profile.value.allowed_host_ports
          content {
            port_start = allowed_host_ports.value.port_start
            port_end   = allowed_host_ports.value.port_end
            protocol   = lookup(local.host_port_protocol_map, coalesce(allowed_host_ports.value.protocol, "_"), null)
          }
        }
        application_security_group_ids = length(node_network_profile.value.application_security_group_ids) > 0 ? node_network_profile.value.application_security_group_ids : null
        node_public_ip_tags            = length(node_network_profile.value.node_public_ip_tags) > 0 ? node_network_profile.value.node_public_ip_tags : null
      }
    }

    dynamic "upgrade_settings" {
      for_each = var.spec.default_node_pool.upgrade_settings != null ? [var.spec.default_node_pool.upgrade_settings] : []
      content {
        max_surge                     = upgrade_settings.value.max_surge
        drain_timeout_in_minutes      = upgrade_settings.value.drain_timeout_in_minutes
        node_soak_duration_in_minutes = upgrade_settings.value.node_soak_duration_in_minutes
        undrainable_node_behavior     = lookup(local.undrainable_node_behavior_map, coalesce(upgrade_settings.value.undrainable_node_behavior, "_"), null)
      }
    }

    tags = length(var.spec.default_node_pool.tags) > 0 ? merge(local.final_tags, var.spec.default_node_pool.tags) : local.final_tags
  }

  # The control plane's managed identity. System-assigned unless the spec
  # brings a user-assigned identity (needed when grants must pre-exist:
  # BYO private DNS zone, BYO subnet).
  identity {
    type         = local.identity_type
    identity_ids = local.identity_type == "UserAssigned" ? var.spec.identity.identity_ids : null
  }

  # The kubelet identity nodes use for image pulls and Azure access. All
  # three fields describe the same user-assigned identity and require a
  # user-assigned cluster identity.
  dynamic "kubelet_identity" {
    for_each = var.spec.kubelet_identity != null ? [var.spec.kubelet_identity] : []
    content {
      client_id                 = kubelet_identity.value.client_id
      object_id                 = kubelet_identity.value.object_id
      user_assigned_identity_id = kubelet_identity.value.user_assigned_identity_id
    }
  }

  dynamic "api_server_access_profile" {
    for_each = var.spec.api_server_access_profile != null ? [var.spec.api_server_access_profile] : []
    content {
      authorized_ip_ranges                = length(api_server_access_profile.value.authorized_ip_ranges) > 0 ? api_server_access_profile.value.authorized_ip_ranges : null
      virtual_network_integration_enabled = api_server_access_profile.value.virtual_network_integration_enabled
      subnet_id                           = api_server_access_profile.value.subnet_id
    }
  }

  # Entra ID (Azure AD) integration -- cluster admission by AAD group
  # membership, optionally with Azure RBAC as the authorization source.
  dynamic "azure_active_directory_role_based_access_control" {
    for_each = var.spec.azure_active_directory_role_based_access_control != null ? [var.spec.azure_active_directory_role_based_access_control] : []
    content {
      tenant_id              = azure_active_directory_role_based_access_control.value.tenant_id
      azure_rbac_enabled     = azure_active_directory_role_based_access_control.value.azure_rbac_enabled
      admin_group_object_ids = length(azure_active_directory_role_based_access_control.value.admin_group_object_ids) > 0 ? azure_active_directory_role_based_access_control.value.admin_group_object_ids : null
    }
  }

  # The network fabric. Written explicitly even when the spec leaves it
  # unset: azurerm's implicit fallback is deprecated kubenet, while the
  # modern AKS default is Azure CNI overlay -- the module makes the good
  # default the actual default.
  network_profile {
    network_plugin      = local.network_plugin
    network_plugin_mode = local.network_plugin_mode
    network_policy      = local.network_policy
    network_data_plane  = local.network_data_plane

    dns_service_ip = var.spec.network_profile != null ? var.spec.network_profile.dns_service_ip : null
    service_cidr   = var.spec.network_profile != null ? var.spec.network_profile.service_cidr : null
    service_cidrs  = var.spec.network_profile != null && length(var.spec.network_profile.service_cidrs) > 0 ? var.spec.network_profile.service_cidrs : null
    pod_cidr       = var.spec.network_profile != null ? var.spec.network_profile.pod_cidr : null
    pod_cidrs      = var.spec.network_profile != null && length(var.spec.network_profile.pod_cidrs) > 0 ? var.spec.network_profile.pod_cidrs : null
    ip_versions    = local.ip_versions

    outbound_type = local.outbound_type != null ? local.outbound_type : "loadBalancer"

    dynamic "load_balancer_profile" {
      for_each = var.spec.network_profile != null && var.spec.network_profile.load_balancer_profile != null ? [var.spec.network_profile.load_balancer_profile] : []
      content {
        outbound_ports_allocated    = load_balancer_profile.value.outbound_ports_allocated
        idle_timeout_in_minutes     = load_balancer_profile.value.idle_timeout_in_minutes
        managed_outbound_ip_count   = load_balancer_profile.value.managed_outbound_ip_count
        managed_outbound_ipv6_count = load_balancer_profile.value.managed_outbound_ipv6_count
        outbound_ip_prefix_ids      = length(load_balancer_profile.value.outbound_ip_prefix_ids) > 0 ? load_balancer_profile.value.outbound_ip_prefix_ids : null
        outbound_ip_address_ids     = length(load_balancer_profile.value.outbound_ip_address_ids) > 0 ? load_balancer_profile.value.outbound_ip_address_ids : null
        backend_pool_type           = local.backend_pool_type
      }
    }

    dynamic "nat_gateway_profile" {
      for_each = var.spec.network_profile != null && var.spec.network_profile.nat_gateway_profile != null ? [var.spec.network_profile.nat_gateway_profile] : []
      content {
        idle_timeout_in_minutes   = nat_gateway_profile.value.idle_timeout_in_minutes
        managed_outbound_ip_count = nat_gateway_profile.value.managed_outbound_ip_count
      }
    }

    dynamic "advanced_networking" {
      for_each = var.spec.network_profile != null && var.spec.network_profile.advanced_networking != null ? [var.spec.network_profile.advanced_networking] : []
      content {
        observability_enabled = advanced_networking.value.observability_enabled
        security_enabled      = advanced_networking.value.security_enabled
      }
    }
  }

  dynamic "auto_scaler_profile" {
    for_each = var.spec.auto_scaler_profile != null ? [var.spec.auto_scaler_profile] : []
    content {
      balance_similar_node_groups                   = auto_scaler_profile.value.balance_similar_node_groups
      daemonset_eviction_for_empty_nodes_enabled    = auto_scaler_profile.value.daemonset_eviction_for_empty_nodes_enabled
      daemonset_eviction_for_occupied_nodes_enabled = auto_scaler_profile.value.daemonset_eviction_for_occupied_nodes_enabled
      expander                                      = local.autoscaler_expander
      ignore_daemonsets_utilization_enabled         = auto_scaler_profile.value.ignore_daemonsets_utilization_enabled
      max_graceful_termination_sec                  = auto_scaler_profile.value.max_graceful_termination_sec
      max_node_provisioning_time                    = auto_scaler_profile.value.max_node_provisioning_time
      max_unready_nodes                             = auto_scaler_profile.value.max_unready_nodes
      max_unready_percentage                        = auto_scaler_profile.value.max_unready_percentage
      new_pod_scale_up_delay                        = auto_scaler_profile.value.new_pod_scale_up_delay
      scan_interval                                 = auto_scaler_profile.value.scan_interval
      scale_down_delay_after_add                    = auto_scaler_profile.value.scale_down_delay_after_add
      scale_down_delay_after_delete                 = auto_scaler_profile.value.scale_down_delay_after_delete
      scale_down_delay_after_failure                = auto_scaler_profile.value.scale_down_delay_after_failure
      scale_down_unneeded                           = auto_scaler_profile.value.scale_down_unneeded
      scale_down_unready                            = auto_scaler_profile.value.scale_down_unready
      scale_down_utilization_threshold              = auto_scaler_profile.value.scale_down_utilization_threshold
      empty_bulk_delete_max                         = auto_scaler_profile.value.empty_bulk_delete_max
      skip_nodes_with_local_storage                 = auto_scaler_profile.value.skip_nodes_with_local_storage
      skip_nodes_with_system_pods                   = auto_scaler_profile.value.skip_nodes_with_system_pods
    }
  }

  # The three maintenance surfaces: the legacy hour-of-week window for
  # routine maintenance, and the two schedule-based windows that govern
  # WHEN the upgrade channels actually apply their work.
  dynamic "maintenance_window" {
    for_each = var.spec.maintenance_window != null ? [var.spec.maintenance_window] : []
    content {
      dynamic "allowed" {
        for_each = maintenance_window.value.allowed
        content {
          day   = local.week_day_map[allowed.value.day]
          hours = allowed.value.hours
        }
      }
      dynamic "not_allowed" {
        for_each = maintenance_window.value.not_allowed
        content {
          start = not_allowed.value.start
          end   = not_allowed.value.end
        }
      }
    }
  }

  dynamic "maintenance_window_auto_upgrade" {
    for_each = var.spec.maintenance_window_auto_upgrade != null ? [var.spec.maintenance_window_auto_upgrade] : []
    content {
      frequency    = local.frequency_map[maintenance_window_auto_upgrade.value.frequency]
      interval     = maintenance_window_auto_upgrade.value.interval
      duration     = maintenance_window_auto_upgrade.value.duration
      day_of_week  = maintenance_window_auto_upgrade.value.day_of_week != null ? local.week_day_map[maintenance_window_auto_upgrade.value.day_of_week] : null
      week_index   = maintenance_window_auto_upgrade.value.week_index != null ? local.week_index_map[maintenance_window_auto_upgrade.value.week_index] : null
      day_of_month = maintenance_window_auto_upgrade.value.day_of_month
      start_date   = maintenance_window_auto_upgrade.value.start_date
      start_time   = maintenance_window_auto_upgrade.value.start_time
      utc_offset   = maintenance_window_auto_upgrade.value.utc_offset

      dynamic "not_allowed" {
        for_each = maintenance_window_auto_upgrade.value.not_allowed
        content {
          start = not_allowed.value.start
          end   = not_allowed.value.end
        }
      }
    }
  }

  dynamic "maintenance_window_node_os" {
    for_each = var.spec.maintenance_window_node_os != null ? [var.spec.maintenance_window_node_os] : []
    content {
      frequency    = local.frequency_map[maintenance_window_node_os.value.frequency]
      interval     = maintenance_window_node_os.value.interval
      duration     = maintenance_window_node_os.value.duration
      day_of_week  = maintenance_window_node_os.value.day_of_week != null ? local.week_day_map[maintenance_window_node_os.value.day_of_week] : null
      week_index   = maintenance_window_node_os.value.week_index != null ? local.week_index_map[maintenance_window_node_os.value.week_index] : null
      day_of_month = maintenance_window_node_os.value.day_of_month
      start_date   = maintenance_window_node_os.value.start_date
      start_time   = maintenance_window_node_os.value.start_time
      utc_offset   = maintenance_window_node_os.value.utc_offset

      dynamic "not_allowed" {
        for_each = maintenance_window_node_os.value.not_allowed
        content {
          start = not_allowed.value.start
          end   = not_allowed.value.end
        }
      }
    }
  }

  # Add-ons. Each block only renders when the spec configures it, so an
  # unset spec and Azure's addon-off default deploy identically.
  dynamic "oms_agent" {
    for_each = var.spec.oms_agent != null ? [var.spec.oms_agent] : []
    content {
      log_analytics_workspace_id      = oms_agent.value.log_analytics_workspace_id
      msi_auth_for_monitoring_enabled = oms_agent.value.msi_auth_for_monitoring_enabled
    }
  }

  dynamic "key_vault_secrets_provider" {
    for_each = var.spec.key_vault_secrets_provider != null && coalesce(var.spec.key_vault_secrets_provider.enabled, true) ? [var.spec.key_vault_secrets_provider] : []
    content {
      secret_rotation_enabled  = key_vault_secrets_provider.value.secret_rotation_enabled
      secret_rotation_interval = key_vault_secrets_provider.value.secret_rotation_interval
    }
  }

  dynamic "microsoft_defender" {
    for_each = var.spec.microsoft_defender != null ? [var.spec.microsoft_defender] : []
    content {
      log_analytics_workspace_id = microsoft_defender.value.log_analytics_workspace_id
    }
  }

  dynamic "monitor_metrics" {
    for_each = var.spec.monitor_metrics != null && coalesce(var.spec.monitor_metrics.enabled, true) ? [var.spec.monitor_metrics] : []
    content {
      annotations_allowed = monitor_metrics.value.annotations_allowed
      labels_allowed      = monitor_metrics.value.labels_allowed
    }
  }

  dynamic "ingress_application_gateway" {
    for_each = var.spec.ingress_application_gateway != null ? [var.spec.ingress_application_gateway] : []
    content {
      gateway_id   = ingress_application_gateway.value.gateway_id
      gateway_name = ingress_application_gateway.value.gateway_name
      subnet_cidr  = ingress_application_gateway.value.subnet_cidr
      subnet_id    = ingress_application_gateway.value.subnet_id
    }
  }

  dynamic "aci_connector_linux" {
    for_each = var.spec.aci_connector_linux != null ? [var.spec.aci_connector_linux] : []
    content {
      subnet_name = aci_connector_linux.value.subnet_name
    }
  }

  dynamic "confidential_computing" {
    for_each = var.spec.confidential_computing != null && coalesce(var.spec.confidential_computing.enabled, true) ? [var.spec.confidential_computing] : []
    content {
      sgx_quote_helper_enabled = confidential_computing.value.sgx_quote_helper_enabled
    }
  }

  dynamic "web_app_routing" {
    for_each = var.spec.web_app_routing != null ? [var.spec.web_app_routing] : []
    content {
      dns_zone_ids             = web_app_routing.value.dns_zone_ids
      default_nginx_controller = lookup(local.nginx_controller_map, coalesce(web_app_routing.value.default_nginx_controller, "_"), null)
    }
  }

  dynamic "service_mesh_profile" {
    for_each = var.spec.service_mesh_profile != null ? [var.spec.service_mesh_profile] : []
    content {
      mode                             = "Istio"
      revisions                        = service_mesh_profile.value.revisions
      internal_ingress_gateway_enabled = service_mesh_profile.value.internal_ingress_gateway_enabled
      external_ingress_gateway_enabled = service_mesh_profile.value.external_ingress_gateway_enabled

      dynamic "certificate_authority" {
        for_each = service_mesh_profile.value.certificate_authority != null ? [service_mesh_profile.value.certificate_authority] : []
        content {
          key_vault_id           = certificate_authority.value.key_vault_id
          root_cert_object_name  = certificate_authority.value.root_cert_object_name
          cert_chain_object_name = certificate_authority.value.cert_chain_object_name
          cert_object_name       = certificate_authority.value.cert_object_name
          key_object_name        = certificate_authority.value.key_object_name
        }
      }
    }
  }

  dynamic "storage_profile" {
    for_each = var.spec.storage_profile != null ? [var.spec.storage_profile] : []
    content {
      blob_driver_enabled         = storage_profile.value.blob_driver_enabled
      disk_driver_enabled         = storage_profile.value.disk_driver_enabled
      file_driver_enabled         = storage_profile.value.file_driver_enabled
      snapshot_controller_enabled = storage_profile.value.snapshot_controller_enabled
    }
  }

  dynamic "workload_autoscaler_profile" {
    for_each = var.spec.workload_autoscaler_profile != null ? [var.spec.workload_autoscaler_profile] : []
    content {
      keda_enabled                    = workload_autoscaler_profile.value.keda_enabled
      vertical_pod_autoscaler_enabled = workload_autoscaler_profile.value.vertical_pod_autoscaler_enabled
    }
  }

  dynamic "key_management_service" {
    for_each = var.spec.key_management_service != null ? [var.spec.key_management_service] : []
    content {
      key_vault_key_id         = key_management_service.value.key_vault_key_id
      key_vault_network_access = local.kms_network_access != null ? local.kms_network_access : "Public"
    }
  }

  dynamic "http_proxy_config" {
    for_each = var.spec.http_proxy_config != null ? [var.spec.http_proxy_config] : []
    content {
      http_proxy  = http_proxy_config.value.http_proxy
      https_proxy = http_proxy_config.value.https_proxy
      no_proxy    = length(http_proxy_config.value.no_proxy) > 0 ? http_proxy_config.value.no_proxy : null
      trusted_ca  = http_proxy_config.value.trusted_ca
    }
  }

  dynamic "linux_profile" {
    for_each = var.spec.linux_profile != null ? [var.spec.linux_profile] : []
    content {
      admin_username = linux_profile.value.admin_username
      ssh_key {
        key_data = linux_profile.value.ssh_public_key
      }
    }
  }

  # Windows credentials -- the prerequisite for any Windows
  # AzureAksNodePool joining this cluster.
  dynamic "windows_profile" {
    for_each = var.spec.windows_profile != null ? [var.spec.windows_profile] : []
    content {
      admin_username = windows_profile.value.admin_username
      admin_password = windows_profile.value.admin_password
      license        = local.windows_license

      dynamic "gmsa" {
        for_each = windows_profile.value.gmsa != null ? [windows_profile.value.gmsa] : []
        content {
          dns_server  = gmsa.value.dns_server
          root_domain = gmsa.value.root_domain
        }
      }
    }
  }

  dynamic "bootstrap_profile" {
    for_each = var.spec.bootstrap_profile != null ? [var.spec.bootstrap_profile] : []
    content {
      artifact_source       = local.bootstrap_artifact_source != null ? local.bootstrap_artifact_source : "Direct"
      container_registry_id = bootstrap_profile.value.container_registry_id
    }
  }

  # The provider requires this block even when the spec leaves node
  # provisioning unset; Manual/Auto are ARM's own defaults, so an absent
  # spec block and this explicit one deploy identically.
  node_provisioning_profile {
    mode               = local.node_provisioning_mode != null ? local.node_provisioning_mode : "Manual"
    default_node_pools = local.node_provisioning_default_pools != null ? local.node_provisioning_default_pools : "Auto"
  }

  dynamic "upgrade_override" {
    for_each = var.spec.upgrade_override != null ? [var.spec.upgrade_override] : []
    content {
      force_upgrade_enabled = upgrade_override.value.force_upgrade_enabled
      effective_until       = upgrade_override.value.effective_until
    }
  }

  tags = local.final_tags
}
