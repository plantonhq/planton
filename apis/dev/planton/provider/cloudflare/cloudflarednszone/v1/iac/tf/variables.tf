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
  description = "CloudflareDnsZoneSpec defines a Cloudflare DNS zone"
  type = object({
    # (Required) The fully qualified domain name of the DNS zone (e.g., "example.com").
    zone_name = string

    # (Required) The Cloudflare account identifier under which to create the zone.
    account_id = string

    # (Optional) Whether the zone is created paused (DNS-only, no proxy/CDN/WAF).
    paused = optional(bool, false)

    # (Optional) The zone deployment type: "full", "partial", "secondary", "internal".
    # Defaults to "full".
    type = optional(string, "full")

    # (Optional) Custom (vanity) name servers (Business/Enterprise plans).
    vanity_name_servers = optional(list(string), [])

    # (Optional) DNS records managed alongside the zone (the lean inline model).
    records = optional(list(object({
      name     = string
      type     = string
      content  = string
      proxied  = optional(bool, false)
      ttl      = optional(number, 1)
      priority = optional(number, 0)
      comment  = optional(string, "")
    })), [])

    # (Optional) Zone-wide DNS settings. Unset fields keep Cloudflare's defaults.
    dns_settings = optional(object({
      flatten_all_cnames  = optional(bool)
      foundation_dns      = optional(bool)
      multi_provider      = optional(bool)
      secondary_overrides = optional(bool)
      ns_ttl              = optional(number)
      zone_mode           = optional(string)
      soa = optional(object({
        expire  = optional(number)
        min_ttl = optional(number)
        mname   = optional(string)
        refresh = optional(number)
        retry   = optional(number)
        rname   = optional(string)
        ttl     = optional(number)
      }))
      nameservers = optional(object({
        ns_set = optional(number)
        type   = optional(string)
      }))
      internal_dns = optional(object({
        # StringValueOrRef is flattened to a plain string by the tfvars converter.
        reference_zone_id = optional(string)
      }))
    }))

    # (Optional) DNSSEC configuration. When enabled, Cloudflare signs the zone and
    # the DS material is published as stack outputs.
    dnssec = optional(object({
      enabled      = optional(bool, false)
      multi_signer = optional(bool)
      presigned    = optional(bool)
      use_nsec3    = optional(bool)
    }))
  })
}
