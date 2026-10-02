# The organization's custom constraint: a definition, enforced by however
# many GcpOrgPolicy resources reference it.
#
# The name (with its custom. prefix), the parent organization, and the
# resource types are immutable -- a change to any recreates the constraint,
# and every policy enforcing the old name lapses -- while the condition,
# action, methods, display name, and description update in place. Optional
# strings are sent only when set; deletion_policy is sent only when set so
# the provider's default (DELETE) stays the provider's.
resource "google_org_policy_custom_constraint" "this" {
  name   = local.constraint_name
  parent = local.parent

  resource_types = var.spec.resource_types
  method_types   = var.spec.method_types
  condition      = var.spec.condition
  action_type    = var.spec.action_type

  display_name    = var.spec.display_name != "" ? var.spec.display_name : null
  description     = var.spec.description != "" ? var.spec.description : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
