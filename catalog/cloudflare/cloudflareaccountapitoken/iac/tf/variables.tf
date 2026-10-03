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
  description = "CloudflareAccountApiToken specification"
  type = object({
    account_id = string
    name = string
    policies = list(object({
      effect = string
      permission_group_ids = list(string)
      resources = optional(map(object({
        permission = optional(string, "")
        subresources = optional(map(string), {})
      })), {})
    }))
    expires_on = optional(string, "")
    not_before = optional(string, "")
    condition = optional(object({
      request_ip = optional(object({
        in_cidrs = optional(list(string), [])
        not_in_cidrs = optional(list(string), [])
      }))
    }))
    status = optional(string, "")
  })
}