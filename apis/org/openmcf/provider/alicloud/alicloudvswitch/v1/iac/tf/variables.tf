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
  description = "Alibaba Cloud VSwitch specification"
  type = object({
    region              = string
    vpc_id              = string
    zone_id             = string
    cidr_block          = string
    vswitch_name        = string
    description         = optional(string, "")
    enable_ipv6         = optional(bool, false)
    ipv6_cidr_block_mask = optional(number, 0)
    tags                = optional(map(string), {})
  })

  validation {
    condition     = length(var.spec.vswitch_name) >= 1 && length(var.spec.vswitch_name) <= 128
    error_message = "vswitch_name must be between 1 and 128 characters."
  }

  validation {
    condition     = length(var.spec.cidr_block) > 0
    error_message = "cidr_block is required."
  }

  validation {
    condition     = length(var.spec.zone_id) > 0
    error_message = "zone_id is required."
  }

  validation {
    condition     = length(var.spec.vpc_id) > 0
    error_message = "vpc_id is required."
  }
}
