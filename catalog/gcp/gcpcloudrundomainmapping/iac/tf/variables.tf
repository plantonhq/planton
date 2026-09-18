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
  description = "GcpCloudRunDomainMapping specification"
  type = object({
    # The GCP project that owns the domain mapping. Can be a literal
    # project ID or a reference to a GcpProject resource. If omitted, the
    # provider's default project is used. The domain must be verified by
    # an identity with access to this project.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # GCP region of the Cloud Run service being mapped (e.g. us-central1).
    # Domain mappings are regional and must be created in the SAME region
    # as their target service. Immutable: changing it replaces the mapping.
    region = string

    # The custom domain to map — this IS the mapping's name in GCP (e.g.
    # "app.example.com"). MUST already be verified by the provisioning
    # identity (Search Console / `gcloud domains verify`); GCP rejects the
    # create otherwise. Subdomains of a verified domain need no separate
    # verification. Immutable: changing it replaces the mapping.
    domain = string

    # The Cloud Run service this domain routes to. Reference a GcpCloudRun
    # resource or provide the service name literally. The service must
    # exist, in this same region and project, before the mapping is
    # created. Immutable: repointing the domain at a different service
    # replaces the mapping.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    route = string

    # How the domain's TLS certificate is provided:
    #   "AUTOMATIC" -- Cloud Run provisions and renews a managed
    #                  certificate (the default, and what almost every
    #                  mapping wants)
    #   "NONE"      -- no managed certificate; the domain serves without
    #                  TLS until one exists (used for migrations where the
    #                  DNS records must be published before certificate
    #                  issuance can succeed)
    # Immutable: changing it replaces the mapping.
    certificate_mode = optional(string, "")

    # When true, this mapping overrides any existing mapping of the same
    # domain without warning. Leave unset for the safe behavior: GCP then
    # fails the create with a conflict error instead of silently stealing
    # a domain another mapping already serves. Set it only after such a
    # conflict error confirmed the override is intended. Immutable.
    force_override = optional(bool, false)

    # The Cloud Run namespace for the mapping — GCP requires it to equal
    # the project ID or the project NUMBER. Leave empty for the sensible
    # default (the module uses the project ID); set it only when a
    # numbered-namespace convention requires the project number instead.
    # Immutable.
    namespace = optional(string, "")

    # Labels stored on the mapping object (Knative-style metadata labels,
    # grouped with the platform's own resource labels). Non-authoritative:
    # the module manages only the labels declared here plus the platform
    # set. Immutable: changing them replaces the mapping.
    labels = optional(map(string), {})

    # Annotations stored on the mapping object. Non-authoritative, and the
    # Cloud Run API adds server-side annotations of its own (those never
    # show up as drift — the module manages only the entries declared
    # here). Immutable: changing them replaces the mapping.
    annotations = optional(map(string), {})

    # Deletion policy for the domain mapping:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- destroying the resource deletes the mapping; the
    #                domain simply stops routing (the service is untouched)
    #   "PREVENT" -- destroy FAILS; protects a mapping production traffic
    #                depends on
    #   "ABANDON" -- the mapping is removed from management but keeps
    #                serving in GCP
    # This is the one field that updates in place (it never touches the
    # API object).
    deletion_policy = optional(string, "")
  })
}
