variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareDnsZone specification"
  type = object({
    zone_name = string
    account_id = string
    paused = optional(bool, false)
    records = optional(list(object({
      name = string
      type = string
      content = optional(string, "")
      proxied = optional(bool, false)
      ttl = optional(number, 0)
      priority = optional(number, 0)
      comment = optional(string, "")
      caa = optional(object({
        flags = optional(number, 0)
        tag = string
        value = string
      }))
      cert = optional(object({
        type = optional(number, 0)
        key_tag = optional(number, 0)
        algorithm = optional(number, 0)
        certificate = string
      }))
      dnskey = optional(object({
        flags = optional(number, 0)
        protocol = optional(number, 0)
        algorithm = optional(number, 0)
        public_key = string
      }))
      ds = optional(object({
        key_tag = optional(number, 0)
        algorithm = optional(number, 0)
        digest_type = optional(number, 0)
        digest = string
      }))
      https = optional(object({
        priority = optional(number, 0)
        target = string
        value = optional(string, "")
      }))
      loc = optional(object({
        lat_direction = optional(string, "")
        lat_degrees = optional(number, 0)
        lat_minutes = optional(number, 0)
        lat_seconds = optional(number, 0)
        long_direction = optional(string, "")
        long_degrees = optional(number, 0)
        long_minutes = optional(number, 0)
        long_seconds = optional(number, 0)
        altitude = optional(number, 0)
        size = optional(number, 0)
        precision_horz = optional(number, 0)
        precision_vert = optional(number, 0)
      }))
      naptr = optional(object({
        order = optional(number, 0)
        preference = optional(number, 0)
        flags = optional(string, "")
        service = optional(string, "")
        regex = optional(string, "")
        replacement = optional(string, "")
      }))
      smimea = optional(object({
        usage = optional(number, 0)
        selector = optional(number, 0)
        matching_type = optional(number, 0)
        certificate = string
      }))
      srv = optional(object({
        priority = optional(number, 0)
        weight = optional(number, 0)
        port = optional(number, 0)
        target = string
      }))
      sshfp = optional(object({
        algorithm = optional(number, 0)
        type = optional(number, 0)
        fingerprint = string
      }))
      svcb = optional(object({
        priority = optional(number, 0)
        target = string
        value = optional(string, "")
      }))
      tlsa = optional(object({
        usage = optional(number, 0)
        selector = optional(number, 0)
        matching_type = optional(number, 0)
        certificate = string
      }))
      uri = optional(object({
        priority = optional(number, 0)
        weight = optional(number, 0)
        target = string
      }))
      tags = optional(list(string), [])
      settings = optional(object({
        ipv4_only = optional(bool, false)
        ipv6_only = optional(bool, false)
        flatten_cname = optional(bool, false)
      }))
      private_routing = optional(bool, false)
    })), [])
    type = optional(string, "")
    vanity_name_servers = optional(list(string), [])
    dns_settings = optional(object({
      flatten_all_cnames = optional(bool, false)
      foundation_dns = optional(bool, false)
      multi_provider = optional(bool, false)
      secondary_overrides = optional(bool, false)
      ns_ttl = optional(number, 0)
      zone_mode = optional(string, "")
      soa = optional(object({
        expire = optional(number, 0)
        min_ttl = optional(number, 0)
        mname = optional(string, "")
        refresh = optional(number, 0)
        retry = optional(number, 0)
        rname = optional(string, "")
        ttl = optional(number, 0)
      }))
      nameservers = optional(object({
        ns_set = optional(number, 0)
        type = optional(string, "")
      }))
      internal_dns = optional(object({
        reference_zone_id = optional(string, "")
      }))
    }))
    dnssec = optional(object({
      enabled = optional(bool, false)
      multi_signer = optional(bool, false)
      presigned = optional(bool, false)
      use_nsec3 = optional(bool, false)
    }))
    hold = optional(object({
      enabled = optional(bool, false)
      include_subdomains = optional(bool, false)
      hold_after = optional(string, "")
    }))
    subscription = optional(object({
      rate_plan = optional(string, "")
      frequency = optional(string, "")
      scope = optional(string, "")
    }))
  })
}