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
  description = "AwsBedrockInvocationLogging specification"
  type = object({
    region                          = string
    text_data_delivery_enabled      = optional(bool)
    image_data_delivery_enabled     = optional(bool)
    embedding_data_delivery_enabled = optional(bool)
    video_data_delivery_enabled     = optional(bool)
    cloudwatch = optional(object({
      log_group_name = string
      role_arn       = string
      large_data_delivery_s3 = optional(object({
        bucket_name = string
        key_prefix  = optional(string, "")
      }))
    }))
    s3 = optional(object({
      bucket_name = string
      key_prefix  = optional(string, "")
    }))
  })
}
