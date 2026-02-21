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
  description = "Alibaba Cloud DNS domain specification"
  type = object({
    region            = string
    domain_name       = string
    group_id          = optional(string, "")
    remark            = optional(string, "")
    resource_group_id = optional(string, "")
    tags              = optional(map(string), {})
  })

  validation {
    condition     = length(var.spec.domain_name) >= 1 && length(var.spec.domain_name) <= 253
    error_message = "domain_name must be between 1 and 253 characters."
  }
}
