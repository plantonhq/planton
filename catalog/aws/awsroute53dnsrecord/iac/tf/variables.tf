variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsRoute53DnsRecord specification"
  type = object({
    # The AWS region where the resource will be created.
    # Route 53 is a global service; this selects the region used for provider
    # API calls.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The Route 53 hosted zone that owns this record. Create-time immutable
    # (ForceNew). Can reference an AwsRoute53Zone resource — the default field
    # path wires to "status.outputs.zone_id".
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    zone_id = string

    # The record name (fully qualified domain name or subdomain). Create-time
    # immutable (ForceNew).
    # Examples:
    #   - "example.com" for the zone apex
    #   - "www.example.com" for a subdomain
    #   - "*.example.com" for a wildcard (catch-all subdomains)
    #   - "_dmarc.example.com" / "_sip._tcp.example.com" — underscore-prefixed
    #     labels, the convention for records that configure services rather
    #     than name hosts (DMARC/DKIM email authentication, SRV service
    #     discovery, ACME challenges); Route 53 accepts them like any label.
    # Route 53 normalizes the trailing dot automatically.
    name = string

    # The DNS record type. All Route 53 resource record set types are
    # supported:
    # - "A" / "AAAA": IPv4 / IPv6 addresses (also the alias record types).
    # - "CNAME": canonical-name redirection (not at the zone apex — use an
    #   A/AAAA alias there).
    # - "MX": mail exchangers ("<priority> <host>" values).
    # - "TXT": text data (SPF, DKIM, domain verification).
    # - "NS" / "SOA": delegation and authority records (usually managed by the
    #   zone itself; override TTLs only with care).
    # - "SRV": service locators ("<priority> <weight> <port> <target>").
    # - "PTR": reverse-DNS pointers.
    # - "CAA": certificate-authority authorization.
    # - "DS": delegation signer, for DNSSEC-signed child zone delegation.
    # - "NAPTR": name authority pointers (telephony/SIP).
    # - "SPF": legacy sender-policy type (RFC 7208 deprecates it — use TXT).
    # - "HTTPS" / "SVCB": service binding records (protocol hints, ECH).
    # - "SSHFP": SSH host key fingerprints.
    # - "TLSA": DANE TLS certificate association.
    type = string

    # Time to live in seconds — how long resolvers cache the record. Required
    # for standard records; must be omitted for alias records (the target's
    # TTL applies). AWS accepts 0 to 2147483647; an explicit 0 means "never
    # cache" (every lookup hits Route 53 — valid, but billed per query).
    # Common values: 60 (fast cutover during incidents, and the recommended
    # ceiling for health-checked records), 300 (general default), 86400
    # (static records like MX/NS).
    ttl = optional(number)

    # The record data for standard records. Format depends on type:
    #   - A: IPv4 addresses (e.g. ["192.0.2.1", "192.0.2.2"])
    #   - AAAA: IPv6 addresses (e.g. ["2001:db8::1"])
    #   - CNAME: one target hostname (e.g. ["target.example.com"])
    #   - MX: priority + mail server (e.g. ["10 mail1.example.com"])
    #   - TXT: text values (e.g. ["v=spf1 include:_spf.google.com ~all"])
    # Each value is at most 4,000 characters (AWS's per-value limit).
    # Mutually exclusive with alias_target — a record is standard or alias,
    # never both.
    values = optional(list(string), [])

    # Alias target for A/AAAA alias records: point this name at an AWS
    # resource (ALB, NLB, CloudFront, S3 website, API Gateway, another record
    # in the zone) instead of literal addresses. Works at the zone apex,
    # queries are free, and Route 53 tracks the target's address changes
    # automatically. Mutually exclusive with values (and ttl).
    alias_target = optional(object({
      # The DNS name of the target resource.
      # Example literals:
      #   - CloudFront: "d1234abcd.cloudfront.net"
      #   - ALB: "my-alb-1234567890.us-east-1.elb.amazonaws.com"
      #   - S3 website: "my-bucket.s3-website-us-east-1.amazonaws.com"
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      dns_name = string

      # The target AWS service's hosted zone ID for the target's region.
      # Example literals: CloudFront "Z2FDTNDATAQYW2" (global), ALB us-east-1
      # "Z35SXDOTRQ7X7K", S3 website us-east-1 "Z3AQBSTGFYJSTF".
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      zone_id = string

      # When true, Route 53 checks the target's own health (e.g. an ALB with no
      # healthy targets is treated as failed) before answering with this record.
      # The building block for alias-based failover.
      evaluate_target_health = optional(bool, false)
    }))

    # Routing policy for advanced traffic management. At most one policy per
    # record; records with the same name and type but different set_identifier
    # values combine into one routing group. Omit for simple routing.
    routing_policy = optional(object({
      # Weighted: split traffic across group members proportionally to their
      # weights. Blue/green deployments, canary releases, gradual migrations.
      weighted = optional(object({
        # Relative weight (0–255). Higher gets more traffic; 0 drains this record
        # (it is answered only if every group member is at 0 or unhealthy).
        weight = optional(number, 0)
      }))

      # Latency: answer with the group member whose AWS region has the lowest
      # measured latency to the resolver. Multi-region applications.
      latency = optional(object({
        # The AWS region this record's resource lives in — the region whose
        # latency measurements represent this group member.
        # Example: "us-east-1", "eu-west-1", "ap-southeast-1"
        #
        # Format-checked rather than enumerated on purpose: the provider's region
        # enum grows with every AWS region launch, and freezing today's list here
        # would reject tomorrow's regions until the next pin sweep. AWS itself
        # rejects unknown regions at change time.
        region = string
      }))

      # Failover: active-passive — the PRIMARY answers while healthy, the
      # SECONDARY takes over when the primary's health check fails.
      failover = optional(object({
        # This record's role in the failover pair:
        # - "PRIMARY": answered while its health check passes.
        # - "SECONDARY": answered when the primary is unhealthy.
        failover_type = string
      }))

      # Geolocation: answer based on WHERE the user is (continent, country, or
      # US state). Compliance boundaries, localized content.
      geolocation = optional(object({
        # Two-letter continent code: "AF", "AN", "AS", "EU", "OC", "NA", "SA"
        # (a frozen set — continents don't churn). Use continent OR country, not
        # both.
        continent = optional(string, "")

        # Two-letter ISO 3166-1 alpha-2 country code (e.g. "US", "GB", "DE"), or
        # "*" for the default record that answers unmatched locations.
        country = optional(string, "")

        # Subdivision code — US states only (e.g. "CA", "NY"); requires country
        # "US". At most 3 characters (AWS's limit).
        subdivision = optional(string, "")
      }))

      # Geoproximity: answer based on DISTANCE between the user and the
      # resource, with a bias dial to grow or shrink each resource's catchment
      # area. Traffic shifting between nearby regions.
      geoproximity = optional(object({
        # The AWS region hosting the resource (for resources in AWS regions).
        # Format-checked rather than enumerated (the region list grows; AWS
        # rejects unknown regions at change time).
        aws_region = optional(string, "")

        # Latitude/longitude of the resource (for resources outside AWS).
        coordinates = optional(object({
          # Latitude in decimal degrees, "-90" to "90" (e.g. "40.71").
          latitude = string

          # Longitude in decimal degrees, "-180" to "180" (e.g. "-74.01").
          longitude = string
        }))

        # The AWS Local Zone group hosting the resource (for Local Zone
        # deployments). Example: "us-east-1-bue-1".
        local_zone_group = optional(string, "")

        # Expands (positive) or shrinks (negative) the geographic area this
        # resource answers for, from -99 to 99. Use it to shift traffic gradually
        # between neighboring locations.
        bias = optional(number, 0)
      }))

      # CIDR: answer based on the resolver's IP block, using a Route 53 CIDR
      # collection. Fine-grained per-network routing (e.g. ISP or office
      # egress ranges).
      cidr = optional(object({
        # ID of the Route 53 CIDR collection holding the location's CIDR blocks.
        collection_id = string

        # Name of the location within the collection this record answers for, or
        # "*" for the collection's default location.
        location_name = string
      }))

      # Multivalue answer: return up to eight healthy records so clients pick
      # among them — poor-man's load balancing with health-check awareness.
      # Not compatible with alias records.
      multivalue_answer = optional(object({}))
    }))

    # Health check gating this record's answers: Route 53 only serves the
    # record while the health check passes. Most commonly paired with failover
    # routing (primary/secondary), but valid with any non-simple routing
    # policy — e.g. weighted records that drop out of rotation when unhealthy.
    # Can reference an AwsRoute53HealthCheck resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    health_check_id = optional(string, "")

    # Distinguishes this record among records with the same name and type in a
    # routing group. Required by every routing policy; must be unique within
    # the group. Examples: "primary", "secondary", "us-east-1", "weight-70".
    set_identifier = optional(string, "")

    # When true, creating this record overwrites an existing record set with
    # the same name and type instead of failing. Useful when adopting records
    # that were created outside the resource graph (e.g. a zone's auto-created
    # records or a manually-created record being brought under management).
    allow_overwrite = optional(bool, false)
  })
}
