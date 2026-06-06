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
    "resource_kind" = "redis_kubernetes"
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

  # Labels for Redis pods
  redis_pod_selector_labels = {
    "app.kubernetes.io/component" = "master"
    "app.kubernetes.io/instance"  = local.resource_id
    "app.kubernetes.io/name"      = "redis"
  }

  # Get namespace from spec
  namespace = var.spec.namespace

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Users can prefix metadata.name with component type if needed (e.g., "redis-my-cache")
  password_secret_name     = "${var.metadata.name}-password"
  external_lb_service_name = "${var.metadata.name}-external-lb"

  # Service name
  kube_service_name = "${var.metadata.name}-master"

  # Internal DNS name for the service
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local"

  # Handy port-forward command
  kube_port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} 8080:8080"

  # Safely handle optional ingress values
  ingress_is_enabled        = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)

  # Redis image configuration (using legacy Bitnami repository)
  redis_image_registry   = "docker.io"
  redis_image_repository = "bitnamilegacy/redis"
  redis_image_tag        = "8.2.1-debian-12-r0"
}
