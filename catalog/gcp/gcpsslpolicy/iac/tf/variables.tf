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
  description = "GcpSslPolicy specification"
  type = object({
    # The GCP project that owns the SSL policy.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the policy.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the SSL policy in GCP. Must be 1-63 characters: lowercase
    # letters, digits, and hyphens; must start with a letter and end with a
    # letter or digit. If not specified, defaults to metadata.name.
    # Immutable: changing it destroys and recreates the policy, briefly
    # breaking every proxy that references the old self_link.
    ssl_policy_name = optional(string, "")

    # Region for a REGIONAL SSL policy (e.g. "us-central1"), used by regional
    # external and internal Application Load Balancer proxies. Leave empty for
    # a GLOBAL policy — the right scope for global external load balancers.
    # Immutable: a policy cannot move between scopes or regions.
    region = optional(string, "")

    # Why this policy exists and which proxies should use it — write it for
    # the operator auditing TLS posture later. Immutable: changing it
    # destroys and recreates the policy (unusual for a description — a GCP
    # API quirk on this resource).
    description = optional(string, "")

    # The cipher-suite profile negotiated with clients (default COMPATIBLE).
    # COMPATIBLE allows the widest client range; MODERN drops broken ciphers
    # while keeping broad reach; RESTRICTED narrows to ciphers with modern
    # security guarantees (and is required when the TLS floor is raised beyond
    # what other profiles allow); CUSTOM hand-picks cipher suites via
    # custom_features; FIPS_202205 pins the FIPS 140-2/3 validated suite set
    # (and requires min_tls_version TLS_1_2 — the only floor that profile
    # supports). Mutable — tightening the profile applies to every proxy
    # referencing this policy on its next handshake.
    profile = optional(string, "")

    # The minimum TLS protocol version clients may negotiate (default
    # TLS_1_0). Raise to TLS_1_2 for PCI DSS and most modern compliance
    # regimes; TLS_1_3 is the strictest floor and requires the RESTRICTED
    # profile. GCP has no maximum-version control — TLS 1.3 is always
    # negotiable when the client supports it, whatever the floor. Mutable.
    min_tls_version = optional(string, "")

    # Exact cipher suites to allow — required with (and only valid with) the
    # CUSTOM profile. Names are IANA-style suite identifiers from GCP's
    # supported set (e.g. TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256); GCP rejects
    # unknown names at deploy time. TLS 1.3 suites are not listable — GCP
    # always enables them regardless of this list. Mutable.
    custom_features = optional(list(string), [])

    # Post-quantum key exchange (X25519MLKEM768) posture for TLS handshakes
    # (default DEFAULT). GCP rolls the hybrid post-quantum group out on its
    # own schedule; this dial controls the rollout stance rather than a
    # static on/off:
    #   DEFAULT  -- follow GCP's rollout timeline
    #   ENABLED  -- allow post-quantum key exchange now
    #   DEFERRED -- opt out until GCP's later mandatory date
    # Mutable.
    post_quantum_key_exchange = optional(string, "")

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the policy is deleted (GCP refuses while any proxy
    #                still references it, so destroy fails rather than
    #                silently loosening TLS floors)
    #   "PREVENT" -- destroy FAILS; protects a compliance-mandated TLS
    #                posture from accidental teardown
    #   "ABANDON" -- the policy is removed from management but left in
    #                GCP, still enforced by every proxy referencing it
    deletion_policy = optional(string, "")
  })
}
