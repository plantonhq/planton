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
  description = "Specification for the GCP DNS Record"
  type = object({

    # The ID of the GCP project where the Managed Zone exists.
    # Supports StringValueOrRef pattern - use {value: "project-id"} for literal values.
    project_id = object({
      value = string
    })

    # The name of the Managed Zone where this DNS record will be created.
    # Supports StringValueOrRef pattern - use {value: "zone-name"} for literal values.
    managed_zone = object({
      value = string
    })

    # The DNS record type to create.
    # Supported types: A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA, SOA.
    type = string

    # The fully qualified domain name for this record.
    # Must end with a trailing dot to indicate FQDN.
    name = string

    # The values/targets for the DNS record.
    # Multiple values create a round-robin record set.
    values = list(string)

    # Time to live (TTL) for the DNS record in seconds.
    # Default: 300 seconds (5 minutes).
    ttl_seconds = optional(number, 300)
  })
}
