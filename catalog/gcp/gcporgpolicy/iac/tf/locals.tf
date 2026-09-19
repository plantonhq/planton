locals {
  # The constraint the policy configures: the spec's literal constraint, or
  # the resolved custom_constraint reference (the custom.<name> handle).
  # Exactly one is set (proto-CEL-enforced).
  constraint = var.spec.constraint != "" ? var.spec.constraint : var.spec.custom_constraint

  # Scope selection. Exactly one arm renders Google's parent; an empty scope
  # means the provider's default project, resolved once through the
  # data.google_project lookup below (the same fallback the Pulumi module
  # takes through the provider's client config).
  scope_project_id = var.spec.scope != null ? var.spec.scope.project_id : ""
  scope_folder_id  = var.spec.scope != null ? var.spec.scope.folder_id : ""
  scope_org_id     = var.spec.scope != null ? var.spec.scope.organization_id : ""

  # Count-gating the lookup keeps every explicitly-scoped plan credential-free:
  # the data source runs only when the manifest names no scope at all.
  needs_project_lookup = local.scope_project_id == "" && local.scope_folder_id == "" && local.scope_org_id == ""

  parent = (
    local.scope_project_id != ""
    ? (startswith(local.scope_project_id, "projects/") ? local.scope_project_id : "projects/${local.scope_project_id}")
    : local.scope_folder_id != ""
    ? (startswith(local.scope_folder_id, "folders/") ? local.scope_folder_id : "folders/${local.scope_folder_id}")
    : local.scope_org_id != ""
    ? "organizations/${local.scope_org_id}"
    : "projects/${data.google_project.this[0].project_id}"
  )

  # PARITY: Google's API models a rule's verdict as booleans in a one-of and
  # the provider flattens that into the tri-state strings "TRUE" / "FALSE" /
  # unset. The spec keeps the API's shape (a oneof of bools, null when the
  # arm is not chosen) and this module renders the string form for exactly
  # the arm that is set -- the Pulumi module does the same -- so
  # `enforce: false` reaches Google as "FALSE" (a real rule) and an unset arm
  # is never sent.
  policy_rules = var.spec.policy != null ? [
    for r in var.spec.policy.rules : {
      allow_all  = r.allow_all == null ? null : (r.allow_all ? "TRUE" : "FALSE")
      deny_all   = r.deny_all == null ? null : (r.deny_all ? "TRUE" : "FALSE")
      enforce    = r.enforce == null ? null : (r.enforce ? "TRUE" : "FALSE")
      values     = r.values
      condition  = r.condition
      parameters = r.parameters != "" ? r.parameters : null
    }
  ] : []

  dry_run_rules = var.spec.dry_run_policy != null ? [
    for r in var.spec.dry_run_policy.rules : {
      allow_all  = r.allow_all == null ? null : (r.allow_all ? "TRUE" : "FALSE")
      deny_all   = r.deny_all == null ? null : (r.deny_all ? "TRUE" : "FALSE")
      enforce    = r.enforce == null ? null : (r.enforce ? "TRUE" : "FALSE")
      values     = r.values
      condition  = r.condition
      parameters = r.parameters != "" ? r.parameters : null
    }
  ] : []
}
