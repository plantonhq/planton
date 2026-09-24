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
  description = "GcpApiKey specification"
  type = object({
    project_id            = optional(string, "")
    key_id                = string
    display_name          = optional(string, "")
    service_account_email = optional(string, "")
    restrictions = optional(object({
      android_key_restrictions = optional(object({
        allowed_applications = list(object({
          package_name     = string
          sha1_fingerprint = string
        }))
      }))
      ios_key_restrictions = optional(object({
        allowed_bundle_ids = list(string)
      }))
      browser_key_restrictions = optional(object({
        allowed_referrers = list(string)
      }))
      server_key_restrictions = optional(object({
        allowed_ips = list(string)
      }))
      api_targets = optional(list(object({
        service = string
        methods = optional(list(string), [])
      })), [])
    }))
    deletion_policy = optional(string, "")
  })
}
