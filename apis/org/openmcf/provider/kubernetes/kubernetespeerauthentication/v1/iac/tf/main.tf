resource "kubernetes_manifest" "peer_authentication" {
  manifest = {
    apiVersion = "security.istio.io/v1"
    kind       = "PeerAuthentication"

    metadata = {
      name      = local.peer_authentication_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.peer_authentication_spec
  }
}
