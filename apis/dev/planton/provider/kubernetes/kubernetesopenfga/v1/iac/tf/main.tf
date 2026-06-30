# Conditionally create namespace based on create_namespace flag
resource "kubernetes_namespace" "openfga_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}
