# A single ADDITIVE IAM grant ON a Cloud Storage bucket: one role, to one
# member, on one bucket. Additive means the provider merges this (role,
# member) pair into the bucket's IAM policy without touching any other
# member's bindings on the same role, and destroy subtracts only this pair.
#
# The grant is its own resource for grantees that depend on the bucket (a
# logging sink exporting into it): declared in the bucket's own iam_members,
# such a grant would be a dependency cycle.
#
# Every argument is immutable (ForceNew): an IAM grant has no update -- any
# change replaces it, which is also how the API behaves.
resource "google_storage_bucket_iam_member" "this" {
  bucket = var.spec.bucket
  role   = var.spec.role
  member = var.spec.member

  # An IAM Condition is part of the grant's identity: the same role granted
  # with and without a condition are two independent bindings in the policy.
  dynamic "condition" {
    for_each = var.spec.condition != null ? [var.spec.condition] : []
    content {
      title       = condition.value.title
      expression  = condition.value.expression
      description = condition.value.description != "" ? condition.value.description : null
    }
  }
}
