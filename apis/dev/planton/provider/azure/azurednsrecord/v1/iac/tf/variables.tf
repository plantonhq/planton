variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "AzureDnsRecord specification"
  type = object({
    # The Azure Resource Group where the DNS Zone exists.
    resource_group = string

    # The name of the DNS Zone where this record will be created.
    # Can be a literal value or resolved from an AzureDnsZone resource.
    zone_name = string

    # The DNS record type to create (A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA).
    type = string

    # The name of the DNS record (relative to the zone).
    # Use "@" for zone apex or specify a subdomain name.
    name = string

    # The values/targets for the DNS record.
    values = list(string)

    # Time to live (TTL) for the DNS record in seconds.
    # Default: 300 seconds (5 minutes).
    ttl_seconds = optional(number, 300)

    # MX record specific: Priority value for mail exchange records.
    # Only applicable when record_type is MX.
    mx_priority = optional(number, 10)
  })
}
