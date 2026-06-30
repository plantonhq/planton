# Conditionally create namespace for MongoDB if create_namespace is true
resource "kubernetes_namespace_v1" "mongodb_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}
