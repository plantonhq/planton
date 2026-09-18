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
  description = "AwsAlb specification"
  type = object({
    # The AWS region where the ALB is created.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The subnets the ALB places its nodes in -- at least two, in different
    # Availability Zones (an AWS requirement that also buys zonal redundancy).
    # Public subnets for internet-facing ALBs, private for internal ones.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnets = list(string)

    # Security groups controlling traffic to and from the ALB. When omitted,
    # AWS attaches the VPC's default security group -- fine for a first boot,
    # wrong for production; attach explicit groups that open exactly the
    # listener ports.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_groups = optional(list(string), [])

    # When true, the ALB is internal (reachable only inside the VPC); when
    # false (default), it is internet-facing with public DNS. Immutable:
    # changing the scheme replaces the load balancer.
    internal = optional(bool, false)

    # The address family of the ALB's nodes. Valid values: "ipv4" (default),
    # "dualstack" (IPv4 + IPv6), "dualstack-without-public-ipv4" (public IPv6
    # with private IPv4 -- avoids public-IPv4 charges for IPv6-capable
    # clients).
    ip_address_type = optional(string, "")

    # IPAM pool that allocates the ALB's public IPv4 addresses (instead of
    # AWS-assigned addresses) -- lets organizations front the ALB with their
    # own BYOIP ranges managed in VPC IPAM. Supply a literal ipam-pool-id;
    # there is no IPAM pool catalog kind yet. Only meaningful for
    # internet-facing ALBs with an IPv4 address family; removing it moves the
    # ALB back to AWS-assigned addresses in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ipv4_ipam_pool_id = optional(string, "")

    # Reserved load balancer capacity, in Load Balancer Capacity Units (LCUs).
    # Pre-provisions the ALB for a known traffic surge (product launch, ticket
    # sale) instead of waiting for organic scaling. BILLS for the reserved
    # LCUs while set, on top of normal LCU usage -- set it for the event
    # window, then remove it (0 / unset releases the reservation). Minimum and
    # maximum reservable units depend on the account's service quotas.
    minimum_load_balancer_capacity_units = optional(number, 0)

    # Prevents deletion of the ALB while enabled. Recommended for production:
    # deleting an ALB silently orphans every listener and rule attached to it.
    delete_protection_enabled = optional(bool, false)

    # Seconds an idle connection stays open. Range 1-4000. AWS default: 60.
    # Raise it above the application's slowest response time to avoid 504s on
    # long-running requests; keep it above any upstream keep-alive interval.
    idle_timeout_seconds = optional(number, 0)

    # Seconds an HTTP client connection may stay alive across requests.
    # Range 60-604800. AWS default: 3600.
    client_keep_alive_seconds = optional(number, 0)

    # Whether HTTP/2 is offered to clients. AWS default: true. Optional rather
    # than plain bool so that false ("disable HTTP/2") is distinguishable from
    # unset ("keep the AWS default").
    http2_enabled = optional(bool)

    # What happens to requests when an attached WAF is unreachable: when true,
    # requests pass through ("fail open"); when false (AWS default), they are
    # rejected ("fail closed"). A deliberate availability-versus-security call.
    # Only meaningful when web_acl_arn attaches a WAF to this ALB.
    waf_fail_open_enabled = optional(bool, false)

    # The REGIONAL-scope WAFv2 web ACL protecting this ALB, by ARN — the
    # modules create the web-ACL association alongside the load balancer.
    # An ALB has at most one web ACL; leave unset for no WAF. The web ACL
    # must live in the same region as the ALB. Can reference an
    # AwsWafWebAcl resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    web_acl_arn = optional(string, "")

    # Allows Amazon Application Recovery Controller to shift this ALB's
    # traffic away from an impaired Availability Zone.
    zonal_shift_enabled = optional(bool, false)

    # Drop request headers with names that are not valid HTTP header fields
    # instead of forwarding them. Hardens against header-smuggling tricks;
    # AWS default: false.
    drop_invalid_header_fields = optional(bool, false)

    # Forward the client's original Host header to the target unchanged
    # instead of rewriting it to the target address. AWS default: false.
    preserve_host_header = optional(bool, false)

    # Append the client's source port to the X-Forwarded-For header. AWS
    # default: false.
    xff_client_port_enabled = optional(bool, false)

    # How the ALB handles the X-Forwarded-For header. Valid values: "append"
    # (AWS default -- add the client IP), "preserve" (pass it through
    # untouched), "remove" (strip it). "preserve"/"remove" matter when the ALB
    # sits behind another proxy layer whose XFF chain must win.
    xff_header_processing_mode = optional(string, "")

    # Protection level against HTTP desync (request-smuggling) attacks.
    # Valid values: "monitor" (classify only), "defensive" (AWS default --
    # block ambiguous requests likely to poison caches), "strictest" (block
    # everything not RFC 7230 compliant).
    desync_mitigation_mode = optional(string, "")

    # Inject the negotiated TLS version and cipher suite as request headers
    # (x-amzn-tls-version, x-amzn-tls-cipher-suite) toward targets, for
    # applications that audit their clients' TLS posture. AWS default: false.
    tls_version_and_cipher_suite_headers_enabled = optional(bool, false)

    # Access logs: one entry per request, delivered to S3. The bucket must
    # carry the ELB log-delivery bucket policy. When omitted, access logging
    # is off.
    access_logs = optional(object({
      # The S3 bucket receiving the logs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      bucket = string

      # Key prefix inside the bucket (e.g. "alb/production"), for sharing one
      # bucket across several load balancers or log types.
      prefix = optional(string, "")
    }))

    # Connection logs: one entry per client connection (TLS handshake
    # details, client address) -- the place TLS negotiation failures that
    # never become requests show up. Delivered to S3; when omitted,
    # connection logging is off.
    connection_logs = optional(object({
      # The S3 bucket receiving the logs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      bucket = string

      # Key prefix inside the bucket (e.g. "alb/production"), for sharing one
      # bucket across several load balancers or log types.
      prefix = optional(string, "")
    }))

    # Health-check logs: one entry per health-check result, for debugging
    # flapping targets without packet captures. Delivered to S3; when
    # omitted, health-check logging is off.
    health_check_logs = optional(object({
      # The S3 bucket receiving the logs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      bucket = string

      # Key prefix inside the bucket (e.g. "alb/production"), for sharing one
      # bucket across several load balancers or log types.
      prefix = optional(string, "")
    }))

    # Optional Route53 DNS: alias A records pointing the given hostnames at
    # the ALB. Alias records are preferred over CNAMEs because they work at
    # the zone apex, cost nothing per query, and inherit the ALB's health.
    dns = optional(object({
      # When true, creates Route53 alias records for the ALB.
      enabled = optional(bool, false)

      # Route53 hosted zone ID where alias records are created.
      # Required when enabled is true.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      route53_zone_id = optional(string, "")

      # Domain names that will point to the ALB via Route53 alias records.
      # Each hostname gets its own A record aliased to the ALB's DNS name.
      hostnames = optional(list(string), [])
    }))
  })
}
