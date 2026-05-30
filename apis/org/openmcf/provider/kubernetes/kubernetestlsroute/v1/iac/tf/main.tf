resource "kubernetes_manifest" "tls_route" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "TLSRoute"

    metadata = {
      name      = local.route_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.tls_route_spec
  }
}
