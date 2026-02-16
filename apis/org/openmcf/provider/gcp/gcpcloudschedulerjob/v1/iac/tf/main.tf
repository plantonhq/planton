resource "google_cloud_scheduler_job" "this" {
  name     = local.job_name
  region   = local.location
  project  = local.project_id
  schedule = var.spec.schedule

  time_zone        = var.spec.time_zone != "" ? var.spec.time_zone : null
  description      = var.spec.description != "" ? var.spec.description : null
  attempt_deadline = var.spec.attempt_deadline != "" ? var.spec.attempt_deadline : null
  paused           = var.spec.paused ? true : null

  dynamic "http_target" {
    for_each = var.spec.http_target != null ? [var.spec.http_target] : []
    content {
      uri         = http_target.value.uri
      http_method = http_target.value.http_method != "" ? http_target.value.http_method : null
      body        = http_target.value.body != "" ? http_target.value.body : null
      headers     = length(http_target.value.headers) > 0 ? http_target.value.headers : null

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
    }
  }

  dynamic "pubsub_target" {
    for_each = var.spec.pubsub_target != null ? [var.spec.pubsub_target] : []
    content {
      topic_name = pubsub_target.value.topic_name.value
      data       = pubsub_target.value.data != "" ? pubsub_target.value.data : null
      attributes = length(pubsub_target.value.attributes) > 0 ? pubsub_target.value.attributes : null
    }
  }

  dynamic "app_engine_http_target" {
    for_each = var.spec.app_engine_http_target != null ? [var.spec.app_engine_http_target] : []
    content {
      relative_uri = app_engine_http_target.value.relative_uri
      http_method  = app_engine_http_target.value.http_method != "" ? app_engine_http_target.value.http_method : null
      body         = app_engine_http_target.value.body != "" ? app_engine_http_target.value.body : null
      headers      = length(app_engine_http_target.value.headers) > 0 ? app_engine_http_target.value.headers : null

      dynamic "app_engine_routing" {
        for_each = app_engine_http_target.value.app_engine_routing != null ? [app_engine_http_target.value.app_engine_routing] : []
        content {
          service  = app_engine_routing.value.service != "" ? app_engine_routing.value.service : null
          version  = app_engine_routing.value.version != "" ? app_engine_routing.value.version : null
          instance = app_engine_routing.value.instance != "" ? app_engine_routing.value.instance : null
        }
      }
    }
  }

  dynamic "retry_config" {
    for_each = var.spec.retry_config != null ? [var.spec.retry_config] : []
    content {
      retry_count          = retry_config.value.retry_count != 0 ? retry_config.value.retry_count : null
      max_retry_duration   = retry_config.value.max_retry_duration != "" ? retry_config.value.max_retry_duration : null
      min_backoff_duration = retry_config.value.min_backoff_duration != "" ? retry_config.value.min_backoff_duration : null
      max_backoff_duration = retry_config.value.max_backoff_duration != "" ? retry_config.value.max_backoff_duration : null
      max_doublings        = retry_config.value.max_doublings != 0 ? retry_config.value.max_doublings : null
    }
  }
}
