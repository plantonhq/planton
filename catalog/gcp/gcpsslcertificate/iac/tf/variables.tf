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
  description = "GcpSslCertificate specification"
  type = object({
    # The GCP project that owns the certificate.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the certificate.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the certificate in GCP. Must be 1-63 characters: lowercase
    # letters, digits, and hyphens; must start with a letter and end with a
    # letter or digit. If not specified, defaults to metadata.name.
    # Self-managed and Google-managed certificates share one namespace per
    # scope — a name used by either kind is taken. Immutable: changing it
    # destroys and recreates the certificate, briefly breaking every proxy
    # that references the old self_link (rotate create-before-destroy
    # instead).
    certificate_name = optional(string, "")

    # Region for a REGIONAL certificate (e.g. "us-central1"), used by regional
    # external and internal Application Load Balancer proxies. Leave empty for
    # a GLOBAL certificate — the right scope for global external load
    # balancers. Immutable: a certificate cannot move between scopes or
    # regions.
    region = optional(string, "")

    # What this certificate secures and where it came from (issuing CA,
    # rotation cadence) — write it for the operator planning the next
    # rotation. Immutable.
    description = optional(string, "")

    # The certificate chain in PEM format: the leaf certificate first,
    # followed by intermediates (at least one intermediate; at most 5
    # certificates total — GCP rejects longer chains). This is PUBLIC
    # handshake material presented to every client, so it is deliberately not
    # marked sensitive; only the private key is secret. Immutable.
    certificate = string

    # The private key matching the certificate, in PEM format (-----BEGIN
    # PRIVATE KEY----- / -----BEGIN RSA PRIVATE KEY----- / -----BEGIN EC
    # PRIVATE KEY-----). GCP accepts RSA-2048 (and larger) and ECDSA P-256
    # keys; the key must be unencrypted (no passphrase). Write-only in GCP —
    # the API never returns it, and it never appears in stack outputs.
    # Immutable. The PEM framing is taught here rather than enforced by a
    # validation rule, because sensitive fields hold a managed-secret
    # reference on consuming platforms and a content-shape rule would
    # reject every reference.
    private_key = string

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the certificate is deleted (GCP refuses while any
    #                proxy still references it, so destroy fails rather
    #                than dropping TLS)
    #   "PREVENT" -- destroy FAILS; a guard rail for a certificate whose
    #                replacement is not yet serving
    #   "ABANDON" -- the certificate is removed from management but left
    #                in GCP (useful mid-rotation handoff)
    deletion_policy = optional(string, "")
  })
}
