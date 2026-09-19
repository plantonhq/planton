# The provider's default project, read only when the manifest names no scope
# (see locals.needs_project_lookup) so every explicitly-scoped plan stays
# credential-free.
data "google_project" "this" {
  count = local.needs_project_lookup ? 1 : 0
}

# The organization policy: the rules for one constraint at one scope.
#
# Its identity is the name, {parent}/policies/{constraint}; both halves are
# immutable, so a change to the scope or the constraint recreates the policy.
# The two rule sets are the same shape -- `spec` is what Google ENFORCES,
# `dry_run_spec` what it only AUDITS -- and each renders only when the
# manifest carries it. Optional scalars are sent only when set so the
# provider's defaults stay the provider's; deletion_policy likewise.
resource "google_org_policy_policy" "this" {
  name   = "${local.parent}/policies/${local.constraint}"
  parent = local.parent

  dynamic "spec" {
    for_each = var.spec.policy != null ? [var.spec.policy] : []
    content {
      inherit_from_parent = spec.value.inherit_from_parent ? true : null
      reset               = spec.value.reset ? true : null

      dynamic "rules" {
        for_each = local.policy_rules
        content {
          allow_all  = rules.value.allow_all
          deny_all   = rules.value.deny_all
          enforce    = rules.value.enforce
          parameters = rules.value.parameters

          dynamic "values" {
            for_each = rules.value.values != null ? [rules.value.values] : []
            content {
              allowed_values = length(values.value.allowed_values) > 0 ? values.value.allowed_values : null
              denied_values  = length(values.value.denied_values) > 0 ? values.value.denied_values : null
            }
          }

          dynamic "condition" {
            for_each = rules.value.condition != null ? [rules.value.condition] : []
            content {
              expression  = condition.value.expression
              title       = condition.value.title != "" ? condition.value.title : null
              description = condition.value.description != "" ? condition.value.description : null
              location    = condition.value.location != "" ? condition.value.location : null
            }
          }
        }
      }
    }
  }

  dynamic "dry_run_spec" {
    for_each = var.spec.dry_run_policy != null ? [var.spec.dry_run_policy] : []
    content {
      inherit_from_parent = dry_run_spec.value.inherit_from_parent ? true : null
      reset               = dry_run_spec.value.reset ? true : null

      dynamic "rules" {
        for_each = local.dry_run_rules
        content {
          allow_all  = rules.value.allow_all
          deny_all   = rules.value.deny_all
          enforce    = rules.value.enforce
          parameters = rules.value.parameters

          dynamic "values" {
            for_each = rules.value.values != null ? [rules.value.values] : []
            content {
              allowed_values = length(values.value.allowed_values) > 0 ? values.value.allowed_values : null
              denied_values  = length(values.value.denied_values) > 0 ? values.value.denied_values : null
            }
          }

          dynamic "condition" {
            for_each = rules.value.condition != null ? [rules.value.condition] : []
            content {
              expression  = condition.value.expression
              title       = condition.value.title != "" ? condition.value.title : null
              description = condition.value.description != "" ? condition.value.description : null
              location    = condition.value.location != "" ? condition.value.location : null
            }
          }
        }
      }
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
