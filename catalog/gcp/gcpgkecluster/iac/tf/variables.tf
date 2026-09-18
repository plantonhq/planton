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
  description = "Specification for the GCP GKE cluster"
  type = object({
    # The GCP project that owns the cluster. The CLI's tfvars converter
    # resolves StringValueOrRef fields to their literal string before the
    # module runs, so this arrives as a plain string. If empty, the
    # provider's default project is used (see locals.tf).
    project_id = optional(string, "")

    # Cluster name in GCP. Empty means "use metadata.name". Immutable.
    cluster_name = optional(string, "")

    # Region (regional cluster) or zone (zonal cluster). Immutable.
    location = string

    # Human-readable description. Immutable.
    description = optional(string, "")

    # VPC network self link; arrives as a plain string after ref
    # resolution. Immutable.
    network = string

    # Subnetwork self link; arrives as a plain string. Immutable.
    subnetwork = string

    # Zones nodes run in (narrows a regional cluster / widens a zonal one).
    node_locations = optional(list(string), [])

    # Extra GCE resource labels merged with the platform labels.
    resource_labels = optional(map(string), {})

    # Engine-side destroy guard (middleware default true, matching GCP).
    deletion_protection = optional(bool, true)

    # VPC-native pod/service IP assignment. Null lets GKE create and manage
    # the secondary ranges.
    ip_allocation = optional(object({
      # Named existing ranges (plain strings after ref resolution) XOR CIDR
      # blocks for GKE-created ranges — spec-level CEL enforces exclusivity.
      cluster_secondary_range_name    = optional(string, "")
      services_secondary_range_name   = optional(string, "")
      cluster_ipv4_cidr_block         = optional(string, "")
      services_ipv4_cidr_block        = optional(string, "")
      stack_type                      = optional(string, "IPV4")
      additional_pod_range_names      = optional(list(string), [])
      pod_cidr_overprovision_disabled = optional(bool, false)

      # Extra subnetworks whose secondary ranges pods may draw from.
      additional_ip_ranges = optional(list(object({
        # Subnetwork self link (plain string after ref resolution).
        subnetwork           = string
        pod_ipv4_range_names = optional(list(string), [])
        status               = optional(string, "")
      })), [])

      # GKE-planned pod/service ranges (no manual CIDR planning).
      auto_ipam_enabled = optional(bool, false)

      # Network tier for the cluster's IP allocation.
      network_tier = optional(string, "")
    }), null)

    # LEGACY_DATAPATH or ADVANCED_DATAPATH (Dataplane V2). Immutable.
    datapath_provider = optional(string, "")

    # Cluster-wide default pods-per-node (8-256). Immutable.
    default_max_pods_per_node = optional(number, null)

    enable_intranode_visibility              = optional(bool, false)
    enable_l4_ilb_subsetting                 = optional(bool, false)
    enable_fqdn_network_policy               = optional(bool, false)
    enable_cilium_clusterwide_network_policy = optional(bool, false)
    enable_multi_networking                  = optional(bool, false)

    # PRIVATE_IPV6_GOOGLE_ACCESS_* value or empty for the GCP default.
    private_ipv6_google_access = optional(string, "")

    # IN_TRANSIT_ENCRYPTION_* value or empty for the GCP default.
    in_transit_encryption = optional(string, "")

    # Disables default SNAT for routable-pod designs.
    disable_default_snat = optional(bool, false)

    # Calico NetworkPolicy enforcement (legacy path; Dataplane V2 enforces
    # natively).
    enable_network_policy = optional(bool, false)

    # Cloud DNS / kube-dns selection and scope.
    dns_config = optional(object({
      cluster_dns                   = optional(string, "")
      cluster_dns_scope             = optional(string, "")
      cluster_dns_domain            = optional(string, "")
      additive_vpc_scope_dns_domain = optional(string, "")
    }), null)

    # CHANNEL_DISABLED or CHANNEL_STANDARD; empty leaves Gateway API alone.
    gateway_api_channel = optional(string, "")

    enable_service_external_ips = optional(bool, false)

    # TIER_1 unlocks high-bandwidth egress on supported machine families.
    total_egress_bandwidth_tier = optional(string, "")

    disable_l4_lb_firewall_reconciliation = optional(bool, false)

    # Private nodes / private control-plane endpoint.
    private_cluster = optional(object({
      enable_private_nodes    = optional(bool, false)
      enable_private_endpoint = optional(bool, false)
      # /28 for peering-based control planes; empty for PSC-based clusters.
      master_ipv4_cidr_block = optional(string, "")
      # Subnetwork self link (plain string) for PSC-based endpoint placement.
      private_endpoint_subnetwork = optional(string, "")
      enable_master_global_access = optional(bool, false)
    }), null)

    # Control-plane API CIDR allowlist.
    master_authorized_networks = optional(object({
      cidr_blocks = optional(list(object({
        cidr_block   = string
        display_name = optional(string, "")
      })), [])
      gcp_public_cidrs_access_enabled      = optional(bool, null)
      private_endpoint_enforcement_enabled = optional(bool, null)
    }), null)

    # DNS endpoint / IP endpoints posture.
    control_plane_endpoints = optional(object({
      dns_endpoint_allow_external_traffic = optional(bool, false)
      ip_endpoints_enabled                = optional(bool, true)
      enable_k8s_tokens_via_dns           = optional(bool, null)
      enable_k8s_certs_via_dns            = optional(bool, null)
    }), null)

    # Release channel arrives as the proto enum NAME (string):
    # RAPID / REGULAR / STABLE / NONE / EXTENDED.
    release_channel = optional(string, "REGULAR")

    # Minimum control-plane Kubernetes version; empty lets the channel drive.
    min_master_version = optional(string, "")

    # Maintenance windows and exclusions.
    maintenance_policy = optional(object({
      daily_window = optional(object({
        start_time = string
      }), null)
      recurring_window = optional(object({
        start_time = string
        end_time   = string
        recurrence = string
      }), null)
      # Time-of-day + duration form of the recurring window, optionally
      # delayed until a calendar date.
      recurring_time_window = optional(object({
        window_start_time = object({
          hours   = optional(number, 0)
          minutes = optional(number, 0)
          seconds = optional(number, 0)
        })
        window_duration = string
        recurrence      = string
        delay_until = optional(object({
          year  = number
          month = number
          day   = number
        }), null)
      }), null)
      exclusions = optional(list(object({
        exclusion_name    = string
        start_time        = string
        end_time          = string
        scope             = optional(string, "")
        end_time_behavior = optional(string, "")
      })), [])

      # Minimum spacing between disruptive maintenance events.
      disruption_budget = optional(object({
        minor_version_disruption_interval = optional(string, "")
        patch_version_disruption_interval = optional(string, "")
      }), null)
    }), null)

    # Node auto-provisioning (NAP).
    cluster_autoscaling = optional(object({
      enabled = optional(bool, false)
      resource_limits = optional(list(object({
        resource_type = string
        minimum       = optional(number, 0)
        maximum       = number
      })), [])
      autoscaling_profile         = optional(string, "BALANCED")
      auto_provisioning_locations = optional(list(string), [])
      auto_provisioning_defaults = optional(object({
        # Service account email (plain string after ref resolution).
        service_account  = optional(string, "")
        oauth_scopes     = optional(list(string), [])
        disk_size_gb     = optional(number, null)
        disk_type        = optional(string, "")
        image_type       = optional(string, "")
        min_cpu_platform = optional(string, "")
        # KMS key path (plain string after ref resolution).
        boot_disk_kms_key           = optional(string, "")
        enable_secure_boot          = optional(bool, false)
        enable_integrity_monitoring = optional(bool, true)
        auto_upgrade                = optional(bool, true)
        auto_repair                 = optional(bool, true)

        # Upgrade rollout for NAP-created pools (surge or blue-green).
        upgrade_settings = optional(object({
          max_surge       = optional(number, null)
          max_unavailable = optional(number, null)
          strategy        = optional(string, "")
          blue_green_settings = optional(object({
            standard_rollout_policy = optional(object({
              batch_percentage    = optional(number, null)
              batch_node_count    = optional(number, null)
              batch_soak_duration = optional(string, "")
            }), null)
            node_pool_soak_duration = optional(string, "")
          }), null)
        }), null)
      }), null)

      # NAP through default ComputeClass definitions.
      default_compute_class_enabled = optional(bool, null)
    }), null)

    enable_vertical_pod_autoscaling = optional(bool, false)

    # HPA profile: NONE or PERFORMANCE; empty for the GCP default.
    hpa_profile = optional(string, "")

    # Workload Identity Federation for GKE (middleware default true).
    workload_identity_enabled = optional(bool, true)

    # Shielded GKE nodes; null leaves the GCP default (true) in place and
    # keeps Autopilot clusters (which reject the field) clean.
    enable_shielded_nodes = optional(bool, null)

    # CMEK envelope encryption of Kubernetes secrets.
    database_encryption = optional(object({
      state = string
      # KMS key path (plain string after ref resolution).
      key_name = optional(string, "")
    }), null)

    # DISABLED or PROJECT_SINGLETON_POLICY_ENFORCE; empty leaves it off.
    binary_authorization_evaluation_mode = optional(string, "")

    # Security Posture dashboard modes.
    security_posture = optional(object({
      mode               = optional(string, "")
      vulnerability_mode = optional(string, "")
    }), null)

    # gke-security-groups@DOMAIN Google Group for RBAC.
    authenticator_security_group = optional(string, "")

    enable_legacy_abac        = optional(bool, false)
    enable_mesh_certificates  = optional(bool, false)
    enable_secret_manager_csi = optional(bool, false)

    # Rotation cadence for CSI-mounted Secret Manager secrets.
    secret_manager_rotation = optional(object({
      enabled           = optional(bool, false)
      rotation_interval = optional(string, "")
    }), null)

    # The Secret Manager SYNC add-on (secrets into Kubernetes Secrets).
    secret_sync = optional(object({
      enabled           = optional(bool, false)
      rotation_enabled  = optional(bool, false)
      rotation_interval = optional(string, "")
    }), null)

    # Two-step (rollback-safe) control-plane upgrades: soak duration in
    # seconds format ("21600s".."604800s"); desired_emulated_version
    # ("major.minor") completes the upgrade after the soak.
    rollback_safe_upgrade = optional(object({
      control_plane_soak_duration = optional(string, "")
    }), null)
    desired_emulated_version = optional(string, "")

    # Confidential GKE nodes (hardware memory encryption). Immutable.
    confidential_nodes = optional(object({
      enabled                    = optional(bool, false)
      confidential_instance_type = optional(string, "")
    }), null)

    # ENABLED or LIMITED; empty for the GCP default.
    anonymous_authentication_mode = optional(string, "")

    enable_identity_service = optional(bool, false)

    # Cloud Logging components; null leaves GKE defaults.
    logging = optional(object({
      enabled    = optional(bool)
      components = optional(list(string), [])
    }), null)

    # Cloud Monitoring components + managed Prometheus; null leaves GKE
    # defaults.
    monitoring = optional(object({
      components                        = optional(list(string), [])
      managed_prometheus_enabled        = optional(bool, true)
      auto_monitoring_scope             = optional(string, "")
      advanced_datapath_metrics_enabled = optional(bool, false)
      advanced_datapath_relay_enabled   = optional(bool, false)
    }), null)

    # Cluster lifecycle notifications to Pub/Sub.
    notification_pubsub = optional(object({
      enabled = optional(bool, false)
      # Topic ID (plain string after ref resolution).
      topic       = optional(string, "")
      event_types = optional(list(string), [])
    }), null)

    enable_cost_management = optional(bool, false)

    # Resource-usage metering into BigQuery.
    resource_usage_export = optional(object({
      # Dataset ID (plain string after ref resolution).
      bigquery_dataset_id                  = string
      enable_network_egress_metering       = optional(bool, false)
      enable_resource_consumption_metering = optional(bool, true)
    }), null)

    # Addon toggles; null applies GKE defaults (HTTP LB, HPA, PD CSI on).
    addons = optional(object({
      http_load_balancing_enabled            = optional(bool, true)
      horizontal_pod_autoscaling_enabled     = optional(bool, true)
      gce_persistent_disk_csi_driver_enabled = optional(bool, true)
      gcp_filestore_csi_driver_enabled       = optional(bool, false)
      gcs_fuse_csi_driver_enabled            = optional(bool, false)
      gke_backup_agent_enabled               = optional(bool, false)
      dns_cache_enabled                      = optional(bool, false)
      config_connector_enabled               = optional(bool, false)
      stateful_ha_enabled                    = optional(bool, false)
      ray_operator_enabled                   = optional(bool, false)
      ray_cluster_logging_enabled            = optional(bool, false)
      ray_cluster_monitoring_enabled         = optional(bool, false)
      cloudrun_enabled                       = optional(bool, false)
      cloudrun_load_balancer_type            = optional(string, "")
      parallelstore_csi_driver_enabled       = optional(bool, false)
      lustre_csi_driver_enabled              = optional(bool, false)
      lustre_csi_legacy_port_enabled         = optional(bool, false)
      lustre_csi_disable_multi_nic           = optional(bool, false)
      pod_snapshot_enabled                   = optional(bool, false)
      agent_sandbox_enabled                  = optional(bool, false)
      slice_controller_enabled               = optional(bool, false)
      slurm_operator_enabled                 = optional(bool, false)
      high_scale_checkpointing_enabled       = optional(bool, false)
      node_readiness_controller_enabled      = optional(bool, false)
    }), null)

    # Autopilot mode (GKE manages nodes; no GcpGkeNodePool resources).
    enable_autopilot = optional(bool, false)

    # Autopilot only: permit NET_ADMIN workloads.
    allow_net_admin = optional(bool, false)

    # Fleet registration project and membership type.
    fleet_project         = optional(string, "")
    fleet_membership_type = optional(string, "")

    # Engine-side destroy stance: DELETE (default), PREVENT, or ABANDON.
    deletion_policy = optional(string, "")

    # Read-side performance switches for large clusters.
    ignore_node_count_changes = optional(bool, false)
    skip_node_pool_refresh    = optional(bool, false)

    # ALPHA cluster (30-day lifetime, no SLA). Immutable.
    enable_kubernetes_alpha = optional(bool, false)

    # Kubernetes BETA API groups to enable.
    k8s_beta_apis = optional(list(string), [])

    # Dataplane optimization mode (pass-through). Immutable.
    dataplane_optimization_mode = optional(string, "")

    # Legacy client-certificate issuance; null takes no stance.
    issue_client_certificate = optional(bool, null)

    # VIA_KUBELET or VIA_CONTROL_PLANE. Immutable.
    node_creation_mode = optional(string, "")

    # ACCELERATED patch auto-upgrades; empty keeps channel pacing.
    gke_auto_upgrade_patch_mode = optional(string, "")

    # Legacy catch-all RBAC subject lockdown.
    rbac_binding_config = optional(object({
      enable_insecure_binding_system_authenticated   = optional(bool, null)
      enable_insecure_binding_system_unauthenticated = optional(bool, null)
    }), null)

    # Autopilot conversion/posture policies.
    autopilot_policy = optional(object({
      no_standard_node_pools  = optional(bool, null)
      no_system_impersonation = optional(bool, null)
      no_system_mutation      = optional(bool, null)
      no_unsafe_webhooks      = optional(bool, null)
    }), null)

    # Autopilot privileged-workload allowlist paths (gke:// or gs://).
    autopilot_privileged_admission_paths = optional(list(string), [])

    # Node settings for Autopilot-managed pools.
    node_pool_auto_config = optional(object({
      network_tags                           = optional(list(string), [])
      resource_manager_tags                  = optional(map(string), {})
      cgroup_mode                            = optional(string, "")
      node_kernel_module_loading_policy      = optional(string, "")
      insecure_kubelet_readonly_port_enabled = optional(string, "")
    }), null)

    # Creation-time defaults for node pools on Standard clusters.
    node_pool_defaults = optional(object({
      gcfs_enabled                           = optional(bool, null)
      insecure_kubelet_readonly_port_enabled = optional(string, "")
      logging_variant                        = optional(string, "")
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

    # Customer-managed control-plane CAs and KMS keys.
    user_managed_keys = optional(object({
      cluster_ca     = optional(string, "")
      etcd_api_ca    = optional(string, "")
      etcd_peer_ca   = optional(string, "")
      aggregation_ca = optional(string, "")
      # KMS key paths (plain strings after ref resolution).
      control_plane_disk_encryption_key = optional(string, "")
      gkeops_etcd_backup_encryption_key = optional(string, "")
      service_account_signing_keys      = optional(list(string), [])
      service_account_verification_keys = optional(list(string), [])
    }), null)
  })

  validation {
    condition     = var.spec.maintenance_policy == null || ((var.spec.maintenance_policy.daily_window != null ? 1 : 0) + (var.spec.maintenance_policy.recurring_window != null ? 1 : 0) + (var.spec.maintenance_policy.recurring_time_window != null ? 1 : 0) == 1)
    error_message = "maintenance_policy must set exactly one of daily_window, recurring_window, or recurring_time_window."
  }

  validation {
    condition     = var.spec.desired_emulated_version == "" || can(regex("^[0-9]+\\.[0-9]+$", var.spec.desired_emulated_version))
    error_message = "desired_emulated_version must be in major.minor format, e.g. \"1.33\"."
  }
}
