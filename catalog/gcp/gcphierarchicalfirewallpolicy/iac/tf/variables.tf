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
  description = "GcpHierarchicalFirewallPolicy specification"
  type = object({
    # Where the policy lives: directly under the organization or inside a
    # folder. Exactly one arm. This is the policy's home for IAM and
    # quota purposes; WHERE IT IS ENFORCED is `associations`, which may
    # name the same node or others. Immutable: changing the parent
    # recreates the policy (and with it every rule and association).
    parent = object({
      # An organization-level policy: the numeric organization ID (from
      # `gcloud organizations list`), without the `organizations/` prefix.
      organization_id = optional(string, "")

      # A folder-level policy: the folder's numeric ID -- a literal, or a
      # reference to a GcpFolder resource (its folder_id output). The
      # reference is how a chart builds a landing zone: the policy waits for
      # its folder to exist and lives inside it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      folder_id = optional(string, "")
    })

    # The policy's user-facing name, unique among the hierarchical policies
    # of the organization. Defaults to metadata.name when empty. RFC 1035:
    # 1-63 characters, lowercase letters, digits, and hyphens, starting with
    # a letter and not ending with a hyphen. Immutable: Google identifies
    # the policy by a server-assigned numeric ID (the `policy_id` output)
    # and the short name is fixed at creation.
    short_name = optional(string, "")

    # What this policy enforces and who owns it -- the operator reading a
    # denied connection's log sees the policy, not this manifest. Mutable.
    description = optional(string, "")

    # The ordered rule set, evaluated from the LOWEST priority number up:
    # the first rule whose match fits the packet decides (allow, deny, or
    # goto_next, which delegates to the next policy level). Each entry is
    # its own Google resource keyed by `priority`, so changing a rule's
    # priority RECREATES that rule (the old priority is freed, the new one
    # is created) while every other rule is left alone; change a rule's
    # content in place and it is updated in place. An empty list is a valid
    # policy that decides nothing (Google's implied goto_next rules apply).
    rules = optional(list(object({
      # The rule's position in the evaluation order: 0 is evaluated first,
      # 2147483647 last. Unique within the policy, and a rule's identity in
      # Google: changing it recreates the rule. Leave a gap between rules
      # (1000, 2000, ...) so a rule can be slotted in later without
      # renumbering. Priorities 2147483646 and 2147483647 are Google's implied
      # goto_next rules -- do not use them.
      priority = number

      # What happens to a matching packet:
      #   "allow"     -- permit it; evaluation of every lower level stops
      #   "deny"      -- drop it; evaluation stops
      #   "goto_next" -- delegate the decision to the next level (the child
      #                  folder's policy, then the network's policies, then
      #                  legacy VPC rules); the way an organization carves
      #                  out exceptions for a branch of the tree
      #   "apply_security_profile_group" -- send the traffic through the
      #                  Cloud NGFW security profile group named in
      #                  security_profile_group for layer-7 inspection, which
      #                  then allows or denies it
      # Lowercase, exactly as Google's API spells them.
      action = string

      # Which traffic the rule looks at: INGRESS (arriving at the target VMs;
      # `match` describes the SOURCE with src_* fields and may name the
      # destination) or EGRESS (leaving the target VMs; `match` describes the
      # DESTINATION with dest_* fields and may name the source). Immutable in
      # meaning: an ingress rule's src_secure_tags and src_networks describe
      # where traffic comes from, so they carry no meaning on an egress rule
      # and Google ignores them there.
      direction = string

      # What this rule is for, in the words an operator reading a firewall
      # log will understand. Mutable.
      description = optional(string, "")

      # When true the rule is kept but not enforced -- traffic behaves as if
      # the rule did not exist. The safe way to switch a rule off without
      # losing it (and its priority slot). Mutable.
      disabled = optional(bool, false)

      # Log every connection this rule decides, to Cloud Logging (and from
      # there to BigQuery or Pub/Sub through a GcpLoggingSink). Logs carry the
      # rule, the policy, both endpoints, and the decision. Cannot be set on
      # a goto_next rule: Google logs only the rule that decides. Mutable.
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

      # Restrict the rule to VMs in these VPC networks -- references to
      # GcpVpcNetwork resources (their network_self_link output) or network
      # self-links as literals. Empty means every VM in every network beneath
      # the association's node. The way one rule in an organization policy
      # targets one shared VPC without a folder of its own.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_resources = optional(list(string), [])

      # Restrict the rule to VMs carrying one of these secure tags --
      # references to GcpTagValue resources (their `name` output,
      # `tagValues/{numeric_id}`) or those names as literals. Maximum 256.
      # A secure tag is a Resource Manager tag value bound to the VM; a tag
      # whose value or network was deleted is INEFFECTIVE, and a rule whose
      # target tags are all ineffective is ignored. Cannot be combined with
      # target_service_accounts. Empty (with target_service_accounts also
      # empty) means every VM the rule can reach.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_secure_tags = optional(list(string), [])

      # Restrict the rule to VMs running as one of these service accounts --
      # references to GcpServiceAccount resources (their email output) or
      # email addresses as literals. Cannot be combined with
      # target_secure_tags.
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

    # Where the policy is ENFORCED: each entry attaches the policy to the
    # organization or to one folder, so every VPC network in every project
    # beneath that node evaluates these rules. A node can carry only one
    # hierarchical policy association at a time -- associating a second
    # policy with the same folder fails until the first is detached. Each
    # entry is its own Google resource; adding or removing one touches only
    # that association. A policy with no association exists but governs
    # nothing.
    associations = optional(list(object({
      # The association's name, unique within the policy. Defaults to
      # `<short_name>-<n>` (n = the entry's position, starting at 1) when
      # empty. Shown by `gcloud compute firewall-policies associations list`.
      name = optional(string, "")

      # The organization or folder the policy is enforced on. Exactly one arm.
      target = object({
        # Enforce on the whole organization: its numeric ID, without the
        # `organizations/` prefix.
        organization_id = optional(string, "")

        # Enforce on one folder and everything beneath it: the folder's numeric
        # ID -- a literal, or a reference to a GcpFolder (its folder_id output).
        # An attachment edge, not a placement: the policy does not live in this
        # folder.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        folder_id = optional(string, "")
      })
    })), [])

    # What destroying this resource does in GCP, applied to the policy, to
    # every rule, and to every association together:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- associations are detached, rules removed, and the
    #                policy deleted; traffic falls through to the next
    #                policy level as if the policy had never existed
    #   "PREVENT" -- destroy FAILS; the guard for an organization's baseline
    #                deny rules
    #   "ABANDON" -- everything is removed from management but keeps
    #                existing and enforcing in GCP
    deletion_policy = optional(string, "")
  })
}
