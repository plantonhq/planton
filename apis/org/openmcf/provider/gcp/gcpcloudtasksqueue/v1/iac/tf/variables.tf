variable "spec" {
  description = "GcpCloudTasksQueue spec"
  type = object({
    project_id    = object({ value = string })
    queue_name    = string
    location      = string
    desired_state = optional(string, "")

    http_target = optional(object({
      http_method = optional(string, "")

      header_overrides = optional(list(object({
        key   = string
        value = string
      })), [])

      oauth_token = optional(object({
        service_account_email = object({ value = string })
        scope                 = optional(string, "")
      }), null)

      oidc_token = optional(object({
        service_account_email = object({ value = string })
        audience              = optional(string, "")
      }), null)

      uri_override = optional(object({
        scheme       = optional(string, "")
        host         = optional(string, "")
        port         = optional(string, "")
        path         = optional(string, "")
        query_params = optional(string, "")
        enforce_mode = optional(string, "")
      }), null)
    }), null)

    rate_limits = optional(object({
      max_dispatches_per_second  = optional(number, 0)
      max_concurrent_dispatches = optional(number, 0)
    }), null)

    retry_config = optional(object({
      max_attempts       = optional(number, 0)
      max_retry_duration = optional(string, "")
      min_backoff        = optional(string, "")
      max_backoff        = optional(string, "")
      max_doublings      = optional(number, 0)
    }), null)

    stackdriver_logging_config = optional(object({
      sampling_ratio = number
    }), null)
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key_base64 = optional(string, "")
  })
  default = { service_account_key_base64 = "" }
}
