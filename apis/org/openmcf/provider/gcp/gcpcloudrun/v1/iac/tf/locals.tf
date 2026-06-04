# locals.tf - Local value transformations for GCP Cloud Run deployment

locals {
  # Service name: use spec.service_name if provided, otherwise metadata.name
  service_name = var.spec.service_name != null ? var.spec.service_name : var.metadata.name

  # Container image URI
  image = "${var.spec.container.image.repo}:${var.spec.container.image.tag}"

  # Memory in format expected by Cloud Run (e.g., "512Mi")
  memory = "${var.spec.container.memory}Mi"

  # CPU as string
  cpu = tostring(var.spec.container.cpu)

  # Container port (defaults to 8080)
  port = var.spec.container.port

  # Timeout in format expected by Cloud Run (e.g., "300s")
  timeout = "${var.spec.timeout_seconds}s"

  # Service account: use provided or null for default
  service_account = var.spec.service_account

  # Convert enum values to GCP API strings
  ingress = var.spec.ingress == "INGRESS_TRAFFIC_INTERNAL_ONLY" ? "INGRESS_TRAFFIC_INTERNAL_ONLY" : (
    var.spec.ingress == "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER" ? "INGRESS_TRAFFIC_INTERNAL_AND_CLOUD_LOAD_BALANCING" : "INGRESS_TRAFFIC_ALL"
  )

  execution_environment = var.spec.execution_environment == "EXECUTION_ENVIRONMENT_GEN1" ? "EXECUTION_ENVIRONMENT_GEN1" : "EXECUTION_ENVIRONMENT_GEN2"

  # Labels for resource tagging
  labels = merge(
    {
      resource      = "true"
      resource_name = var.metadata.name
      resource_kind = "gcpcloudrun"
    },
    var.metadata.id != null ? { resource_id = var.metadata.id } : {},
    var.metadata.org != null ? { organization = var.metadata.org } : {},
    var.metadata.env != null ? { environment = var.metadata.env } : {},
    var.metadata.labels != null ? var.metadata.labels : {}
  )

  # Environment variables: plain variables
  env_vars = var.spec.container.env != null && var.spec.container.env.variables != null ? [
    for k, v in var.spec.container.env.variables : {
      name  = k
      value = v
    }
  ] : []

  # Environment variables: secrets from Secret Manager
  env_secrets = var.spec.container.env != null && var.spec.container.env.secrets != null ? [
    for k, v in var.spec.container.env.secrets : {
      name   = k
      secret = v
    }
  ] : []

  # DNS configuration
  dns_enabled      = var.spec.dns != null ? var.spec.dns.enabled : false
  dns_hostnames    = var.spec.dns != null ? var.spec.dns.hostnames : []
  dns_managed_zone = var.spec.dns != null ? var.spec.dns.managed_zone.value : ""

  # VPC access configuration
  has_vpc_access = var.spec.vpc_access != null && (
    (var.spec.vpc_access.network != null && var.spec.vpc_access.network.value != null && var.spec.vpc_access.network.value != "") ||
    (var.spec.vpc_access.subnet != null && var.spec.vpc_access.subnet.value != null && var.spec.vpc_access.subnet.value != "")
  )
  vpc_network = var.spec.vpc_access != null && var.spec.vpc_access.network != null ? var.spec.vpc_access.network.value : null
  vpc_subnet  = var.spec.vpc_access != null && var.spec.vpc_access.subnet != null ? var.spec.vpc_access.subnet.value : null
}
