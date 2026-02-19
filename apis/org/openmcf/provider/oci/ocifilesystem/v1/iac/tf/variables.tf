variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "OciFileSystem specification"
  type = object({
    compartment_id = object({
      value = string
    })

    availability_domain = string

    display_name = optional(string, "")

    kms_key_id = optional(object({
      value = string
    }), null)

    filesystem_snapshot_policy_id = optional(object({
      value = string
    }), null)

    mount_target = object({
      subnet_id = object({
        value = string
      })

      display_name    = optional(string, "")
      hostname_label  = optional(string, "")
      ip_address      = optional(string, "")

      nsg_ids = optional(list(object({
        value = string
      })), [])

      requested_throughput = optional(number, 0)
      max_fs_stat_bytes    = optional(number, 0)
      max_fs_stat_files    = optional(number, 0)
    })

    exports = list(object({
      path = string

      export_options = optional(list(object({
        source                        = string
        access                        = optional(string, "")
        identity_squash               = optional(string, "")
        require_privileged_source_port = optional(bool, false)
        is_anonymous_access_allowed   = optional(bool, false)
        anonymous_uid                 = optional(number, 0)
        anonymous_gid                 = optional(number, 0)
      })), [])
    }))
  })
}
