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
  description = "CloudflareZeroTrustDnsLocation specification"
  type = object({
    account_id             = string
    name                   = string
    client_default         = optional(bool)
    ecs_support            = optional(bool)
    dns_destination_ips_id = optional(string, "")
    endpoints = optional(object({
      doh = object({
        enabled       = optional(bool)
        require_token = optional(bool)
        networks = optional(list(object({
          network = string
        })), [])
      })
      dot = object({
        enabled = optional(bool)
        networks = optional(list(object({
          network = string
        })), [])
      })
      ipv4 = object({
        enabled = optional(bool)
      })
      ipv6 = object({
        enabled = optional(bool)
        networks = optional(list(object({
          network = string
        })), [])
      })
    }))
    networks = optional(list(object({
      network = string
    })), [])
    max_ttl = optional(object({
      mode     = string
      ttl_secs = optional(number)
    }))
  })
}