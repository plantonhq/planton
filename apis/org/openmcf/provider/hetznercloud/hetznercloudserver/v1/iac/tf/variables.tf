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
  description = "HetznerCloudServer specification"
  type = object({
    server_type = string
    image       = string
    location    = string
    ssh_keys    = optional(list(string))
    user_data   = optional(string)
    placement_group_id = optional(string)
    firewall_ids       = optional(list(string))
    public_net = optional(object({
      ipv4_enabled = optional(bool)
      ipv6_enabled = optional(bool)
      ipv4         = optional(string)
      ipv6         = optional(string)
    }))
    networks = optional(list(object({
      network_id = string
      ip         = optional(string)
      alias_ips  = optional(list(string))
    })))
    backups                  = optional(bool)
    keep_disk                = optional(bool)
    delete_protection        = optional(bool)
    rebuild_protection       = optional(bool)
    shutdown_before_deletion = optional(bool)
    dns_ptr                  = optional(string)
  })
}

variable "hcloud_token" {
  description = "Hetzner Cloud API token for authentication"
  type        = string
  sensitive   = true
}
