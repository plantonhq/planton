resource "kubernetes_manifest" "grpc_route" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "GRPCRoute"

    metadata = {
      name      = var.metadata.name
      namespace = var.spec.namespace
      labels    = local.labels
    }

    spec = local.manifest_spec
  }
}
