resource "alicloud_oss_bucket" "main" {
  bucket          = var.spec.bucket_name
  acl             = var.spec.acl
  storage_class   = var.spec.storage_class
  redundancy_type = var.spec.redundancy_type
  force_destroy   = var.spec.force_destroy
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags            = local.final_tags

  dynamic "versioning" {
    for_each = var.spec.versioning_enabled ? [1] : []
    content {
      status = "Enabled"
    }
  }

  dynamic "server_side_encryption_rule" {
    for_each = var.spec.server_side_encryption != null ? [var.spec.server_side_encryption] : []
    content {
      sse_algorithm      = server_side_encryption_rule.value.sse_algorithm
      kms_master_key_id  = server_side_encryption_rule.value.kms_master_key_id != "" ? server_side_encryption_rule.value.kms_master_key_id : null
    }
  }

  dynamic "lifecycle_rule" {
    for_each = var.spec.lifecycle_rules
    content {
      id      = "rule-${lifecycle_rule.key}"
      prefix  = lifecycle_rule.value.prefix
      enabled = lifecycle_rule.value.enabled

      dynamic "expiration" {
        for_each = lifecycle_rule.value.expiration_days > 0 ? [lifecycle_rule.value.expiration_days] : []
        content {
          days = expiration.value
        }
      }

      dynamic "transitions" {
        for_each = lifecycle_rule.value.transitions
        content {
          days          = transitions.value.days
          storage_class = transitions.value.storage_class
        }
      }

      dynamic "abort_multipart_upload" {
        for_each = lifecycle_rule.value.abort_multipart_upload_days > 0 ? [lifecycle_rule.value.abort_multipart_upload_days] : []
        content {
          days = abort_multipart_upload.value
        }
      }

      dynamic "noncurrent_version_expiration" {
        for_each = lifecycle_rule.value.noncurrent_version_expiration_days > 0 ? [lifecycle_rule.value.noncurrent_version_expiration_days] : []
        content {
          days = noncurrent_version_expiration.value
        }
      }
    }
  }

  dynamic "cors_rule" {
    for_each = var.spec.cors_rules
    content {
      allowed_origins = cors_rule.value.allowed_origins
      allowed_methods = cors_rule.value.allowed_methods
      allowed_headers = cors_rule.value.allowed_headers
      expose_headers  = cors_rule.value.expose_headers
      max_age_seconds = cors_rule.value.max_age_seconds > 0 ? cors_rule.value.max_age_seconds : null
    }
  }

  dynamic "logging" {
    for_each = var.spec.logging != null ? [var.spec.logging] : []
    content {
      target_bucket = logging.value.target_bucket
      target_prefix = logging.value.target_prefix != "" ? logging.value.target_prefix : null
    }
  }
}
