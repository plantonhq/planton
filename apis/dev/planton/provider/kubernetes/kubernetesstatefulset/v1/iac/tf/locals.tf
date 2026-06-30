##############################################
# locals.tf
#
# Includes logic for:
#  - Deriving resource_id, labels, namespace
#  - Determining ingress hostnames
#  - Computing service names
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
    "resource_kind" = "kubernetes_statefulset"
    "app"           = var.metadata.name
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null && try(var.metadata.env, "") != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Selector labels (subset used for pod selection)
  selector_labels = {
    "app" = var.metadata.name
  }

  # Get namespace from spec
  namespace = var.spec.namespace

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  env_secret_name        = "${var.metadata.name}-env-secrets"
  image_pull_secret_name = "${var.metadata.name}-image-pull"
  headless_service_name  = "${var.metadata.name}-headless"

  # The client service name
  kube_service_name = var.metadata.name

  # Internal DNS name for the service
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local"

  # Handy port-forward command
  kube_port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} 8080:8080"

  # Replicas (default to 1)
  replicas = try(var.spec.availability.replicas, 1)

  # Pod management policy (default to OrderedReady)
  pod_management_policy = try(var.spec.pod_management_policy, "OrderedReady") != "" ? try(var.spec.pod_management_policy, "OrderedReady") : "OrderedReady"

  # Safely handle optional ingress values
  ingress_is_enabled = try(var.spec.ingress.enabled, false)
  ingress_hostname   = try(var.spec.ingress.hostname, "")

  # External hostname (null if not applicable)
  ingress_external_hostname = (
    local.ingress_is_enabled && local.ingress_hostname != ""
  ) ? local.ingress_hostname : null

  # Internal hostname (null if not applicable)
  ingress_internal_hostname = (
    local.ingress_is_enabled && local.ingress_hostname != ""
  ) ? "internal-${local.ingress_hostname}" : null
}
