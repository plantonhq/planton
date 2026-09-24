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
  description = "GcpFirebaseAppleApp specification"
  type = object({
    project_id   = optional(string, "")
    display_name = string
    bundle_id    = string
    app_store_id = optional(string, "")
    team_id      = optional(string, "")
    api_key_id   = optional(string, "")
    app_check = optional(object({
      app_attest = optional(object({
        enabled   = optional(bool)
        token_ttl = optional(string, "")
      }))
      device_check = optional(object({
        key_id      = string
        private_key = string
        token_ttl   = optional(string, "")
      }))
      debug_tokens = optional(list(object({
        display_name = string
        token        = string
      })), [])
    }))
    deletion_policy = optional(string, "")
  })
}
