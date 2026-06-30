resource "aws_athena_workgroup" "this" {
  name          = local.workgroup_name
  description   = ""
  state         = "ENABLED"
  force_destroy = var.spec.force_destroy
  tags          = local.tags

  configuration {
    bytes_scanned_cutoff_per_query = var.spec.bytes_scanned_cutoff_per_query > 0 ? var.spec.bytes_scanned_cutoff_per_query : null

    enforce_workgroup_configuration         = var.spec.enforce_workgroup_configuration
    publish_cloudwatch_metrics_enabled      = var.spec.publish_cloudwatch_metrics_enabled
    requester_pays_enabled                  = var.spec.requester_pays_enabled
    enable_minimum_encryption_configuration = var.spec.enable_minimum_encryption_configuration

    execution_role = var.spec.execution_role != "" ? var.spec.execution_role : null

    dynamic "engine_version" {
      for_each = local.has_engine_version ? [1] : []
      content {
        selected_engine_version = var.spec.selected_engine_version
      }
    }

    dynamic "result_configuration" {
      for_each = local.has_result_config ? [1] : []
      content {
        output_location       = var.spec.result_configuration.output_location != "" ? var.spec.result_configuration.output_location : null
        expected_bucket_owner = var.spec.result_configuration.expected_bucket_owner != "" ? var.spec.result_configuration.expected_bucket_owner : null

        dynamic "encryption_configuration" {
          for_each = local.has_encryption ? [1] : []
          content {
            encryption_option = var.spec.result_configuration.encryption_option
            kms_key_arn       = var.spec.result_configuration.kms_key_arn != "" ? var.spec.result_configuration.kms_key_arn : null
          }
        }

        dynamic "acl_configuration" {
          for_each = local.has_acl ? [1] : []
          content {
            s3_acl_option = var.spec.result_configuration.s3_acl_option
          }
        }
      }
    }
  }
}
