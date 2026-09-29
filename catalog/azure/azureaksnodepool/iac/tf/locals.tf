locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureaksnodepool"
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

  # Enum-name-string -> ARM value maps. null lets Azure apply its default;
  # only an explicit choice is ever sent, so an unspecified spec and
  # Azure's default deploy identically on both engines.
  mode = (
    var.spec.mode == "SYSTEM" ? "System" :
    var.spec.mode == "USER" ? "User" : "User"
  )

  os_type = (
    var.spec.os_type == "WINDOWS" ? "Windows" : "Linux"
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

  # Spot economics only render for SPOT pools; ARM rejects them on
  # Regular pools (spec-level validation enforces the pairing too).
  priority = var.spec.priority == "SPOT" ? "Spot" : null

  eviction_policy = (
    var.spec.priority != "SPOT" ? null :
    var.spec.eviction_policy == "EVICTION_DELETE" ? "Delete" :
    var.spec.eviction_policy == "EVICTION_DEALLOCATE" ? "Deallocate" : null
  )

  # Unset spot_max_price on a spot pool means -1: pay up to the on-demand
  # price and never be evicted on price -- the setting nearly everyone
  # wants.
  spot_max_price = (
    var.spec.priority != "SPOT" ? null :
    var.spec.spot_max_price != null && var.spec.spot_max_price != 0 ? var.spec.spot_max_price : -1
  )

  os_disk_type = (
    var.spec.os_disk_type == "MANAGED" ? "Managed" :
    var.spec.os_disk_type == "EPHEMERAL" ? "Ephemeral" : null
  )

  kubelet_disk_type = (
    var.spec.kubelet_disk_type == "OS" ? "OS" :
    var.spec.kubelet_disk_type == "TEMPORARY" ? "Temporary" : null
  )

  gpu_instance_map = {
    "MIG1G" = "MIG1g"
    "MIG2G" = "MIG2g"
    "MIG3G" = "MIG3g"
    "MIG4G" = "MIG4g"
    "MIG7G" = "MIG7g"
  }

  gpu_driver = (
    var.spec.gpu_driver == "INSTALL" ? "Install" :
    var.spec.gpu_driver == "NONE" ? "None" : null
  )

  scale_down_mode = (
    var.spec.scale_down_mode == "DELETE" ? "Delete" :
    var.spec.scale_down_mode == "DEALLOCATE" ? "Deallocate" : null
  )

  workload_runtime = (
    var.spec.workload_runtime == "OCI_CONTAINER" ? "OCIContainer" :
    var.spec.workload_runtime == "KATA_MSHV_VM_ISOLATION" ? "KataMshvVmIsolation" :
    var.spec.workload_runtime == "WASM_WASI" ? "WasmWasi" : null
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
