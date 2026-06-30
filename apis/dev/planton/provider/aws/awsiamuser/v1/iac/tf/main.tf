resource "aws_iam_user" "this" {
  name = local.resource_name
  tags = local.tags
}

resource "aws_iam_user_policy_attachment" "managed" {
  for_each  = toset(local.managed_policy_arns)
  user      = aws_iam_user.this.name
  policy_arn = each.value
}

resource "aws_iam_user_policy" "inline" {
  for_each = local.inline_policies_map
  name     = each.key
  user     = aws_iam_user.this.name
  policy   = each.value
}

resource "aws_iam_access_key" "this" {
  count   = local.disable_access_keys ? 0 : 1
  user    = aws_iam_user.this.name
  pgp_key = null
}



