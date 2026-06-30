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
  description = "OpenStackVolumeSpec defines the configuration for a Cinder block storage volume"
  type = object({
    # (Optional) Human-readable description of the volume.
    description = optional(string, "")

    # (Required) Volume size in gigabytes. Must be > 0.
    size = number

    # (Optional) Cinder volume type (e.g., "SSD", "HDD").
    volume_type = optional(string, "")

    # (Optional) Availability zone. ForceNew.
    availability_zone = optional(string, "")

    # (Optional) Snapshot UUID to restore from. ForceNew. Mutually exclusive with source_vol_id and image_id.
    snapshot_id = optional(string, "")

    # (Optional) Volume UUID to clone from. ForceNew. Mutually exclusive with snapshot_id and image_id.
    source_vol_id = optional(string, "")

    # (Optional) Image ID to create bootable volume from.
    # Supports StringValueOrRef pattern - use {value: "image-uuid"} for literal values.
    # ForceNew. Mutually exclusive with snapshot_id and source_vol_id.
    image_id = optional(object({
      value = string
    }))

    # (Optional) Key-value metadata for the volume.
    metadata = optional(map(string), {})

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
