resource "google_cloud_tasks_queue" "this" {
  name     = local.queue_name
  location = local.location
  project  = local.project_id

  dynamic "http_target" {
    for_each = var.spec.http_target != null ? [var.spec.http_target] : []
    content {
      http_method = http_target.value.http_method != "" ? http_target.value.http_method : null

      dynamic "header_overrides" {
        for_each = http_target.value.header_overrides
        content {
          header {
            key   = header_overrides.value.key
            value = header_overrides.value.value
          }
        }
      }

      dynamic "oauth_token" {
        for_each = http_target.value.oauth_token != null ? [http_target.value.oauth_token] : []
        content {
          service_account_email = oauth_token.value.service_account_email.value
          scope                 = oauth_token.value.scope != "" ? oauth_token.value.scope : null
        }
      }

      dynamic "oidc_token" {
        for_each = http_target.value.oidc_token != null ? [http_target.value.oidc_token] : []
        content {
          service_account_email = oidc_token.value.service_account_email.value
          audience              = oidc_token.value.audience != "" ? oidc_token.value.audience : null
        }
      }

      dynamic "uri_override" {
        for_each = http_target.value.uri_override != null ? [http_target.value.uri_override] : []
        content {
          scheme                   = uri_override.value.scheme != "" ? uri_override.value.scheme : null
          host                     = uri_override.value.host != "" ? uri_override.value.host : null
          port                     = uri_override.value.port != "" ? uri_override.value.port : null
          uri_override_enforce_mode = uri_override.value.enforce_mode != "" ? uri_override.value.enforce_mode : null

          dynamic "path_override" {
            for_each = uri_override.value.path != "" ? [uri_override.value.path] : []
            content {
              path = path_override.value
            }
          }

          dynamic "query_override" {
            for_each = uri_override.value.query_params != "" ? [uri_override.value.query_params] : []
            content {
              query_params = query_override.value
            }
          }
        }
      }
    }
  }

  dynamic "rate_limits" {
    for_each = var.spec.rate_limits != null ? [var.spec.rate_limits] : []
    content {
      max_dispatches_per_second  = rate_limits.value.max_dispatches_per_second > 0 ? rate_limits.value.max_dispatches_per_second : null
      max_concurrent_dispatches = rate_limits.value.max_concurrent_dispatches > 0 ? rate_limits.value.max_concurrent_dispatches : null
    }
  }

  dynamic "retry_config" {
    for_each = var.spec.retry_config != null ? [var.spec.retry_config] : []
    content {
      max_attempts       = retry_config.value.max_attempts != 0 ? retry_config.value.max_attempts : null
      max_retry_duration = retry_config.value.max_retry_duration != "" ? retry_config.value.max_retry_duration : null
      min_backoff        = retry_config.value.min_backoff != "" ? retry_config.value.min_backoff : null
      max_backoff        = retry_config.value.max_backoff != "" ? retry_config.value.max_backoff : null
      max_doublings      = retry_config.value.max_doublings != 0 ? retry_config.value.max_doublings : null
    }
  }

  dynamic "stackdriver_logging_config" {
    for_each = var.spec.stackdriver_logging_config != null ? [var.spec.stackdriver_logging_config] : []
    content {
      sampling_ratio = stackdriver_logging_config.value.sampling_ratio
    }
  }
}
