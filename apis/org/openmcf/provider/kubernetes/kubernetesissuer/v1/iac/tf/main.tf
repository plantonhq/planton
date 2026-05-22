resource "kubernetes_manifest" "issuer" {
  manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "Issuer"

    metadata = {
      name      = local.issuer_name
      namespace = local.namespace
      labels    = var.metadata.labels
    }

    spec = local.issuer_spec
  }
}
