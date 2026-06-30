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
  description = "HetznerCloudDnsZone specification"
  type = object({
    domain_name       = string
    mode              = string
    ttl               = optional(number)
    delete_protection = optional(bool)

    primary_nameservers = optional(list(object({
      address        = string
      port           = optional(number)
      tsig_algorithm = optional(string)
      tsig_key       = optional(string, "")
    })))

    record_sets = optional(list(object({
      name = string
      type = string
      ttl  = optional(number)
      records = list(object({
        value   = string
        comment = optional(string)
      }))
    })))
  })
}

variable "hcloud_token" {
  description = "Hetzner Cloud API token for authentication"
  type        = string
  sensitive   = true
}
