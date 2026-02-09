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
  description = "OpenStackImageSpec defines the configuration for a Glance image"
  type = object({
    # (Required) Container format: "bare", "ovf", "aki", "ari", "ami", "ova", "docker", "compressed"
    container_format = string

    # (Required) Disk format: "raw", "vhd", "vhdx", "vmdk", "vdi", "iso", "ploop", "qcow2", "aki", "ari", "ami"
    disk_format = string

    # (Optional) URL to download image data from.
    image_source_url = optional(string, "")

    # (Optional) Minimum disk size in GB. Default: 0.
    min_disk_gb = optional(number, 0)

    # (Optional) Minimum RAM in MB. Default: 0.
    min_ram_mb = optional(number, 0)

    # (Optional) Prevent deletion. Default: false.
    protected = optional(bool, false)

    # (Optional) Hide from default listings. Default: false.
    hidden = optional(bool, false)

    # (Optional) Tags for filtering.
    tags = optional(list(string), [])

    # (Optional) Access control: "private", "shared", "community", "public". Default: "private".
    visibility = optional(string, "private")

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
