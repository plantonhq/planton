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
  description = "Alibaba Cloud Container Registry specification"
  type = object({
    region            = string
    instance_name     = string
    instance_type     = string
    payment_type      = optional(string, "Subscription")
    period            = optional(number, 0)
    password          = optional(string, "")
    resource_group_id = optional(string, "")
    namespaces = optional(list(object({
      name               = string
      auto_create        = optional(bool, false)
      default_visibility = optional(string, "PRIVATE")
    })), [])
  })

  validation {
    condition     = contains(["Basic", "Standard", "Advanced"], var.spec.instance_type)
    error_message = "instance_type must be one of: Basic, Standard, Advanced."
  }

  validation {
    condition     = contains(["Subscription", "PayAsYouGo"], var.spec.payment_type)
    error_message = "payment_type must be one of: Subscription, PayAsYouGo."
  }

  validation {
    condition = alltrue([
      for ns in var.spec.namespaces : length(ns.name) >= 2 && length(ns.name) <= 120
    ])
    error_message = "Each namespace name must be between 2 and 120 characters."
  }

  validation {
    condition = alltrue([
      for ns in var.spec.namespaces : contains(["PUBLIC", "PRIVATE"], ns.default_visibility)
    ])
    error_message = "Each namespace default_visibility must be one of: PUBLIC, PRIVATE."
  }
}
