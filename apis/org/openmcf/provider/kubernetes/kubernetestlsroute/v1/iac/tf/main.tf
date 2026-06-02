resource "kubernetes_manifest" "tls_route" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "TLSRoute"

    metadata = {
      name      = var.metadata.name
      namespace = var.spec.namespace
      labels    = local.labels
    }

    spec = local.manifest_spec
  }
}
