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
  description = "Alibaba Cloud RAM Policy specification"
  type = object({
    region          = string
    policy_name     = string
    description     = optional(string, "")
    policy_document = string
    rotate_strategy = optional(string)
    tags            = optional(map(string), {})
    force           = optional(bool, false)
  })

  validation {
    condition     = length(var.spec.policy_name) >= 1 && length(var.spec.policy_name) <= 128
    error_message = "policy_name must be between 1 and 128 characters."
  }

  validation {
    condition     = var.spec.rotate_strategy == null || contains(["None", "DeleteOldestNonDefaultVersionWhenLimitExceeded"], var.spec.rotate_strategy)
    error_message = "rotate_strategy must be \"None\" or \"DeleteOldestNonDefaultVersionWhenLimitExceeded\"."
  }
}
