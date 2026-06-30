locals {
  # Naming and tags
  resource_name = coalesce(try(var.metadata.name, null), coalesce(try(var.spec.function.handler, null), "aws-lambda-function"))
  tags          = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # VPC networking
  safe_subnet_ids = try(var.spec.function.vpc_config.subnet_ids, [])
  safe_security_group_ids = try(var.spec.function.vpc_config.security_group_ids, [])

  # Package type and runtime/handler handling
  package_type = coalesce(try(var.spec.function.package_type, null), "Zip")
  is_image     = local.package_type == "Image"
  runtime      = local.is_image ? null : coalesce(try(var.spec.function.runtime, null), "nodejs18.x")
  handler      = local.is_image ? null : coalesce(try(var.spec.function.handler, null), "index.handler")

  # Code sources
  image_uri        = try(var.spec.function.image_uri, null)
  s3_bucket        = try(var.spec.function.s3_bucket, null)
  s3_key           = try(var.spec.function.s3_key, null)
  s3_object_version = try(var.spec.function.s3_object_version, null)

  # Whether we have enough inputs to create the function
  create_function = local.is_image ? (local.image_uri != null) : (local.s3_bucket != null && local.s3_key != null)

  # KMS for env vars
  kms_key_arn = try(var.spec.function.kms_key_arn, null)

  # Environment variables as single key/value → convert to map(string) if present
  env_var_object = try(var.spec.function.variables, null)
  env_var_map    = local.env_var_object != null ? tomap({ (local.env_var_object.key) = local.env_var_object.value }) : null

  # Layers
  layer_arns = try(var.spec.function.layers, [])

  # CloudWatch log group settings
  log_retention_days = try(var.spec.cloudwatch_log_group.retention_in_days, null)
  log_kms_key_arn    = try(var.spec.cloudwatch_log_group.kms_key_arn, null)

  # IAM role toggles
  enable_lambda_at_edge              = coalesce(try(var.spec.iam_role.lambda_at_edge_enabled, null), false)
  enable_cloudwatch_lambda_insights  = coalesce(try(var.spec.iam_role.cloudwatch_lambda_insights_enabled, null), false)
  custom_policy_arns                 = try(var.spec.iam_role.custom_iam_policy_arns, [])
  ssm_parameter_names                = try(var.spec.iam_role.ssm_parameter_names, [])
  inline_policy_json_raw             = try(var.spec.iam_role.inline_iam_policy, "")
  inline_policy_json                 = length(trimspace(local.inline_policy_json_raw)) > 0 ? trimspace(local.inline_policy_json_raw) : null
}


