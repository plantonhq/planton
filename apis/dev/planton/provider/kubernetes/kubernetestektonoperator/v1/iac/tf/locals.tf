##############################################
# locals.tf
#
# Includes logic for:
#  - Deriving resource_id and labels
#  - Tekton operator configuration
#  - Component selection logic
#
# IMPORTANT: Namespace Behavior
# Tekton Operator manages its own namespaces:
# - 'tekton-operator' for the operator itself
# - 'tekton-pipelines' for Tekton components
# These are fixed and cannot be customized.
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
    "resource_kind" = "kubernetes_tekton_operator"
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

  # Tekton Operator uses fixed namespaces (managed by the operator)
  operator_namespace   = "tekton-operator"
  components_namespace = "tekton-pipelines"
  tekton_config_name   = "config"

  # Operator release URL (uses version from spec)
  # https://github.com/tektoncd/operator/releases
  operator_release_url = "https://storage.googleapis.com/tekton-releases/operator/previous/${var.spec.operator_version}/release.yaml"

  # Determine profile based on enabled components
  tekton_profile = (
    var.spec.components.pipelines && var.spec.components.triggers && var.spec.components.dashboard ? "all" :
    var.spec.components.pipelines && var.spec.components.triggers ? "basic" :
    "lite"
  )
}
