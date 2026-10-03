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
  description = "AwsS3VectorBucket specification"
  type = object({
    region = string
    encryption = optional(object({
      sse_type = optional(string, "")
      kms_key_arn = optional(string, "")
    }))
    force_destroy = optional(bool, false)
    policy = optional(string, "")
    indexes = optional(list(object({
      name = string
      dimension = optional(number, 0)
      distance_metric = optional(string, "")
      non_filterable_metadata_keys = optional(list(string), [])
      encryption = optional(object({
        sse_type = optional(string, "")
        kms_key_arn = optional(string, "")
      }))
    })), [])
  })
}