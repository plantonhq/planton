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
  description = "GcpCertManagerCert specification"
  type = object({
    # The GCP project to create the certificate in.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing the project destroys and recreates the certificate.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the certificate in GCP. Must be 1-64 characters: start with a
    # letter, then letters, digits, hyphens, or underscores. Certificate
    # names must be unique per location. Immutable.
    # If not specified, defaults to metadata.name.
    cert_name = optional(string, "")

    # Human-readable description of the certificate.
    description = optional(string, "")

    # The Certificate Manager location. Defaults to "global" — the correct
    # choice for classic external HTTPS load balancers. Regional
    # certificates serve regional load balancers only. Immutable.
    location = optional(string, "")

    # Where the certificate is served from. Immutable.
    #   DEFAULT: core Google data centers (choose this if unsure).
    #   EDGE_CACHE: Edge Points of Presence (Media CDN).
    #   ALL_REGIONS: every GCP region (global certificates only —
    #     cross-region internal Application Load Balancers).
    #   CLIENT_AUTH: presented BY the load balancer to the backend when
    #     backend mTLS is configured.
    scope = optional(string, "")

    # Google-managed arm: provisioned and renewed automatically.
    managed = optional(object({
      # The domains this certificate covers (e.g. "example.com",
      # "*.example.com"). Wildcards are supported only with DNS authorizations
      # or an issuance config. Immutable.
      domains = list(string)

      # DNS authorizations proving control of the domains — one per distinct
      # domain (an authorization covers its domain and that domain's
      # wildcard). Each entry is the authorization's fully-qualified resource
      # ID; reference GcpCertManagerDnsAuthorization resources.
      # Omit (with no issuance_config) for load-balancer authorization.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      dns_authorizations = optional(list(string), [])

      # Private-PKI issuance: the CertificateIssuanceConfig resource name
      # (projects/*/locations/*/certificateIssuanceConfigs/*) that signs
      # certificates from your own CA instead of a public one.
      # Mutually exclusive with dns_authorizations.
      issuance_config = optional(string, "")
    }))

    # Self-managed arm: bring-your-own PEM certificate and key.
    self_managed = optional(object({
      # The certificate chain in PEM form: leaf certificate first, followed by
      # any intermediates. Certificate material is public — only the key is
      # secret. Updating the pair in place rotates the certificate.
      pem_certificate = string

      # The leaf certificate's private key in PEM form — a PRIVATE KEY block
      # (take care not to swap the certificate and the key). Secret material —
      # never logged, masked in outputs. The PEM framing is taught here rather
      # than enforced by a validation rule, because sensitive fields hold a
      # managed-secret reference on consuming platforms and a content-shape
      # rule would reject every reference.
      pem_private_key = string
    }))

    # User labels merged onto the certificate beneath the platform's
    # attribution labels (platform keys win on conflicts).
    # Keys/values: lowercase letters, digits, underscores, hyphens.
    labels = optional(map(string), {})

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the certificate is deleted (GCP refuses while a proxy
    #                or certificate map still references it)
    #   "PREVENT" -- destroy FAILS; a guard rail for a certificate whose
    #                replacement is not yet serving
    #   "ABANDON" -- the certificate is removed from management but left
    #                in GCP (useful mid-rotation handoff)
    deletion_policy = optional(string, "")
  })
}
