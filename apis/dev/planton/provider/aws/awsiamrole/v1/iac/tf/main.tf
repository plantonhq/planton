resource "aws_iam_role" "this" {
  name               = local.resource_name
  path               = local.path
  assume_role_policy = local.trust_policy_json
  description        = local.description

  tags = local.tags
}

resource "aws_iam_role_policy_attachment" "managed" {
  for_each   = toset(local.managed_policy_arns)
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

# Always create an instance profile that wraps this role. Instance profiles are
# free and idempotent in AWS, and EC2 requires one (not a bare role) to assume a
# role. Its ARN is what an AwsEc2Instance.iam_instance_profile_arn references.
resource "aws_iam_instance_profile" "this" {
  name = local.resource_name
  role = aws_iam_role.this.name
  tags = local.tags
}



