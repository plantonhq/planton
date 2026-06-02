resource "aws_iam_role" "this" {
  name               = local.resource_name
  path               = local.path
  assume_role_policy = local.trust_policy_json
  description        = local.description

  tags = local.tags
}

resource "aws_iam_role_policy_attachment" "managed" {
  for_each = toset(local.managed_policy_arns)
  role     = aws_iam_role.this.name
  policy_arn = each.value
}

resource "aws_iam_role_policy" "inline" {
  for_each = local.inline_policies_map
  name     = each.key
  role     = aws_iam_role.this.id
  # each.value is a free-form JSON object (google.protobuf.Struct); aws_iam_role_policy
  # wants policy as a JSON string.
  policy = jsonencode(each.value)
}



