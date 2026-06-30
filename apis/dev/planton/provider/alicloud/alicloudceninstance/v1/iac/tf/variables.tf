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
  description = "Alibaba Cloud CEN instance specification"
  type = object({
    region            = string
    cen_instance_name = string
    description       = optional(string, "")
    protection_level  = optional(string, "")
    resource_group_id = optional(string, "")
    tags              = optional(map(string), {})
    attachments = optional(list(object({
      child_instance_id        = string
      child_instance_type      = optional(string, "VPC")
      child_instance_region_id = string
    })), [])
  })

  validation {
    condition     = length(var.spec.cen_instance_name) >= 2 && length(var.spec.cen_instance_name) <= 128
    error_message = "cen_instance_name must be between 2 and 128 characters."
  }

  validation {
    condition     = var.spec.protection_level == "" || var.spec.protection_level == "REDUCED"
    error_message = "protection_level must be empty or 'REDUCED'."
  }

  validation {
    condition = alltrue([
      for a in var.spec.attachments : contains(["VPC", "VBR", "CCN"], a.child_instance_type)
    ])
    error_message = "Each attachment child_instance_type must be one of: VPC, VBR, CCN."
  }
}
