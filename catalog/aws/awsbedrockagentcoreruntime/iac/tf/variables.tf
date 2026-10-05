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
  description = "AwsBedrockAgentCoreRuntime specification"
  type = object({
    region       = string
    runtime_name = string
    description  = optional(string, "")
    role_arn     = string
    artifact = object({
      container = optional(object({
        image_uri = string
      }))
      code = optional(object({
        runtime     = optional(string, "")
        entry_point = list(string)
        s3 = object({
          bucket     = string
          prefix     = string
          version_id = optional(string, "")
        })
      }))
    })
    network = object({
      mode = optional(string, "")
      vpc_config = optional(object({
        subnets         = list(string)
        security_groups = list(string)
      }))
    })
    server_protocol       = optional(string, "")
    environment_variables = optional(map(string), {})
    lifecycle = optional(object({
      idle_runtime_session_timeout_seconds = optional(number, 0)
      max_lifetime_seconds                 = optional(number, 0)
    }))
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
    request_header_allowlist = optional(list(string), [])
    filesystems = optional(list(object({
      mount_path                = string
      efs_access_point_arn      = optional(string, "")
      s3_files_access_point_arn = optional(string, "")
      session_storage           = optional(bool, false)
    })), [])
    endpoints = optional(list(object({
      name                  = string
      description           = optional(string, "")
      agent_runtime_version = optional(string, "")
    })), [])
    resource_policy = optional(any)
  })
}
