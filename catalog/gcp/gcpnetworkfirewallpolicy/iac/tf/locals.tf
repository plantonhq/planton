locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null lets the google provider resolve its own
  # project from configuration or the GOOGLE_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The policy's name defaults to metadata.name when the spec leaves it
  # empty -- the same naming basis every kind uses.
  policy_name = var.spec.policy_name != "" ? var.spec.policy_name : var.metadata.name

  # Empty region -> the GLOBAL resource family; a region -> the REGIONAL
  # family. The two families share every argument but `region`, so every
  # resource below is count-gated on this one switch.
  is_regional = var.spec.region != ""

  # Rules keyed by priority: Google identifies a rule by its priority, so the
  # key makes a renumbered rule exactly what Google makes it -- a replaced
  # rule -- while every other rule is left untouched. Priorities are unique
  # by the spec's CEL. Each family gets the whole map or nothing.
  rules_by_priority = { for rule in var.spec.rules : tostring(rule.priority) => rule }
  global_rules      = local.is_regional ? {} : local.rules_by_priority
  regional_rules    = local.is_regional ? local.rules_by_priority : {}

  # Associations keyed by name (defaulting to <policy_name>-<n>). Names are
  # unique by the spec's CEL.
  associations_by_name = {
    for i, association in var.spec.associations :
    (association.name != "" ? association.name : "${local.policy_name}-${i + 1}") => association.network
  }
  global_associations   = local.is_regional ? {} : local.associations_by_name
  regional_associations = local.is_regional ? local.associations_by_name : {}

  # Declaration-order names for the output (a map loses the order).
  association_names = [
    for i, association in var.spec.associations :
    association.name != "" ? association.name : "${local.policy_name}-${i + 1}"
  ]

  # Empty defers to the provider default (DELETE); applied to the policy,
  # every rule, and every association alike.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
