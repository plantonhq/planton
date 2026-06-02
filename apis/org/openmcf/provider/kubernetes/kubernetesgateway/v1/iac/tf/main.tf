resource "kubernetes_manifest" "gateway" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "Gateway"

    metadata = {
      name      = var.metadata.name
      namespace = var.spec.namespace
      labels    = local.labels
    }

    spec = local.manifest_spec
  }
}
