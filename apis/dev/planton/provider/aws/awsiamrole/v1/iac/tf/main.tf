# An IAM role is an assumable identity: the trust policy controls WHO can
# assume it, the attached/inline policies control WHAT it can do once assumed,
# and an optional permissions boundary caps the maximum it can ever do. Name
# and path are create-only (changing them replaces the role); everything else
# updates in place.
resource "aws_iam_role" "this" {
  name               = var.metadata.name
  path               = var.spec.path != "" ? var.spec.path : null
  assume_role_policy = local.trust_policy_json
  description        = var.spec.description != "" ? var.spec.description : null

  # 0 means "unset" (proto3 zero value); AWS then applies its 3600s default.
  max_session_duration = var.spec.max_session_duration != 0 ? var.spec.max_session_duration : null

  # The boundary is a ceiling, not a grant: effective permissions are the
  # intersection of this policy and the role's permission policies.
  permissions_boundary = var.spec.permissions_boundary != "" ? var.spec.permissions_boundary : null

  # When enabled, deletion force-detaches policies still attached to the role
  # (including attachments made outside this resource) instead of failing.
  force_detach_policies = var.spec.force_detach_policies

  tags = local.aws_tags
}

# Each managed-policy attachment is its own resource (not the deprecated
# exclusive managed_policy_arns role argument) so attachments reconcile
# individually: adding or removing an entry attaches or detaches just that
# policy, and attachments made outside this resource are left alone.
resource "aws_iam_role_policy_attachment" "managed" {
  for_each   = toset(var.spec.managed_policy_arns)
  role       = aws_iam_role.this.name
  policy_arn = each.value
}

resource "aws_iam_role_policy" "inline" {
  for_each = local.inline_policies_json
  name     = each.key
  role     = aws_iam_role.this.id
  # each.value is already a JSON-encoded policy string (see locals.inline_policies_json).
  policy = each.value
}
