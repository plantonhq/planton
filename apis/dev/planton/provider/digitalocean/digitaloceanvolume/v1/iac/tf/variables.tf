variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the DigitalOcean Volume"
  type = object({
    volume_name     = string
    description     = optional(string, "")
    region          = string
    size_gib        = number
    filesystem_type = optional(string, "NONE")  # NONE, EXT4, or XFS
    snapshot_id     = optional(string, "")
    tags            = optional(list(string), [])
  })
}