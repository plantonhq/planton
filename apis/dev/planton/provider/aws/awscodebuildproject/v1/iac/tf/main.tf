# --- CodeBuild Project ---

resource "aws_codebuild_project" "this" {
  name                   = var.metadata.id
  description            = var.spec.description != "" ? var.spec.description : null
  service_role           = var.spec.service_role.value
  build_timeout          = var.spec.build_timeout
  queued_timeout         = var.spec.queued_timeout
  concurrent_build_limit = var.spec.concurrent_build_limit > 0 ? var.spec.concurrent_build_limit : null
  encryption_key         = try(var.spec.encryption_key.value, null)
  source_version         = var.spec.source_version != "" ? var.spec.source_version : null
  tags                   = local.tags

  source {
    type            = var.spec.source.type
    location        = var.spec.source.location != "" ? var.spec.source.location : null
    buildspec       = var.spec.source.buildspec != "" ? var.spec.source.buildspec : null
    git_clone_depth = var.spec.source.git_clone_depth > 0 ? var.spec.source.git_clone_depth : null

    report_build_status = var.spec.source.report_build_status

    dynamic "git_submodules_config" {
      for_each = var.spec.source.fetch_submodules ? [1] : []
      content {
        fetch_submodules = true
      }
    }
  }

  environment {
    type                        = var.spec.environment.type
    compute_type                = var.spec.environment.compute_type
    image                       = var.spec.environment.image
    privileged_mode             = var.spec.environment.privileged_mode
    image_pull_credentials_type = var.spec.environment.image_pull_credentials_type

    dynamic "environment_variable" {
      for_each = var.spec.environment.environment_variables
      content {
        name  = environment_variable.value.name
        value = environment_variable.value.value
        type  = environment_variable.value.type
      }
    }

    dynamic "registry_credential" {
      for_each = var.spec.environment.registry_credential != null ? [var.spec.environment.registry_credential] : []
      content {
        credential          = registry_credential.value.credential
        credential_provider = registry_credential.value.credential_provider
      }
    }
  }

  artifacts {
    type                = var.spec.artifacts.type
    location            = try(var.spec.artifacts.location.value, null)
    name                = var.spec.artifacts.name != "" ? var.spec.artifacts.name : null
    path                = var.spec.artifacts.path != "" ? var.spec.artifacts.path : null
    packaging           = var.spec.artifacts.packaging != "" ? var.spec.artifacts.packaging : null
    namespace_type      = var.spec.artifacts.namespace_type != "" ? var.spec.artifacts.namespace_type : null
    encryption_disabled = var.spec.artifacts.encryption_disabled
  }

  dynamic "cache" {
    for_each = local.has_cache ? [var.spec.cache] : []
    content {
      type     = cache.value.type
      location = try(cache.value.location.value, null)
      modes    = cache.value.type == "LOCAL" ? cache.value.modes : null
    }
  }

  dynamic "logs_config" {
    for_each = local.has_logs_config ? [var.spec.logs_config] : []
    content {
      dynamic "cloudwatch_logs" {
        for_each = logs_config.value.cloudwatch_logs != null ? [logs_config.value.cloudwatch_logs] : []
        content {
          status      = cloudwatch_logs.value.status
          group_name  = try(cloudwatch_logs.value.group_name.value, null)
          stream_name = cloudwatch_logs.value.stream_name != "" ? cloudwatch_logs.value.stream_name : null
        }
      }
      dynamic "s3_logs" {
        for_each = logs_config.value.s3_logs != null ? [logs_config.value.s3_logs] : []
        content {
          status              = s3_logs.value.status
          location            = try(s3_logs.value.location.value, null)
          encryption_disabled = s3_logs.value.encryption_disabled
        }
      }
    }
  }

  dynamic "vpc_config" {
    for_each = local.has_vpc_config ? [var.spec.vpc_config] : []
    content {
      vpc_id             = vpc_config.value.vpc_id.value
      subnets            = [for s in vpc_config.value.subnet_ids : s.value]
      security_group_ids = [for sg in vpc_config.value.security_group_ids : sg.value]
    }
  }
}

# --- Webhook (optional) ---

resource "aws_codebuild_webhook" "this" {
  count        = local.has_webhook ? 1 : 0
  project_name = aws_codebuild_project.this.name

  build_type = var.spec.webhook.build_type != "" ? var.spec.webhook.build_type : null

  dynamic "filter_group" {
    for_each = try(var.spec.webhook.filter_groups, [])
    content {
      dynamic "filter" {
        for_each = filter_group.value.filters
        content {
          type                    = filter.value.type
          pattern                 = filter.value.pattern
          exclude_matched_pattern = filter.value.exclude_matched_pattern
        }
      }
    }
  }
}

# --- Outputs ---

output "project_arn" {
  description = "ARN of the CodeBuild project"
  value       = aws_codebuild_project.this.arn
}

output "project_name" {
  description = "Name of the CodeBuild project"
  value       = aws_codebuild_project.this.name
}

output "service_role_arn" {
  description = "IAM service role ARN used by the project"
  value       = aws_codebuild_project.this.service_role
}

output "webhook_url" {
  description = "Webhook URL (empty if no webhook)"
  value       = local.has_webhook ? aws_codebuild_webhook.this[0].url : ""
}

output "webhook_payload_url" {
  description = "Webhook payload URL (empty if no webhook)"
  value       = local.has_webhook ? aws_codebuild_webhook.this[0].payload_url : ""
}
