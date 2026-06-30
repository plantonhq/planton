##############################################
# locals.tf
#
# Includes logic for:
#  - Deriving resource_id and labels
#  - Tekton manifest URLs
#  - Ingress configuration
#
# IMPORTANT: Namespace Behavior
# Tekton manifests install to 'tekton-pipelines' namespace.
# This is fixed and cannot be customized.
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
    "resource_kind" = "kubernetes_tekton"
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

  # Tekton uses fixed namespace (created by manifests)
  namespace = "tekton-pipelines"

  # Version normalization
  pipeline_version  = coalesce(var.spec.pipeline_version, "latest")
  dashboard_version = try(coalesce(var.spec.dashboard.version, "latest"), "latest")

  # Dashboard enabled flag
  dashboard_enabled = try(var.spec.dashboard.enabled, false)

  # Manifest URLs
  pipeline_manifest_url  = "https://storage.googleapis.com/tekton-releases/pipeline/${local.pipeline_version}/release.yaml"
  dashboard_manifest_url = "https://infra.tekton.dev/tekton-releases/dashboard/${local.dashboard_version}/release.yaml"

  # Dashboard service configuration (fixed by Tekton Dashboard manifest)
  dashboard_service_name = "tekton-dashboard"
  dashboard_service_port = 9097

  # Dashboard internal endpoint
  dashboard_internal_endpoint = "${local.dashboard_service_name}.${local.namespace}.svc.cluster.local:${local.dashboard_service_port}"

  # Port forward command
  port_forward_dashboard_command = "kubectl port-forward -n ${local.namespace} service/${local.dashboard_service_name} 9097:9097"

  # Cloud events configuration
  cloud_events_enabled  = var.spec.cloud_events != null && try(var.spec.cloud_events.sink_url, "") != ""
  cloud_events_sink_url = local.cloud_events_enabled ? var.spec.cloud_events.sink_url : ""

  # Ingress configuration
  ingress_enabled  = try(var.spec.dashboard.ingress.enabled, false) && try(var.spec.dashboard.ingress.hostname, "") != ""
  ingress_hostname = local.ingress_enabled ? var.spec.dashboard.ingress.hostname : ""

  # Extract domain from hostname for ClusterIssuer name
  ingress_hostname_parts = local.ingress_enabled ? split(".", local.ingress_hostname) : []
  cluster_issuer_name    = length(local.ingress_hostname_parts) > 1 ? join(".", slice(local.ingress_hostname_parts, 1, length(local.ingress_hostname_parts))) : ""

  # Computed resource names (for multi-instance support)
  cert_secret_name         = "${var.metadata.name}-dashboard-cert"
  gateway_name             = "${var.metadata.name}-dashboard-external"
  http_redirect_route_name = "${var.metadata.name}-dashboard-http-redirect"
  https_route_name         = "${var.metadata.name}-dashboard-https"

  # Gateway API configuration
  istio_ingress_namespace              = "istio-ingress"
  gateway_ingress_class_name           = "istio"
  gateway_external_lb_service_hostname = "ingress-external.istio-ingress.svc.cluster.local"
}
