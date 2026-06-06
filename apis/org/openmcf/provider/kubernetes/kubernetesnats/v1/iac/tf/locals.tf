locals {
  # Use 'metadata.id' if set, otherwise fall back to 'metadata.name'.
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels for all resources
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "nats_kubernetes"
    "resource_name" = var.metadata.name
  }

  # Organization label only if non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ? { "organization" = var.metadata.org }
    : {}
  )

  # Environment label only if non-empty
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
    ? { "environment" = var.metadata.env }
    : {}
  )

  # Merge all labels
  labels = merge(local.base_labels, local.org_label, local.env_label)

  # Get namespace from spec
  namespace = var.spec.namespace

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  # Users can prefix metadata.name with component type if needed (e.g., "nats-my-bus")
  auth_secret_name         = "${var.metadata.name}-auth"
  no_auth_user_secret_name = "${var.metadata.name}-no-auth-user"
  tls_secret_name          = "${var.metadata.name}-tls"
  external_lb_service_name = "${var.metadata.name}-external-lb"

  # NATS service name (created by Helm chart)
  nats_service_name = var.metadata.name

  # Fixed NATS client port
  nats_client_port = 4222

  # Internal FQDN for NATS service
  internal_client_url = "nats://${local.nats_service_name}.${local.namespace}.svc.cluster.local:${local.nats_client_port}"

  # Ingress configuration
  ingress_is_enabled = try(var.spec.ingress.enabled, false)
  ingress_hostname   = try(var.spec.ingress.hostname, null)

  # NATS Helm chart version (from spec or default)
  nats_helm_chart_version = try(var.spec.nats_helm_chart_version, "2.12.3")

  # ============================================================================
  # NACK (NATS Controllers for Kubernetes) Configuration
  # ============================================================================

  # NACK controller enabled flag
  nack_controller_enabled = try(var.spec.nack_controller.enabled, false)

  # NACK Helm chart version (from spec or default)
  nack_helm_chart_version = try(var.spec.nack_controller.helm_chart_version, "0.31.1")

  # NACK app version (GitHub release tag) - differs from chart version
  # Used for fetching CRDs from GitHub
  nack_app_version = try(var.spec.nack_controller.app_version, "0.21.1")

  # NACK CRDs URL - fetched from GitHub using app version (not chart version!)
  nack_crds_url = "https://raw.githubusercontent.com/nats-io/nack/v${local.nack_app_version}/deploy/crds.yml"

  # NACK Helm chart repository
  nack_helm_chart_repo_url = "https://nats-io.github.io/k8s/helm/charts"

  # Enable control-loop mode for KeyValue, ObjectStore, and Account support
  nack_enable_control_loop = try(var.spec.nack_controller.enable_control_loop, false)

  # Admin username for basic auth
  admin_username = "nats"

  # NATS URL for NACK controller (with or without credentials)
  nack_nats_url = (
    try(var.spec.auth.enabled, false) && try(var.spec.auth.scheme, "") == "basic_auth"
    ? "nats://${local.admin_username}:${random_password.nats_admin_password[0].result}@${local.nats_service_name}.${local.namespace}.svc.cluster.local:${local.nats_client_port}"
    : local.internal_client_url
  )

  # Streams configuration
  streams = try(var.spec.streams, [])
}
