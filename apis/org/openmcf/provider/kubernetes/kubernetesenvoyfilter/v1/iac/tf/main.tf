resource "kubernetes_manifest" "envoy_filter" {
  manifest = {
    apiVersion = "networking.istio.io/v1alpha3"
    kind       = "EnvoyFilter"

    metadata = {
      name      = local.envoy_filter_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.envoy_filter_spec
  }
}
