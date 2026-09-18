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
  description = "GcpCertManagerDnsAuthorization specification"
  type = object({
    # The GCP project to create the authorization in.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the authorization in GCP. Must be 1-64 characters: start with
    # a letter, then letters, digits, hyphens, or underscores.
    # If not specified, defaults to metadata.name.
    # Immutable: renaming destroys and recreates the authorization.
    authorization_name = optional(string, "")

    # The domain being authorized. Covers the domain itself and its
    # wildcard: authorizing "example.com" issues certificates for
    # "example.com" and "*.example.com". Immutable.
    # Bare domain only — no "*." prefix and no trailing dot.
    domain = string

    # Human-readable description of the authorization.
    description = optional(string, "")

    # The Certificate Manager location. Defaults to "global" — the correct
    # choice for classic external HTTPS load balancers. Regional
    # authorizations pair with regional certificates only.
    # Immutable: changing the location destroys and recreates the
    # authorization.
    location = optional(string, "")

    # How the validation record is scoped:
    #   FIXED_RECORD (default for global): the classic DNS-01 style record,
    #     one per (domain, authorization).
    #   PER_PROJECT_RECORD: one record per (domain, project) — lets multiple
    #     Certificate Manager resources across projects share the same
    #     validation record, and is the default for non-global locations.
    # If omitted, GCP picks the location-appropriate default. Immutable.
    type = optional(string, "")

    # User labels merged onto the authorization beneath the platform's
    # attribution labels (platform keys win on conflicts).
    # Keys/values: lowercase letters, digits, underscores, hyphens.
    labels = optional(map(string), {})

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the authorization is deleted (certificates already
    #                issued against it keep serving; renewal for domains
    #                not yet serving through the LB stops validating)
    #   "PREVENT" -- destroy FAILS; protects the validation chain a
    #                certificate migration depends on
    #   "ABANDON" -- the authorization is removed from management but
    #                left in GCP, still validating its domain
    deletion_policy = optional(string, "")
  })
}
