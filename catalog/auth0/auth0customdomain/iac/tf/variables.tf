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
  description = "Auth0CustomDomain specification"
  type = object({
    domain                   = string
    type                     = string
    custom_client_ip_header  = optional(string, "")
    tls_policy               = optional(string, "")
    domain_metadata          = optional(map(string), {})
    relying_party_identifier = optional(string, "")
  })
}
