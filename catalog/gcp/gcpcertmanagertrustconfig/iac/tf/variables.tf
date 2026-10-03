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
  description = "GcpCertManagerTrustConfig specification"
  type = object({
    # The GCP project to create the trust config in: a literal project ID or a
    # GcpProject reference. Empty uses the provider's default project.
    # Immutable: changing it recreates the trust config.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the trust config in GCP: 1-64 characters, starting with a letter,
    # then letters, digits, hyphens, or underscores. Defaults to metadata.name.
    # Immutable.
    trust_config_name = optional(string, "")

    # Human-readable description.
    description = optional(string, "")

    # The Certificate Manager location. Defaults to "global", which serves
    # global external and cross-region internal Application Load Balancers; a
    # regional load balancer needs a trust config in its own region (for
    # example "us-central1"). Immutable.
    location = optional(string, "")

    # The trust stores client certificates are validated against. Google
    # allows exactly one today ("Only one TrustStore specified is currently
    # allowed" -- Certificate Manager API reference).
    trust_stores = optional(list(object({
      # Root CA certificates, each PEM-encoded ("-----BEGIN CERTIFICATE-----").
      # A client chain is valid when it builds up to one of them.
      trust_anchors = optional(list(string), [])

      # Intermediate CA certificates, each PEM-encoded, used to build a chain
      # from a client certificate to a trust anchor when the client does not
      # send its intermediates.
      intermediate_cas = optional(list(string), [])
    })), [])

    # Certificates accepted regardless of chain, each a PEM-encoded X.509
    # certificate ("-----BEGIN CERTIFICATE-----..."). Use for self-signed
    # device certificates or a single partner certificate you do not want to
    # trust a whole CA for. A matching certificate must still parse, prove
    # possession of its private key, and satisfy its SAN constraints.
    allowlisted_certificates = optional(list(string), [])

    # User labels merged onto the trust config beneath the platform's
    # attribution labels (platform keys win on conflicts).
    labels = optional(map(string), {})

    # What happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the trust config is deleted (Google refuses while a TLS
    #                policy still references it)
    #   "PREVENT" -- destroy FAILS; a guard rail for a trust config live
    #                mutual-TLS traffic depends on
    #   "ABANDON" -- removed from management but left in GCP
    deletion_policy = optional(string, "")
  })
}
