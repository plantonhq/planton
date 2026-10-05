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
  description = "AwsSagemakerModel specification"
  type = object({
    region = string
    execution_role_arn = string
    enable_network_isolation = optional(bool, false)
    primary_container = optional(object({
      image = optional(string, "")
      model_package_arn = optional(string, "")
      container_hostname = optional(string, "")
      environment = optional(map(string), {})
      mode = optional(string, "")
      model_data_url = optional(string, "")
      model_data_source = optional(object({
        s3_uri = string
        s3_data_type = optional(string, "")
        compression_type = optional(string, "")
        accept_eula = optional(bool, false)
      }))
      additional_model_data_sources = optional(list(object({
        channel_name = string
        source = object({
          s3_uri = string
          s3_data_type = optional(string, "")
          compression_type = optional(string, "")
          accept_eula = optional(bool, false)
        })
      })), [])
      inference_specification_name = optional(string, "")
      multi_model_cache = optional(string, "")
      image_config = optional(object({
        repository_access_mode = optional(string, "")
        repository_credentials_provider_arn = optional(string, "")
      }))
    }))
    containers = optional(list(object({
      image = optional(string, "")
      model_package_arn = optional(string, "")
      container_hostname = optional(string, "")
      environment = optional(map(string), {})
      mode = optional(string, "")
      model_data_url = optional(string, "")
      model_data_source = optional(object({
        s3_uri = string
        s3_data_type = optional(string, "")
        compression_type = optional(string, "")
        accept_eula = optional(bool, false)
      }))
      additional_model_data_sources = optional(list(object({
        channel_name = string
        source = object({
          s3_uri = string
          s3_data_type = optional(string, "")
          compression_type = optional(string, "")
          accept_eula = optional(bool, false)
        })
      })), [])
      inference_specification_name = optional(string, "")
      multi_model_cache = optional(string, "")
      image_config = optional(object({
        repository_access_mode = optional(string, "")
        repository_credentials_provider_arn = optional(string, "")
      }))
    })), [])
    inference_execution_mode = optional(string, "")
    vpc_config = optional(object({
      subnet_ids = list(string)
      security_group_ids = list(string)
    }))
  })
}