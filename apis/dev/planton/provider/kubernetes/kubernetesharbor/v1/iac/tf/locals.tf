locals {
  namespace = (
    var.harbor_kubernetes.spec.namespace != null && var.harbor_kubernetes.spec.namespace != ""
    ? var.harbor_kubernetes.spec.namespace
    : var.harbor_kubernetes.metadata.name
  )

  labels = {
    "app.kubernetes.io/name"     = "harbor"
    "app.kubernetes.io/instance" = var.harbor_kubernetes.metadata.name
  }

  core_service_name     = "${var.harbor_kubernetes.metadata.name}-harbor-core"
  portal_service_name   = "${var.harbor_kubernetes.metadata.name}-harbor-portal"
  registry_service_name = "${var.harbor_kubernetes.metadata.name}-harbor-registry"

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  ingress_cert_secret_name  = "${var.harbor_kubernetes.metadata.name}-ingress-cert"
  ingress_certificate_name  = "${var.harbor_kubernetes.metadata.name}-ingress-cert"
  ingress_gateway_name      = "${var.harbor_kubernetes.metadata.name}-external"
  ingress_http_route_name   = "${var.harbor_kubernetes.metadata.name}-https-external"
}

