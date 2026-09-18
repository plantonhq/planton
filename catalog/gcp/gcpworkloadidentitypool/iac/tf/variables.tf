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
  description = "GcpWorkloadIdentityPool specification"
  type = object({
    # The ID for the pool, which becomes the final component of its resource
    # name (projects/<number>/locations/global/workloadIdentityPools/<id>).
    # 4-32 characters of lowercase letters, digits, and hyphens; the prefix
    # "gcp-" is reserved by Google. Immutable: changing it destroys and
    # recreates the pool, invalidating every principal and grant that
    # references the old pool name.
    workload_identity_pool_id = string

    # The GCP project that owns this pool. Federation quotas and IAM principals
    # are scoped to this project.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the pool.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Human-readable name shown in the GCP console (max 32 characters). Mutable.
    display_name = optional(string, "")

    # What this pool federates and who owns it — write it for the operator
    # auditing trust boundaries later (max 256 characters). Mutable.
    description = optional(string, "")

    # Emergency kill switch: a disabled pool rejects all token exchanges and
    # existing tokens stop granting access; re-enabling restores them. Prefer
    # disabling over deleting when rotating or investigating — deletion starts
    # the 30-day soft-delete clock and blocks ID reuse. Mutable.
    disabled = optional(bool, false)

    # How the pool operates. FEDERATION_ONLY (the default) federates external
    # identities into Google Cloud and is the right mode for keyless CI/CD and
    # cross-cloud auth. TRUST_DOMAIN assigns managed identities to Google Cloud
    # workloads (SPIFFE-style ns/<namespace>/sa/<workload> subjects) — pools in
    # this mode cannot hold providers. (A third mode, SYSTEM_TRUST_DOMAIN,
    # exists only for pools Google itself manages and cannot be created, so it
    # is not accepted here.)
    # Immutable in the API: the console may accept an edit attempt but the
    # update fails server-side — create a new pool to change modes.
    mode = optional(string)

    # Configuration for issuing mutual-TLS (mTLS) workload certificates to the
    # identities in this pool — the certificate half of a TRUST_DOMAIN pool.
    # Leave unset for token-exchange federation (FEDERATION_ONLY pools).
    inline_certificate_issuance_config = optional(object({
      # Maps a cloud region to the Certificate Authority Service CA pool (full
      # resource path projects/<project>/locations/<location>/caPools/<pool>)
      # that issues certificates for workloads in that region. The region in the
      # key must match the CA pool's own region. Exactly one of ca_pools or
      # use_default_shared_ca supplies the signing authority.
      ca_pools = optional(map(string), {})

      # Key algorithm for the generated certificate key pairs. Defaults
      # server-side to ECDSA_P256 — the right choice unless a legacy verifier
      # requires RSA.
      key_algorithm = optional(string)

      # Lifetime of issued workload certificates in seconds, formatted like
      # "86400s". Must be between 86400s (24 hours) and 2592000s (30 days);
      # defaults server-side to 86400s. Shorter lifetimes shrink the blast radius
      # of a leaked certificate at the cost of more rotation traffic.
      lifetime = optional(string)

      # Percentage of remaining certificate lifetime at which rotation begins.
      # Must be between 50 and 80; defaults server-side to 50 (rotate at half
      # life). Raise it only if workloads tolerate very tight rotation windows.
      rotation_window_percentage = optional(number)

      # Issue certificates from the GCP-provisioned default shared CA in the
      # workload's own region instead of your own CA Service pools — the
      # zero-setup path for managed-identity trust domains. Exactly one of
      # ca_pools or use_default_shared_ca supplies the signing authority.
      use_default_shared_ca = optional(bool, false)
    }))

    # Additional trust domains whose certificates this pool's trust domain
    # accepts. A trust domain always trusts itself; list only foreign domains.
    inline_trust_config = optional(object({
      # Trust bundles keyed by foreign trust domain (e.g. "example.com").
      # Maximum 10 entries. A trust domain automatically trusts itself and must
      # not be listed here.
      additional_trust_bundles = list(object({
        # The foreign trust domain being trusted (e.g. "example.com").
        trust_domain = string

        # Trust anchors for the domain: incoming end-entity certificates must chain
        # up to one of these. GCP requires at least one PEM anchor per bundle even
        # when trust_default_shared_ca is enabled — the shared CA is ADDED to the
        # bundle, never a substitute for it.
        trust_anchors = list(object({
          # PEM certificate of the PKI used for validation. Must contain exactly one
          # CA certificate (root or intermediate). This is public key material, not a
          # secret.
          pem_certificate = string
        }))

        # Additionally include the GCP-managed regional root certificates (the
        # default shared CA) in this bundle's trust set, so certificates issued
        # by use_default_shared_ca pools in the foreign domain are accepted.
        # Only meaningful for managed-identity (TRUST_DOMAIN) pools.
        trust_default_shared_ca = optional(bool, false)
      }))
    }))

    # Which workloads may RECEIVE a managed identity from this pool: each
    # rule names a single Google Cloud workload resource (for example
    # "//run.googleapis.com/projects/123/type/Service/*"), and matching
    # workloads are issued the identity. GCP caps a pool at 50 rules.
    # Mutable — but note GCP applies the rules through a separate API call
    # after the pool itself is created, so a failed apply can leave a pool
    # without its rules; re-apply converges.
    attestation_rules = optional(list(object({
      # A single workload operating on Google Cloud, as a full resource name
      # with optional trailing wildcard — for example
      # "//run.googleapis.com/projects/123/type/Service/*".
      google_cloud_resource = string
    })), [])

    # What happens to the pool in GCP when this resource is destroyed.
    #   "DELETE"  -- (GCP's default when unset) the pool is soft-deleted:
    #                token exchanges stop immediately, the pool is
    #                restorable for ~30 days, and its ID stays reserved
    #                until permanent deletion
    #   "PREVENT" -- destroy FAILS; protects the trust boundary every
    #                keyless-CI grant in the project references
    #   "ABANDON" -- the pool is removed from management but keeps
    #                federating in GCP
    deletion_policy = optional(string, "")
  })
}
