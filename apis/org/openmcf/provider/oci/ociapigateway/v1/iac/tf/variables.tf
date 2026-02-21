variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "OciApiGateway specification"
  type = object({
    compartment_id = object({
      value = string
    })

    endpoint_type = string

    subnet_id = object({
      value = string
    })

    display_name = optional(string, "")

    certificate_id = optional(string, "")

    network_security_group_ids = optional(list(object({
      value = string
    })), [])

    deployment = object({
      path_prefix  = string
      display_name = optional(string, "")

      logging_policies = optional(object({
        access_log = optional(object({
          is_enabled = bool
        }), null)
        execution_log = optional(object({
          is_enabled = bool
          log_level  = optional(string, "")
        }), null)
      }), null)

      request_policies = optional(object({
        authentication = optional(object({
          issuers                     = optional(list(string), [])
          audiences                   = optional(list(string), [])
          token_header                = optional(string, "")
          token_query_param           = optional(string, "")
          token_auth_scheme           = optional(string, "")
          max_clock_skew_in_seconds   = optional(number, null)
          is_anonymous_access_allowed = optional(bool, null)
          public_keys = object({
            type                        = string
            uri                         = optional(string, "")
            is_ssl_verify_disabled      = optional(bool, null)
            max_cache_duration_in_hours = optional(number, null)
            keys = optional(list(object({
              kid    = string
              format = string
              key    = optional(string, "")
              kty    = optional(string, "")
              alg    = optional(string, "")
              n      = optional(string, "")
              e      = optional(string, "")
              use    = optional(string, "")
            })), [])
          })
          verify_claims = optional(list(object({
            key         = optional(string, "")
            values      = optional(list(string), [])
            is_required = optional(bool, null)
          })), [])
        }), null)

        cors = optional(object({
          allowed_origins              = list(string)
          allowed_methods              = optional(list(string), [])
          allowed_headers              = optional(list(string), [])
          exposed_headers              = optional(list(string), [])
          is_allow_credentials_enabled = optional(bool, null)
          max_age_in_seconds           = optional(number, null)
        }), null)

        rate_limiting = optional(object({
          rate_in_requests_per_second = number
          rate_key                    = string
        }), null)
      }), null)

      routes = list(object({
        path    = string
        methods = optional(list(string), [])
        backend = object({
          type                       = string
          url                        = optional(string, "")
          function_id                = optional(string, "")
          status                     = optional(number, 0)
          body                       = optional(string, "")
          connect_timeout_in_seconds = optional(number, null)
          read_timeout_in_seconds    = optional(number, null)
          send_timeout_in_seconds    = optional(number, null)
          is_ssl_verify_disabled     = optional(bool, null)
          headers = optional(list(object({
            name  = string
            value = string
          })), [])
        })
        authorization = optional(object({
          type          = string
          allowed_scope = optional(list(string), [])
        }), null)
        logging_policies = optional(object({
          access_log = optional(object({
            is_enabled = bool
          }), null)
          execution_log = optional(object({
            is_enabled = bool
            log_level  = optional(string, "")
          }), null)
        }), null)
      }))
    })
  })
}
