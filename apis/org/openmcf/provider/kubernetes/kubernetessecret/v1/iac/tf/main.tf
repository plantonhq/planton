# Kubernetes Secret Terraform Module
# Creates a Kubernetes Secret with type-safe data variants

resource "kubernetes_secret_v1" "secret" {
  metadata {
    name        = var.spec.name
    namespace   = var.spec.namespace
    labels      = local.labels
    annotations = local.annotations
  }

  type      = local.secret_type
  data      = local.secret_data
  immutable = var.spec.immutable
}
