resource "kubernetes_manifest" "service_entry" {
  manifest = {
    apiVersion = "networking.istio.io/v1"
    kind       = "ServiceEntry"

    metadata = {
      name      = var.metadata.name
      namespace = var.spec.namespace
      labels    = local.labels
    }

    spec = local.manifest_spec
  }
}
