locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "mongodb_kubernetes"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null && try(var.metadata.env, "") != ""
    ) ? {
    "environment" = var.metadata.env
  } : {}

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Use resource_id as the namespace name
  namespace = local.resource_id

  # The service name for MongoDB (Helm chart overrides "fullname")
  kube_service_name = var.metadata.name

  # Construct labels for the MongoDB pods
  mongodb_pod_selector_labels = {
    "app.kubernetes.io/component" = "mongodb"
    "app.kubernetes.io/instance"  = local.resource_id
    "app.kubernetes.io/name"      = "mongodb"
  }

  # Ingress configuration
  ingress_is_enabled        = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  password_secret_name     = "${var.metadata.name}-password"
  external_lb_service_name = "${var.metadata.name}-external-lb"
}
