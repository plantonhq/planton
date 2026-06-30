resource "kubernetes_manifest" "reference_grant" {
  manifest = {
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "ReferenceGrant"

    metadata = {
      name      = var.metadata.name
      namespace = var.spec.namespace
      labels    = local.labels
    }

    spec = local.manifest_spec
  }
}
