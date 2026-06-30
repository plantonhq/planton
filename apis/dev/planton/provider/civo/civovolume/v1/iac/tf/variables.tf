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
  description = "CivoVolume specification"
  type = object({
    # The name of the volume (lowercase letters, numbers, and hyphens only)
    volume_name = string

    # The Civo region where the volume will be created
    region = string

    # The size of the volume in GiB (1-16000)
    size_gib = number

    # The initial filesystem to format the volume with (optional)
    # Values: "NONE" (default), "EXT4", "XFS"
    # Note: Civo provider doesn't expose filesystem formatting,
    # so this is informational only
    filesystem_type = optional(string, "NONE")

    # An optional snapshot ID to create this volume from (optional)
    # Note: Snapshot functionality is not available on public Civo cloud
    snapshot_id = optional(string, "")

    # A list of tags to apply to the volume (optional)
    # Note: Civo Volume provider doesn't currently support tags
    tags = optional(list(string), [])
  })
}