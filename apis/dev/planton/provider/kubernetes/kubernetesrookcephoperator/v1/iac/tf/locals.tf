##############################################
# locals.tf
#
# Includes logic for:
#  - Deriving resource_id and labels
#  - Rook Ceph Operator configuration
#  - Computed resource names
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
    "resource_kind" = "kubernetes_rook_ceph_operator"
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

  # Rook Ceph Operator configuration
  namespace       = var.spec.namespace
  helm_chart_name = "rook-ceph"
  helm_chart_repo = "https://charts.rook.io/release"

  # Helm chart version (strip 'v' prefix if present)
  helm_chart_version = trimprefix(var.spec.operator_version, "v")

  # Computed resource names
  helm_release_name = var.metadata.name

  # CSI configuration with defaults
  csi_config = var.spec.csi != null ? var.spec.csi : {
    enable_rbd_driver       = true
    enable_cephfs_driver    = true
    disable_csi_driver      = false
    enable_csi_host_network = true
    provisioner_replicas    = 2
    enable_csi_addons       = false
    enable_nfs_driver       = false
  }
}
