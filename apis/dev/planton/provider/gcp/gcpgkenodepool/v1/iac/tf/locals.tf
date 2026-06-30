locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.spec.node_pool_name
  )

  # Base GCP labels (following Planton conventions)
  base_gcp_labels = {
    "resource"      = "true"
    "resource_kind" = "gcp-gke-node-pool"
    "resource_name" = var.spec.node_pool_name
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? {
    "environment" = var.metadata.env
  } : {}

  # Merge base, org, environment labels with resource_id
  final_gcp_labels = merge(
    local.base_gcp_labels,
    { "resource_id" = local.resource_id },
    local.org_label,
    local.env_label
  )

  # Merge Planton labels with user-specified node labels
  merged_node_labels = merge(
    local.final_gcp_labels,
    var.spec.node_labels
  )

  # Determine node management settings (invert disable flags to enable flags)
  auto_upgrade_enabled = !var.spec.management.disable_auto_upgrade
  auto_repair_enabled  = !var.spec.management.disable_auto_repair

  # Network tag for GKE cluster (follows "gke-<cluster-name>" convention)
  network_tag = "gke-${var.spec.cluster_name.value}"

  # OAuth scopes for node service account (recommended minimal set)
  oauth_scopes = [
    "https://www.googleapis.com/auth/monitoring",
    "https://www.googleapis.com/auth/logging.write",
    "https://www.googleapis.com/auth/devstorage.read_only"
  ]
}

