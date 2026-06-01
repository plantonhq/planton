resource "kubernetes_manifest" "http_route" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"

    metadata = {
      name      = local.route_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.http_route_spec
  }
}
