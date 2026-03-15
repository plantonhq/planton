# variables.tf

variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareZeroTrustAccessApplicationSpec defines the configuration for the Access Application"
  type = object({
    application_name         = string
    zone_id = object({
      value = string
    })
    hostname                 = string
    policy_type              = optional(string, "ALLOW")
    allowed_emails           = optional(list(string), [])
    session_duration_minutes = optional(number, 1440)
    require_mfa              = optional(bool, false)
    allowed_google_groups    = optional(list(string), [])
  })
}
