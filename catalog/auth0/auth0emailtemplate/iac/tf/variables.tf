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
  description = "Auth0EmailTemplate specification"
  type = object({
    template                  = string
    from                      = string
    subject                   = string
    body                      = string
    syntax                    = optional(string)
    result_url                = optional(string, "")
    url_lifetime_in_seconds   = optional(number)
    enabled                   = optional(bool)
    include_email_in_redirect = optional(bool)
  })
}
