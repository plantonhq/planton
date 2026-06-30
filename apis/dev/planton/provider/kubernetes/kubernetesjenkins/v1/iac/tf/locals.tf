###########################
# locals.tf
###########################

locals {
  # Use metadata.id if provided and non-empty; otherwise, fallback to metadata.name
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "jenkins_kubernetes"
  }

  # Organization label only if org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment label only if env and env.id are provided
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
    ) ? {
    "organization" = var.metadata.env
  } : {}
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Use namespace from spec with fallback to resource_id
  namespace = (
    var.spec.namespace != null && var.spec.namespace != ""
    ? var.spec.namespace
    : local.resource_id
  )

  # Handle optional fields in var.spec.ingress
  ingress_is_enabled = try(var.spec.ingress.is_enabled, false)
  ingress_dns_domain = try(var.spec.ingress.dns_domain, "")

  # If ingress is enabled and domain is non-empty, define external/internal hostnames
  ingress_external_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}.${local.ingress_dns_domain}" : null

  ingress_internal_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}-internal.${local.ingress_dns_domain}" : null

  # Official Jenkins helm chart repository & chart name
  jenkins_chart_repo = "https://charts.jenkins.io"
  jenkins_chart_name = "jenkins"
  # Pin a specific chart version, or allow the user to override in var.spec.helm_values
  # You could also expose this as a variable if you prefer
  jenkins_chart_version = "5.8.8"

  # Example placeholders. Adjust if needed, or move to variables.
  istio_ingress_namespace      = "istio-ingress"
  gateway_ingress_class_name   = "istio"
  gateway_external_lb_hostname = "ingress-external.istio-ingress.svc.cluster.local"

  # For the certificate we need a list of hostnames:
  ingress_hostnames = compact([
    local.ingress_external_hostname,
    local.ingress_internal_hostname,
  ])

  # The cluster-issuer name typically matches your DNS domain or a known issuer
  ingress_cert_cluster_issuer_name = local.ingress_dns_domain

  # The name of the TLS secret that cert-manager will create
  ingress_cert_secret_name = local.resource_id

  # The Jenkins service name (port 80) you want to route traffic to.
  # If you install Jenkins via Helm with a particular name, set that here.
  # If your helm chart sets the service name as "<chartName>-jenkins", adapt accordingly.
  jenkins_kube_service_name = "${local.resource_id}-jenkins"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  # Users can prefix metadata.name with component type if needed (e.g., "jenkins-my-ci")
  admin_credentials_secret_name = "${var.metadata.name}-admin-credentials"
  ingress_certificate_name      = "${var.metadata.name}-ingress-cert"
  external_gateway_name         = "${var.metadata.name}-external"
  http_redirect_route_name      = "${var.metadata.name}-http-redirect"
  https_route_name              = "${var.metadata.name}-https"
  tls_secret_name               = "${var.metadata.name}-tls"
}
