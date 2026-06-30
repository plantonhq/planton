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
  description = "spec"
  type = object({

    # The DNS zone name (e.g., "example.com"). Do not include a trailing dot.
    zone_name = string

    # The Azure Resource Group in which to create the DNS zone.
    # This resource group must exist in the target subscription.
    resource_group = string

    # (Optional) DNS records to pre-populate in the zone.
    # Each record includes type, name, values, and TTL.
    records = optional(list(object({

      # Required.** The DNS record type (e.g., A, AAAA, CNAME, TXT, MX, CAA, SRV, NS, PTR).
      record_type = string

      # Required.** The name of the DNS record.
      # This can be a fully qualified domain name (ending with a dot, e.g., "www.example.com.")
      # or a relative name within the zone (e.g., "www" for "www.example.com").
      # An empty name "@" denotes the zone root.
      name = string

      # The list of values for the DNS record.
      # For example, IP addresses for A/AAAA records, or hostnames for CNAME records.
      values = list(string)

      # The Time To Live (TTL) for the DNS record, in seconds.
      ttl_seconds = optional(number, 60)
    })), [])
  })
}
