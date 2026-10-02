locals {
  # Honor the spec contract: an empty project_id falls back to the
  # provider's default project. Passing null (instead of "") lets the
  # google provider resolve its own project from configuration.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The scope ID defaults to metadata.name -- identical to the Pulumi
  # module.
  scope_id        = var.spec.scope_id != "" ? var.spec.scope_id : var.metadata.name
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, on the scope
  # and on every namespace, role binding, and cluster binding (never on the
  # Kubernetes namespace_labels). User labels merge in first so the
  # platform attribution labels can never be clobbered by a spec label with
  # the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpgkefleetscope"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  attribution_labels = merge(local.base_labels, local.org_label, local.env_label, local.id_label)
  final_labels       = merge(var.spec.labels, local.attribution_labels)

  # The folded children, each keyed by its own declared ID, so adding or
  # removing one never renames or recreates its siblings.
  namespaces         = { for namespace in var.spec.namespaces : namespace.scope_namespace_id => namespace }
  rbac_role_bindings = { for binding in var.spec.rbac_role_bindings : binding.scope_rbac_role_binding_id => binding }

  # A cluster binding addresses its membership by location and ID, parsed
  # from the membership's full name
  # ("projects/{p}/locations/{l}/memberships/{id}"); the binding lives in
  # the scope's fleet project, which Google requires the membership to
  # share.
  membership_bindings = {
    for binding in var.spec.membership_bindings : binding.membership_binding_id => {
      location      = split("/", binding.membership)[3]
      membership_id = split("/", binding.membership)[5]
      labels        = binding.labels
    }
  }
}
