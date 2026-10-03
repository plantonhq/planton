variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpCertManagerIssuanceConfig specification"
  type = object({
    # The GCP project to create the issuance config in: a literal project ID
    # or a GcpProject reference. Empty uses the provider's default project.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the issuance config in GCP: 1-64 characters, starting with a
    # letter, then letters, digits, hyphens, or underscores. Defaults to
    # metadata.name. Immutable.
    issuance_config_name = optional(string, "")

    # Human-readable description. Immutable.
    description = optional(string, "")

    # The Certificate Manager location. Empty means "global" (the provider's
    # default), which serves global certificates; a regional certificate
    # needs an issuance config in its own region. Immutable.
    location = optional(string, "")

    # The Certificate Authority Service pool that issues the certificates, by
    # full name: projects/{project}/locations/{location}/caPools/{pool}.
    # Reference a GcpPrivateCaPool -- its `name` output is exactly this value.
    # The pool needs an enabled authority, and the Certificate Manager service
    # agent needs roles/privateca.certificateRequester on it. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ca_pool = string

    # Key algorithm for the private key Google generates for each
    # certificate: RSA_2048 (widest client compatibility) or ECDSA_P256
    # (smaller, faster handshakes; every modern client). Immutable.
    key_algorithm = string

    # Lifetime of each issued certificate, as a duration in seconds with up
    # to nine fractional digits and an "s" suffix: from 21 days ("1814400s")
    # to 30 days ("2592000s"). Shorter lifetimes limit the damage of a leaked
    # key; renewal is automatic either way. Immutable.
    lifetime = string

    # How far into a certificate's lifetime Google renews it, as a
    # percentage from 1 to 99. Google requires renewal at least 7 days after
    # issuance and at least 7 days before expiry, so the usable range depends
    # on the lifetime: 34-66 for 21 days, 24-76 for 30 days. Immutable.
    rotation_window_percentage = number

    # User labels merged onto the issuance config beneath the platform's
    # attribution labels (platform keys win on conflicts).
    labels = optional(map(string), {})

    # What happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the config is deleted (Google refuses while a
    #                certificate still references it)
    #   "PREVENT" -- destroy FAILS; a guard rail while certificates depend
    #                on it
    #   "ABANDON" -- removed from management but left in GCP
    deletion_policy = optional(string, "")
  })
}
