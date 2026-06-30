data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role" "lambda" {
  name               = "${local.resource_name}-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Principal = {
          Service = ["lambda.amazonaws.com"]
        }
      }
    ]
  })

  permissions_boundary = try(var.spec.iam_role.permissions_boundary, null)
  tags                 = local.tags
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "cloudwatch_lambda_insights" {
  count      = local.enable_cloudwatch_lambda_insights ? 1 : 0
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLambdaInsightsExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "custom" {
  for_each  = toset(local.custom_policy_arns)
  role      = aws_iam_role.lambda.name
  policy_arn = each.value
}

resource "aws_iam_role_policy" "inline" {
  count  = local.inline_policy_json != null ? 1 : 0
  name   = "${local.resource_name}-inline"
  role   = aws_iam_role.lambda.id
  policy = coalesce(local.inline_policy_json, jsonencode({
    Version   = "2012-10-17"
    Statement = []
  }))
}

resource "aws_lambda_function" "this" {
  count = 1

  function_name = local.resource_name
  role          = aws_iam_role.lambda.arn

  package_type = local.package_type
  image_uri    = local.is_image ? coalesce(local.image_uri, "111111111111.dkr.ecr.us-east-1.amazonaws.com/dummy:latest") : null
  runtime      = local.runtime
  handler      = local.handler

  architectures = try(var.spec.function.architectures, null)
  publish       = try(var.spec.function.publish, null)
  memory_size   = try(var.spec.function.memory_size, null)
  timeout       = try(var.spec.function.timeout, null)

  kms_key_arn = local.kms_key_arn

  dynamic "environment" {
    for_each = local.env_var_map != null ? [1] : []
    content {
      variables = local.env_var_map
    }
  }

  dynamic "image_config" {
    for_each = local.is_image && try(var.spec.function.image_config, null) != null ? [1] : []
    content {
      command           = try(var.spec.function.image_config.commands, null)
      entry_point       = try(var.spec.function.image_config.entry_points, null)
      working_directory = try(var.spec.function.image_config.working_directory, null)
    }
  }

  dynamic "vpc_config" {
    for_each = length(local.safe_subnet_ids) > 0 || length(local.safe_security_group_ids) > 0 ? [1] : []
    content {
      subnet_ids         = local.safe_subnet_ids
      security_group_ids = local.safe_security_group_ids
    }
  }

  dynamic "file_system_config" {
    for_each = try(var.spec.function.file_system_config.arn, null) != null || try(var.spec.function.file_system_config.local_mount_path, null) != null ? [1] : []
    content {
      arn              = try(var.spec.function.file_system_config.arn, null)
      local_mount_path = try(var.spec.function.file_system_config.local_mount_path, null)
    }
  }

  layers = local.layer_arns

  s3_bucket         = local.is_image ? null : coalesce(local.s3_bucket, "dummy-bucket")
  s3_key            = local.is_image ? null : coalesce(local.s3_key, "dummy.zip")
  s3_object_version = local.is_image ? null : local.s3_object_version

  source_code_hash = try(var.spec.function.source_code_hash, null)

  dynamic "tracing_config" {
    for_each = try(var.spec.function.tracing_config_mode, null) != null ? [1] : []
    content {
      mode = var.spec.function.tracing_config_mode
    }
  }

  dynamic "dead_letter_config" {
    for_each = try(var.spec.function.dead_letter_config_target_arn, null) != null ? [1] : []
    content {
      target_arn = var.spec.function.dead_letter_config_target_arn
    }
  }

  dynamic "ephemeral_storage" {
    for_each = try(var.spec.function.ephemeral_storage_size, null) != null ? [1] : []
    content {
      size = var.spec.function.ephemeral_storage_size
    }
  }

  tags = local.tags
}

resource "aws_cloudwatch_log_group" "lambda" {
  count            = 1
  name             = "/aws/lambda/${aws_lambda_function.this[0].function_name}"
  retention_in_days = local.log_retention_days
  kms_key_id       = local.log_kms_key_arn
  tags             = local.tags
}

resource "aws_lambda_permission" "invoke" {
  for_each      = { for idx, p in coalesce(try(var.spec.invoke_function_permissions, null), []) : idx => p }
  statement_id  = "AllowExecutionFrom-${each.key}"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this[0].arn
  principal     = each.value.principal
  source_arn    = try(each.value.source_arn, null)
}


