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
  description = "AwsBedrockAgentCoreIdentity specification"
  type = object({
    region = string
    workload_identities = optional(list(object({
      name                                = string
      allowed_resource_oauth2_return_urls = optional(list(string), [])
    })), [])
    api_key_credential_providers = optional(list(object({
      name    = string
      api_key = string
    })), [])
    oauth2_credential_providers = optional(list(object({
      name          = string
      vendor        = optional(string, "")
      client_id     = string
      client_secret = string
      oauth_discovery = optional(object({
        discovery_url = optional(string, "")
        authorization_server_metadata = optional(object({
          issuer                 = string
          authorization_endpoint = string
          token_endpoint         = string
          response_types         = optional(list(string), [])
        }))
      }))
    })), [])
    policy_engine = optional(object({
      engine_name        = string
      description        = optional(string, "")
      encryption_key_arn = optional(string, "")
      policies = optional(list(object({
        name            = string
        description     = optional(string, "")
        cedar_statement = string
        validation_mode = optional(string, "")
      })), [])
    }))
  })
}
