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
  description = "AwsEventBridgeApiDestination specification"
  type = object({
    region = string
    connection = optional(object({
      name = string
      description = optional(string, "")
      api_key = optional(object({
        key = string
        value = string
      }))
      basic = optional(object({
        username = string
        password = string
      }))
      oauth = optional(object({
        authorization_endpoint = string
        http_method = optional(string, "")
        client_id = string
        client_secret = string
        oauth_http_parameters = object({
          body = optional(list(object({
            key = string
            value = string
            is_value_secret = optional(bool, false)
          })), [])
          header = optional(list(object({
            key = string
            value = string
            is_value_secret = optional(bool, false)
          })), [])
          query_string = optional(list(object({
            key = string
            value = string
            is_value_secret = optional(bool, false)
          })), [])
        })
      }))
      invocation_http_parameters = optional(object({
        body = optional(list(object({
          key = string
          value = string
          is_value_secret = optional(bool, false)
        })), [])
        header = optional(list(object({
          key = string
          value = string
          is_value_secret = optional(bool, false)
        })), [])
        query_string = optional(list(object({
          key = string
          value = string
          is_value_secret = optional(bool, false)
        })), [])
      }))
      private_invocation_endpoint = optional(object({
        resource_configuration_arn = string
      }))
      private_authorization_endpoint = optional(object({
        resource_configuration_arn = string
      }))
      kms_key_identifier = optional(string, "")
    }))
    destination = optional(object({
      name = string
      description = optional(string, "")
      connection_arn = optional(string, "")
      invocation_endpoint = string
      http_method = optional(string, "")
      invocation_rate_limit_per_second = optional(number)
    }))
  })
}