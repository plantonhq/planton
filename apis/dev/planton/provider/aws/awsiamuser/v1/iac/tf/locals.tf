locals {
  resource_name = coalesce(try(var.spec.user_name, null), try(var.metadata.name, null), "aws-iam-user")
  tags          = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  managed_policy_arns = try(var.spec.managed_policy_arns, [])
  inline_policies_map = try(var.spec.inline_policies, {})

  disable_access_keys = try(var.spec.disable_access_keys, false)
}



