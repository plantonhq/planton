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
  description = "CloudflareDnsRecordSpec defines the configuration for creating a DNS record"
  type = object({
    # (Required) The Cloudflare Zone ID where this DNS record will be created.
    # StringValueOrRef is flattened to a plain string by the tfvars converter.
    zone_id = optional(string)

    # (Required) The name of the DNS record (e.g., "www", "api", "@" for root)
    name = string

    # (Required) The type of DNS record
    # Valid values: "A", "AAAA", "CNAME", "MX", "TXT", "SRV", "NS", "CAA"
    type = string

    # (Required) The value/target of the DNS record
    value = string

    # (Optional) Whether the record is proxied through Cloudflare (orange cloud)
    # Only applicable to A, AAAA, and CNAME records
    # Defaults to false
    proxied = optional(bool, false)

    # (Optional) Time to live (TTL) in seconds
    # 1 = automatic, or 60-86400 seconds
    # Defaults to 1 (automatic)
    ttl = optional(number, 1)

    # (Optional) Priority for MX and SRV records
    # Lower values indicate higher priority
    # Range: 0-65535
    priority = optional(number, 0)

    # (Optional) Comment/note for the DNS record (max 100 characters)
    comment = optional(string, "")
  })
}
