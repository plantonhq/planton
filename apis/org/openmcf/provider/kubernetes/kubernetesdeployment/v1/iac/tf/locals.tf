##############################################
# locals.tf
#
# Includes logic for:
#  - Deriving resource_id, labels, namespace
#  - Determining ingress hostnames
#  - Creating image_pull_secret_data from
#    docker_credential (if provider == "gcp_artifact_registry").
##############################################

locals {
  # Derive a stable resource ID (prefer `metadata.id`, fallback to `metadata.name`)
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Selector labels use metadata.name to ensure each deployment's pods are uniquely identified.
  # This prevents traffic routing conflicts when multiple deployments share a namespace.
  selector_labels = {
    "app"           = var.metadata.name
    "resource_name" = var.metadata.name
  }

  # Base labels include selector labels plus metadata
  base_labels = merge(local.selector_labels, {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "microservice_kubernetes"
  })

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

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  # Users can prefix metadata.name with component type if needed (e.g., "deploy-my-app")
  env_secret_name                   = "${var.metadata.name}-env-secrets"
  image_pull_secret_name            = "${var.metadata.name}-image-pull"
  ingress_certificate_name          = "${var.metadata.name}-ingress-cert"
  external_gateway_name             = "${var.metadata.name}-external"
  internal_gateway_name             = "${var.metadata.name}-internal"
  http_external_redirect_route_name = "${var.metadata.name}-http-external-redirect"
  https_external_route_name         = "${var.metadata.name}-https-external"
  http_internal_redirect_route_name = "${var.metadata.name}-http-internal-redirect"
  https_internal_route_name         = "${var.metadata.name}-https-internal"

  # Use metadata.name for service name to avoid conflicts when multiple deployments share a namespace
  kube_service_name = var.metadata.name

  # Internal DNS name for the service
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local"

  # Handy port-forward command
  kube_port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} 8080:8080"

  # Safely handle optional ingress values
  ingress_is_enabled = try(var.spec.ingress.is_enabled, false)
  ingress_dns_domain = try(var.spec.ingress.dns_domain, "")

  # External hostname (null if not applicable)
  ingress_external_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}.${local.ingress_dns_domain}" : null

  # Internal hostname (null if not applicable)
  ingress_internal_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}-internal.${local.ingress_dns_domain}" : null

  # For certificate creation
  ingress_cert_cluster_issuer_name = local.ingress_dns_domain != "" ? local.ingress_dns_domain : null
  ingress_cert_secret_name         = local.resource_id
}
