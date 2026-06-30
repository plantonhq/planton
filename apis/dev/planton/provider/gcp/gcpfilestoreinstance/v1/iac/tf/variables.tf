variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = optional(string, "")
  })
  default = {}
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
    id   = optional(string, "")
    org  = optional(string, "")
    env  = optional(string, "")
  })
}

variable "spec" {
  description = "GcpFilestoreInstance specification"
  type = object({
    project_id = object({
      value = string
    })
    instance_name = string
    location      = string
    tier          = string
    description   = optional(string, "")
    protocol      = optional(string, "")
    kms_key_name  = optional(object({ value = string }), null)

    deletion_protection_enabled = optional(bool, false)
    deletion_protection_reason  = optional(string, "")

    file_share = object({
      name        = string
      capacity_gb = number
      nfs_export_options = optional(list(object({
        ip_ranges   = optional(list(string), [])
        access_mode = optional(string, "")
        squash_mode = optional(string, "")
        anon_uid    = optional(number, null)
        anon_gid    = optional(number, null)
      })), [])
    })

    network_config = object({
      network = object({
        value = string
      })
      connect_mode      = optional(string, "")
      reserved_ip_range = optional(string, "")
    })

    performance_config = optional(object({
      fixed_iops = optional(object({
        max_iops = number
      }), null)
      iops_per_tb = optional(object({
        max_iops_per_tb = number
      }), null)
    }), null)
  })
}
