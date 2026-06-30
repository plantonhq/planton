variable "metadata" {
  description = "Planton resource metadata"
  type = object({
    name    = string
    id      = optional(string, "")
    org     = optional(string, "")
    env     = optional(string, "")
    labels  = optional(map(string), {})
    tags    = optional(list(string), [])
    version = optional(string, "")
  })
}

variable "spec" {
  description = "GcpCloudArmorPolicy spec"
  type = object({
    project_id  = object({ value = string })
    policy_name = optional(string, "")
    description = optional(string, "")
    type        = optional(string, "")

    adaptive_protection_config = optional(object({
      enable_layer_7_ddos_defense = optional(bool, false)
      rule_visibility             = optional(string, "")
    }), null)

    advanced_options_config = optional(object({
      json_parsing                 = optional(string, "")
      log_level                    = optional(string, "")
      user_ip_request_headers      = optional(list(string), [])
      request_body_inspection_size = optional(string, "")
    }), null)

    rules = optional(list(object({
      action      = string
      priority    = number
      description = optional(string, "")
      preview     = optional(bool, false)

      match = object({
        versioned_expr = optional(string, "")
        src_ip_ranges  = optional(list(string), [])
        expression     = optional(string, "")
      })

      rate_limit_options = optional(object({
        conform_action       = string
        exceed_action        = string
        enforce_on_key       = optional(string, "")
        enforce_on_key_name  = optional(string, "")
        rate_limit_threshold = object({
          count        = number
          interval_sec = number
        })
        ban_threshold = optional(object({
          count        = number
          interval_sec = number
        }), null)
        ban_duration_sec = optional(number, 0)
        exceed_redirect_options = optional(object({
          type   = string
          target = optional(string, "")
        }), null)
      }), null)

      redirect_options = optional(object({
        type   = string
        target = optional(string, "")
      }), null)

      header_action = optional(object({
        request_headers_to_adds = list(object({
          header_name  = string
          header_value = optional(string, "")
        }))
      }), null)

      preconfigured_waf_config = optional(object({
        exclusions = list(object({
          target_rule_set  = string
          target_rule_ids  = optional(list(string), [])
          request_headers = optional(list(object({
            operator = string
            value    = optional(string, "")
          })), [])
          request_cookies = optional(list(object({
            operator = string
            value    = optional(string, "")
          })), [])
          request_uris = optional(list(object({
            operator = string
            value    = optional(string, "")
          })), [])
          request_query_params = optional(list(object({
            operator = string
            value    = optional(string, "")
          })), [])
        }))
      }), null)
    })), [])
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = optional(string, "")
  })
  default = { service_account_key = "" }
}
