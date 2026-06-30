variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name = string,
    id = optional(string),
    org = optional(string),
    env = optional(string),
    labels = optional(map(string)),
    tags = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}


variable "spec" {
  description = "Specification for the DigitalOcean Firewall"
  type = object({
    name = string
    
    # Inbound rules: traffic allowed *to* Droplets
    inbound_rules = optional(list(object({
      protocol                   = string
      port_range                 = optional(string)
      source_addresses           = optional(list(string))
      source_droplet_ids         = optional(list(number))
      source_tags                = optional(list(string))
      source_kubernetes_ids      = optional(list(string))
      source_load_balancer_uids  = optional(list(string))
    })), [])
    
    # Outbound rules: traffic allowed *from* Droplets
    outbound_rules = optional(list(object({
      protocol                        = string
      port_range                      = optional(string)
      destination_addresses           = optional(list(string))
      destination_droplet_ids         = optional(list(number))
      destination_tags                = optional(list(string))
      destination_kubernetes_ids      = optional(list(string))
      destination_load_balancer_uids  = optional(list(string))
    })), [])
    
    # The Droplet IDs to which this firewall is applied (max 10)
    droplet_ids = optional(list(number), [])
    
    # The names of Droplet tags to which this firewall is applied
    tags = optional(list(string), [])
  })
}