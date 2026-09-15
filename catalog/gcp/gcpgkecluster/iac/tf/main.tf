# Enable the Kubernetes Engine API so a fresh project can host the cluster.
# disable_on_destroy is false: tearing down one cluster must never disable
# the API for everything else in the project.
resource "google_project_service" "container_api" {
  project = local.project_id
  service = "container.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A GKE cluster — the managed Kubernetes control plane plus cluster-wide
# configuration. Node pools are separate GcpGkeNodePool resources: the
# default node pool is always removed at create time on Standard clusters,
# so every pool is an explicitly managed, first-class node. Autopilot
# clusters manage nodes themselves and take no node pools.
#
# Lifecycle notes the API enforces:
#   - name, location, description, network, subnetwork, the whole
#     ip_allocation_policy, datapath_provider, default_max_pods_per_node,
#     confidential_nodes, enable_autopilot, and the private control-plane
#     placement fields are immutable — changing any of them replaces the
#     cluster (and everything running on it).
#   - enable_l4_ilb_subsetting is one-way: it can be turned on in place but
#     never off.
#   - deletion_protection is an engine-side guard: while true, a destroy
#     plan fails before touching the cluster.
resource "google_container_cluster" "this" {
  name        = local.cluster_name
  project     = local.project_id
  location    = var.spec.location
  description = local.description

  network    = var.spec.network
  subnetwork = var.spec.subnetwork

  # Nodes may span fewer (regional) or more (zonal) zones than the control
  # plane; empty defers to GKE's location defaults.
  node_locations = length(var.spec.node_locations) > 0 ? var.spec.node_locations : null

  deletion_protection = var.spec.deletion_protection

  # Engine-side destroy stance, layered UNDER deletion_protection: the
  # GKE-native guard blocks the API call; deletion_policy governs what the
  # engine even attempts (PREVENT fails the plan, ABANDON drops state).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Read-side performance switches for large clusters: skip the per-pool
  # IGM node-count queries and the inline node-pool refresh. The refresh
  # skip is safe in this composition because pools are always separate
  # GcpGkeNodePool resources, never inline blocks here.
  ignore_node_count_changes = var.spec.ignore_node_count_changes ? true : null
  skip_node_pool_refresh    = var.spec.skip_node_pool_refresh ? true : null

  # ALPHA cluster: every alpha feature gate on, no SLA, deleted by GKE
  # after 30 days. Strictly for short-lived evaluation.
  enable_kubernetes_alpha = var.spec.enable_kubernetes_alpha ? true : null

  dynamic "enable_k8s_beta_apis" {
    for_each = length(var.spec.k8s_beta_apis) > 0 ? [1] : []
    content {
      enabled_apis = var.spec.k8s_beta_apis
    }
  }

  dataplane_optimization_mode = var.spec.dataplane_optimization_mode != "" ? var.spec.dataplane_optimization_mode : null

  # Clusters are always VPC-native (alias IP). The ip_allocation_policy
  # block is emitted even when the spec omits ip_allocation: an empty block
  # tells GKE to create and manage the pod/service secondary ranges itself,
  # while named ranges pin allocation to planned subnetwork ranges.
  networking_mode = "VPC_NATIVE"
  ip_allocation_policy {
    cluster_secondary_range_name  = try(var.spec.ip_allocation.cluster_secondary_range_name, "") != "" ? var.spec.ip_allocation.cluster_secondary_range_name : null
    services_secondary_range_name = try(var.spec.ip_allocation.services_secondary_range_name, "") != "" ? var.spec.ip_allocation.services_secondary_range_name : null
    cluster_ipv4_cidr_block       = try(var.spec.ip_allocation.cluster_ipv4_cidr_block, "") != "" ? var.spec.ip_allocation.cluster_ipv4_cidr_block : null
    services_ipv4_cidr_block      = try(var.spec.ip_allocation.services_ipv4_cidr_block, "") != "" ? var.spec.ip_allocation.services_ipv4_cidr_block : null
    stack_type                    = try(var.spec.ip_allocation.stack_type, "IPV4")

    dynamic "additional_pod_ranges_config" {
      for_each = length(try(var.spec.ip_allocation.additional_pod_range_names, [])) > 0 ? [1] : []
      content {
        pod_range_names = var.spec.ip_allocation.additional_pod_range_names
      }
    }

    dynamic "pod_cidr_overprovision_config" {
      for_each = try(var.spec.ip_allocation.pod_cidr_overprovision_disabled, false) ? [1] : []
      content {
        disabled = true
      }
    }

    dynamic "additional_ip_ranges_config" {
      for_each = try(var.spec.ip_allocation.additional_ip_ranges, [])
      content {
        subnetwork           = additional_ip_ranges_config.value.subnetwork
        pod_ipv4_range_names = length(additional_ip_ranges_config.value.pod_ipv4_range_names) > 0 ? additional_ip_ranges_config.value.pod_ipv4_range_names : null
        status               = additional_ip_ranges_config.value.status != "" ? additional_ip_ranges_config.value.status : null
      }
    }

    dynamic "auto_ipam_config" {
      for_each = try(var.spec.ip_allocation.auto_ipam_enabled, false) ? [1] : []
      content {
        enabled = true
      }
    }

    dynamic "network_tier_config" {
      for_each = try(var.spec.ip_allocation.network_tier, "") != "" ? [1] : []
      content {
        network_tier = var.spec.ip_allocation.network_tier
      }
    }
  }

  # Standard clusters: drop the API-mandated default node pool immediately —
  # node pools are composed as GcpGkeNodePool resources. Autopilot rejects
  # both fields (GKE owns node management there).
  remove_default_node_pool = local.is_autopilot ? null : true
  initial_node_count       = local.is_autopilot ? null : 1

  enable_autopilot = local.is_autopilot ? true : null
  allow_net_admin  = var.spec.allow_net_admin ? true : null

  datapath_provider         = local.datapath_provider
  default_max_pods_per_node = var.spec.default_max_pods_per_node
  # Sent only when true: the provider rejects the argument as CONFIGURED
  # (even false) on Autopilot clusters, and null and false mean the same
  # thing to GKE.
  enable_intranode_visibility              = var.spec.enable_intranode_visibility ? true : null
  enable_l4_ilb_subsetting                 = var.spec.enable_l4_ilb_subsetting
  enable_fqdn_network_policy               = var.spec.enable_fqdn_network_policy
  enable_cilium_clusterwide_network_policy = var.spec.enable_cilium_clusterwide_network_policy
  enable_multi_networking                  = var.spec.enable_multi_networking
  private_ipv6_google_access               = local.private_ipv6_google_access
  in_transit_encryption_config             = local.in_transit_encryption
  disable_l4_lb_firewall_reconciliation    = var.spec.disable_l4_lb_firewall_reconciliation

  # Calico NetworkPolicy enforcement is two coupled settings: the
  # enforcement block here and the addon below. Both follow the single
  # spec toggle so they can never drift apart. Omitted entirely on
  # Autopilot (which enforces NetworkPolicy natively via Dataplane V2).
  dynamic "network_policy" {
    for_each = var.spec.enable_network_policy ? [1] : []
    content {
      enabled  = true
      provider = "CALICO"
    }
  }

  dynamic "default_snat_status" {
    for_each = var.spec.disable_default_snat ? [1] : []
    content {
      disabled = true
    }
  }

  dynamic "network_performance_config" {
    for_each = var.spec.total_egress_bandwidth_tier != "" ? [1] : []
    content {
      total_egress_bandwidth_tier = var.spec.total_egress_bandwidth_tier
    }
  }

  dynamic "dns_config" {
    for_each = var.spec.dns_config != null ? [var.spec.dns_config] : []
    content {
      cluster_dns                   = dns_config.value.cluster_dns != "" ? dns_config.value.cluster_dns : null
      cluster_dns_scope             = dns_config.value.cluster_dns_scope != "" ? dns_config.value.cluster_dns_scope : null
      cluster_dns_domain            = dns_config.value.cluster_dns_domain != "" ? dns_config.value.cluster_dns_domain : null
      additive_vpc_scope_dns_domain = dns_config.value.additive_vpc_scope_dns_domain != "" ? dns_config.value.additive_vpc_scope_dns_domain : null
    }
  }

  dynamic "gateway_api_config" {
    for_each = var.spec.gateway_api_channel != "" ? [1] : []
    content {
      channel = var.spec.gateway_api_channel
    }
  }

  dynamic "service_external_ips_config" {
    for_each = var.spec.enable_service_external_ips ? [1] : []
    content {
      enabled = true
    }
  }

  # Private topology: private nodes need Cloud NAT (a GcpRouterNat on the
  # network) for outbound internet; a private-only endpoint removes public
  # kubectl access entirely.
  dynamic "private_cluster_config" {
    for_each = var.spec.private_cluster != null ? [var.spec.private_cluster] : []
    content {
      enable_private_nodes        = private_cluster_config.value.enable_private_nodes
      enable_private_endpoint     = private_cluster_config.value.enable_private_endpoint
      master_ipv4_cidr_block      = private_cluster_config.value.master_ipv4_cidr_block != "" ? private_cluster_config.value.master_ipv4_cidr_block : null
      private_endpoint_subnetwork = private_cluster_config.value.private_endpoint_subnetwork != "" ? private_cluster_config.value.private_endpoint_subnetwork : null

      dynamic "master_global_access_config" {
        for_each = private_cluster_config.value.enable_master_global_access ? [1] : []
        content {
          enabled = true
        }
      }
    }
  }

  dynamic "master_authorized_networks_config" {
    for_each = var.spec.master_authorized_networks != null ? [var.spec.master_authorized_networks] : []
    content {
      gcp_public_cidrs_access_enabled      = master_authorized_networks_config.value.gcp_public_cidrs_access_enabled
      private_endpoint_enforcement_enabled = master_authorized_networks_config.value.private_endpoint_enforcement_enabled

      dynamic "cidr_blocks" {
        for_each = master_authorized_networks_config.value.cidr_blocks
        content {
          cidr_block   = cidr_blocks.value.cidr_block
          display_name = cidr_blocks.value.display_name != "" ? cidr_blocks.value.display_name : null
        }
      }
    }
  }

  dynamic "control_plane_endpoints_config" {
    for_each = var.spec.control_plane_endpoints != null ? [var.spec.control_plane_endpoints] : []
    content {
      dns_endpoint_config {
        allow_external_traffic    = control_plane_endpoints_config.value.dns_endpoint_allow_external_traffic
        enable_k8s_tokens_via_dns = control_plane_endpoints_config.value.enable_k8s_tokens_via_dns
        enable_k8s_certs_via_dns  = control_plane_endpoints_config.value.enable_k8s_certs_via_dns
      }
      ip_endpoints_config {
        enabled = control_plane_endpoints_config.value.ip_endpoints_enabled
      }
    }
  }

  # Legacy client-certificate issuance: certificate auth bypasses IAM and
  # cannot be revoked short of rotating the cluster CA — emitted only when
  # a manifest explicitly takes a stance.
  dynamic "master_auth" {
    for_each = var.spec.issue_client_certificate != null ? [1] : []
    content {
      client_certificate_config {
        issue_client_certificate = var.spec.issue_client_certificate
      }
    }
  }

  dynamic "node_creation_config" {
    for_each = var.spec.node_creation_mode != "" ? [1] : []
    content {
      node_creation_mode = var.spec.node_creation_mode
    }
  }

  dynamic "gke_auto_upgrade_config" {
    for_each = var.spec.gke_auto_upgrade_patch_mode != "" ? [1] : []
    content {
      patch_mode = var.spec.gke_auto_upgrade_patch_mode
    }
  }

  dynamic "rbac_binding_config" {
    for_each = var.spec.rbac_binding_config != null ? [var.spec.rbac_binding_config] : []
    content {
      enable_insecure_binding_system_authenticated   = rbac_binding_config.value.enable_insecure_binding_system_authenticated
      enable_insecure_binding_system_unauthenticated = rbac_binding_config.value.enable_insecure_binding_system_unauthenticated
    }
  }

  dynamic "autopilot_cluster_policy_config" {
    for_each = var.spec.autopilot_policy != null ? [var.spec.autopilot_policy] : []
    content {
      no_standard_node_pools  = autopilot_cluster_policy_config.value.no_standard_node_pools
      no_system_impersonation = autopilot_cluster_policy_config.value.no_system_impersonation
      no_system_mutation      = autopilot_cluster_policy_config.value.no_system_mutation
      no_unsafe_webhooks      = autopilot_cluster_policy_config.value.no_unsafe_webhooks
    }
  }

  autopilot_privileged_admission = length(var.spec.autopilot_privileged_admission_paths) > 0 ? var.spec.autopilot_privileged_admission_paths : null

  # Node settings GKE applies to the pools IT manages on Autopilot — the
  # Autopilot counterpart of per-pool node_config.
  dynamic "node_pool_auto_config" {
    for_each = var.spec.node_pool_auto_config != null ? [var.spec.node_pool_auto_config] : []
    content {
      resource_manager_tags = length(node_pool_auto_config.value.resource_manager_tags) > 0 ? node_pool_auto_config.value.resource_manager_tags : null

      dynamic "network_tags" {
        for_each = length(node_pool_auto_config.value.network_tags) > 0 ? [1] : []
        content {
          tags = node_pool_auto_config.value.network_tags
        }
      }

      dynamic "linux_node_config" {
        for_each = (node_pool_auto_config.value.cgroup_mode != "" || node_pool_auto_config.value.node_kernel_module_loading_policy != "") ? [1] : []
        content {
          cgroup_mode = node_pool_auto_config.value.cgroup_mode != "" ? node_pool_auto_config.value.cgroup_mode : null

          dynamic "node_kernel_module_loading" {
            for_each = node_pool_auto_config.value.node_kernel_module_loading_policy != "" ? [1] : []
            content {
              policy = node_pool_auto_config.value.node_kernel_module_loading_policy
            }
          }
        }
      }

      dynamic "node_kubelet_config" {
        for_each = node_pool_auto_config.value.insecure_kubelet_readonly_port_enabled != "" ? [1] : []
        content {
          insecure_kubelet_readonly_port_enabled = node_pool_auto_config.value.insecure_kubelet_readonly_port_enabled
        }
      }
    }
  }

  # Creation-time defaults inherited by every node pool on a Standard
  # cluster; a pool's own node_config overrides these.
  dynamic "node_pool_defaults" {
    for_each = var.spec.node_pool_defaults != null ? [var.spec.node_pool_defaults] : []
    content {
      node_config_defaults {
        insecure_kubelet_readonly_port_enabled = node_pool_defaults.value.insecure_kubelet_readonly_port_enabled != "" ? node_pool_defaults.value.insecure_kubelet_readonly_port_enabled : null
        logging_variant                        = node_pool_defaults.value.logging_variant != "" ? node_pool_defaults.value.logging_variant : null

        dynamic "gcfs_config" {
          for_each = node_pool_defaults.value.gcfs_enabled != null ? [1] : []
          content {
            enabled = node_pool_defaults.value.gcfs_enabled
          }
        }

        dynamic "containerd_config" {
          for_each = node_pool_defaults.value.containerd_config != null ? [node_pool_defaults.value.containerd_config] : []
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
      }
    }
  }

  # Bring-your-own control-plane trust: customer CAs and KMS keys for the
  # control plane's disks and ServiceAccount JWT signing.
  dynamic "user_managed_keys_config" {
    for_each = var.spec.user_managed_keys != null ? [var.spec.user_managed_keys] : []
    content {
      cluster_ca                        = user_managed_keys_config.value.cluster_ca != "" ? user_managed_keys_config.value.cluster_ca : null
      etcd_api_ca                       = user_managed_keys_config.value.etcd_api_ca != "" ? user_managed_keys_config.value.etcd_api_ca : null
      etcd_peer_ca                      = user_managed_keys_config.value.etcd_peer_ca != "" ? user_managed_keys_config.value.etcd_peer_ca : null
      aggregation_ca                    = user_managed_keys_config.value.aggregation_ca != "" ? user_managed_keys_config.value.aggregation_ca : null
      control_plane_disk_encryption_key = user_managed_keys_config.value.control_plane_disk_encryption_key != "" ? user_managed_keys_config.value.control_plane_disk_encryption_key : null
      gkeops_etcd_backup_encryption_key = user_managed_keys_config.value.gkeops_etcd_backup_encryption_key != "" ? user_managed_keys_config.value.gkeops_etcd_backup_encryption_key : null
      service_account_signing_keys      = length(user_managed_keys_config.value.service_account_signing_keys) > 0 ? user_managed_keys_config.value.service_account_signing_keys : null
      service_account_verification_keys = length(user_managed_keys_config.value.service_account_verification_keys) > 0 ? user_managed_keys_config.value.service_account_verification_keys : null
    }
  }

  release_channel {
    channel = local.release_channel
  }

  min_master_version = local.min_master_version

  dynamic "maintenance_policy" {
    for_each = var.spec.maintenance_policy != null ? [var.spec.maintenance_policy] : []
    content {
      dynamic "daily_maintenance_window" {
        for_each = maintenance_policy.value.daily_window != null ? [maintenance_policy.value.daily_window] : []
        content {
          start_time = daily_maintenance_window.value.start_time
        }
      }

      dynamic "recurring_window" {
        for_each = maintenance_policy.value.recurring_window != null ? [maintenance_policy.value.recurring_window] : []
        content {
          start_time = recurring_window.value.start_time
          end_time   = recurring_window.value.end_time
          recurrence = recurring_window.value.recurrence
        }
      }

      dynamic "disruption_budget" {
        for_each = maintenance_policy.value.disruption_budget != null ? [maintenance_policy.value.disruption_budget] : []
        content {
          minor_version_disruption_interval = disruption_budget.value.minor_version_disruption_interval != "" ? disruption_budget.value.minor_version_disruption_interval : null
          patch_version_disruption_interval = disruption_budget.value.patch_version_disruption_interval != "" ? disruption_budget.value.patch_version_disruption_interval : null
        }
      }

      dynamic "maintenance_exclusion" {
        for_each = maintenance_policy.value.exclusions
        content {
          exclusion_name = maintenance_exclusion.value.exclusion_name
          start_time     = maintenance_exclusion.value.start_time
          end_time       = maintenance_exclusion.value.end_time

          dynamic "exclusion_options" {
            for_each = maintenance_exclusion.value.scope != "" ? [1] : []
            content {
              scope             = maintenance_exclusion.value.scope
              end_time_behavior = maintenance_exclusion.value.end_time_behavior != "" ? maintenance_exclusion.value.end_time_behavior : null
            }
          }
        }
      }
    }
  }

  # Node auto-provisioning: GKE creates/deletes node pools within the
  # resource limits — bounded by spec-level validation so an enabled NAP
  # always carries limits (an unbounded NAP is an unbounded bill).
  dynamic "cluster_autoscaling" {
    for_each = var.spec.cluster_autoscaling != null ? [var.spec.cluster_autoscaling] : []
    content {
      enabled                       = cluster_autoscaling.value.enabled
      autoscaling_profile           = cluster_autoscaling.value.autoscaling_profile
      auto_provisioning_locations   = length(cluster_autoscaling.value.auto_provisioning_locations) > 0 ? cluster_autoscaling.value.auto_provisioning_locations : null
      default_compute_class_enabled = cluster_autoscaling.value.default_compute_class_enabled

      dynamic "resource_limits" {
        for_each = cluster_autoscaling.value.resource_limits
        content {
          resource_type = resource_limits.value.resource_type
          minimum       = resource_limits.value.minimum
          maximum       = resource_limits.value.maximum
        }
      }

      dynamic "auto_provisioning_defaults" {
        for_each = cluster_autoscaling.value.auto_provisioning_defaults != null ? [cluster_autoscaling.value.auto_provisioning_defaults] : []
        content {
          service_account   = auto_provisioning_defaults.value.service_account != "" ? auto_provisioning_defaults.value.service_account : null
          oauth_scopes      = length(auto_provisioning_defaults.value.oauth_scopes) > 0 ? auto_provisioning_defaults.value.oauth_scopes : null
          disk_size         = auto_provisioning_defaults.value.disk_size_gb
          disk_type         = auto_provisioning_defaults.value.disk_type != "" ? auto_provisioning_defaults.value.disk_type : null
          image_type        = auto_provisioning_defaults.value.image_type != "" ? auto_provisioning_defaults.value.image_type : null
          min_cpu_platform  = auto_provisioning_defaults.value.min_cpu_platform != "" ? auto_provisioning_defaults.value.min_cpu_platform : null
          boot_disk_kms_key = auto_provisioning_defaults.value.boot_disk_kms_key != "" ? auto_provisioning_defaults.value.boot_disk_kms_key : null

          shielded_instance_config {
            enable_secure_boot          = auto_provisioning_defaults.value.enable_secure_boot
            enable_integrity_monitoring = auto_provisioning_defaults.value.enable_integrity_monitoring
          }

          management {
            auto_upgrade = auto_provisioning_defaults.value.auto_upgrade
            auto_repair  = auto_provisioning_defaults.value.auto_repair
          }

          dynamic "upgrade_settings" {
            for_each = auto_provisioning_defaults.value.upgrade_settings != null ? [auto_provisioning_defaults.value.upgrade_settings] : []
            content {
              max_surge       = upgrade_settings.value.max_surge
              max_unavailable = upgrade_settings.value.max_unavailable
              strategy        = upgrade_settings.value.strategy != "" ? upgrade_settings.value.strategy : null

              dynamic "blue_green_settings" {
                for_each = upgrade_settings.value.blue_green_settings != null ? [upgrade_settings.value.blue_green_settings] : []
                content {
                  node_pool_soak_duration = blue_green_settings.value.node_pool_soak_duration != "" ? blue_green_settings.value.node_pool_soak_duration : null

                  dynamic "standard_rollout_policy" {
                    for_each = blue_green_settings.value.standard_rollout_policy != null ? [blue_green_settings.value.standard_rollout_policy] : []
                    content {
                      batch_percentage    = standard_rollout_policy.value.batch_percentage
                      batch_node_count    = standard_rollout_policy.value.batch_node_count
                      batch_soak_duration = standard_rollout_policy.value.batch_soak_duration != "" ? standard_rollout_policy.value.batch_soak_duration : null
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  dynamic "vertical_pod_autoscaling" {
    for_each = var.spec.enable_vertical_pod_autoscaling ? [1] : []
    content {
      enabled = true
    }
  }

  dynamic "pod_autoscaling" {
    for_each = var.spec.hpa_profile != "" ? [1] : []
    content {
      hpa_profile = var.spec.hpa_profile
    }
  }

  # Workload Identity Federation for GKE: the pool name is fixed by the API
  # to PROJECT_ID.svc.id.goog. Autopilot clusters have it always on — the
  # block is suppressed there to avoid a permanent diff.
  dynamic "workload_identity_config" {
    for_each = var.spec.workload_identity_enabled && !local.is_autopilot ? [1] : []
    content {
      workload_pool = local.workload_pool
    }
  }

  enable_shielded_nodes = var.spec.enable_shielded_nodes

  dynamic "database_encryption" {
    for_each = var.spec.database_encryption != null ? [var.spec.database_encryption] : []
    content {
      state    = database_encryption.value.state
      key_name = database_encryption.value.key_name != "" ? database_encryption.value.key_name : null
    }
  }

  dynamic "binary_authorization" {
    for_each = var.spec.binary_authorization_evaluation_mode != "" ? [1] : []
    content {
      evaluation_mode = var.spec.binary_authorization_evaluation_mode
    }
  }

  dynamic "security_posture_config" {
    for_each = var.spec.security_posture != null ? [var.spec.security_posture] : []
    content {
      mode               = security_posture_config.value.mode != "" ? security_posture_config.value.mode : null
      vulnerability_mode = security_posture_config.value.vulnerability_mode != "" ? security_posture_config.value.vulnerability_mode : null
    }
  }

  dynamic "authenticator_groups_config" {
    for_each = var.spec.authenticator_security_group != "" ? [1] : []
    content {
      security_group = var.spec.authenticator_security_group
    }
  }

  enable_legacy_abac = var.spec.enable_legacy_abac

  dynamic "mesh_certificates" {
    for_each = var.spec.enable_mesh_certificates ? [1] : []
    content {
      enable_certificates = true
    }
  }

  dynamic "secret_manager_config" {
    for_each = var.spec.enable_secret_manager_csi ? [1] : []
    content {
      enabled = true

      dynamic "rotation_config" {
        for_each = var.spec.secret_manager_rotation != null ? [var.spec.secret_manager_rotation] : []
        content {
          enabled           = rotation_config.value.enabled
          rotation_interval = rotation_config.value.rotation_interval != "" ? rotation_config.value.rotation_interval : null
        }
      }
    }
  }

  # The Secret Manager SYNC add-on (secrets into Kubernetes Secret
  # objects) — a separate add-on from the CSI mount path above.
  dynamic "secret_sync_config" {
    for_each = var.spec.secret_sync != null ? [var.spec.secret_sync] : []
    content {
      enabled = secret_sync_config.value.enabled

      dynamic "rotation_config" {
        for_each = secret_sync_config.value.rotation_enabled || secret_sync_config.value.rotation_interval != "" ? [1] : []
        content {
          enabled           = secret_sync_config.value.rotation_enabled
          rotation_interval = secret_sync_config.value.rotation_interval != "" ? secret_sync_config.value.rotation_interval : null
        }
      }
    }
  }

  dynamic "confidential_nodes" {
    for_each = var.spec.confidential_nodes != null ? [var.spec.confidential_nodes] : []
    content {
      enabled                    = confidential_nodes.value.enabled
      confidential_instance_type = confidential_nodes.value.confidential_instance_type != "" ? confidential_nodes.value.confidential_instance_type : null
    }
  }

  dynamic "anonymous_authentication_config" {
    for_each = var.spec.anonymous_authentication_mode != "" ? [1] : []
    content {
      mode = var.spec.anonymous_authentication_mode
    }
  }

  dynamic "identity_service_config" {
    for_each = var.spec.enable_identity_service ? [1] : []
    content {
      enabled = true
    }
  }

  # Logging switched off is written as the empty component list (GKE's
  # spelling of "no Cloud Logging integration"); the API already refuses a
  # switched-on block with no components. Monitoring's presence carries
  # nothing by itself (an empty component list is GKE's default there).
  dynamic "logging_config" {
    for_each = var.spec.logging != null ? [var.spec.logging] : []
    content {
      enable_components = coalesce(logging_config.value.enabled, true) ? logging_config.value.components : []
    }
  }

  dynamic "monitoring_config" {
    for_each = var.spec.monitoring != null ? [var.spec.monitoring] : []
    content {
      enable_components = length(monitoring_config.value.components) > 0 ? monitoring_config.value.components : null

      managed_prometheus {
        enabled = monitoring_config.value.managed_prometheus_enabled

        dynamic "auto_monitoring_config" {
          for_each = monitoring_config.value.auto_monitoring_scope != "" ? [1] : []
          content {
            scope = monitoring_config.value.auto_monitoring_scope
          }
        }
      }

      dynamic "advanced_datapath_observability_config" {
        for_each = (monitoring_config.value.advanced_datapath_metrics_enabled || monitoring_config.value.advanced_datapath_relay_enabled) ? [1] : []
        content {
          enable_metrics = monitoring_config.value.advanced_datapath_metrics_enabled
          enable_relay   = monitoring_config.value.advanced_datapath_relay_enabled
        }
      }
    }
  }

  dynamic "notification_config" {
    for_each = var.spec.notification_pubsub != null ? [var.spec.notification_pubsub] : []
    content {
      pubsub {
        enabled = notification_config.value.enabled
        topic   = notification_config.value.topic != "" ? notification_config.value.topic : null

        dynamic "filter" {
          for_each = length(notification_config.value.event_types) > 0 ? [1] : []
          content {
            event_type = notification_config.value.event_types
          }
        }
      }
    }
  }

  dynamic "cost_management_config" {
    for_each = var.spec.enable_cost_management ? [1] : []
    content {
      enabled = true
    }
  }

  dynamic "resource_usage_export_config" {
    for_each = var.spec.resource_usage_export != null ? [var.spec.resource_usage_export] : []
    content {
      enable_network_egress_metering       = resource_usage_export_config.value.enable_network_egress_metering
      enable_resource_consumption_metering = resource_usage_export_config.value.enable_resource_consumption_metering

      bigquery_destination {
        dataset_id = resource_usage_export_config.value.bigquery_dataset_id
      }
    }
  }

  # Addons: emitted when the spec configures them or when Calico network
  # policy needs its companion addon. The network-policy addon toggle always
  # mirrors enable_network_policy (never set on Autopilot).
  dynamic "addons_config" {
    for_each = (var.spec.addons != null || (var.spec.enable_network_policy && !local.is_autopilot)) ? [1] : []
    content {
      dynamic "network_policy_config" {
        for_each = local.is_autopilot ? [] : [1]
        content {
          disabled = !var.spec.enable_network_policy
        }
      }

      http_load_balancing {
        disabled = !try(var.spec.addons.http_load_balancing_enabled, true)
      }

      horizontal_pod_autoscaling {
        disabled = !try(var.spec.addons.horizontal_pod_autoscaling_enabled, true)
      }

      gce_persistent_disk_csi_driver_config {
        enabled = try(var.spec.addons.gce_persistent_disk_csi_driver_enabled, true)
      }

      dynamic "gcp_filestore_csi_driver_config" {
        for_each = try(var.spec.addons.gcp_filestore_csi_driver_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "gcs_fuse_csi_driver_config" {
        for_each = try(var.spec.addons.gcs_fuse_csi_driver_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "gke_backup_agent_config" {
        for_each = try(var.spec.addons.gke_backup_agent_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "dns_cache_config" {
        for_each = try(var.spec.addons.dns_cache_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "config_connector_config" {
        for_each = try(var.spec.addons.config_connector_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "stateful_ha_config" {
        for_each = try(var.spec.addons.stateful_ha_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "ray_operator_config" {
        for_each = try(var.spec.addons.ray_operator_enabled, false) ? [1] : []
        content {
          enabled = true

          dynamic "ray_cluster_logging_config" {
            for_each = try(var.spec.addons.ray_cluster_logging_enabled, false) ? [1] : []
            content {
              enabled = true
            }
          }

          dynamic "ray_cluster_monitoring_config" {
            for_each = try(var.spec.addons.ray_cluster_monitoring_enabled, false) ? [1] : []
            content {
              enabled = true
            }
          }
        }
      }

      # The Cloud Run addon's argument is inverted (disabled) — the spec
      # keeps the affirmative form every other addon uses.
      dynamic "cloudrun_config" {
        for_each = try(var.spec.addons.cloudrun_enabled, false) ? [1] : []
        content {
          disabled           = false
          load_balancer_type = try(var.spec.addons.cloudrun_load_balancer_type, "") != "" ? var.spec.addons.cloudrun_load_balancer_type : null
        }
      }

      dynamic "parallelstore_csi_driver_config" {
        for_each = try(var.spec.addons.parallelstore_csi_driver_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "lustre_csi_driver_config" {
        for_each = try(var.spec.addons.lustre_csi_driver_enabled, false) ? [1] : []
        content {
          enabled                   = true
          enable_legacy_lustre_port = try(var.spec.addons.lustre_csi_legacy_port_enabled, false) ? true : null
          disable_multi_nic         = try(var.spec.addons.lustre_csi_disable_multi_nic, false) ? true : null
        }
      }

      dynamic "pod_snapshot_config" {
        for_each = try(var.spec.addons.pod_snapshot_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "agent_sandbox_config" {
        for_each = try(var.spec.addons.agent_sandbox_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "slice_controller_config" {
        for_each = try(var.spec.addons.slice_controller_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }

      dynamic "slurm_operator_config" {
        for_each = try(var.spec.addons.slurm_operator_enabled, false) ? [1] : []
        content {
          enabled = true
        }
      }
    }
  }

  dynamic "fleet" {
    for_each = (var.spec.fleet_project != "" || var.spec.fleet_membership_type != "") ? [1] : []
    content {
      project         = var.spec.fleet_project != "" ? var.spec.fleet_project : null
      membership_type = var.spec.fleet_membership_type != "" ? var.spec.fleet_membership_type : null
    }
  }

  resource_labels = local.final_labels

  depends_on = [google_project_service.container_api]
}
