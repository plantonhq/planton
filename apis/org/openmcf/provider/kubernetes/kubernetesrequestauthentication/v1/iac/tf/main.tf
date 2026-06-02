resource "kubernetes_manifest" "request_authentication" {
  manifest = {
    apiVersion = "security.istio.io/v1"
    kind       = "RequestAuthentication"

    metadata = {
      name      = local.request_authentication_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.request_authentication_spec
  }
}
