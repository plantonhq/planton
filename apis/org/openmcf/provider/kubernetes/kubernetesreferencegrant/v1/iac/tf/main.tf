resource "kubernetes_manifest" "reference_grant" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "ReferenceGrant"

    metadata = {
      name      = local.reference_grant_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = local.reference_grant_spec
  }
}
