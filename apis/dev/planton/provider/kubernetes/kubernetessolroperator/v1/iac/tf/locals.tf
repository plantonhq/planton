##############################################
# locals.tf
#
# Includes logic for:
#  - Deriving resource_id and labels
#  - Solr Operator configuration
#  - Computed resource names for namespace sharing
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
    "resource_kind" = "kubernetes_solr_operator"
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

  # Solr Operator configuration
  # Namespace comes from input (var.spec.namespace)
  namespace       = var.spec.namespace
  helm_chart_name = "solr-operator"
  helm_chart_repo = "https://solr.apache.org/charts"

  # Helm chart version (strip 'v' prefix if present)
  # The default version (v0.9.1) is set in variables.tf
  helm_chart_version = trimprefix(var.spec.operator_version, "v")

  # CRD manifest URL (must match the operator version)
  crd_manifest_url = "https://solr.apache.org/operator/downloads/crds/${var.spec.operator_version}/all-with-dependencies.yaml"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # The Helm release name uses metadata.name to ensure uniqueness
  helm_release_name  = var.metadata.name
  crds_resource_name = "${var.metadata.name}-crds"
}
