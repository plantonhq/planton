variable "metadata" {
  description = "Catalog object metadata"
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
  description = "DigitalOceanVolume specification"
  type = object({
    volume_name              = string
    description              = optional(string, "")
    region                   = string
    size_gib                 = number
    filesystem_type          = optional(string, "")
    initial_filesystem_label = optional(string, "")
    snapshot_id              = optional(string, "")
    tags                     = optional(list(string), [])
  })
}