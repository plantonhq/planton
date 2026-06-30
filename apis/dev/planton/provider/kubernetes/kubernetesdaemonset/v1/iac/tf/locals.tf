##############################################
# locals.tf
#
# Includes logic for:
#  - Deriving resource_id, labels, namespace
#  - Computing resource names
#  - ServiceAccount name derivation
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
    "resource_kind" = "kubernetes_daemonset"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Selector labels (subset for pod selection)
  selector_labels = {
    "resource_id"   = local.resource_id
    "resource_kind" = "kubernetes_daemonset"
  }

  # Get namespace from spec
  namespace = var.spec.namespace

  # Computed resource names to avoid conflicts
  env_secret_name        = "${var.metadata.name}-env-secrets"
  image_pull_secret_name = "${var.metadata.name}-image-pull"

  # ServiceAccount name: use provided name, or default to DaemonSet name
  service_account_name = (
    var.spec.service_account_name != null && var.spec.service_account_name != ""
    ? var.spec.service_account_name
    : var.metadata.name
  )
}

