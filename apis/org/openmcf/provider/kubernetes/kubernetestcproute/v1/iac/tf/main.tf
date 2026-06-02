resource "kubernetes_manifest" "tcp_route" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1alpha2"
    kind       = "TCPRoute"

    metadata = {
      name      = var.metadata.name
      namespace = var.spec.namespace
      labels    = local.labels
    }

    spec = local.manifest_spec
  }
}
