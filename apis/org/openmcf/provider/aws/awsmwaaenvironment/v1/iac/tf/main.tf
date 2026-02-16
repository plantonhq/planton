# --- Managed Security Group ---

resource "aws_security_group" "environment" {
  count = local.has_ingress_refs ? 1 : 0

  name        = var.metadata.id
  description = "Managed security group for MWAA environment"
  vpc_id      = var.spec.vpc_id
  tags        = local.tags
}

# Self-referencing inbound rule: MWAA VPC endpoints must communicate with each other
resource "aws_security_group_rule" "ingress_self" {
  count = local.has_ingress_refs ? 1 : 0

  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  source_security_group_id = aws_security_group.environment[0].id
  security_group_id        = aws_security_group.environment[0].id
}

# HTTPS (443) from source security groups for Airflow UI access
resource "aws_security_group_rule" "ingress_https_sg" {
  for_each = local.has_ingress_refs ? toset(var.spec.security_group_ids) : toset([])

  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  source_security_group_id = each.value
  security_group_id        = aws_security_group.environment[0].id
}

# HTTPS (443) from CIDR blocks
resource "aws_security_group_rule" "ingress_https_cidr" {
  count = local.has_ingress_refs && length(var.spec.allowed_cidr_blocks) > 0 ? 1 : 0

  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = var.spec.allowed_cidr_blocks
  security_group_id = aws_security_group.environment[0].id
}

# Allow all outbound traffic
resource "aws_security_group_rule" "egress_all" {
  count = local.has_ingress_refs ? 1 : 0

  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.environment[0].id
}

# --- MWAA Environment ---

resource "aws_mwaa_environment" "this" {
  name               = var.metadata.id
  dag_s3_path        = var.spec.dag_s3_path
  execution_role_arn = var.spec.execution_role_arn
  source_bucket_arn  = var.spec.source_bucket_arn
  tags               = local.tags

  network_configuration {
    subnet_ids         = var.spec.subnet_ids
    security_group_ids = local.effective_security_group_ids
  }

  # Airflow version
  airflow_version = var.spec.airflow_version != "" ? var.spec.airflow_version : null

  # Airflow config overrides
  airflow_configuration_options = length(var.spec.airflow_configuration_options) > 0 ? var.spec.airflow_configuration_options : null

  # S3 artifacts
  plugins_s3_path                  = var.spec.plugins_s3_path != "" ? var.spec.plugins_s3_path : null
  plugins_s3_object_version        = var.spec.plugins_s3_object_version != "" ? var.spec.plugins_s3_object_version : null
  requirements_s3_path             = var.spec.requirements_s3_path != "" ? var.spec.requirements_s3_path : null
  requirements_s3_object_version   = var.spec.requirements_s3_object_version != "" ? var.spec.requirements_s3_object_version : null
  startup_script_s3_path           = var.spec.startup_script_s3_path != "" ? var.spec.startup_script_s3_path : null
  startup_script_s3_object_version = var.spec.startup_script_s3_object_version != "" ? var.spec.startup_script_s3_object_version : null

  # KMS encryption
  kms_key = var.spec.kms_key_arn != "" ? var.spec.kms_key_arn : null

  # Environment sizing
  environment_class = var.spec.environment_class != "" ? var.spec.environment_class : null
  min_workers       = var.spec.min_workers > 0 ? var.spec.min_workers : null
  max_workers       = var.spec.max_workers > 0 ? var.spec.max_workers : null
  min_webservers    = var.spec.min_webservers > 0 ? var.spec.min_webservers : null
  max_webservers    = var.spec.max_webservers > 0 ? var.spec.max_webservers : null
  schedulers        = var.spec.schedulers > 0 ? var.spec.schedulers : null

  # Access and networking
  webserver_access_mode = var.spec.webserver_access_mode
  endpoint_management   = var.spec.endpoint_management != "" ? var.spec.endpoint_management : null

  # Maintenance
  weekly_maintenance_window_start = var.spec.weekly_maintenance_window_start != "" ? var.spec.weekly_maintenance_window_start : null

  # Worker replacement strategy
  worker_replacement_strategy = var.spec.worker_replacement_strategy != "" ? var.spec.worker_replacement_strategy : null

  # Logging configuration
  dynamic "logging_configuration" {
    for_each = var.spec.logging_configuration != null ? [var.spec.logging_configuration] : []
    content {
      dynamic "dag_processing_logs" {
        for_each = logging_configuration.value.dag_processing_logs != null ? [logging_configuration.value.dag_processing_logs] : []
        content {
          enabled   = dag_processing_logs.value.enabled
          log_level = dag_processing_logs.value.log_level
        }
      }

      dynamic "scheduler_logs" {
        for_each = logging_configuration.value.scheduler_logs != null ? [logging_configuration.value.scheduler_logs] : []
        content {
          enabled   = scheduler_logs.value.enabled
          log_level = scheduler_logs.value.log_level
        }
      }

      dynamic "task_logs" {
        for_each = logging_configuration.value.task_logs != null ? [logging_configuration.value.task_logs] : []
        content {
          enabled   = task_logs.value.enabled
          log_level = task_logs.value.log_level
        }
      }

      dynamic "webserver_logs" {
        for_each = logging_configuration.value.webserver_logs != null ? [logging_configuration.value.webserver_logs] : []
        content {
          enabled   = webserver_logs.value.enabled
          log_level = webserver_logs.value.log_level
        }
      }

      dynamic "worker_logs" {
        for_each = logging_configuration.value.worker_logs != null ? [logging_configuration.value.worker_logs] : []
        content {
          enabled   = worker_logs.value.enabled
          log_level = worker_logs.value.log_level
        }
      }
    }
  }
}
