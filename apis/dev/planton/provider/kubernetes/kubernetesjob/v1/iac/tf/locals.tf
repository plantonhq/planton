##############################################
# locals.tf
#
# Includes logic for:
#  - Deriving resource_id, labels, namespace
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
    "resource_kind" = "kubernetes_job"
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

  # Get namespace from spec
  namespace = var.spec.namespace

  # Namespace name to use in all resources (created or existing)
  namespace_name = var.spec.create_namespace ? kubernetes_namespace.this[0].metadata[0].name : data.kubernetes_namespace.existing[0].metadata[0].name

  # The job name is used for identification
  kube_service_name = var.metadata.name

  # Internal DNS name for the job
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  env_secrets_secret_name = "${var.metadata.name}-env-secrets"
  image_pull_secret_name  = "${var.metadata.name}-image-pull-secret"
}
