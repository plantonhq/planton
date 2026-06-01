resource "kubernetes_manifest" "grpc_route" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "GRPCRoute"

    metadata = {
      name      = local.route_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.grpc_route_spec
  }
}
