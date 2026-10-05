variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsEbsSnapshot specification"
  type = object({
    region = string
    volume_id = optional(string, "")
    copy_from = optional(object({
      source_snapshot_id = string
      source_region = string
      encrypted = optional(bool, false)
      kms_key_id = optional(string, "")
      completion_duration_minutes = optional(number, 0)
    }))
    import_from = optional(object({
      disk_container = object({
        format = optional(string, "")
        description = optional(string, "")
        url = optional(string, "")
        s3_bucket = optional(string, "")
        s3_key = optional(string, "")
      })
      role_name = optional(string, "")
      encrypted = optional(bool, false)
      kms_key_id = optional(string, "")
    }))
    description = optional(string, "")
    storage_tier = optional(string, "")
    permanent_restore = optional(bool, false)
    temporary_restore_days = optional(number, 0)
    fast_restore_availability_zones = optional(list(string), [])
    share_with_account_ids = optional(list(string), [])
  })
}