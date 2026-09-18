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
  description = "GcpDnsRecord specification"
  type = object({
    # The GCP project that hosts the managed zone.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used (ambient credentials
    # decide).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The name of the managed zone this record lives in.
    # This is the zone RESOURCE name (e.g. "prod-example-zone"), not the DNS
    # name — Cloud DNS addresses zones by resource name.
    # Can be a literal value or a reference to a GcpDnsZone resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    managed_zone = string

    # The DNS record type, uppercase (e.g. "A", "AAAA", "CNAME", "MX", "TXT",
    # "SRV", "NS", "PTR", "CAA", "SOA", "HTTPS", "SVCB", "DS", "DNSKEY",
    # "TLSA", "SSHFP", "NAPTR"). Cloud DNS accepts any record type the API
    # supports; the string is passed through as-is, so new types need no
    # spec change.
    type = string

    # The fully qualified domain name this record set applies to.
    # Must end with a trailing dot (e.g. "www.example.com.").
    # A leading "*." creates a wildcard record; leading underscores support
    # service labels such as "_dmarc" and "_acme-challenge".
    # Can be a literal FQDN or a reference — compose validation records from
    # GcpCertManagerDnsAuthorization's dns_record_name output.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    name = string

    # Static values (RRDATA) for the record set — the meaning depends on type:
    #   A: IPv4 addresses ("192.0.2.1")
    #   AAAA: IPv6 addresses ("2001:db8::1")
    #   CNAME: target hostname WITH trailing dot ("target.example.com.")
    #   MX: "priority mailserver." ("10 mail.example.com.")
    #   TXT: text values; values containing spaces need surrounding \" quotes,
    #        and single values longer than 255 characters (DKIM keys) must be
    #        split with "" between chunks.
    # Multiple values answer as a round-robin set. Mutually exclusive with
    # routing_policy. Each entry can be a literal or a reference — compose
    # validation targets from GcpCertManagerDnsAuthorization's
    # dns_record_data output.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    values = optional(list(string), [])

    # Time to live in seconds — how long resolvers cache this record.
    # Common values: 60 (fast failover), 300 (default), 3600, 86400; NS
    # records conventionally use 172800 (2 days). Lower TTLs propagate
    # changes faster at the cost of more query load.
    ttl_seconds = optional(number)

    # Query-steering policy — Cloud DNS answers each query based on weights,
    # caller geography, or target health instead of returning static values.
    # Mutually exclusive with values.
    routing_policy = optional(object({
      # Weighted round robin: traffic splits across entries in proportion to
      # their weights. Useful for canary rollouts and A/B traffic splitting.
      wrr = optional(list(object({
        # The ratio of traffic routed to this entry, relative to the sum of all
        # weights. A weight of 0 receives no traffic (useful for staging an
        # entry before shifting traffic onto it) — declared optional so the
        # explicit 0 is expressible while the field itself stays required.
        weight = number

        # Static values (RRDATA) answered for this entry.
        # If the zone has DNSSEC enabled, an entry may set only one of values or
        # health_checked_targets; otherwise both may be combined.
        values = optional(list(string), [])

        # Load-balancer targets health-checked for this entry (A/AAAA records
        # only). Unhealthy targets are withdrawn from answers automatically.
        health_checked_targets = optional(object({
          # Internal load balancer frontends to health check. Cloud DNS reads the
          # load balancer's own health signal — no separate health check resource
          # is needed for these.
          internal_load_balancers = optional(list(object({
            # The frontend IP address of the load balancer.
            # Can be a literal IP or a reference to a GcpAddress resource (the
            # reserved internal VIP the forwarding rule serves on).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            ip_address = string

            # The IP protocol the load balancer frontend is configured for.
            # Case-sensitive: "tcp" or "udp".
            ip_protocol = string

            # The type of load balancer. Case-sensitive: "regionalL4ilb",
            # "regionalL7ilb", or "globalL7ilb". If omitted, Cloud DNS infers it.
            load_balancer_type = optional(string, "")

            # The fully qualified self-link URL of the VPC network the load balancer
            # belongs to (e.g. "https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}").
            # Can be a literal URL or a reference to a GcpVpcNetwork resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            network_url = string

            # The configured port of the load balancer frontend.
            port = string

            # The ID of the project the load balancer belongs to — may differ from
            # the record's project in Shared-VPC and cross-project topologies.
            # Can be a literal project ID or a reference to a GcpProject resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            project = string

            # The region of the load balancer. Required for regional load balancers,
            # omitted for global ones.
            region = optional(string, "")
          })), [])

          # Public internet IP addresses to health check. Requires the routing
          # policy's health_check to be set.
          external_endpoints = optional(list(string), [])
        }))
      })), [])

      # Geolocation routing: each entry answers queries originating nearest to
      # its location. Useful for latency-sensitive multi-region serving.
      geo = optional(list(object({
        # The Google Cloud location name this entry serves (e.g. "us-east1",
        # "europe-west3"). Queries are routed to the entry nearest the caller.
        location = string

        # Static values (RRDATA) answered for this location.
        values = optional(list(string), [])

        # Load-balancer targets health-checked for this location (A/AAAA records
        # only). Unhealthy targets are withdrawn from answers automatically.
        health_checked_targets = optional(object({
          # Internal load balancer frontends to health check. Cloud DNS reads the
          # load balancer's own health signal — no separate health check resource
          # is needed for these.
          internal_load_balancers = optional(list(object({
            # The frontend IP address of the load balancer.
            # Can be a literal IP or a reference to a GcpAddress resource (the
            # reserved internal VIP the forwarding rule serves on).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            ip_address = string

            # The IP protocol the load balancer frontend is configured for.
            # Case-sensitive: "tcp" or "udp".
            ip_protocol = string

            # The type of load balancer. Case-sensitive: "regionalL4ilb",
            # "regionalL7ilb", or "globalL7ilb". If omitted, Cloud DNS infers it.
            load_balancer_type = optional(string, "")

            # The fully qualified self-link URL of the VPC network the load balancer
            # belongs to (e.g. "https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}").
            # Can be a literal URL or a reference to a GcpVpcNetwork resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            network_url = string

            # The configured port of the load balancer frontend.
            port = string

            # The ID of the project the load balancer belongs to — may differ from
            # the record's project in Shared-VPC and cross-project topologies.
            # Can be a literal project ID or a reference to a GcpProject resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            project = string

            # The region of the load balancer. Required for regional load balancers,
            # omitted for global ones.
            region = optional(string, "")
          })), [])

          # Public internet IP addresses to health check. Requires the routing
          # policy's health_check to be set.
          external_endpoints = optional(list(string), [])
        }))
      })), [])

      # When true, geo queries are fenced: a location with unhealthy targets
      # keeps answering (with the unhealthy answer) instead of failing over to
      # the next-closest location. Applies to geo routing only.
      enable_geo_fencing = optional(bool, false)

      # Failover routing: queries are answered with the primary targets while
      # any are healthy, then fall back to a regional geo policy.
      primary_backup = optional(object({
        # The global primary targets. Queries are answered from these while any
        # target is healthy.
        primary = object({
          # Internal load balancer frontends to health check. Cloud DNS reads the
          # load balancer's own health signal — no separate health check resource
          # is needed for these.
          internal_load_balancers = optional(list(object({
            # The frontend IP address of the load balancer.
            # Can be a literal IP or a reference to a GcpAddress resource (the
            # reserved internal VIP the forwarding rule serves on).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            ip_address = string

            # The IP protocol the load balancer frontend is configured for.
            # Case-sensitive: "tcp" or "udp".
            ip_protocol = string

            # The type of load balancer. Case-sensitive: "regionalL4ilb",
            # "regionalL7ilb", or "globalL7ilb". If omitted, Cloud DNS infers it.
            load_balancer_type = optional(string, "")

            # The fully qualified self-link URL of the VPC network the load balancer
            # belongs to (e.g. "https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}").
            # Can be a literal URL or a reference to a GcpVpcNetwork resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            network_url = string

            # The configured port of the load balancer frontend.
            port = string

            # The ID of the project the load balancer belongs to — may differ from
            # the record's project in Shared-VPC and cross-project topologies.
            # Can be a literal project ID or a reference to a GcpProject resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            project = string

            # The region of the load balancer. Required for regional load balancers,
            # omitted for global ones.
            region = optional(string, "")
          })), [])

          # Public internet IP addresses to health check. Requires the routing
          # policy's health_check to be set.
          external_endpoints = optional(list(string), [])
        })

        # The regional failover policy used when no primary target is healthy.
        backup_geo = list(object({
          # The Google Cloud location name this entry serves (e.g. "us-east1",
          # "europe-west3"). Queries are routed to the entry nearest the caller.
          location = string

          # Static values (RRDATA) answered for this location.
          values = optional(list(string), [])

          # Load-balancer targets health-checked for this location (A/AAAA records
          # only). Unhealthy targets are withdrawn from answers automatically.
          health_checked_targets = optional(object({
            # Internal load balancer frontends to health check. Cloud DNS reads the
            # load balancer's own health signal — no separate health check resource
            # is needed for these.
            internal_load_balancers = optional(list(object({
              # The frontend IP address of the load balancer.
              # Can be a literal IP or a reference to a GcpAddress resource (the
              # reserved internal VIP the forwarding rule serves on).
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              ip_address = string

              # The IP protocol the load balancer frontend is configured for.
              # Case-sensitive: "tcp" or "udp".
              ip_protocol = string

              # The type of load balancer. Case-sensitive: "regionalL4ilb",
              # "regionalL7ilb", or "globalL7ilb". If omitted, Cloud DNS infers it.
              load_balancer_type = optional(string, "")

              # The fully qualified self-link URL of the VPC network the load balancer
              # belongs to (e.g. "https://www.googleapis.com/compute/v1/projects/{project}/global/networks/{network}").
              # Can be a literal URL or a reference to a GcpVpcNetwork resource.
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              network_url = string

              # The configured port of the load balancer frontend.
              port = string

              # The ID of the project the load balancer belongs to — may differ from
              # the record's project in Shared-VPC and cross-project topologies.
              # Can be a literal project ID or a reference to a GcpProject resource.
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              project = string

              # The region of the load balancer. Required for regional load balancers,
              # omitted for global ones.
              region = optional(string, "")
            })), [])

            # Public internet IP addresses to health check. Requires the routing
            # policy's health_check to be set.
            external_endpoints = optional(list(string), [])
          }))
        }))

        # Ratio of traffic (0.0–1.0) trickled to the backup targets even while
        # the primaries are healthy — keeps backup paths warm and verifiable.
        trickle_ratio = optional(number)

        # When true, backup geo queries are fenced (see
        # GcpDnsRecordRoutingPolicy.enable_geo_fencing).
        enable_geo_fencing_for_backups = optional(bool, false)
      }))

      # Health check used for public-IP health checking of routing-policy
      # targets. Can be a literal self-link or a reference to a GcpHealthCheck
      # resource. Internal load balancer targets carry their own implicit
      # health checking and do not need this.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      health_check = optional(string, "")
    }))

    # Deletion policy for the record set — what happens on destroy:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the record set is deleted; resolvers stop getting
    #                answers for this name/type as caches expire
    #   "PREVENT" -- destroy FAILS; protects a name other systems resolve
    #                (mail routing, domain verification, service discovery)
    #   "ABANDON" -- the record set is removed from management but keeps
    #                answering queries in GCP
    deletion_policy = optional(string, "")
  })
}
