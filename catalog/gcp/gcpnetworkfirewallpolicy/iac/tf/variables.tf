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
  description = "GcpNetworkFirewallPolicy specification"
  type = object({
    # The GCP project the policy and its rules are created in, and whose
    # networks it can be associated with: a reference to a GcpProject or
    # the project ID as a literal. Empty means the provider's default
    # project. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The policy's name in GCP, unique among the project's network firewall
    # policies of the same scope. Defaults to metadata.name when empty. RFC
    # 1035: 1-63 characters, lowercase letters, digits, and hyphens,
    # starting with a letter and not ending with a hyphen. Immutable: a
    # rename recreates the policy with every rule and association.
    policy_name = optional(string, "")

    # Empty for a GLOBAL policy (all regions of the associated networks; the
    # right default). A region name such as us-central1 for a REGIONAL
    # policy, which governs only that region's traffic -- pick regional when
    # the rules target an internal managed load balancer (INTERNAL_MANAGED_LB
    # rules are regional) or need an RDMA or ULL policy type. Immutable: a
    # policy cannot move between scopes or regions.
    region = optional(string, "")

    # What this policy enforces and who owns it -- the operator reading a
    # denied connection's log sees the policy, not this manifest. Mutable.
    description = optional(string, "")

    # Which kind of network the policy can be associated with; a network
    # accepts a policy only when its network profile carries the matching
    # type. VPC_POLICY is the ordinary VPC network and Google's default when
    # unset -- leave this empty unless the network is one of the special
    # profiles. Regional policies only: RDMA_ROCE_POLICY (RoCE RDMA
    # networks), RDMA_FALCON_POLICY (Falcon RDMA networks), ULL_POLICY
    # (ultra-low-latency networks). Immutable. Both engines send it only
    # when set, so an unset value never fights Google's default on re-plan.
    policy_type = optional(string)

    # The ordered rule set, evaluated from the LOWEST priority number up:
    # the first rule whose match fits the packet decides (allow, deny, or
    # goto_next, which delegates to the next level). Each entry is its own
    # Google resource keyed by `priority`, so changing a rule's priority
    # RECREATES that rule (the old priority is freed, the new one is
    # created) while every other rule is left alone; change a rule's content
    # in place and it is updated in place. An empty list is a valid policy
    # that decides nothing.
    rules = optional(list(object({
      # The rule's position in the evaluation order: 0 is evaluated first,
      # 2147483647 last. Unique within the policy, and a rule's identity in
      # Google: changing it recreates the rule. Leave a gap between rules
      # (1000, 2000, ...) so a rule can be slotted in later without
      # renumbering.
      priority = number

      # What happens to a matching packet:
      #   "allow"     -- permit it; evaluation of every lower level stops
      #   "deny"      -- drop it; evaluation stops
      #   "goto_next" -- delegate the decision to the next level (the
      #                  regional policy after a global one, then legacy VPC
      #                  rules)
      #   "apply_security_profile_group" -- send the traffic through the
      #                  Cloud NGFW security profile group named in
      #                  security_profile_group for layer-7 inspection, which
      #                  then allows or denies it
      # Lowercase, exactly as Google's API spells them.
      action = string

      # Which traffic the rule looks at: INGRESS (arriving at the targets;
      # `match` describes the SOURCE with src_* fields and may name the
      # destination) or EGRESS (leaving the targets; `match` describes the
      # DESTINATION with dest_* fields and may name the source). An ingress
      # rule's src_secure_tags and src_networks describe where traffic comes
      # from, so they carry no meaning on an egress rule and Google ignores
      # them there.
      direction = string

      # What this rule is for, in the words an operator reading a firewall
      # log will understand. Mutable.
      description = optional(string, "")

      # A label for the rule shown by `gcloud` and the console beside its
      # priority. Mutable, and NOT the rule's identity (priority is), so it
      # can be renamed freely. Optional.
      rule_name = optional(string, "")

      # When true the rule is kept but not enforced -- traffic behaves as if
      # the rule did not exist. The safe way to switch a rule off without
      # losing it (and its priority slot). Mutable.
      disabled = optional(bool, false)

      # Log every connection this rule decides, to Cloud Logging (and from
      # there to BigQuery or Pub/Sub through a GcpLoggingSink). Logs carry the
      # rule, the policy, both endpoints, and the decision. Cannot be set on a
      # goto_next rule: Google logs only the rule that decides. Mutable.
      enable_logging = optional(bool, false)

      # The condition a packet must satisfy for this rule to decide it. At
      # least one layer4_configs entry is required; every other field narrows
      # the match and an empty field matches everything.
      match = object({
        # Protocols and ports the rule applies to. At least one entry; a packet
        # matches when it fits any entry.
        layer4_configs = list(object({
          # The IP protocol: a well-known name (tcp, udp, icmp, esp, ah, ipip,
          # sctp), `all` for every protocol, or a protocol number (0-255).
          ip_protocol = string

          # Ports or port ranges the rule applies to, for tcp and udp only:
          # "22", "80", "8000-8999". Empty means every port of the protocol.
          ports = optional(list(string), [])
        }))

        # Source addresses in CIDR form (IPv4 or IPv6, e.g. 10.0.0.0/8,
        # 0.0.0.0/0, 2001:db8::/32). Maximum 5000. On an INGRESS rule the
        # everyday selector ("from the internet", "from on-premises").
        src_ip_ranges = optional(list(string), [])

        # Destination addresses in CIDR form. Maximum 5000. On an EGRESS rule
        # the everyday selector ("to the internet", "to the database subnet").
        dest_ip_ranges = optional(list(string), [])

        # Address groups whose addresses count as the source, as Network
        # Security address-group names
        # (`projects/{project}/locations/global/addressGroups/{name}` or
        # `organizations/{id}/locations/global/addressGroups/{name}`). Maximum
        # 10. Literals: address groups are not yet a catalog kind.
        src_address_groups = optional(list(string), [])

        # Address groups whose addresses count as the destination. Maximum 10.
        dest_address_groups = optional(list(string), [])

        # Fully qualified domain names whose resolved addresses count as the
        # source (Google resolves them continuously). Maximum 100.
        src_fqdns = optional(list(string), [])

        # Fully qualified domain names whose resolved addresses count as the
        # destination -- the way an egress rule allows `api.example.com` without
        # pinning its IPs. Maximum 100.
        dest_fqdns = optional(list(string), [])

        # Countries whose IP space counts as the source, as ISO 3166-1 alpha-2
        # codes (US, DE, ...). Maximum 5000. Geo-blocking on ingress.
        src_region_codes = optional(list(string), [])

        # Countries whose IP space counts as the destination, as ISO 3166-1
        # alpha-2 codes. Maximum 5000.
        dest_region_codes = optional(list(string), [])

        # Google Threat Intelligence lists whose addresses count as the source,
        # by list name: iplist-tor-exit-nodes, iplist-known-malicious-ips,
        # iplist-search-engines-crawlers, iplist-vpn-providers,
        # iplist-anon-proxies, iplist-crypto-miners, iplist-public-clouds and
        # the per-cloud lists (iplist-public-clouds-aws, -azure, -gcp). A deny
        # rule on iplist-known-malicious-ips is the classic use.
        src_threat_intelligences = optional(list(string), [])

        # Google Threat Intelligence lists whose addresses count as the
        # destination, by list name (the same names as src_threat_intelligences).
        dest_threat_intelligences = optional(list(string), [])

        # Source VMs carrying one of these secure tags -- references to
        # GcpTagValue resources (their `name` output, `tagValues/{numeric_id}`)
        # or those names as literals. Maximum 256. INGRESS rules only: on an
        # ingress rule with src_secure_tags and no src_ip_ranges, if every tag
        # is INEFFECTIVE the rule is ignored. Micro-segmentation without IP
        # ranges: "frontend-tagged VMs may reach backend-tagged VMs on 8080".
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        src_secure_tags = optional(list(string), [])

        # Source VPC networks -- references to GcpVpcNetwork resources (their
        # network_self_link output) or network URLs as literals. INGRESS rules
        # only: matches traffic that originates inside one of these networks
        # (peered or shared), the selector behind src_network_context
        # VPC_NETWORKS.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        src_networks = optional(list(string), [])

        # Where the traffic comes from, as a class rather than an address:
        # INTERNET (outside Google Cloud), INTRA_VPC (inside this network),
        # NON_INTERNET (any non-internet source), VPC_NETWORKS (the networks in
        # src_networks), or UNSPECIFIED. Unset means the rule does not narrow on
        # context (Google's default); both engines send it only when set.
        src_network_context = optional(string)

        # Where the traffic is going, as a class: INTERNET, INTRA_VPC,
        # NON_INTERNET, VPC_NETWORKS, or UNSPECIFIED. Unset means no narrowing
        # (Google's default); sent only when set. The egress-side twin of
        # src_network_context -- "to the internet" without listing 0.0.0.0/0.
        dest_network_context = optional(string)
      })

      # What the rule applies to: INSTANCES (VMs on the associated networks;
      # Google's default when unset) or INTERNAL_MANAGED_LB (the internal
      # Application Load Balancers named in target_forwarding_rules -- a
      # regional policy's feature). Both engines send it only when set, so an
      # unset value never fights Google's default on re-plan.
      target_type = optional(string)

      # The internal managed load balancers the rule applies to when
      # target_type is INTERNAL_MANAGED_LB -- references to forwarding-rule
      # resources (a GcpGlobalForwardingRule's self_link output) or
      # forwarding-rule self-links as literals. Required for that target
      # type, forbidden for INSTANCES. An internal managed load balancer's
      # forwarding rule is REGIONAL, so today the reference form resolves
      # fully once the referenced kind carries its regional arm; until then
      # give the regional forwarding rule's self-link as a literal.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_forwarding_rules = optional(list(string), [])

      # Restrict the rule to VMs carrying one of these secure tags --
      # references to GcpTagValue resources (their `name` output,
      # `tagValues/{numeric_id}`) or those names as literals. Maximum 256. A
      # secure tag is a Resource Manager tag value bound to the VM; a tag
      # whose value or network was deleted is INEFFECTIVE, and a rule whose
      # target tags are all ineffective is ignored. Empty (with
      # target_service_accounts also empty) means every VM on the associated
      # networks.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_secure_tags = optional(list(string), [])

      # Restrict the rule to VMs running as one of these service accounts --
      # references to GcpServiceAccount resources (their email output) or
      # email addresses as literals.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_service_accounts = optional(list(string), [])

      # The Cloud NGFW security profile group that inspects the traffic when
      # action is apply_security_profile_group, as its full resource URL:
      # `//networksecurity.googleapis.com/projects/{project}/locations/global/securityProfileGroups/{name}`
      # (organization-owned groups use `organizations/{id}` in place of
      # `projects/{project}`). Required for that action, forbidden for every
      # other. A literal: security profile groups are not yet a catalog kind.
      security_profile_group = optional(string, "")

      # Decrypt TLS traffic before the security profile group inspects it
      # (Cloud NGFW TLS inspection; needs a TLS inspection policy on the
      # network). Only with action apply_security_profile_group.
      tls_inspect = optional(bool, false)
    })), [])

    # The VPC networks this policy is enforced on -- each entry is one
    # association resource attaching the policy to one network in the
    # same project. A network carries at most one global and, per region,
    # one regional network firewall policy at a time; associating a second
    # fails until the first is detached. A policy with no association
    # exists but governs nothing.
    associations = optional(list(object({
      # The association's name, unique within the policy. Defaults to
      # `<policy_name>-<n>` (n = the entry's position, starting at 1) when
      # empty.
      name = optional(string, "")

      # The VPC network the policy is enforced on: a reference to a
      # GcpVpcNetwork resource (its network_self_link output) or the network's
      # self-link as a literal. Must be in the policy's project. An attachment
      # edge, not a placement: the policy does not live in the network.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = string
    })), [])

    # What destroying this resource does in GCP, applied to the policy, to
    # every rule, and to every association together:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- associations are detached, rules removed, and the
    #                policy deleted; traffic falls through to the next
    #                level as if the policy had never existed
    #   "PREVENT" -- destroy FAILS; the guard for a network's baseline deny
    #   "ABANDON" -- everything is removed from management but keeps
    #                existing and enforcing in GCP
    deletion_policy = optional(string, "")
  })
}
