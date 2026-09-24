# secret_environment values are kept in AWS Secrets Manager, one secret per
# variable named "<family>/<container>/<name>", so the task definition
# carries the secret's ARN and never the value. Matches the Pulumi module:
#   - the resource policy lets only the execution role read the secret (the
#     ECS agent fetches it as that role at task start), so the author's role
#     needs no edit;
#   - the recovery window is zero: the copy is derived from whoever supplied
#     the value, and a scheduled deletion would reserve the name and fail the
#     next deploy of the same family;
#   - valueFrom pins the version id ("<arn>:::<version-id>"), so a changed
#     value registers a new revision and a referencing service rolls.

locals {
  # Keyed "<container>/<name>" -- the address the container definitions in
  # locals.tf look each variable up by.
  secret_environment = merge([
    for container in var.spec.containers : {
      for name, value in container.secret_environment : "${container.name}/${name}" => {
        secret_name = "${local.family}/${container.name}/${name}"
        container   = container.name
        name        = name
        value       = value
      }
    }
  ]...)

  # Per container, name -> pinned valueFrom, merged over the author's own
  # secrets map when the container definitions are rendered.
  stored_secret_value_from = {
    for container in var.spec.containers : container.name => {
      for name, value in container.secret_environment :
      name => "${aws_secretsmanager_secret.env["${container.name}/${name}"].arn}:::${aws_secretsmanager_secret_version.env["${container.name}/${name}"].version_id}"
    }
  }
}

resource "aws_secretsmanager_secret" "env" {
  for_each = local.secret_environment

  name                    = each.value.secret_name
  description             = "Value of ${each.value.name} for container ${each.value.container} of ECS task definition family ${local.family}"
  recovery_window_in_days = 0
  tags                    = local.aws_tags

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "ExecutionRoleReadsAtTaskStart"
      Effect    = "Allow"
      Principal = { AWS = local.execution_role_arn }
      Action    = "secretsmanager:GetSecretValue"
      Resource  = "*"
    }]
  })

  lifecycle {
    precondition {
      condition     = can(regex("^[A-Za-z0-9/_+=.@-]{1,512}$", each.value.secret_name))
      error_message = "secret_environment variable ${each.value.name} in container ${each.value.container} cannot be stored: Secrets Manager secret ${each.value.secret_name} may hold only letters, digits and /_+=.@- and at most 512 characters -- rename the variable."
    }
  }
}

resource "aws_secretsmanager_secret_version" "env" {
  for_each = local.secret_environment

  secret_id     = aws_secretsmanager_secret.env[each.key].id
  secret_string = each.value.value
}
