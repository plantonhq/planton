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
  description = "OpenStackInstanceSpec defines the configuration for a compute instance"
  type = object({
    # (Optional) Human-readable flavor name (e.g., "m1.medium"). Mutually exclusive with flavor_id.
    flavor_name = optional(string, "")

    # (Optional) Flavor UUID. Mutually exclusive with flavor_name.
    flavor_id = optional(string, "")

    # (Optional) Image name (e.g., "ubuntu-22.04"). Not needed if booting from block_device.
    image_name = optional(string, "")

    # (Optional) Image UUID. Not needed if booting from block_device.
    image_id = optional(string, "")

    # (Optional) SSH keypair name. Supports StringValueOrRef pattern.
    key_pair = optional(object({
      value = string
    }))

    # (Required) Network attachments. At least one required.
    networks = list(object({
      # Network UUID (mutually exclusive with port). Supports StringValueOrRef.
      uuid = optional(object({
        value = string
      }))
      # Port UUID (mutually exclusive with uuid). Supports StringValueOrRef.
      port = optional(object({
        value = string
      }))
      # Specific IPv4 address to request.
      fixed_ip_v4 = optional(string, "")
      # Mark as access network.
      access_network = optional(bool, false)
    }))

    # (Optional) Security group names. Supports StringValueOrRef.
    security_groups = optional(list(object({
      value = string
    })), [])

    # (Optional) Block device mappings for boot-from-volume and additional storage.
    block_device = optional(list(object({
      source_type          = string
      uuid                 = optional(string, "")
      destination_type     = optional(string, "")
      boot_index           = optional(number, -1)
      volume_size          = optional(number, 0)
      delete_on_termination = optional(bool, false)
      volume_type          = optional(string, "")
    })), [])

    # (Optional) Cloud-init or user data script.
    user_data = optional(string, "")

    # (Optional) Instance metadata key-value pairs.
    metadata = optional(map(string), {})

    # (Optional) Enable config drive for metadata delivery.
    config_drive = optional(bool)

    # (Optional) Server group UUID for placement control. Supports StringValueOrRef.
    server_group_id = optional(object({
      value = string
    }))

    # (Optional) Availability zone.
    availability_zone = optional(string, "")

    # (Optional) Tags for the instance.
    tags = optional(list(string), [])

    # (Optional) Region override.
    region = optional(string, "")
  })
}
