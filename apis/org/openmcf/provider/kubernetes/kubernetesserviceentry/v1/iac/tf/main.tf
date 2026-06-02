resource "kubernetes_manifest" "service_entry" {
  manifest = {
    apiVersion = "networking.istio.io/v1"
    kind       = "ServiceEntry"

    metadata = {
      name      = local.service_entry_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.service_entry_spec
  }
}
