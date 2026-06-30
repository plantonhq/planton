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
  description = "Alibaba Cloud NAT Gateway specification"
  type = object({
    region              = string
    vpc_id              = string
    vswitch_id          = string
    nat_gateway_name    = string
    description         = optional(string, "")
    nat_type            = optional(string, "Enhanced")
    payment_type        = optional(string, "PayAsYouGo")
    internet_charge_type = optional(string, "PayByLcu")
    specification       = optional(string, "")
    deletion_protection = optional(bool, false)
    tags                = optional(map(string), {})
    eip_id              = string
    snat_entries = optional(list(object({
      source_vswitch_id = optional(string, "")
      source_cidr       = optional(string, "")
      snat_entry_name   = optional(string, "")
    })), [])
  })

  validation {
    condition     = length(var.spec.nat_gateway_name) >= 2 && length(var.spec.nat_gateway_name) <= 128
    error_message = "nat_gateway_name must be between 2 and 128 characters."
  }

  validation {
    condition     = contains(["Enhanced", "Normal"], var.spec.nat_type)
    error_message = "nat_type must be one of: Enhanced, Normal."
  }

  validation {
    condition     = contains(["PayAsYouGo", "Subscription"], var.spec.payment_type)
    error_message = "payment_type must be one of: PayAsYouGo, Subscription."
  }

  validation {
    condition     = contains(["PayByLcu", "PayBySpec"], var.spec.internet_charge_type)
    error_message = "internet_charge_type must be one of: PayByLcu, PayBySpec."
  }

  validation {
    condition = (
      var.spec.specification == "" ||
      contains(["Small", "Middle", "Large", "XLarge.1"], var.spec.specification)
    )
    error_message = "specification must be one of: Small, Middle, Large, XLarge.1."
  }
}
