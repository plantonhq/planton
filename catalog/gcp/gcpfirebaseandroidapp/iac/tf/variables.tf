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
  description = "GcpFirebaseAndroidApp specification"
  type = object({
    project_id    = optional(string, "")
    display_name  = string
    package_name  = string
    sha1_hashes   = optional(list(string), [])
    sha256_hashes = optional(list(string), [])
    api_key_id    = optional(string, "")
    app_check = optional(object({
      play_integrity = optional(object({
        enabled   = optional(bool)
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
