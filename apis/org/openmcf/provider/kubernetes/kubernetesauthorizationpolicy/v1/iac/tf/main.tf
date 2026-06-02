resource "kubernetes_manifest" "authorization_policy" {
  manifest = {
    apiVersion = "security.istio.io/v1"
    kind       = "AuthorizationPolicy"

    metadata = {
      name      = local.authorization_policy_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.authorization_policy_spec
  }
}
