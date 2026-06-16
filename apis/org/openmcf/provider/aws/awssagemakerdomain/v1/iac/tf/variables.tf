variable "metadata" {
  description = "Resource metadata"
  type = object({
    id   = string
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
  })
}

variable "spec" {
  description = "AwsSagemakerDomainSpec - desired state of the SageMaker domain"
  type = object({
    # The AWS region where the resource will be created.
    region                 = string
    auth_mode              = string
    vpc_id                 = string
    subnet_ids             = list(string)
    kms_key_id             = optional(string, "")
    app_network_access_type = optional(string, "PublicInternetOnly")

    domain_security_group_ids = optional(list(string), [])

    docker_settings = optional(object({
      enable_docker_access      = optional(string, "DISABLED")
      vpc_only_trusted_accounts = optional(list(string), [])
    }), null)

    default_user_settings = object({
      execution_role      = string
      security_groups     = optional(list(string), [])
      default_landing_uri = optional(string, "")
      studio_web_portal   = optional(string, "")

      jupyter_lab_app_settings = optional(object({
        default_resource_spec = optional(object({
          instance_type        = optional(string, "")
          lifecycle_config_arn = optional(string, "")
          sagemaker_image_arn  = optional(string, "")
        }), null)
        lifecycle_config_arns = optional(list(string), [])
        custom_image = optional(list(object({
          app_image_config_name = string
          image_name            = string
          image_version_number  = optional(number, null)
        })), [])
        code_repository = optional(list(object({
          repository_url = string
        })), [])
        app_lifecycle_management = optional(object({
          idle_settings = optional(object({
            idle_timeout_in_minutes  = optional(number, null)
            lifecycle_management     = optional(string, "")
            max_idle_timeout_in_minutes = optional(number, null)
            min_idle_timeout_in_minutes = optional(number, null)
          }), null)
        }), null)
      }), null)

      kernel_gateway_app_settings = optional(object({
        default_resource_spec = optional(object({
          instance_type        = optional(string, "")
          lifecycle_config_arn = optional(string, "")
          sagemaker_image_arn  = optional(string, "")
        }), null)
        lifecycle_config_arns = optional(list(string), [])
        custom_image = optional(list(object({
          app_image_config_name = string
          image_name            = string
          image_version_number  = optional(number, null)
        })), [])
      }), null)

      sharing_settings = optional(object({
        notebook_output_option = optional(string, "")
        s3_kms_key_id          = optional(string, "")
        s3_output_path         = optional(string, "")
      }), null)

      space_storage_settings = optional(object({
        default_ebs_storage_settings = optional(object({
          default_ebs_volume_size_in_gb = number
          maximum_ebs_volume_size_in_gb = number
        }), null)
      }), null)
    })
  })
}
