locals {
  labels = {
    "app.kubernetes.io/name"       = "reference-grant"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "planton"
    "app.kubernetes.io/component"  = "reference-grant"
  }

  # The CR spec is var.spec minus the Planton "namespace" foreign key, which maps to
  # metadata.namespace rather than into the CR spec. target_cluster is already
  # dropped by the converter, which also emits camelCase, null-pruned keys, so no
  # other transformation is needed.
  manifest_spec = { for k, v in var.spec : k => v if k != "namespace" }
}
