resource "kubernetes_manifest" "envoy_filter" {
  manifest = {
    apiVersion = "networking.istio.io/v1alpha3"
    kind       = "EnvoyFilter"

    metadata = {
      name      = var.metadata.name
      namespace = var.spec.namespace
      labels    = local.labels
    }

    spec = local.manifest_spec
  }
}
