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
    "resource_kind" = "ingress_nginx_kubernetes"
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

  # Namespace from spec or default
  namespace = (
    var.spec.namespace != null && var.spec.namespace != ""
    ? var.spec.namespace
    : "kubernetes-ingress-nginx"
  )

  # Helm chart configuration
  helm_chart_name       = "kubernetes-ingress-nginx"
  helm_chart_repo       = "https://kubernetes.github.io/ingress-nginx"
  default_chart_version = "4.11.1"

  # Use specified chart version or default
  chart_version = (
    var.spec.chart_version != null && var.spec.chart_version != ""
    ? var.spec.chart_version
    : local.default_chart_version
  )

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  # Users can prefix metadata.name with component type if needed (e.g., "nginx-public")
  release_name = var.metadata.name
  service_name = "${var.metadata.name}-controller"
  service_type = "LoadBalancer"

  # Determine service annotations based on cloud provider and internal flag
  gke_annotations = var.spec.gke != null ? (
    var.spec.internal
    ? { "cloud.google.com/load-balancer-type" = "internal" }
    : { "cloud.google.com/load-balancer-type" = "external" }
  ) : {}

  eks_annotations = var.spec.eks != null ? (
    var.spec.internal
    ? { "service.beta.kubernetes.io/aws-load-balancer-scheme" = "internal" }
    : { "service.beta.kubernetes.io/aws-load-balancer-scheme" = "internet-facing" }
  ) : {}

  aks_annotations = var.spec.aks != null && var.spec.internal ? {
    "service.beta.kubernetes.io/azure-load-balancer-internal" = "true"
  } : {}

  # Merge all cloud-specific annotations
  service_annotations = merge(
    local.gke_annotations,
    local.eks_annotations,
    local.aks_annotations
  )
}

