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
  description = "Alibaba Cloud NAS file system specification"
  type = object({
    region           = string
    file_system_type = optional(string, "standard")
    protocol_type    = string
    storage_type     = string
    description      = optional(string, "")

    encryption = optional(object({
      encrypt_type = number
      kms_key_id   = optional(string, "")
    }))

    capacity = optional(number, 0)
    zone_id  = optional(string, "")

    vpc_id    = string
    vswitch_id = string

    access_rules = optional(list(object({
      source_cidr_ip   = string
      rw_access_type   = optional(string, "RDWR")
      user_access_type = optional(string, "no_squash")
      priority         = optional(number, 1)
    })), [])

    resource_group_id = optional(string, "")
    tags              = optional(map(string), {})
  })

  validation {
    condition     = contains(["standard", "extreme"], var.spec.file_system_type)
    error_message = "file_system_type must be one of: standard, extreme."
  }

  validation {
    condition     = contains(["NFS", "SMB"], var.spec.protocol_type)
    error_message = "protocol_type must be one of: NFS, SMB."
  }

  validation {
    condition     = contains(["Performance", "Capacity", "Premium", "standard", "advance"], var.spec.storage_type)
    error_message = "storage_type must be one of: Performance, Capacity, Premium (standard NAS), standard, advance (extreme NAS)."
  }
}
