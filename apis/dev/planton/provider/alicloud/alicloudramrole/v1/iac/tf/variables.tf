variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud RAM Role specification"
  type = object({
    region                       = string
    role_name                    = string
    description                  = optional(string, "")
    assume_role_policy_document = string
    max_session_duration         = optional(number, 3600)
    tags                         = optional(map(string), {})
    force                        = optional(bool, false)
    policy_attachments = optional(list(object({
      policy_name = string
      policy_type = optional(string, "System")
    })), [])
  })

  validation {
    condition     = length(var.spec.role_name) >= 1 && length(var.spec.role_name) <= 64
    error_message = "role_name must be between 1 and 64 characters."
  }

  validation {
    condition     = var.spec.max_session_duration >= 3600 && var.spec.max_session_duration <= 43200
    error_message = "max_session_duration must be between 3600 and 43200 seconds."
  }
}
