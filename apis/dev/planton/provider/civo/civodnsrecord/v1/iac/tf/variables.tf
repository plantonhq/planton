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
  description = "CivoDnsRecordSpec defines the configuration for creating a DNS record"
  type = object({
    # (Required) The Civo Zone ID where this DNS record will be created
    zone_id = string

    # (Required) The name of the DNS record (e.g., "www", "api", "@" for root)
    name = string

    # (Required) The type of DNS record
    # Valid values: "A", "AAAA", "CNAME", "MX", "TXT", "SRV", "NS"
    type = string

    # (Required) The value/target of the DNS record
    value = string

    # (Optional) Time to live (TTL) in seconds
    # Valid range: 60-86400 seconds
    # Defaults to 3600 (1 hour)
    ttl = optional(number, 3600)

    # (Optional) Priority for MX and SRV records
    # Lower values indicate higher priority
    # Range: 0-65535
    priority = optional(number, 0)
  })
}
