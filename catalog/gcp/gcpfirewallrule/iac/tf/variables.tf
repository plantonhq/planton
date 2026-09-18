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
  description = "GcpFirewallRule specification"
  type = object({
    # The GCP project ID in which to create this firewall rule.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # VPC network to which this firewall rule applies.
    # Accepts a network name (e.g., "default") or a full self-link URL.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = string

    # Name of the firewall rule in GCP.
    # Must be 1-63 characters, lowercase letters, numbers, or hyphens.
    # Must start with a lowercase letter and end with a letter or number.
    # Example: "allow-http-ingress", "deny-all-egress"
    rule_name = string

    # Traffic direction this rule applies to.
    # "INGRESS" matches inbound traffic to instances; "EGRESS" matches outbound traffic.
    direction = string

    # Action to take when the rule matches traffic.
    # "ALLOW" permits the matched traffic; "DENY" blocks it.
    action = string

    # Protocol and port combinations this rule matches.
    # At least one rule must be specified.
    rules = list(object({
      # IP protocol to match. Accepted values: "tcp", "udp", "icmp", "esp", "ah", "sctp", "ipip", "all",
      # or an IANA protocol number (e.g., "6" for TCP).
      protocol = string

      # Ports or port ranges to match. Only applicable when protocol is "tcp",
      # "udp", or "sctp" — including their IANA numbers ("6", "17", "132").
      # Each entry is a single port (e.g., "80") or a range (e.g., "8000-9000").
      # Omit for protocols that do not use ports (icmp, esp, ah, etc.).
      ports = optional(list(string), [])
    }))

    # Rule priority. Range: 0-65535. Lower values indicate higher priority.
    # At the same priority, DENY rules take precedence over ALLOW rules.
    # Default: 1000
    priority = optional(number)

    # Human-readable description for the firewall rule.
    # Example: "Allow HTTP/HTTPS traffic from the internet"
    description = optional(string, "")

    # Source IPv4 or IPv6 CIDR ranges for INGRESS rules.
    # Traffic is only matched when the source IP falls within one of these ranges.
    # For INGRESS rules, at least one of source_ranges, source_tags, or source_service_accounts is required.
    # Example: ["0.0.0.0/0"] to match all IPv4 traffic, ["10.0.0.0/8"] for internal.
    source_ranges = optional(list(string), [])

    # Destination IPv4 or IPv6 CIDR ranges for EGRESS rules.
    # Traffic is only matched when the destination IP falls within one of these ranges.
    # If omitted on EGRESS rules, GCP defaults to ["0.0.0.0/0"] (all destinations).
    destination_ranges = optional(list(string), [])

    # Source instance network tags for INGRESS rules.
    # Traffic from instances with any of these tags is matched.
    # Cannot be combined with source_service_accounts or target_service_accounts.
    source_tags = optional(list(string), [])

    # Target instance network tags.
    # The rule applies only to instances that have one of these tags.
    # If omitted, the rule applies to all instances in the network.
    # Cannot be combined with source_service_accounts or target_service_accounts.
    target_tags = optional(list(string), [])

    # Source service accounts for INGRESS rules. Max 10.
    # Traffic from instances running as any of these service accounts is matched.
    # Cannot be combined with source_tags or target_tags.
    source_service_accounts = optional(list(string), [])

    # Target service accounts. Max 10.
    # The rule applies only to instances running as one of these service accounts.
    # If omitted, the rule applies to all instances in the network.
    # Cannot be combined with source_tags or target_tags.
    target_service_accounts = optional(list(string), [])

    # Whether the firewall rule is disabled.
    # A disabled rule exists in the configuration but is not enforced.
    # Useful for temporarily suspending a rule without deleting it.
    disabled = optional(bool, false)

    # Logging configuration. When present, firewall logging is enabled.
    # Omit this field to disable logging (the default).
    log_config = optional(object({
      # Metadata inclusion mode for firewall logs.
      metadata = string
    }))

    # Resource Manager tags bound to the firewall rule at create time, for
    # org-policy conditions and IAM scoping. Keys are "tagKeys/{tag_key_id}"
    # and values "tagValues/{tag_value_id}" (IDs, not short names). Changing
    # tags REPLACES the firewall rule (the provider sends them through a
    # create-only params block) -- plan tag changes deliberately.
    resource_manager_tags = optional(map(string), {})

    # What happens to the firewall rule in GCP when this resource is destroyed.
    #   "DELETE"  -- (GCP's default when unset) the rule is deleted; traffic
    #                it allowed is cut over to the next matching rule
    #   "PREVENT" -- destroy FAILS; protects a rule other teams' traffic
    #                may silently depend on
    #   "ABANDON" -- the rule is removed from management but keeps enforcing
    #                in GCP (free at rest; clean it up manually)
    deletion_policy = optional(string, "")
  })
}
