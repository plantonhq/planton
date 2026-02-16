# -----------------------------------------------------------------------------
# VPC Connector (created inline when subnet_ids are provided)
# -----------------------------------------------------------------------------
resource "aws_apprunner_vpc_connector" "this" {
  count = local.create_inline_vpc_connector ? 1 : 0

  vpc_connector_name = "${local.service_name}-vpc"
  subnets            = var.spec.subnet_ids
  security_groups    = var.spec.security_group_ids
  tags               = local.tags
}

# -----------------------------------------------------------------------------
# Auto Scaling Configuration (created when auto_scaling block is provided)
# -----------------------------------------------------------------------------
resource "aws_apprunner_auto_scaling_configuration_version" "this" {
  count = var.spec.auto_scaling != null ? 1 : 0

  auto_scaling_configuration_name = "${local.service_name}-asc"
  min_size                        = var.spec.auto_scaling.min_size
  max_size                        = var.spec.auto_scaling.max_size
  max_concurrency                 = var.spec.auto_scaling.max_concurrency
  tags                            = local.tags
}

# -----------------------------------------------------------------------------
# App Runner Service
# -----------------------------------------------------------------------------
resource "aws_apprunner_service" "this" {
  service_name = local.service_name

  # --- Source configuration ---------------------------------------------------
  source_configuration {
    auto_deployments_enabled = var.spec.auto_deployments_enabled

    dynamic "authentication_configuration" {
      for_each = local.needs_auth_config ? [1] : []
      content {
        access_role_arn = var.spec.image_source != null ? (
          var.spec.image_source.access_role_arn != "" ? var.spec.image_source.access_role_arn : null
        ) : null
        connection_arn = var.spec.code_source != null ? var.spec.code_source.connection_arn : null
      }
    }

    # Image-based source
    dynamic "image_repository" {
      for_each = var.spec.image_source != null ? [var.spec.image_source] : []
      content {
        image_identifier      = image_repository.value.image_identifier
        image_repository_type = image_repository.value.image_repository_type

        image_configuration {
          port                          = var.spec.port
          start_command                 = var.spec.start_command != "" ? var.spec.start_command : null
          runtime_environment_variables = length(var.spec.environment_variables) > 0 ? var.spec.environment_variables : null
          runtime_environment_secrets   = length(var.spec.environment_secrets) > 0 ? var.spec.environment_secrets : null
        }
      }
    }

    # Code-based source
    dynamic "code_repository" {
      for_each = var.spec.code_source != null ? [var.spec.code_source] : []
      content {
        repository_url   = code_repository.value.repository_url
        source_directory = code_repository.value.source_directory != "" ? code_repository.value.source_directory : null

        source_code_version {
          type  = "BRANCH"
          value = code_repository.value.branch
        }

        code_configuration {
          configuration_source = code_repository.value.configuration_source

          dynamic "code_configuration_values" {
            for_each = code_repository.value.configuration_source == "API" ? [1] : []
            content {
              runtime                       = code_repository.value.runtime
              build_command                 = code_repository.value.build_command != "" ? code_repository.value.build_command : null
              start_command                 = var.spec.start_command != "" ? var.spec.start_command : null
              port                          = var.spec.port
              runtime_environment_variables = length(var.spec.environment_variables) > 0 ? var.spec.environment_variables : null
              runtime_environment_secrets   = length(var.spec.environment_secrets) > 0 ? var.spec.environment_secrets : null
            }
          }
        }
      }
    }
  }

  # --- Instance configuration ------------------------------------------------
  instance_configuration {
    cpu               = var.spec.cpu
    memory            = var.spec.memory
    instance_role_arn = var.spec.instance_role_arn != "" ? var.spec.instance_role_arn : null
  }

  # --- Health check configuration --------------------------------------------
  dynamic "health_check_configuration" {
    for_each = var.spec.health_check != null ? [var.spec.health_check] : []
    content {
      protocol            = health_check_configuration.value.protocol
      path                = health_check_configuration.value.protocol == "HTTP" ? health_check_configuration.value.path : null
      interval            = health_check_configuration.value.interval_seconds
      timeout             = health_check_configuration.value.timeout_seconds
      healthy_threshold   = health_check_configuration.value.healthy_threshold
      unhealthy_threshold = health_check_configuration.value.unhealthy_threshold
    }
  }

  # --- Network configuration (dynamic: only when non-default) ----------------
  dynamic "network_configuration" {
    for_each = (
      local.egress_type != "DEFAULT" ||
      !var.spec.is_publicly_accessible ||
      var.spec.ip_address_type != "IPV4"
    ) ? [1] : []
    content {
      egress_configuration {
        egress_type       = local.egress_type
        vpc_connector_arn = local.effective_vpc_connector_arn
      }
      ingress_configuration {
        is_publicly_accessible = var.spec.is_publicly_accessible
      }
      ip_address_type = var.spec.ip_address_type
    }
  }

  # --- Encryption configuration (dynamic: only when KMS key provided) --------
  dynamic "encryption_configuration" {
    for_each = var.spec.kms_key_arn != "" ? [1] : []
    content {
      kms_key = var.spec.kms_key_arn
    }
  }

  # --- Observability configuration (dynamic: only when enabled) ---------------
  dynamic "observability_configuration" {
    for_each = var.spec.observability_enabled ? [1] : []
    content {
      observability_enabled           = true
      observability_configuration_arn = var.spec.observability_configuration_arn
    }
  }

  # --- Auto scaling -----------------------------------------------------------
  auto_scaling_configuration_arn = (
    var.spec.auto_scaling != null
    ? aws_apprunner_auto_scaling_configuration_version.this[0].arn
    : null
  )

  tags = local.tags
}
