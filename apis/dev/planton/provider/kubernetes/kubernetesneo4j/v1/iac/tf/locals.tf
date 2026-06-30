locals {
  # Derive a stable resource ID (prefer metadata.id if set, otherwise fallback to metadata.name)
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels applied to all Kubernetes resources
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "neo4j_kubernetes"
    "resource_name" = var.metadata.name
  }

  # Organization label (only if non-empty)
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ? { "organization" = var.metadata.org }
    : {}
  )

  # Environment label (only if non-empty)
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
    ? { "environment" = var.metadata.env }
    : {}
  )

  # Merge all labels
  labels = merge(local.base_labels, local.org_label, local.env_label)

  # Get namespace from spec
  namespace = var.spec.namespace

  # Neo4j Helm chart constants
  neo4j_helm_chart_name    = "neo4j"
  neo4j_helm_chart_repo    = "https://helm.neo4j.com/neo4j"
  neo4j_helm_chart_version = "2025.03.0"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # The Neo4j Helm chart creates a secret named "<release>-auth" for the password
  password_secret_name = "${var.metadata.name}-auth"

  # Service name created by the Helm chart (follows Neo4j chart naming pattern)
  service_name = "${var.metadata.name}-neo4j"

  # Fully qualified service name for in-cluster connections
  service_fqdn = "${local.service_name}.${local.namespace}.svc.cluster.local"

  # Bolt URI for database connections (port 7687)
  bolt_uri = "bolt://${local.service_fqdn}:7687"

  # HTTP URI for Neo4j Browser (port 7474)
  http_uri = "http://${local.service_fqdn}:7474"

  # Port-forward command for local development
  port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.service_name} 7474:7474"

  # Ingress configuration
  ingress_enabled         = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, "")

  # Memory configuration (with defaults)
  heap_max    = try(var.spec.memory_config.heap_max, "")
  page_cache  = try(var.spec.memory_config.page_cache, "")
}

