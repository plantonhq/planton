# Amazon Athena workgroup.
#
# One provider resource carries the whole surface; everything interesting
# lives inside the single `configuration` block. Two provider behaviors shape
# the wiring below:
#
#   - Optional nested blocks are presence-driven: the provider treats an
#     absent block as the feature's disabled state (its diff-suppression
#     equates the two), so each spec message maps to a `dynamic` block guarded
#     by a presence local -- never an always-emitted block with zero values,
#     which would pin AWS defaults and create phantom diffs.
#   - Where AWS requires an explicit `enabled` flag INSIDE a block
#     (managed results, the three logging arms), the block's presence in the
#     spec asserts it: enabled is hardcoded true because "present but
#     disabled" and "absent" are the same AWS state.
resource "aws_athena_workgroup" "this" {
  name = local.workgroup_name

  # Empty description and omitted description are the same AWS state; send
  # null so the plan stays clean either way.
  description = var.spec.description != "" ? var.spec.description : null

  # DISABLED rejects new query submissions but keeps configuration, history,
  # and saved queries -- the pause switch, not a teardown.
  state = local.workgroup_state

  # Allows destroy to proceed even when the workgroup still holds named
  # queries or prepared statements.
  force_destroy = var.spec.force_destroy
  tags          = local.tags

  configuration {
    # 0 means "no limit" in the spec; the provider expresses no-limit by
    # omitting the argument, so 0 maps to null.
    bytes_scanned_cutoff_per_query = var.spec.bytes_scanned_cutoff_per_query > 0 ? var.spec.bytes_scanned_cutoff_per_query : null

    # Tri-state dials (proto `optional bool`, spec default true): null falls
    # through to the provider default (also true), so an omitted dial and an
    # explicit true deploy identically while explicit false is representable.
    enforce_workgroup_configuration    = var.spec.enforce_workgroup_configuration
    publish_cloudwatch_metrics_enabled = var.spec.publish_cloudwatch_metrics_enabled

    requester_pays_enabled                  = var.spec.requester_pays_enabled
    enable_minimum_encryption_configuration = var.spec.enable_minimum_encryption_configuration

    # Assumed for Spark workloads and Identity Center-enabled workgroups;
    # plain SQL workgroups leave it unset.
    execution_role = var.spec.execution_role != "" ? var.spec.execution_role : null

    dynamic "engine_version" {
      for_each = local.has_engine_version ? [1] : []
      content {
        selected_engine_version = var.spec.selected_engine_version
      }
    }

    # Customer-managed S3 result storage. Mutually exclusive with managed
    # query results (spec CEL mirrors the provider's own plan-time rule).
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

    # AWS-managed result storage: no bucket to own, 24-hour retention,
    # results retrievable through Athena APIs only.
    dynamic "managed_query_results_configuration" {
      for_each = local.has_managed_results ? [1] : []
      content {
        enabled = coalesce(var.spec.managed_query_results.enabled, true)

        dynamic "encryption_configuration" {
          for_each = var.spec.managed_query_results.kms_key != "" ? [1] : []
          content {
            kms_key = var.spec.managed_query_results.kms_key
          }
        }
      }
    }

    # KMS encryption for Spark notebook content and session data (SQL query
    # results are covered by the result blocks above, not this).
    dynamic "customer_content_encryption_configuration" {
      for_each = local.has_content_encryption ? [1] : []
      content {
        kms_key = var.spec.customer_content_encryption_kms_key
      }
    }

    # IAM Identity Center trusted identity propagation. Create-time settings:
    # changing them replaces the workgroup.
    dynamic "identity_center_configuration" {
      for_each = local.has_identity_center ? [1] : []
      content {
        enable_identity_center       = var.spec.identity_center.enable_identity_center
        identity_center_instance_arn = var.spec.identity_center.identity_center_instance_arn != "" ? var.spec.identity_center.identity_center_instance_arn : null
      }
    }

    # S3 Access Grants credentials for the result location, scoped to the
    # propagated user identity.
    dynamic "query_results_s3_access_grants_configuration" {
      for_each = local.has_access_grants ? [1] : []
      content {
        enable_s3_access_grants  = var.spec.s3_access_grants.enable_s3_access_grants
        authentication_type      = var.spec.s3_access_grants.authentication_type
        create_user_level_prefix = var.spec.s3_access_grants.create_user_level_prefix
      }
    }

    # Log delivery. The three arms are independent and combinable; each is
    # enabled by presence with its required `enabled` flag asserted.
    dynamic "monitoring_configuration" {
      for_each = local.has_monitoring ? [1] : []
      content {
        dynamic "cloud_watch_logging_configuration" {
          for_each = local.has_cw_logging ? [1] : []
          content {
            enabled                = coalesce(var.spec.monitoring.cloud_watch_logging.enabled, true)
            log_group              = var.spec.monitoring.cloud_watch_logging.log_group != "" ? var.spec.monitoring.cloud_watch_logging.log_group : null
            log_stream_name_prefix = var.spec.monitoring.cloud_watch_logging.log_stream_name_prefix != "" ? var.spec.monitoring.cloud_watch_logging.log_stream_name_prefix : null

            # AWS models log selection as a map of log family -> categories
            # (e.g. SPARK -> [DRIVER, EXECUTOR]); the spec's repeated entries
            # map one-to-one onto the provider's log_type set.
            dynamic "log_type" {
              for_each = var.spec.monitoring.cloud_watch_logging.log_types
              content {
                key    = log_type.value.key
                values = log_type.value.values
              }
            }
          }
        }

        dynamic "managed_logging_configuration" {
          for_each = local.has_managed_logging ? [1] : []
          content {
            enabled = coalesce(var.spec.monitoring.managed_logging.enabled, true)
            kms_key = var.spec.monitoring.managed_logging.kms_key != "" ? var.spec.monitoring.managed_logging.kms_key : null
          }
        }

        dynamic "s3_logging_configuration" {
          for_each = local.has_s3_logging ? [1] : []
          content {
            enabled      = coalesce(var.spec.monitoring.s3_logging.enabled, true)
            log_location = var.spec.monitoring.s3_logging.log_location != "" ? var.spec.monitoring.s3_logging.log_location : null
            kms_key      = var.spec.monitoring.s3_logging.kms_key != "" ? var.spec.monitoring.s3_logging.kms_key : null
          }
        }
      }
    }
  }
}
