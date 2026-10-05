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
  description = "AwsEbsVolume specification"
  type = object({
    region = string
    availability_zone = optional(string, "")
    type = optional(string, "")
    size_gb = optional(number, 0)
    snapshot_id = optional(string, "")
    encrypted = optional(bool, false)
    kms_key_id = optional(string, "")
    iops = optional(number, 0)
    throughput_mibps = optional(number, 0)
    multi_attach_enabled = optional(bool, false)
    final_snapshot = optional(bool, false)
    volume_initialization_rate = optional(number, 0)
    copy_from = optional(object({
      source_volume_id = string
    }))
    attachments = optional(list(object({
      device_name = string
      instance_id = string
      force_detach = optional(bool, false)
      skip_destroy = optional(bool, false)
      stop_instance_before_detaching = optional(bool, false)
    })), [])
  })
}