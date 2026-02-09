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
  description = "OpenStackVolumeAttachSpec defines the configuration for a volume attachment"
  type = object({
    # (Required) The ID of the compute instance to attach the volume to.
    # Supports StringValueOrRef pattern - use {value: "instance-uuid"} for literal values.
    instance_id = object({
      value = string
    })

    # (Required) The ID of the Cinder volume to attach.
    # Supports StringValueOrRef pattern - use {value: "volume-uuid"} for literal values.
    volume_id = object({
      value = string
    })

    # (Optional) Device path where the volume appears inside the instance.
    # Example: "/dev/vdb". Computed by OpenStack if omitted.
    device = optional(string, "")

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
