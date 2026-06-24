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
  description = "CloudflareDnsRecordSpec defines a single DNS record in a Cloudflare zone"
  type = object({
    # (Required) The Cloudflare Zone ID where this DNS record will be created.
    # StringValueOrRef is flattened to a plain string by the tfvars converter.
    zone_id = optional(string)

    # (Required) The record name (e.g., "www", "api", "@" for the zone apex).
    name = string

    # (Required) The DNS record type. Determines whether the value comes from
    # content (simple types) or the matching data block (structured types).
    type = string

    # (Optional) Presentation-format value for simple record types
    # (A/AAAA/CNAME/MX/NS/PTR/TXT/OPENPGPKEY). Empty for structured types.
    content = optional(string, "")

    # (Optional) Whether the record is proxied through Cloudflare (orange cloud).
    # Only applicable to A, AAAA, and CNAME records. Defaults to false.
    proxied = optional(bool, false)

    # (Optional) Time to live (TTL) in seconds. 1 = automatic, or 30-86400.
    ttl = optional(number, 1)

    # (Optional) Priority for MX records. Range: 0-65535.
    priority = optional(number, 0)

    # (Optional) Comment/note for the record.
    comment = optional(string, "")

    # (Optional) Custom tags for the record.
    tags = optional(list(string), [])

    # (Optional) Record-level settings (only affect proxied records).
    settings = optional(object({
      ipv4_only     = optional(bool)
      ipv6_only     = optional(bool)
      flatten_cname = optional(bool)
    }))

    # (Optional) Structured data for non-simple record types. Exactly one case
    # is set; locals.tf flattens it into the provider's single data object.
    data = optional(object({
      caa = optional(object({
        flags = optional(number)
        tag   = optional(string)
        value = optional(string)
      }))
      cert = optional(object({
        type        = optional(number)
        key_tag     = optional(number)
        algorithm   = optional(number)
        certificate = optional(string)
      }))
      dnskey = optional(object({
        flags      = optional(number)
        protocol   = optional(number)
        algorithm  = optional(number)
        public_key = optional(string)
      }))
      ds = optional(object({
        key_tag     = optional(number)
        algorithm   = optional(number)
        digest_type = optional(number)
        digest      = optional(string)
      }))
      https = optional(object({
        priority = optional(number)
        target   = optional(string)
        value    = optional(string)
      }))
      loc = optional(object({
        lat_direction  = optional(string)
        lat_degrees    = optional(number)
        lat_minutes    = optional(number)
        lat_seconds    = optional(number)
        long_direction = optional(string)
        long_degrees   = optional(number)
        long_minutes   = optional(number)
        long_seconds   = optional(number)
        altitude       = optional(number)
        size           = optional(number)
        precision_horz = optional(number)
        precision_vert = optional(number)
      }))
      naptr = optional(object({
        order       = optional(number)
        preference  = optional(number)
        flags       = optional(string)
        service     = optional(string)
        regex       = optional(string)
        replacement = optional(string)
      }))
      smimea = optional(object({
        usage         = optional(number)
        selector      = optional(number)
        matching_type = optional(number)
        certificate   = optional(string)
      }))
      srv = optional(object({
        priority = optional(number)
        weight   = optional(number)
        port     = optional(number)
        target   = optional(string)
      }))
      sshfp = optional(object({
        algorithm   = optional(number)
        type        = optional(number)
        fingerprint = optional(string)
      }))
      svcb = optional(object({
        priority = optional(number)
        target   = optional(string)
        value    = optional(string)
      }))
      tlsa = optional(object({
        usage         = optional(number)
        selector      = optional(number)
        matching_type = optional(number)
        certificate   = optional(string)
      }))
      uri = optional(object({
        priority = optional(number)
        weight   = optional(number)
        target   = optional(string)
      }))
    }))
  })
}
