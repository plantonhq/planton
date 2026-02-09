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
  description = "OpenStackDnsZoneSpec defines the configuration for a Designate DNS zone"
  type = object({
    # (Required) The DNS domain name for the zone (e.g., "example.com").
    domain_name = string

    # (Optional) Email address of the zone administrator.
    email = optional(string, "")

    # (Optional) Human-readable description of the DNS zone.
    description = optional(string, "")

    # (Optional) Default TTL (in seconds) for records in this zone.
    ttl = optional(number)

    # (Optional) Zone type: "PRIMARY" or "SECONDARY".
    type = optional(string, "")

    # (Optional) Master nameserver addresses for SECONDARY zones.
    masters = optional(list(string), [])

    # (Optional) Inline DNS records to create alongside the zone.
    records = optional(list(object({
      record_type = number
      record_name = string
      values      = list(string)
      ttl         = optional(number, 60)
    })), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
