locals {
  # Planton identity labels — the planton.ai/* family defined once in
  # pkg/kubernetes/manifestprojection, which the Pulumi projection helper
  # stamps too. Conditional entries use the null-prune idiom: heterogeneous
  # conditional merges fail HCL type unification when sibling entries infer
  # as different object types.
  labels = {
    for k, v in {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesRequestAuthentication"
      "planton.ai/resource-id"   = (var.metadata.id != null && var.metadata.id != "") ? var.metadata.id : null
      "planton.ai/organization"  = (var.metadata.org != null && var.metadata.org != "") ? var.metadata.org : null
      "planton.ai/environment"   = (var.metadata.env != null && var.metadata.env != "") ? var.metadata.env : null
    } : k => v if v != null
  }

  # The CR spec is var.spec minus the Planton "namespace" foreign key, which maps to
  # metadata.namespace rather than into the CR spec. The converter already emits
  # camelCase, null-pruned keys with StringValueOrRef foreign keys resolved to
  # literal strings, so no other transformation is needed.
  manifest_spec = { for k, v in var.spec : k => v if k != "namespace" }
}
