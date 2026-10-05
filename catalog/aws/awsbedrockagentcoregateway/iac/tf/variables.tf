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
  description = "AwsBedrockAgentCoreGateway specification"
  type = object({
    region          = string
    description     = optional(string, "")
    role_arn        = string
    authorizer_type = optional(string, "")
    custom_jwt_authorizer = optional(object({
      discovery_url    = string
      allowed_audience = optional(list(string), [])
      allowed_clients  = optional(list(string), [])
      allowed_scopes   = optional(list(string), [])
      allowed_workloads = optional(object({
        workload_identities      = list(string)
        hosting_environment_arns = list(string)
      }))
      custom_claims = optional(list(object({
        claim_name     = string
        value_type     = optional(string, "")
        match_operator = optional(string, "")
        match_value    = optional(string, "")
        match_values   = optional(list(string), [])
      })), [])
      private_endpoint = optional(object({
        managed_vpc = optional(object({
          vpc_id                   = string
          subnet_ids               = list(string)
          security_group_ids       = optional(list(string), [])
          endpoint_ip_address_type = optional(string, "")
          routing_domain           = optional(string, "")
          tags                     = optional(map(string), {})
        }))
        self_managed_lattice = optional(object({
          resource_configuration_id = string
        }))
      }))
      private_endpoint_overrides = optional(list(object({
        domain = string
        private_endpoint = object({
          managed_vpc = optional(object({
            vpc_id                   = string
            subnet_ids               = list(string)
            security_group_ids       = optional(list(string), [])
            endpoint_ip_address_type = optional(string, "")
            routing_domain           = optional(string, "")
            tags                     = optional(map(string), {})
          }))
          self_managed_lattice = optional(object({
            resource_configuration_id = string
          }))
        })
      })), [])
    }))
    kms_key_arn             = optional(string, "")
    expose_debug_exceptions = optional(bool, false)
    mcp = optional(object({
      instructions              = optional(string, "")
      enable_semantic_search    = optional(bool, false)
      supported_versions        = optional(list(string), [])
      session_timeout_seconds   = optional(number, 0)
      enable_response_streaming = optional(bool, false)
    }))
    interceptors = optional(list(object({
      interception_points  = list(string)
      lambda_arn           = string
      pass_request_headers = optional(bool)
    })), [])
    policy_engine = optional(object({
      policy_engine_arn = string
      mode              = optional(string, "")
    }))
    targets = optional(any, [])
  })
}
