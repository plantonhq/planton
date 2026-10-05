variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareAiGateway specification"
  type = object({
    account_id                 = string
    gateway_id                 = string
    cache_invalidate_on_update = bool
    cache_ttl                  = number
    collect_logs               = bool
    rate_limiting_interval     = number
    rate_limiting_limit        = number
    rate_limiting_technique    = optional(string, "")
    retry = optional(object({
      backoff      = optional(string, "")
      delay        = optional(number)
      max_attempts = optional(number)
    }))
    log_management = optional(object({
      max_records = optional(number)
      strategy    = optional(string, "")
    }))
    authentication          = optional(bool)
    logpush                 = optional(bool)
    logpush_public_key      = optional(string, "")
    zdr                     = optional(bool)
    workers_ai_billing_mode = optional(string, "")
    store_id                = optional(string, "")
    dlp = optional(object({
      enabled  = bool
      action   = optional(string, "")
      profiles = optional(list(string), [])
      policies = optional(list(object({
        id       = string
        enabled  = bool
        action   = string
        check    = list(string)
        profiles = list(string)
      })), [])
    }))
    guardrails = optional(object({
      prompt = object({
        p1  = optional(string, "")
        s1  = optional(string, "")
        s2  = optional(string, "")
        s3  = optional(string, "")
        s4  = optional(string, "")
        s5  = optional(string, "")
        s6  = optional(string, "")
        s7  = optional(string, "")
        s8  = optional(string, "")
        s9  = optional(string, "")
        s10 = optional(string, "")
        s11 = optional(string, "")
        s12 = optional(string, "")
        s13 = optional(string, "")
      })
      response = object({
        p1  = optional(string, "")
        s1  = optional(string, "")
        s2  = optional(string, "")
        s3  = optional(string, "")
        s4  = optional(string, "")
        s5  = optional(string, "")
        s6  = optional(string, "")
        s7  = optional(string, "")
        s8  = optional(string, "")
        s9  = optional(string, "")
        s10 = optional(string, "")
        s11 = optional(string, "")
        s12 = optional(string, "")
        s13 = optional(string, "")
      })
    }))
    otel = optional(list(object({
      url           = string
      headers       = optional(map(string), {})
      authorization = optional(string, "")
      content_type  = optional(string, "")
    })), [])
    stripe = optional(object({
      authorization = string
      usage_events = list(object({
        payload = string
      }))
    }))
    spend_limits = optional(object({
      enabled = optional(bool)
      rules = optional(list(object({
        id         = string
        enabled    = optional(bool)
        limit      = number
        limit_type = string
        window     = number
        technique  = optional(string, "")
        metadata = optional(map(object({
          mode   = string
          values = optional(list(string), [])
        })), {})
        model = optional(object({
          mode   = string
          values = list(string)
        }))
        provider = optional(object({
          mode   = string
          values = list(string)
        }))
      })), [])
    }))
    dynamic_routes = optional(list(object({
      name = string
      elements = list(object({
        id   = string
        type = string
        outputs = object({
          next = optional(object({
            element_id = string
          }))
          on_true = optional(object({
            element_id = string
          }))
          on_false = optional(object({
            element_id = string
          }))
          success = optional(object({
            element_id = string
          }))
          fallback = optional(object({
            element_id = string
          }))
          element_id = optional(string, "")
        })
        properties = optional(object({
          conditions = optional(string, "")
          key        = optional(string, "")
          limit      = optional(number)
          limit_type = optional(string, "")
          window     = optional(number)
          model      = optional(string, "")
          provider   = optional(string, "")
          retries    = optional(number)
          timeout    = optional(number)
        }))
      }))
    })), [])
  })
}
