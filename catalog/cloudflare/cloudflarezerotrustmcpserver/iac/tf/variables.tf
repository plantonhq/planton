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
  description = "CloudflareZeroTrustMcpServer specification"
  type = object({
    account_id       = string
    server_id        = string
    name             = string
    hostname         = string
    auth_type        = string
    auth_credentials = optional(string, "")
    client_secret    = optional(string, "")
    description      = optional(string, "")
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
    is_shared_oauth_callback_enabled = optional(bool)
    secure_web_gateway               = optional(bool)
  })
}