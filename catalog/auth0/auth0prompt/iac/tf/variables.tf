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
  description = "Auth0Prompt specification"
  type = object({
    universal_login_experience     = optional(string, "")
    identifier_first               = optional(bool)
    webauthn_platform_first_factor = optional(bool)
  })
}
