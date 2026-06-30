#############################################
# GCS Bucket
#############################################

resource "google_storage_bucket" "main" {
  name                        = var.spec.bucket_name
  project                     = var.spec.gcp_project_id.value
  location                    = var.spec.location
  labels                      = local.final_labels
  force_destroy               = true
  uniform_bucket_level_access = var.spec.uniform_bucket_level_access_enabled
  storage_class               = var.spec.storage_class
  requester_pays              = var.spec.requester_pays
  public_access_prevention    = var.spec.public_access_prevention

  # Versioning configuration
  dynamic "versioning" {
    for_each = var.spec.versioning_enabled ? [1] : []
    content {
      enabled = true
    }
  }

  # Lifecycle rules
  dynamic "lifecycle_rule" {
    for_each = var.spec.lifecycle_rules
    content {
      action {
        type          = lifecycle_rule.value.action.type
        storage_class = lifecycle_rule.value.action.storage_class
      }
      condition {
        age                   = lifecycle_rule.value.condition.age_days
        created_before        = lifecycle_rule.value.condition.created_before
        with_state            = lifecycle_rule.value.condition.is_live != null ? (lifecycle_rule.value.condition.is_live ? "LIVE" : "ARCHIVED") : null
        num_newer_versions    = lifecycle_rule.value.condition.num_newer_versions
        matches_storage_class = lifecycle_rule.value.condition.matches_storage_class
      }
    }
  }

  # Encryption configuration (CMEK)
  dynamic "encryption" {
    for_each = var.spec.encryption != null ? [var.spec.encryption] : []
    content {
      default_kms_key_name = encryption.value.kms_key_name
    }
  }

  # CORS configuration
  dynamic "cors" {
    for_each = var.spec.cors_rules
    content {
      method          = cors.value.methods
      origin          = cors.value.origins
      response_header = cors.value.response_headers
      max_age_seconds = cors.value.max_age_seconds
    }
  }

  # Website configuration
  dynamic "website" {
    for_each = var.spec.website != null ? [var.spec.website] : []
    content {
      main_page_suffix = website.value.main_page_suffix
      not_found_page   = website.value.not_found_page
    }
  }

  # Retention policy (WORM compliance)
  dynamic "retention_policy" {
    for_each = var.spec.retention_policy != null ? [var.spec.retention_policy] : []
    content {
      retention_period = retention_policy.value.retention_period_seconds
      is_locked        = retention_policy.value.is_locked
    }
  }

  # Logging configuration
  dynamic "logging" {
    for_each = var.spec.logging != null ? [var.spec.logging] : []
    content {
      log_bucket        = logging.value.log_bucket
      log_object_prefix = logging.value.log_object_prefix
    }
  }
}

#############################################
# IAM Bindings
#############################################

resource "google_storage_bucket_iam_binding" "bindings" {
  for_each = { for idx, binding in var.spec.iam_bindings : idx => binding }

  bucket  = google_storage_bucket.main.name
  role    = each.value.role
  members = each.value.members

  dynamic "condition" {
    for_each = each.value.condition != null ? [each.value.condition] : []
    content {
      expression = condition.value
      title      = "condition-${each.key}"
    }
  }
}
