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
    "resource_kind" = "kubernetes_openbao"
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

  # Get namespace from spec
  namespace = var.spec.namespace

  # Helm chart configuration
  helm_chart_name    = "openbao"
  helm_chart_repo    = "https://openbao.github.io/openbao-helm"
  helm_chart_version = coalesce(var.spec.helm_chart_version, "0.23.3")

  # OpenBao ports
  openbao_port         = 8200
  openbao_cluster_port = 8201

  # Service name (uses release name)
  kube_service_name = var.metadata.name

  # Internal DNS name for the service
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local:${local.openbao_port}"

  # API address
  api_address = "http://${local.kube_service_name}.${local.namespace}.svc.cluster.local:${local.openbao_port}"

  # Cluster address (HA mode)
  cluster_address = "https://${local.kube_service_name}-0.${local.kube_service_name}-internal.${local.namespace}.svc.cluster.local:${local.openbao_cluster_port}"

  # Port-forward command
  kube_port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} ${local.openbao_port}:${local.openbao_port}"

  # HA configuration
  ha_enabled  = try(var.spec.high_availability.enabled, false)
  ha_replicas = local.ha_enabled ? coalesce(try(var.spec.high_availability.replicas, null), 3) : 1

  # Server replicas
  server_replicas = var.spec.server_container.replicas

  # UI enabled (default true)
  ui_enabled = coalesce(var.spec.ui_enabled, true)

  # TLS configuration
  tls_enabled = coalesce(var.spec.tls_enabled, false)

  # Injector configuration
  injector_enabled  = try(var.spec.injector.enabled, false)
  injector_replicas = local.injector_enabled ? coalesce(try(var.spec.injector.replicas, null), 1) : 0

  # Ingress configuration
  ingress_is_enabled        = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)

  # Extract domain from hostname for certificate issuer
  # Example: "openbao.example.com" -> "example.com"
  ingress_cert_cluster_issuer_name = local.ingress_external_hostname != null ? (
    join(".", slice(split(".", local.ingress_external_hostname), 1,
      length(split(".", local.ingress_external_hostname))))
  ) : null

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  ingress_cert_secret_name         = "${var.metadata.name}-tls"
  ingress_certificate_name         = "${var.metadata.name}-certificate"
  ingress_gateway_name             = "${var.metadata.name}-external"
  ingress_http_redirect_route_name = "${var.metadata.name}-http-redirect"
  ingress_https_route_name         = "${var.metadata.name}-https"
}
