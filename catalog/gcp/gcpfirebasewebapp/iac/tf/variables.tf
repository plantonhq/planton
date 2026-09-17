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
  description = "GcpFirebaseWebApp specification"
  type = object({
    project_id   = optional(string, "")
    display_name = string
    api_key_id   = optional(string, "")
    app_check = optional(object({
      recaptcha_v3 = optional(object({
        site_secret = string
        token_ttl   = optional(string, "")
      }))
      recaptcha_enterprise = optional(object({
        site_key  = string
        token_ttl = optional(string, "")
      }))
      debug_tokens = optional(list(object({
        display_name = string
        token        = string
      })), [])
    }))
    deletion_policy = optional(string, "")
  })
}
