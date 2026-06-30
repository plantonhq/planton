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
    "resource_kind" = "prometheus_kubernetes"
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

  # Get namespace from spec
  namespace = var.spec.namespace

  # Service name for Prometheus
  kube_service_name = "${var.metadata.name}-prometheus"

  # Fully qualified domain name for the service
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local"

  # Port-forward command for local access
  kube_port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} 9090:9090"

  # Ingress configuration
  ingress_is_enabled = try(var.spec.ingress.is_enabled, false)
  ingress_dns_domain = try(var.spec.ingress.dns_domain, "")
  
  # External hostname for ingress
  external_hostname = local.ingress_is_enabled && local.ingress_dns_domain != "" ? "prometheus.${local.ingress_dns_domain}" : ""
  
  # Internal hostname (if using internal DNS)
  internal_hostname = local.ingress_is_enabled && local.ingress_dns_domain != "" ? "prometheus-internal.${local.ingress_dns_domain}" : ""
}

