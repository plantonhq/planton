resource "kubernetes_namespace" "helm_release_namespace" {
  count = local.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace_name
    labels = local.final_labels
  }
}
