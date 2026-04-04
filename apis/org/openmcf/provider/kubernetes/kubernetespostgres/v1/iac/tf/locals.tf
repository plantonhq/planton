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
    "resource_kind" = "postgres_kubernetes"
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

  # Labels for Postgres pods
  postgres_pod_selector_labels = {
    "planton.ai/resource-kind" = "postgres_kubernetes"
    "planton.ai/resource-id"   = local.resource_id
  }

  # Namespace uses the resource_id
  namespace = local.resource_id

  # Service name
  kube_service_name = "${var.metadata.name}-master"

  # Fully qualified domain name for the service
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local"

  # Handy port-forward command
  kube_port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} 8080:8080"

  # Ingress configuration
  ingress_is_enabled        = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)

  # Namespace reference: use created namespace name if created, otherwise use local.namespace
  namespace_name = var.spec.create_namespace ? kubernetes_namespace_v1.postgres_namespace[0].metadata[0].name : local.namespace

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  external_lb_service_name = "${var.metadata.name}-external-lb"
}
