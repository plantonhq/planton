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
  description = "AwsBedrockAgentCoreTools specification"
  type = object({
    region = string
    browsers = optional(list(object({
      name               = string
      description        = optional(string, "")
      execution_role_arn = optional(string, "")
      network = object({
        mode = optional(string, "")
        vpc_config = optional(object({
          subnets         = list(string)
          security_groups = list(string)
        }))
      })
      signing_enabled = optional(bool)
      recording = optional(object({
        enabled = optional(bool)
        s3_location = optional(object({
          bucket = string
          prefix = string
        }))
      }))
      enterprise_policies = optional(list(object({
        type = optional(string, "")
        s3 = object({
          bucket     = string
          prefix     = string
          version_id = optional(string, "")
        })
      })), [])
      certificates = optional(list(object({
        secret_arn = string
      })), [])
    })), [])
    browser_profiles = optional(list(object({
      name        = string
      description = optional(string, "")
    })), [])
    code_interpreters = optional(list(object({
      name               = string
      description        = optional(string, "")
      execution_role_arn = optional(string, "")
      network = object({
        mode = optional(string, "")
        vpc_config = optional(object({
          subnets         = list(string)
          security_groups = list(string)
        }))
      })
      certificates = optional(list(object({
        secret_arn = string
      })), [])
    })), [])
  })
}
