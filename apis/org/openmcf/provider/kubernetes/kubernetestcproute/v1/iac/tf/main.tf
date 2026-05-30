resource "kubernetes_manifest" "tcp_route" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1alpha2"
    kind       = "TCPRoute"

    metadata = {
      name      = local.route_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.tcp_route_spec
  }
}
