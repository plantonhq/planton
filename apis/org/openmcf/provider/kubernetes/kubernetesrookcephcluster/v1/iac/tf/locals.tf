##############################################
# locals.tf
#
# Computed values for Rook Ceph Cluster deployment
##############################################

locals {
  # Derive a stable resource ID (prefer `metadata.id`, fallback to `metadata.name`)
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "kubernetes_rook_ceph_cluster"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null &&
    try(var.metadata.env, "") != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Rook Ceph Cluster configuration
  namespace          = var.spec.namespace
  operator_namespace = var.spec.operator_namespace
  helm_chart_name    = "rook-ceph-cluster"
  helm_chart_repo    = "https://charts.rook.io/release"

  # Helm chart version (strip 'v' prefix if present)
  helm_chart_version = trimprefix(var.spec.helm_chart_version, "v")

  # Computed resource names
  helm_release_name = var.metadata.name
  ceph_cluster_name = var.metadata.name

  # Ceph image configuration with defaults
  ceph_image = var.spec.ceph_image != null ? var.spec.ceph_image : {
    repository        = "quay.io/ceph/ceph"
    tag               = "v19.2.3"
    allow_unsupported = false
  }

  # Cluster configuration with defaults
  cluster_config = var.spec.cluster != null ? var.spec.cluster : {
    data_dir_host_path = "/var/lib/rook"
    mon = {
      count                   = 3
      allow_multiple_per_node = false
    }
    mgr = {
      count                   = 2
      allow_multiple_per_node = false
    }
    storage = {
      use_all_nodes   = true
      use_all_devices = true
    }
    network   = null
    resources = null
  }

  # Collect storage pool names for outputs
  block_pool_names = var.spec.block_pools != null ? [for bp in var.spec.block_pools : bp.name] : []
  block_storage_class_names = var.spec.block_pools != null ? [
    for bp in var.spec.block_pools : bp.storage_class.name
    if bp.storage_class != null && bp.storage_class.enabled
  ] : []

  filesystem_names = var.spec.filesystems != null ? [for fs in var.spec.filesystems : fs.name] : []
  filesystem_storage_class_names = var.spec.filesystems != null ? [
    for fs in var.spec.filesystems : fs.storage_class.name
    if fs.storage_class != null && fs.storage_class.enabled
  ] : []

  object_store_names = var.spec.object_stores != null ? [for os in var.spec.object_stores : os.name] : []
  object_storage_class_names = var.spec.object_stores != null ? [
    for os in var.spec.object_stores : os.storage_class.name
    if os.storage_class != null && os.storage_class.enabled
  ] : []

  # Build block pools configuration for Helm
  ceph_block_pools = var.spec.block_pools != null ? [
    for bp in var.spec.block_pools : {
      name = bp.name
      spec = {
        failureDomain = bp.failure_domain
        replicated = {
          size = bp.replicated_size
        }
      }
      storageClass = bp.storage_class != null ? {
        enabled              = bp.storage_class.enabled
        name                 = bp.storage_class.name
        isDefault            = bp.storage_class.is_default
        reclaimPolicy        = bp.storage_class.reclaim_policy
        allowVolumeExpansion = bp.storage_class.allow_volume_expansion
        volumeBindingMode    = bp.storage_class.volume_binding_mode
      } : null
    }
  ] : []

  # Build filesystems configuration for Helm
  ceph_filesystems = var.spec.filesystems != null ? [
    for fs in var.spec.filesystems : {
      name = fs.name
      spec = {
        metadataPool = {
          replicated = {
            size = fs.metadata_pool_replicated_size
          }
        }
        dataPools = [{
          failureDomain = fs.failure_domain
          replicated = {
            size = fs.data_pool_replicated_size
          }
          name = "data0"
        }]
        metadataServer = {
          activeCount   = fs.active_mds_count
          activeStandby = fs.active_standby
        }
      }
      storageClass = fs.storage_class != null ? {
        enabled              = fs.storage_class.enabled
        name                 = fs.storage_class.name
        isDefault            = fs.storage_class.is_default
        reclaimPolicy        = fs.storage_class.reclaim_policy
        allowVolumeExpansion = fs.storage_class.allow_volume_expansion
        volumeBindingMode    = fs.storage_class.volume_binding_mode
      } : null
    }
  ] : []

  # Build object stores configuration for Helm
  ceph_object_stores = var.spec.object_stores != null ? [
    for os in var.spec.object_stores : {
      name = os.name
      spec = {
        metadataPool = {
          failureDomain = os.failure_domain
          replicated = {
            size = os.metadata_pool_replicated_size
          }
        }
        dataPool = {
          failureDomain = os.failure_domain
          erasureCoded = {
            dataChunks   = os.data_pool_erasure_data_chunks
            codingChunks = os.data_pool_erasure_coding_chunks
          }
        }
        preservePoolsOnDelete = os.preserve_pools_on_delete
        gateway = {
          port      = os.gateway_port
          instances = os.gateway_instances
        }
      }
      storageClass = os.storage_class != null ? {
        enabled              = os.storage_class.enabled
        name                 = os.storage_class.name
        isDefault            = os.storage_class.is_default
        reclaimPolicy        = os.storage_class.reclaim_policy
        allowVolumeExpansion = os.storage_class.allow_volume_expansion
        volumeBindingMode    = os.storage_class.volume_binding_mode
      } : null
    }
  ] : []

  # Build cephClusterSpec
  ceph_cluster_spec = {
    dataDirHostPath = try(local.cluster_config.data_dir_host_path, "/var/lib/rook")

    dashboard = {
      enabled = var.spec.enable_dashboard
      ssl     = true
    }

    mon = {
      count                = try(local.cluster_config.mon.count, 3)
      allowMultiplePerNode = try(local.cluster_config.mon.allow_multiple_per_node, false)
    }

    mgr = {
      count                = try(local.cluster_config.mgr.count, 2)
      allowMultiplePerNode = try(local.cluster_config.mgr.allow_multiple_per_node, false)
    }

    storage = {
      useAllNodes   = try(local.cluster_config.storage.use_all_nodes, true)
      useAllDevices = try(local.cluster_config.storage.use_all_devices, true)
    }
  }
}
