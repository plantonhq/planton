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
  description = "GcpManagedSslCertificate specification"
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
    # Immutable: changing it destroys and recreates the certificate, briefly
    # breaking every target HTTPS proxy that references the old self_link.
    certificate_name = optional(string, "")

    # What this certificate secures and which proxy uses it — write it for the
    # operator debugging a TLS provisioning issue later. Immutable.
    description = optional(string, "")

    # The domains the certificate is valid for (1-100). Each must be a
    # fully-qualified domain name; a leading "*." wildcard is NOT supported by
    # Google-managed certificates. Every domain must have DNS pointing at the
    # load balancer before the certificate can finish provisioning. Immutable:
    # changing the list destroys and recreates the certificate.
    domains = list(string)

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the certificate is deleted (GCP refuses while any
    #                proxy still references it, so destroy fails rather
    #                than dropping TLS)
    #   "PREVENT" -- destroy FAILS; a guard rail for a certificate whose
    #                replacement is not yet provisioned and serving
    #   "ABANDON" -- the certificate is removed from management but left
    #                in GCP (useful mid-rotation handoff)
    deletion_policy = optional(string, "")
  })
}
