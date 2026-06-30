locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels applied to all resources
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "keycloak_kubernetes"
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

  # Namespace from spec.namespace (StringValueOrRef), with fallback to default pattern
  namespace = try(var.spec.namespace.value, "keycloak-${var.metadata.name}")

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  # Users can prefix metadata.name with component type if needed (e.g., "keycloak-my-auth")
  password_secret_name     = "${var.metadata.name}-password"
  db_password_secret_name  = "${var.metadata.name}-db-password"
  external_lb_service_name = "${var.metadata.name}-external-lb"

  # Service configuration
  # service_name uses just the metadata.name; Helm chart handles its own suffixes
  service_name = var.metadata.name
  service_port = 8080

  # Ingress configuration
  ingress_is_enabled = var.spec.ingress.is_enabled
  ingress_dns_domain = var.spec.ingress.dns_domain

  # Stack outputs
  port_forward_command = "kubectl port-forward -n ${local.namespace} svc/${local.service_name} 8080:8080"
  kube_endpoint        = "${local.service_name}.${local.namespace}.svc.cluster.local:8080"

  # External and internal hostnames (only set if ingress is enabled)
  external_hostname = local.ingress_is_enabled && local.ingress_dns_domain != "" ? "https://${local.ingress_dns_domain}" : ""
  internal_hostname = local.ingress_is_enabled && local.ingress_dns_domain != "" ? "https://${local.ingress_dns_domain}-internal" : ""
}

