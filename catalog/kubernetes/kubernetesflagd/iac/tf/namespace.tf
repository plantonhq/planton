# The installation namespace, created only when create_namespace is set
# (Pulumi twin: namespace.go).
resource "kubernetes_namespace_v1" "flagd" {
  count = try(var.spec.create_namespace, false) ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}
