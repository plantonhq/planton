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
  description = "DigitalOceanDroplet specification"
  type = object({
    droplet_name = string
    region = optional(string, "")
    size = string
    image = string
    vpc = optional(string, "")
    enable_ipv6 = optional(bool, false)
    enable_backups = optional(bool, false)
    volume_ids = optional(list(string), [])
    tags = optional(list(string), [])
    user_data = optional(string, "")
    monitoring = optional(bool, false)
    ssh_keys = optional(list(string), [])
    backup_policy = optional(object({
      plan = optional(string, "")
      weekday = optional(string, "")
      hour = optional(number, 0)
    }))
    droplet_agent = optional(bool)
    graceful_shutdown = optional(bool, false)
    resize_disk = optional(bool)
    public_networking = optional(bool)
    gpu_partition_mode = optional(string, "")
  })
}