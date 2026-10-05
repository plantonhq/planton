locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureakscluster"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over them:
  # user tags deliberately win so an org's governance conventions (cost
  # center, owner) can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # ARM requires exactly one DNS prefix flavor. When the spec sets neither,
  # derive the public prefix from the cluster name -- the behavior nearly
  # everyone wants -- unless this is a private cluster carrying its own
  # private prefix.
  dns_prefix = (
    var.spec.dns_prefix != null && var.spec.dns_prefix != "" ? var.spec.dns_prefix :
    (var.spec.dns_prefix_private_cluster != null && var.spec.dns_prefix_private_cluster != "") ? null :
    var.spec.name
  )
  dns_prefix_private_cluster = (
    var.spec.dns_prefix_private_cluster != null && var.spec.dns_prefix_private_cluster != ""
    ? var.spec.dns_prefix_private_cluster
    : null
  )

  # Enum-name-string -> ARM value maps. null lets Azure apply its default;
  # only an explicit choice is ever sent, so an unspecified spec and
  # Azure's default deploy identically on both engines.
  sku_tier = (
    var.spec.sku_tier == "FREE" ? "Free" :
    var.spec.sku_tier == "STANDARD" ? "Standard" :
    var.spec.sku_tier == "PREMIUM" ? "Premium" : null
  )

  support_plan = (
    var.spec.support_plan == "KUBERNETES_OFFICIAL" ? "KubernetesOfficial" :
    var.spec.support_plan == "AKS_LONG_TERM_SUPPORT" ? "AKSLongTermSupport" : null
  )

  automatic_upgrade_channel = (
    var.spec.automatic_upgrade_channel == "PATCH" ? "patch" :
    var.spec.automatic_upgrade_channel == "STABLE" ? "stable" :
    var.spec.automatic_upgrade_channel == "RAPID" ? "rapid" :
    var.spec.automatic_upgrade_channel == "NODE_IMAGE" ? "node-image" : null
  )

  node_os_upgrade_channel = (
    var.spec.node_os_upgrade_channel == "NODE_OS_NODE_IMAGE" ? "NodeImage" :
    var.spec.node_os_upgrade_channel == "SECURITY_PATCH" ? "SecurityPatch" :
    var.spec.node_os_upgrade_channel == "UNMANAGED" ? "Unmanaged" :
    var.spec.node_os_upgrade_channel == "NODE_OS_NONE" ? "None" : null
  )

  identity_type = (
    var.spec.identity != null && var.spec.identity.type == "USER_ASSIGNED"
    ? "UserAssigned"
    : "SystemAssigned"
  )

  # The modern AKS default the module writes explicitly when the spec's
  # network profile is unset or leaves the plugin unspecified: Azure CNI in
  # overlay mode. Kubenet is deprecated (retires March 2028), and writing
  # the default explicitly keeps both engines and ARM's returned state
  # aligned.
  network_plugin = (
    var.spec.network_profile == null ? "azure" :
    var.spec.network_profile.network_plugin == "KUBENET" ? "kubenet" :
    var.spec.network_profile.network_plugin == "NETWORK_PLUGIN_NONE" ? "none" :
    "azure"
  )

  # Overlay applies when explicitly chosen, or by default for Azure CNI
  # when no pod subnet points at traditional dynamic allocation.
  network_plugin_mode = (
    local.network_plugin != "azure" ? null :
    var.spec.network_profile != null && var.spec.network_profile.network_plugin_mode == "OVERLAY" ? "overlay" :
    var.spec.network_profile != null && var.spec.network_profile.network_plugin_mode != null && var.spec.network_profile.network_plugin_mode != "" ? null :
    (var.spec.default_node_pool.pod_subnet_id != null && var.spec.default_node_pool.pod_subnet_id != "") ? null :
    "overlay"
  )

  network_policy = (
    var.spec.network_profile == null ? null :
    var.spec.network_profile.network_policy == "NETWORK_POLICY_AZURE" ? "azure" :
    var.spec.network_profile.network_policy == "CALICO" ? "calico" :
    var.spec.network_profile.network_policy == "NETWORK_POLICY_CILIUM" ? "cilium" : null
  )

  network_data_plane = (
    var.spec.network_profile == null ? null :
    var.spec.network_profile.network_data_plane == "DATA_PLANE_AZURE" ? "azure" :
    var.spec.network_profile.network_data_plane == "DATA_PLANE_CILIUM" ? "cilium" : null
  )

  outbound_type = (
    var.spec.network_profile == null ? null :
    var.spec.network_profile.outbound_type == "LOAD_BALANCER" ? "loadBalancer" :
    var.spec.network_profile.outbound_type == "MANAGED_NAT_GATEWAY" ? "managedNATGateway" :
    var.spec.network_profile.outbound_type == "USER_ASSIGNED_NAT_GATEWAY" ? "userAssignedNATGateway" :
    var.spec.network_profile.outbound_type == "USER_DEFINED_ROUTING" ? "userDefinedRouting" :
    var.spec.network_profile.outbound_type == "OUTBOUND_NONE" ? "none" : null
  )

  ip_versions = (
    var.spec.network_profile == null ? null :
    length(var.spec.network_profile.ip_versions) == 0 ? null :
    [for v in var.spec.network_profile.ip_versions : v == "IPV6" ? "IPv6" : "IPv4"]
  )

  backend_pool_type = (
    var.spec.network_profile == null || var.spec.network_profile.load_balancer_profile == null ? null :
    var.spec.network_profile.load_balancer_profile.backend_pool_type == "NODE_IP_CONFIGURATION" ? "NodeIPConfiguration" :
    var.spec.network_profile.load_balancer_profile.backend_pool_type == "NODE_IP" ? "NodeIP" : null
  )

  autoscaler_expander = (
    var.spec.auto_scaler_profile == null ? null :
    var.spec.auto_scaler_profile.expander == "LEAST_WASTE" ? "least-waste" :
    var.spec.auto_scaler_profile.expander == "MOST_PODS" ? "most-pods" :
    var.spec.auto_scaler_profile.expander == "PRIORITY" ? "priority" :
    var.spec.auto_scaler_profile.expander == "RANDOM" ? "random" : null
  )

  # Weekday enum name -> ARM's day name, shared by the maintenance blocks.
  week_day_map = {
    "SUNDAY"    = "Sunday"
    "MONDAY"    = "Monday"
    "TUESDAY"   = "Tuesday"
    "WEDNESDAY" = "Wednesday"
    "THURSDAY"  = "Thursday"
    "FRIDAY"    = "Friday"
    "SATURDAY"  = "Saturday"
  }

  frequency_map = {
    "DAILY"            = "Daily"
    "WEEKLY"           = "Weekly"
    "RELATIVE_MONTHLY" = "RelativeMonthly"
    "ABSOLUTE_MONTHLY" = "AbsoluteMonthly"
  }

  week_index_map = {
    "FIRST"  = "First"
    "SECOND" = "Second"
    "THIRD"  = "Third"
    "FOURTH" = "Fourth"
    "LAST"   = "Last"
  }

  nginx_controller_map = {
    "ANNOTATION_CONTROLLED" = "AnnotationControlled"
    "INTERNAL"              = "Internal"
    "EXTERNAL"              = "External"
    "NGINX_NONE"            = "None"
  }

  kms_network_access = (
    var.spec.key_management_service == null ? null :
    var.spec.key_management_service.key_vault_network_access == "KMS_PUBLIC" ? "Public" :
    var.spec.key_management_service.key_vault_network_access == "KMS_PRIVATE" ? "Private" : null
  )

  bootstrap_artifact_source = (
    var.spec.bootstrap_profile == null ? null :
    var.spec.bootstrap_profile.artifact_source == "DIRECT" ? "Direct" :
    var.spec.bootstrap_profile.artifact_source == "CACHE" ? "Cache" : null
  )

  node_provisioning_mode = (
    var.spec.node_provisioning_profile == null ? null :
    var.spec.node_provisioning_profile.mode == "MANUAL" ? "Manual" :
    var.spec.node_provisioning_profile.mode == "AUTO" ? "Auto" : null
  )

  node_provisioning_default_pools = (
    var.spec.node_provisioning_profile == null ? null :
    var.spec.node_provisioning_profile.default_node_pools == "NODE_POOLS_AUTO" ? "Auto" :
    var.spec.node_provisioning_profile.default_node_pools == "NODE_POOLS_NONE" ? "None" : null
  )

  windows_license = (
    var.spec.windows_profile == null ? null :
    var.spec.windows_profile.license == "WINDOWS_SERVER" ? "Windows_Server" : null
  )

  # Default-pool enum maps (shared vocabulary with the standalone
  # AzureAksNodePool module -- the two field shapes deliberately converge).
  pool_os_disk_type = (
    var.spec.default_node_pool.os_disk_type == "MANAGED" ? "Managed" :
    var.spec.default_node_pool.os_disk_type == "EPHEMERAL" ? "Ephemeral" : null
  )

  pool_kubelet_disk_type = (
    var.spec.default_node_pool.kubelet_disk_type == "OS" ? "OS" :
    var.spec.default_node_pool.kubelet_disk_type == "TEMPORARY" ? "Temporary" : null
  )

  os_sku_map = {
    "UBUNTU"        = "Ubuntu"
    "UBUNTU_2204"   = "Ubuntu2204"
    "UBUNTU_2404"   = "Ubuntu2404"
    "AZURE_LINUX"   = "AzureLinux"
    "AZURE_LINUX_3" = "AzureLinux3"
    "WINDOWS_2019"  = "Windows2019"
    "WINDOWS_2022"  = "Windows2022"
  }

  gpu_instance_map = {
    "MIG1G" = "MIG1g"
    "MIG2G" = "MIG2g"
    "MIG3G" = "MIG3g"
    "MIG4G" = "MIG4g"
    "MIG7G" = "MIG7g"
  }

  pool_gpu_driver = (
    var.spec.default_node_pool.gpu_driver == "INSTALL" ? "Install" :
    var.spec.default_node_pool.gpu_driver == "NONE" ? "None" : null
  )

  pool_scale_down_mode = (
    var.spec.default_node_pool.scale_down_mode == "DELETE" ? "Delete" :
    var.spec.default_node_pool.scale_down_mode == "DEALLOCATE" ? "Deallocate" : null
  )

  pool_workload_runtime = (
    var.spec.default_node_pool.workload_runtime == "OCI_CONTAINER" ? "OCIContainer" :
    var.spec.default_node_pool.workload_runtime == "KATA_MSHV_VM_ISOLATION" ? "KataMshvVmIsolation" : null
  )

  cpu_manager_policy_map = {
    "CPU_MANAGER_NONE"   = "none"
    "CPU_MANAGER_STATIC" = "static"
  }

  topology_manager_policy_map = {
    "TOPOLOGY_NONE"    = "none"
    "BEST_EFFORT"      = "best-effort"
    "RESTRICTED"       = "restricted"
    "SINGLE_NUMA_NODE" = "single-numa-node"
  }

  transparent_huge_page_map = {
    "THP_ALWAYS"  = "always"
    "THP_MADVISE" = "madvise"
    "THP_NEVER"   = "never"
  }

  transparent_huge_page_defrag_map = {
    "DEFRAG_ALWAYS"        = "always"
    "DEFRAG_DEFER"         = "defer"
    "DEFRAG_DEFER_MADVISE" = "defer+madvise"
    "DEFRAG_MADVISE"       = "madvise"
    "DEFRAG_NEVER"         = "never"
  }

  undrainable_node_behavior_map = {
    "CORDON"   = "Cordon"
    "SCHEDULE" = "Schedule"
  }

  host_port_protocol_map = {
    "TCP" = "TCP"
    "UDP" = "UDP"
  }
}
