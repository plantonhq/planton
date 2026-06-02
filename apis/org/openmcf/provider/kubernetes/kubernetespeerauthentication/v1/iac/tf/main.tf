resource "kubernetes_manifest" "peer_authentication" {
  manifest = {
    apiVersion = "security.istio.io/v1"
    kind       = "PeerAuthentication"

    metadata = {
      name      = var.metadata.name
      namespace = var.spec.namespace
      labels    = local.labels
    }

    spec = local.manifest_spec
  }
}
