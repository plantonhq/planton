locals {
  # The policy's short name defaults to metadata.name when the spec leaves it
  # empty -- the same naming basis every kind uses.
  short_name = var.spec.short_name != "" ? var.spec.short_name : var.metadata.name

  # Google's parent argument is one string rendered from whichever arm the
  # spec set (exactly one, proto-CEL-enforced): organizations/{id} for an
  # organization-level policy, folders/{id} for a folder-level one. A
  # folder_id that already carries its prefix passes through so a
  # hand-written full name still works -- the same rule as GcpFolder.
  parent = (
    var.spec.parent.organization_id != ""
    ? "organizations/${var.spec.parent.organization_id}"
    : startswith(var.spec.parent.folder_id, "folders/")
    ? var.spec.parent.folder_id
    : "folders/${var.spec.parent.folder_id}"
  )

  # Rules keyed by priority: Google identifies a rule by its priority, so the
  # key makes a renumbered rule exactly what Google makes it -- a replaced
  # rule -- while every other rule is left untouched. Priorities are unique
  # by the spec's CEL.
  rules_by_priority = { for rule in var.spec.rules : tostring(rule.priority) => rule }

  # Associations keyed by name (defaulting to <short_name>-<n>), each with
  # its attachment target rendered like the parent. Names are unique by the
  # spec's CEL.
  associations_by_name = {
    for i, association in var.spec.associations :
    (association.name != "" ? association.name : "${local.short_name}-${i + 1}") => {
      target = (
        association.target.organization_id != ""
        ? "organizations/${association.target.organization_id}"
        : startswith(association.target.folder_id, "folders/")
        ? association.target.folder_id
        : "folders/${association.target.folder_id}"
      )
    }
  }

  # Declaration-order names for the output (a map loses the order).
  association_names = [
    for i, association in var.spec.associations :
    association.name != "" ? association.name : "${local.short_name}-${i + 1}"
  ]

  # Empty defers to the provider default (DELETE); applied to the policy,
  # every rule, and every association alike.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
