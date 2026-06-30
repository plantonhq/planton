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
    "resource_kind" = "signoz_kubernetes"
    "resource_name" = var.metadata.name
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

  # Use resource_id as the namespace name
  namespace = local.resource_id

  # Service names
  signoz_service_name         = "${var.metadata.name}-signoz"
  otel_collector_service_name = "${var.metadata.name}-otel-collector"

  # Kubernetes FQDNs and ports
  signoz_ui_port = 8080
  otel_grpc_port = 4317
  otel_http_port = 4318

  signoz_kube_endpoint         = "${local.signoz_service_name}.${local.namespace}.svc.cluster.local:${local.signoz_ui_port}"
  otel_collector_grpc_endpoint = "${local.otel_collector_service_name}.${local.namespace}.svc.cluster.local:${local.otel_grpc_port}"
  otel_collector_http_endpoint = "${local.otel_collector_service_name}.${local.namespace}.svc.cluster.local:${local.otel_http_port}"

  # Database configuration
  is_external_database = var.spec.database.is_external

  # SigNoz UI ingress
  signoz_ingress_is_enabled        = try(var.spec.ingress.ui.enabled, false)
  signoz_ingress_external_hostname = try(var.spec.ingress.ui.hostname, null)

  # OTel Collector ingress
  otel_collector_ingress_is_enabled     = try(var.spec.ingress.otel_collector.enabled, false)
  otel_collector_external_http_hostname = try(var.spec.ingress.otel_collector.hostname, null)

  # ClickHouse configuration (for self-managed mode)
  clickhouse_endpoint = (
    !local.is_external_database
    ? "${var.metadata.name}-clickhouse.${local.namespace}.svc.cluster.local:8123"
    : null
  )

  # Cluster configuration (for self-managed ClickHouse)
  cluster_is_enabled = (
    !local.is_external_database &&
    try(var.spec.database.managed_database.cluster.is_enabled, false)
  )

  shard_count = local.cluster_is_enabled ? try(
    var.spec.database.managed_database.cluster.shard_count, 1
  ) : 1

  replica_count = local.cluster_is_enabled ? try(
    var.spec.database.managed_database.cluster.replica_count, 1
  ) : 1

  # Zookeeper configuration (for distributed ClickHouse)
  zookeeper_is_enabled = (
    !local.is_external_database &&
    try(var.spec.database.managed_database.zookeeper.is_enabled, false)
  )

  # Ingress resource names (computed to avoid conflicts when multiple instances share a namespace)
  # SigNoz UI ingress resources
  signoz_certificate_name        = "${var.metadata.name}-signoz-cert"
  signoz_gateway_name            = "${var.metadata.name}-signoz-external-gateway"
  signoz_https_route_name        = "${var.metadata.name}-signoz-https-route"
  signoz_http_redirect_route_name = "${var.metadata.name}-signoz-http-redirect"

  # OTel Collector ingress resources
  otel_certificate_name        = "${var.metadata.name}-otel-http-cert"
  otel_gateway_name            = "${var.metadata.name}-otel-http-external-gateway"
  otel_https_route_name        = "${var.metadata.name}-otel-https-route"
  otel_http_redirect_route_name = "${var.metadata.name}-otel-http-redirect"

  # ClusterIssuer name extracted from hostname
  signoz_cert_cluster_issuer_name = (
    local.signoz_ingress_external_hostname != null
    ? join(".", slice(split(".", local.signoz_ingress_external_hostname), 1, length(split(".", local.signoz_ingress_external_hostname))))
    : ""
  )
  otel_cert_cluster_issuer_name = (
    local.otel_collector_external_http_hostname != null
    ? join(".", slice(split(".", local.otel_collector_external_http_hostname), 1, length(split(".", local.otel_collector_external_http_hostname))))
    : ""
  )

  # Istio ingress namespace and gateway service hostname
  istio_ingress_namespace                   = "istio-ingress"
  gateway_external_loadbalancer_service_hostname = "istio-ingress-gateway.istio-ingress.svc.cluster.local"

  # SigNoz frontend port (for routing)
  signoz_frontend_port = 3301
}

