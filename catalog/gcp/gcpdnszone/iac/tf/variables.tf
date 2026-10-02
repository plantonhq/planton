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
  description = "GcpDnsZone specification"
  type = object({
    # The GCP project that owns the managed zone. Accepts a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Authoritative DNS domain for the zone (e.g. "example.com."). Must end with
    # a trailing dot. When omitted, modules derive dns_name from metadata.name
    # plus "." — preserving the legacy default for public zones.
    dns_name = optional(string, "")

    # Human-readable description shown in the GCP console.
    description = optional(string, "")

    # Zone visibility: public (internet-facing) or private (VPC/GKE only).
    visibility = optional(string)

    # VPC/GKE visibility targets for standard private zones. Not used with
    # forwarding or peering zone types.
    private_visibility_config = optional(object({
      networks = optional(list(object({
        # VPC network self-link or resource name. Reference a GcpVpcNetwork or
        # supply the full URL (projects/{project}/global/networks/{network}).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network_url = string
      })), [])
      gke_clusters = optional(list(object({
        # GKE cluster resource path (projects/{p}/locations/{loc}/clusters/{name}).
        # Reference a GcpGkeCluster or supply the path directly.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        gke_cluster_name = string
      })), [])
    }))

    # DNSSEC signing configuration (public zones).
    dnssec_config = optional(object({
      # DNSSEC mode: off, on, or transfer (import existing signed zone).
      state = optional(string)

      # Optional custom initial key specs. Updatable only while state is off.
      default_key_specs = optional(list(object({
        # DNSSEC algorithm. Common values: ecdsap256sha256, rsasha256.
        algorithm = optional(string, "")

        # Key length in bits (e.g. 256 for ECDSA, 2048 for RSA).
        key_length = optional(number, 0)

        # keySigning (KSK) or zoneSigning (ZSK).
        key_type = optional(string, "")
      })), [])

      # Authenticated denial-of-existence mechanism: nsec or nsec3.
      non_existence = optional(string, "")
    }))

    # Outbound forwarding targets (private forwarding zones).
    forwarding_config = optional(object({
      target_name_servers = list(object({
        # IPv4 address of the upstream resolver.
        ipv4_address = optional(string, "")

        # Fully qualified domain name of the forwarding target.
        domain_name = optional(string, "")

        # Query path: default (RFC1918 via VPC, else internet) or private (always VPC).
        forwarding_path = optional(string, "")

        # IPv6 address of the upstream resolver. A target carries one address
        # family — the provider rejects a target with both IPv4 and IPv6 set.
        ipv6_address = optional(string, "")
      }))
    }))

    # DNS peering target network (private peering zones).
    peering_config = optional(object({
      # VPC network to peer with. Reference a GcpVpcNetwork or supply the self-link.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_network = string
    }))

    # Cloud DNS query logging.
    cloud_logging_config = optional(object({
      # When true, log every query received by this managed zone. Off by default.
      enable_logging = optional(bool, false)
    }))

    # When true, delete all record sets in the zone on destroy. Default false.
    force_destroy = optional(bool)

    # Additional GCP labels merged with platform labels.
    labels = optional(map(string), {})

    # Deletion policy for the managed zone — the second of this kind's two
    # destroy levers (force_destroy empties the records first; this decides
    # the zone shell itself):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the zone is deleted (GCP refuses while non-default
    #                record sets remain unless force_destroy is true); its
    #                delegated name servers stop answering
    #   "PREVENT" -- destroy FAILS; protects a zone that registrars and
    #                parent zones delegate to
    #   "ABANDON" -- the zone is removed from management but keeps serving
    #                DNS in GCP
    deletion_policy = optional(string, "")
  })
}
