variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)."
  type = object({
    name   = string
    org    = string
    env    = string
    id     = string
    labels = optional(map(string), {})
  })
}

variable "spec" {
  description = "AwsHttpApiGatewaySpec — desired configuration for the HTTP API Gateway."
  type = object({
    description = optional(string, "")
    cors_configuration = optional(object({
      allow_origins     = optional(list(string), [])
      allow_methods     = optional(list(string), [])
      allow_headers     = optional(list(string), [])
      expose_headers    = optional(list(string), [])
      max_age_seconds   = optional(number, 0)
      allow_credentials = optional(bool, false)
    }), null)
    disable_execute_api_endpoint = optional(bool, false)
    stage = optional(object({
      name        = optional(string, "")
      auto_deploy = optional(bool, true)
      access_log = optional(object({
        destination_arn = object({
          value = string
        })
        format = string
      }), null)
      default_throttle = optional(object({
        burst_limit = optional(number, 0)
        rate_limit  = optional(number, 0)
      }), null)
      stage_variables = optional(map(string), {})
    }), null)
    routes = list(object({
      route_key = string
      integration = object({
        integration_type = string
        integration_uri = object({
          value = string
        })
        payload_format_version = optional(string, "")
        integration_method     = optional(string, "")
        timeout_milliseconds   = optional(number, 0)
      })
      authorization_type   = optional(string, "")
      authorizer_name      = optional(string, "")
      authorization_scopes = optional(list(string), [])
    }))
    authorizers = optional(list(object({
      name            = string
      authorizer_type = string
      jwt_configuration = optional(object({
        issuer    = string
        audiences = optional(list(string), [])
      }), null)
      authorizer_uri = optional(object({
        value = string
      }), null)
      authorizer_credentials_arn = optional(object({
        value = string
      }), null)
      identity_sources                  = optional(list(string), [])
      result_ttl_seconds                = optional(number, 0)
      enable_simple_responses           = optional(bool, false)
      authorizer_payload_format_version = optional(string, "")
    })), [])
  })
}
