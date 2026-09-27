variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "Auth0TenantSettings specification"
  type = object({
    friendly_name = optional(string, "")
    picture_url   = optional(string, "")
    support_email = optional(string, "")
    support_url   = optional(string, "")
  })
}
