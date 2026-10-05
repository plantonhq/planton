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
  description = "CloudflareZeroTrustMcpPortal specification"
  type = object({
    account_id         = string
    portal_id          = string
    hostname           = string
    name               = string
    description        = optional(string, "")
    allow_code_mode    = optional(bool)
    secure_web_gateway = optional(bool)
    servers = optional(list(object({
      server_id        = string
      default_disabled = optional(bool)
      on_behalf        = optional(bool)
      updated_prompts = optional(list(object({
        name        = string
        alias       = optional(string, "")
        description = optional(string, "")
        enabled     = optional(bool)
      })), [])
      updated_tools = optional(list(object({
        name        = string
        alias       = optional(string, "")
        description = optional(string, "")
        enabled     = optional(bool)
      })), [])
    })), [])
  })
}