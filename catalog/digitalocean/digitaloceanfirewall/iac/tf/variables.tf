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
  description = "DigitalOceanFirewall specification"
  type = object({
    firewall_name = string
    inbound_rules = optional(list(object({
      protocol = string
      port_range = optional(string, "")
      source_addresses = optional(list(string), [])
      source_tags = optional(list(string), [])
      source_droplet_ids = optional(list(string), [])
      source_kubernetes_ids = optional(list(string), [])
      source_load_balancer_uids = optional(list(string), [])
    })), [])
    outbound_rules = optional(list(object({
      protocol = string
      port_range = optional(string, "")
      destination_addresses = optional(list(string), [])
      destination_tags = optional(list(string), [])
      destination_droplet_ids = optional(list(string), [])
      destination_kubernetes_ids = optional(list(string), [])
      destination_load_balancer_uids = optional(list(string), [])
    })), [])
    tags = optional(list(string), [])
    droplet_ids = optional(list(string), [])
  })
}