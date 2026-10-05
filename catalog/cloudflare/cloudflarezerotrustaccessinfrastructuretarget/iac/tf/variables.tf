variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareZeroTrustAccessInfrastructureTarget specification"
  type = object({
    account_id = string
    hostname   = string
    ip = object({
      ipv4 = optional(object({
        ip_addr            = string
        virtual_network_id = optional(string, "")
      }))
      ipv6 = optional(object({
        ip_addr            = string
        virtual_network_id = optional(string, "")
      }))
    })
  })
}