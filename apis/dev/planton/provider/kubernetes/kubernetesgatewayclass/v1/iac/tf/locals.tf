locals {
  labels = {
    "app.kubernetes.io/name"       = "gateway-class"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "planton"
    "app.kubernetes.io/component"  = "gateway-class"
  }

  # Cluster-scoped CR: the converter already emits camelCase, null-pruned keys
  # (target_cluster dropped), so the spec is passed through unchanged.
  manifest_spec = var.spec
}
