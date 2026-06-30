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
  description = "Alibaba Cloud VPC specification"
  type = object({
    region            = string
    vpc_name          = string
    cidr_block        = string
    description       = optional(string, "")
    enable_ipv6       = optional(bool, false)
    resource_group_id = optional(string, "")
    tags              = optional(map(string), {})
  })

  validation {
    condition     = length(var.spec.vpc_name) >= 1 && length(var.spec.vpc_name) <= 128
    error_message = "vpc_name must be between 1 and 128 characters."
  }

  validation {
    condition     = length(var.spec.cidr_block) > 0
    error_message = "cidr_block is required."
  }
}
