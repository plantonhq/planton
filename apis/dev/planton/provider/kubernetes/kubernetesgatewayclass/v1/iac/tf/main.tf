resource "kubernetes_manifest" "gateway_class" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "GatewayClass"

    metadata = {
      name   = var.metadata.name
      labels = local.labels
    }

    spec = local.manifest_spec
  }
}
