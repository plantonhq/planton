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
  description = "Auth0EmailProvider specification"
  type = object({
    default_from_address = string
    enabled              = optional(bool)
    smtp = optional(object({
      host     = string
      port     = optional(number, 0)
      user     = string
      password = string
      headers = optional(object({
        x_mc_view_content_link  = optional(string, "")
        x_ses_configuration_set = optional(string, "")
      }))
    }))
    ses = optional(object({
      access_key_id          = string
      secret_access_key      = string
      region                 = string
      configuration_set_name = optional(string, "")
    }))
    sendgrid = optional(object({
      api_key = string
    }))
    sparkpost = optional(object({
      api_key = string
      region  = optional(string, "")
    }))
    mailgun = optional(object({
      api_key = string
      domain  = string
      region  = optional(string, "")
    }))
    mandrill = optional(object({
      api_key           = string
      view_content_link = optional(bool)
    }))
    azure_cs = optional(object({
      connection_string = string
    }))
    ms365 = optional(object({
      tenant_id     = string
      client_id     = string
      client_secret = string
    }))
    custom = optional(object({}))
  })
}
