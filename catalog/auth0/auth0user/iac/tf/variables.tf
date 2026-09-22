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
  description = "Auth0User specification"
  type = object({
    connection_name      = string
    email                = optional(string, "")
    email_verified       = optional(bool, false)
    verify_email         = optional(bool)
    username             = optional(string, "")
    name                 = optional(string, "")
    given_name           = optional(string, "")
    family_name          = optional(string, "")
    nickname             = optional(string, "")
    picture              = optional(string, "")
    phone_number         = optional(string, "")
    phone_verified       = optional(bool, false)
    blocked              = optional(bool, false)
    user_id              = optional(string, "")
    password             = optional(string, "")
    passwordless         = optional(bool, false)
    user_metadata        = optional(any)
    app_metadata         = optional(any)
    custom_domain_header = optional(string, "")
    roles                = optional(list(string), [])
    permissions = optional(list(object({
      name                       = string
      resource_server_identifier = string
    })), [])
  })
}
