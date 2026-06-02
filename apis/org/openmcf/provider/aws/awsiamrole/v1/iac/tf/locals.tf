locals {
  resource_name = coalesce(try(var.metadata.name, null), "aws-iam-role")
  tags          = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  description = try(var.spec.description, null)
  path        = try(var.spec.path, "/")

  # trust_policy is a free-form JSON object (google.protobuf.Struct); aws_iam_role
  # wants assume_role_policy as a JSON string, so encode the object here.
  trust_policy_json = try(jsonencode(var.spec.trust_policy), null)

  managed_policy_arns = try(var.spec.managed_policy_arns, [])

  inline_policies_map = try(var.spec.inline_policies, {})
}



