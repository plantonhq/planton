locals {
  resource_name = coalesce(try(var.metadata.name, null), "aws-iam-role")
  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  description = try(var.spec.description, null)
  path        = try(var.spec.path, "/")

  # trust_policy is a free-form JSON object (google.protobuf.Struct); aws_iam_role
  # wants assume_role_policy as a JSON string, so encode the object here.
  trust_policy_json = try(jsonencode(var.spec.trust_policy), null)

  managed_policy_arns = try(var.spec.managed_policy_arns, [])

  # inline_policies is free-form JSON (map<string, google.protobuf.Struct>), typed `any` in
  # variables.tf because its entries have heterogeneous shapes. Encode each policy document to a
  # JSON string here so the result is a homogeneous map(string): aws_iam_role_policy.for_each
  # accepts a map/set, and converting a heterogeneous object to a map would otherwise fail with
  # "all map elements must have the same type".
  inline_policies_json = {
    for k, v in try(var.spec.inline_policies, {}) : k => jsonencode(v)
  }
}



