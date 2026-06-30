# Conditionally create namespace for OpenBao if create_namespace is true
resource "kubernetes_namespace" "openbao_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}
