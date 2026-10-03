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
  description = "AwsOrganizationPolicy specification"
  type = object({
    region      = string
    policy_name = string
    type        = optional(string, "")
    content     = any
    description = optional(string, "")
    attachments = optional(list(object({
      target_id = string
    })), [])
  })
}