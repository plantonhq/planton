# Local values for Argo CD deployment

locals {
  # Use metadata.id if provided and non-empty; otherwise, fallback to metadata.name
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels for all resources
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "argocd_kubernetes"
  }

  # Organization label only if org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment label only if env is provided
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
    ) ? {
    "environment" = var.metadata.env
  } : {}

  # Merge all labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Namespace from spec
  namespace = var.spec.namespace

  # Namespace reference - either created or existing
  namespace_name = var.spec.create_namespace ? (
    length(kubernetes_namespace.argocd_namespace) > 0 ? kubernetes_namespace.argocd_namespace[0].metadata[0].name : local.namespace
    ) : (
    length(data.kubernetes_namespace.existing) > 0 ? data.kubernetes_namespace.existing[0].metadata[0].name : local.namespace
  )

  # Argo CD Helm chart configuration
  argocd_chart_repo    = "https://argoproj.github.io/argo-helm"
  argocd_chart_name    = "argo-cd"
  argocd_chart_version = "7.7.12" # Pin to stable version

  # Service name follows Helm chart naming convention
  # The argo-cd chart creates a service named <release-name>-argocd-server
  service_name = "${local.resource_id}-argocd-server"

  # Kubernetes service FQDN for internal cluster access
  kube_service_fqdn = "${local.service_name}.${local.namespace}.svc.cluster.local"

  # Port-forward command for local access
  port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.service_name} 8080:80"

  # Handle optional ingress configuration
  ingress_is_enabled = try(var.spec.ingress.is_enabled, false)
  ingress_dns_domain = try(var.spec.ingress.dns_domain, "")

  # External hostname (public ingress)
  ingress_external_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.namespace}.${local.ingress_dns_domain}" : ""

  # Internal hostname (private ingress)
  ingress_internal_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.namespace}-internal.${local.ingress_dns_domain}" : ""

  # Istio ingress configuration (if using istio)
  istio_ingress_namespace      = "istio-ingress"
  gateway_ingress_class_name   = "istio"
  gateway_external_lb_hostname = "ingress-external.istio-ingress.svc.cluster.local"

  # Certificate configuration for ingress
  ingress_cert_cluster_issuer_name = local.ingress_dns_domain
  # Computed TLS secret name to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  ingress_cert_secret_name = "${var.metadata.name}-tls"

  # Hostnames list for certificate
  ingress_hostnames = compact([
    local.ingress_external_hostname,
    local.ingress_internal_hostname,
  ])
}

