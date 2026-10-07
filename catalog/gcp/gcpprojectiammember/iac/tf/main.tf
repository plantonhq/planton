# A single ADDITIVE project-level IAM grant: one role, to one member, on one
# project. Additive means the provider merges this (role, member) pair into
# the project's IAM policy without touching any other member's bindings on the
# same role, and destroy subtracts only this exact pair — grants made by other
# charts, teams, or tools are never clobbered.
#
# Every argument is immutable (ForceNew): IAM grants have no update — any
# change replaces the grant atomically, which is also how the API behaves.
resource "google_project_iam_member" "this" {
  project = local.project
  role    = var.spec.role
  member  = var.spec.member

  # An IAM Condition is part of the grant's identity: the same role granted
  # with and without a condition are two independent bindings in the policy.
  dynamic "condition" {
    for_each = var.spec.condition != null ? [var.spec.condition] : []
    content {
      title       = condition.value.title
      expression  = condition.value.expression
      description = condition.value.description
    }
  }
}
