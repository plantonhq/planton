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
  description = "AwsConfigConformancePack specification"
  type = object({
    region = string
    organization_scope = optional(bool, false)
    delivery_s3_bucket = optional(string, "")
    delivery_s3_key_prefix = optional(string, "")
    template_body = optional(string, "")
    template_s3_uri = optional(string, "")
    input_parameters = optional(list(object({
      parameter_name = string
      parameter_value = string
    })), [])
    excluded_accounts = optional(list(string), [])
  })
}
