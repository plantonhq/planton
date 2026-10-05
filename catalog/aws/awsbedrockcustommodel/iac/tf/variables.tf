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
  description = "AwsBedrockCustomModel specification"
  type = object({
    region = string
    base_model_arn = optional(string, "")
    customization_type = optional(string, "")
    job_name = optional(string, "")
    hyperparameters = optional(map(string), {})
    role_arn = string
    custom_model_kms_key_arn = optional(string, "")
    training_data_s3_uri = optional(string, "")
    output_data_s3_uri = optional(string, "")
    validation_data_s3_uris = optional(list(string), [])
    vpc_config = optional(object({
      subnet_ids = list(string)
      security_group_ids = list(string)
    }))
  })
}