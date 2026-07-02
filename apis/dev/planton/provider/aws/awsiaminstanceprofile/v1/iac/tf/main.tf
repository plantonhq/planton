# An instance profile is the container that delivers an IAM role to EC2:
# instances cannot assume a role directly, they can only be launched with a
# profile that carries one. The profile holds exactly one role (an AWS limit,
# not a provider choice). Name and path are create-only; the role can be
# swapped in place -- AWS removes the old role and adds the new one without
# replacing the profile, so running instances pick up the new role's
# credentials on their next metadata refresh.
#
# The role is attached by NAME (not ARN) -- that is what the underlying
# AddRoleToInstanceProfile API takes. IAM is eventually consistent; the
# provider retries the role attach internally until the freshly-created role
# is visible.
resource "aws_iam_instance_profile" "this" {
  name = var.metadata.name
  path = var.spec.path != "" ? var.spec.path : null
  role = var.spec.role

  tags = local.aws_tags
}
