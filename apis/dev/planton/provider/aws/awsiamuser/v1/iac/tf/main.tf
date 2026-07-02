# An IAM user is a long-lived identity with permanent credentials -- prefer
# roles wherever temporary credentials work. The user name is mutable (AWS
# renames in place) and so is the path; the permissions boundary caps the
# maximum the user can ever do, which matters most on principals whose
# credentials do not expire.
resource "aws_iam_user" "this" {
  name = var.spec.user_name
  path = var.spec.path != "" ? var.spec.path : null

  # The boundary is a ceiling, not a grant: effective permissions are the
  # intersection of this policy and the user's permission policies.
  permissions_boundary = var.spec.permissions_boundary != "" ? var.spec.permissions_boundary : null

  # When enabled, deletion also removes credentials created OUTSIDE this
  # resource (login profile, extra access keys, MFA devices, SSH keys, signing
  # certs) instead of failing on them.
  force_destroy = var.spec.force_destroy

  tags = local.aws_tags
}

# Each managed-policy attachment is its own resource so attachments reconcile
# individually: adding or removing an entry attaches or detaches just that
# policy, and attachments made outside this resource are left alone.
resource "aws_iam_user_policy_attachment" "managed" {
  for_each   = toset(var.spec.managed_policy_arns)
  user       = aws_iam_user.this.name
  policy_arn = each.value
}

resource "aws_iam_user_policy" "inline" {
  for_each = local.inline_policies_json
  name     = each.key
  user     = aws_iam_user.this.name
  # each.value is already a JSON-encoded policy string (see locals.inline_policies_json).
  policy = each.value
}

# One active access key by default -- programmatic access is the usual reason
# an IAM user exists. The secret lands in state and in the (sensitive) stack
# outputs; no PGP key is used because the platform delivers outputs through
# its own secret-handling channel.
resource "aws_iam_access_key" "this" {
  count = var.spec.disable_access_keys ? 0 : 1
  user  = aws_iam_user.this.name
}
