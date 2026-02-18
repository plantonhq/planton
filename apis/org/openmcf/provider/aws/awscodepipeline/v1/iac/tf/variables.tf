variable "metadata" {
  description = "Resource metadata from the manifest"
  type = object({
    name = string
    id   = string
    org  = string
    env  = string
    labels = object({
      key   = string
      value = string
    })
    annotations = object({
      key   = string
      value = string
    })
    tags = list(string)
  })
}

variable "spec" {
  description = "AwsCodePipelineSpec configuration"
  type = object({
    # The AWS region where the resource will be created.
    region = string

    pipeline_type  = string
    execution_mode = string

    role_arn = object({
      value = string
    })

    artifact_stores = list(object({
      location          = object({ value = string })
      region            = string
      encryption_key_id = optional(object({ value = string }))
    }))

    stages = list(object({
      name = string
      actions = list(object({
        name               = string
        category           = string
        owner              = string
        provider           = string
        version            = string
        configuration      = map(string)
        input_artifacts    = list(string)
        output_artifacts   = list(string)
        namespace          = string
        region             = string
        role_arn           = optional(object({ value = string }))
        run_order          = number
        timeout_in_minutes = number
      }))
    }))

    triggers = optional(list(object({
      provider_type = string
      git_configuration = object({
        source_action_name = string
        push = optional(list(object({
          branches   = optional(object({ includes = list(string), excludes = list(string) }))
          file_paths = optional(object({ includes = list(string), excludes = list(string) }))
          tags       = optional(object({ includes = list(string), excludes = list(string) }))
        })))
        pull_request = optional(list(object({
          branches   = optional(object({ includes = list(string), excludes = list(string) }))
          file_paths = optional(object({ includes = list(string), excludes = list(string) }))
          events     = optional(list(string))
        })))
      })
    })))

    variables = optional(list(object({
      name          = string
      default_value = string
      description   = string
    })))
  })
}

variable "provider_config" {
  description = "AWS provider configuration"
  type = object({
    region            = string
    access_key_id     = string
    secret_access_key = string
    session_token     = string
  })
  default = {
    region            = ""
    access_key_id     = ""
    secret_access_key = ""
    session_token     = ""
  }
}
