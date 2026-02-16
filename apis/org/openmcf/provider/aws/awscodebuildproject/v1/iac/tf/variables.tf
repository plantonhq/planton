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
  description = "AwsCodeBuildProjectSpec configuration"
  type = object({
    source = object({
      type                = string
      location            = string
      buildspec           = string
      git_clone_depth     = number
      report_build_status = bool
      fetch_submodules    = bool
    })

    environment = object({
      type                       = string
      compute_type               = string
      image                      = string
      privileged_mode            = bool
      image_pull_credentials_type = string
      environment_variables = list(object({
        name  = string
        value = string
        type  = string
      }))
      registry_credential = optional(object({
        credential          = string
        credential_provider = string
      }))
    })

    artifacts = object({
      type                = string
      location            = optional(object({ value = string }))
      name                = string
      path                = string
      packaging           = string
      namespace_type      = string
      encryption_disabled = bool
    })

    service_role = object({
      value = string
    })

    description            = string
    encryption_key         = optional(object({ value = string }))
    build_timeout          = number
    queued_timeout         = number
    concurrent_build_limit = number
    source_version         = string

    cache = optional(object({
      type     = string
      location = optional(object({ value = string }))
      modes    = list(string)
    }))

    logs_config = optional(object({
      cloudwatch_logs = optional(object({
        status      = string
        group_name  = optional(object({ value = string }))
        stream_name = string
      }))
      s3_logs = optional(object({
        status              = string
        location            = optional(object({ value = string }))
        encryption_disabled = bool
      }))
    }))

    vpc_config = optional(object({
      vpc_id             = object({ value = string })
      subnet_ids         = list(object({ value = string }))
      security_group_ids = list(object({ value = string }))
    }))

    webhook = optional(object({
      build_type = string
      filter_groups = list(object({
        filters = list(object({
          type                    = string
          pattern                 = string
          exclude_matched_pattern = bool
        }))
      }))
    }))
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
