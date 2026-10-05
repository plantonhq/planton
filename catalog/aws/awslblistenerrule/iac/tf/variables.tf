variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsLbListenerRule specification"
  type = object({
    region = string
    listener_arn = string
    priority = optional(number, 0)
    # Condition blocks AND together; exactly one criterion is set per block
    # (enforced by spec validation before the module runs).
    conditions = list(object({
      host_header = optional(object({
        values = optional(list(string), [])
        regex_values = optional(list(string), [])
      }))
      path_pattern = optional(object({
        values = optional(list(string), [])
        regex_values = optional(list(string), [])
      }))
      http_header = optional(object({
        http_header_name = string
        values = optional(list(string), [])
        regex_values = optional(list(string), [])
      }))
      http_request_method = optional(object({
        values = list(string)
      }))
      query_string = optional(object({
        pairs = list(object({
          key = optional(string, "")
          value = string
        }))
      }))
      source_ip = optional(object({
        values = list(string)
      }))
    }))
    # An action chain: every entry carries exactly the configuration object
    # matching its type (enforced by spec validation before the module runs).
    actions = list(object({
      type = string
      order = optional(number, 0)
      forward = optional(object({
        target_groups = list(object({
          arn = string
          weight = optional(number, 0)
        }))
        stickiness = optional(object({
          enabled = optional(bool, false)
          duration_seconds = optional(number, 0)
        }))
      }))
      redirect = optional(object({
        status_code = string
        protocol = optional(string, "")
        port = optional(string, "")
        host = optional(string, "")
        path = optional(string, "")
        query = optional(string, "")
      }))
      fixed_response = optional(object({
        content_type = string
        status_code = optional(string, "")
        message_body = optional(string, "")
      }))
      authenticate_cognito = optional(object({
        user_pool_arn = string
        user_pool_client_id = string
        user_pool_domain = string
        authentication_request_extra_params = optional(map(string), {})
        on_unauthenticated_request = optional(string, "")
        scope = optional(string, "")
        session_cookie_name = optional(string, "")
        session_timeout_seconds = optional(number, 0)
      }))
      authenticate_oidc = optional(object({
        issuer = string
        authorization_endpoint = string
        token_endpoint = string
        user_info_endpoint = string
        client_id = string
        client_secret = string
        authentication_request_extra_params = optional(map(string), {})
        on_unauthenticated_request = optional(string, "")
        scope = optional(string, "")
        session_cookie_name = optional(string, "")
        session_timeout_seconds = optional(number, 0)
      }))
      jwt_validation = optional(object({
        issuer = string
        jwks_endpoint = string
        additional_claims = optional(list(object({
          name = string
          format = string
          values = list(string)
        })), [])
      }))
    }))
    transforms = optional(list(object({
      type = string
      host_header_rewrite = optional(object({
        regex = string
        replace = optional(string, "")
      }))
      url_rewrite = optional(object({
        regex = string
        replace = optional(string, "")
      }))
    })), [])
  })
}
