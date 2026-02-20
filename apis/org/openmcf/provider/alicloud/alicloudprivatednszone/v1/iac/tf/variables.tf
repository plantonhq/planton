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
  description = "Alibaba Cloud Private Zone specification"
  type = object({
    region            = string
    zone_name         = string
    remark            = optional(string, "")
    resource_group_id = optional(string, "")
    vpc_attachments = list(object({
      vpc_id    = string
      region_id = optional(string, "")
    }))
    records = optional(list(object({
      rr       = string
      type     = string
      value    = string
      ttl      = optional(number, 60)
      priority = optional(number, 1)
      remark   = optional(string, "")
    })), [])
    tags = optional(map(string), {})
  })

  validation {
    condition     = length(var.spec.zone_name) >= 1 && length(var.spec.zone_name) <= 253
    error_message = "zone_name must be between 1 and 253 characters."
  }

  validation {
    condition     = length(var.spec.vpc_attachments) >= 1
    error_message = "At least one VPC attachment is required."
  }

  validation {
    condition = alltrue([
      for r in var.spec.records : contains(["A", "CNAME", "MX", "PTR", "SRV", "TXT"], r.type)
    ])
    error_message = "Record type must be one of: A, CNAME, MX, PTR, SRV, TXT."
  }
}
