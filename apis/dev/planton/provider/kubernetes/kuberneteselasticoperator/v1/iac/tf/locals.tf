##############################################
# locals.tf
#
# Includes logic for:
#  - Deriving resource_id and labels
#  - ECK operator configuration
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
    "resource_kind" = "kubernetes_elastic_operator"
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

  # ECK operator configuration
  namespace = (
    var.spec.namespace != null && var.spec.namespace != ""
    ? var.spec.namespace
    : "elastic-system" # fallback to default
  )
  helm_chart_name    = "eck-operator"
  helm_chart_repo    = "https://helm.elastic.co"
  helm_chart_version = "2.14.0"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # The Helm release name uses metadata.name to ensure uniqueness
  helm_release_name = var.metadata.name

  # Labels to inherit in ECK-managed resources
  inherited_labels = [
    "resource",
    "organization",
    "environment",
    "resource_kind",
    "resource_id",
  ]
}

