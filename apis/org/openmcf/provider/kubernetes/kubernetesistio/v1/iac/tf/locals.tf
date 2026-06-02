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
    "resource_kind" = "istio_kubernetes"
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

  # Namespace configuration
  system_namespace  = "istio-system"
  gateway_namespace = "istio-ingress"

  # Helm chart configuration
  helm_repo             = "https://istio-release.storage.googleapis.com/charts"
  base_chart_name       = "base"
  istiod_chart_name     = "istiod"
  gateway_chart_name    = "gateway"
  default_chart_version = "1.26.8"

  # Istio version: driven by spec.version, falling back to the module default.
  # Keep default_chart_version in sync with the SDK/CRD pin for this OpenMCF release.
  chart_version = try(var.spec.version, null) != null ? var.spec.version : local.default_chart_version

  # Computed Helm release names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{chart}
  # Users can prefix metadata.name with component type if needed (e.g., "istio-prod")
  base_release_name    = "${var.metadata.name}-base"
  istiod_release_name  = "${var.metadata.name}-istiod"
  gateway_release_name = "${var.metadata.name}-gateway"

  # Gateway service configuration
  gateway_service_name = local.gateway_release_name
  gateway_port         = 80

  # Port forward command for istiod (uses computed release name)
  port_forward_command = "kubectl port-forward -n ${local.system_namespace} svc/${local.istiod_release_name} 15014:15014"

  # Kubernetes endpoint for istiod (uses computed release name)
  kube_endpoint = "${local.istiod_release_name}.${local.system_namespace}.svc.cluster.local:15012"

  # Ingress endpoint (gateway service endpoint, uses computed release name)
  ingress_endpoint = "${local.gateway_release_name}.${local.gateway_namespace}.svc.cluster.local:${local.gateway_port}"
}

