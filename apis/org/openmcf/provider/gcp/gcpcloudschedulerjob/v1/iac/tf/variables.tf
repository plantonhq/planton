variable "spec" {
  description = "GcpCloudSchedulerJob spec"
  type = object({
    project_id = object({ value = string })
    job_name   = optional(string, "")
    location   = string
    schedule   = string
    time_zone  = optional(string, "")
    description      = optional(string, "")
    attempt_deadline = optional(string, "")
    paused           = optional(bool, false)

    http_target = optional(object({
      uri         = string
      http_method = optional(string, "")
      body        = optional(string, "")
      headers     = optional(map(string), {})

      oauth_token = optional(object({
        service_account_email = object({ value = string })
        scope                 = optional(string, "")
      }), null)

      oidc_token = optional(object({
        service_account_email = object({ value = string })
        audience              = optional(string, "")
      }), null)
    }), null)

    pubsub_target = optional(object({
      topic_name = object({ value = string })
      data       = optional(string, "")
      attributes = optional(map(string), {})
    }), null)

    app_engine_http_target = optional(object({
      relative_uri = string
      http_method  = optional(string, "")
      body         = optional(string, "")
      headers      = optional(map(string), {})

      app_engine_routing = optional(object({
        service  = optional(string, "")
        version  = optional(string, "")
        instance = optional(string, "")
      }), null)
    }), null)

    retry_config = optional(object({
      retry_count          = optional(number, 0)
      max_retry_duration   = optional(string, "")
      min_backoff_duration = optional(string, "")
      max_backoff_duration = optional(string, "")
      max_doublings        = optional(number, 0)
    }), null)
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = optional(string, "")
  })
  default = { service_account_key = "" }
}
