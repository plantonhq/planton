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
  description = "AwsSagemakerNotebookInstance specification"
  type = object({
    region = string
    instance_type = optional(string, "")
    role_arn = string
    volume_size_gb = optional(number)
    subnet_id = optional(string, "")
    security_group_ids = optional(list(string), [])
    kms_key_arn = optional(string, "")
    direct_internet_access = optional(string, "")
    root_access = optional(string, "")
    platform_identifier = optional(string, "")
    default_code_repository = optional(string, "")
    additional_code_repositories = optional(list(string), [])
    imds_minimum_version = optional(string, "")
    lifecycle_config = optional(object({
      on_create = optional(string, "")
      on_start = optional(string, "")
    }))
  })
}