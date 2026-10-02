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
  description = "GcpIamOauthClient specification"
  type = object({
    # The GCP project that owns the OAuth client. Can be a literal project
    # ID or a reference to a GcpProject resource. If omitted, the
    # provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The location of the OAuth client. "global" is the documented home
    # for workforce OAuth clients and the default here. Immutable.
    location = optional(string, "")

    # The client's resource ID (the last segment of its name). Defaults to
    # metadata.name when left empty. Immutable: changing it destroys and
    # recreates the client — consumers must re-register.
    oauth_client_id = optional(string, "")

    # Human-readable name shown in consoles and consent surfaces.
    display_name = optional(string, "")

    # What this client is for — the operator-facing record.
    description = optional(string, "")

    # When true, the client stops accepting new authorizations without
    # being deleted — the reversible kill switch.
    disabled = optional(bool, false)

    # The client's confidentiality model. Only "CONFIDENTIAL_CLIENT"
    # (server-side apps that can keep a secret; manage secrets via
    # credentials below) can be created: GCP's enum also lists
    # PUBLIC_CLIENT (mobile/SPA), but the service rejects creating one
    # with 400 "Client type is not supported" (live-verified at the raw
    # API — no field combination unlocks it). Re-admit PUBLIC_CLIENT here
    # when GCP ships support. Immutable.
    client_type = optional(string, "")

    # OAuth grant types the client may use — Google's API accepts exactly
    # these values (a closed enum in the IAM REST API):
    #   "AUTHORIZATION_CODE_GRANT" -- the standard code flow
    #   "REFRESH_TOKEN_GRANT"      -- long-lived sessions via refresh
    #                                 tokens
    allowed_grant_types = list(string)

    # OAuth scopes the client may request during flows (e.g.
    # "https://www.googleapis.com/auth/cloud-platform", "openid",
    # "email", "profile", "groups").
    allowed_scopes = list(string)

    # Redirect URIs allowed when authorization completes. Each entry is a
    # literal URL or a reference to another resource's URL output (e.g. a
    # GcpCloudRun service's url) — referencing kills the drift between the
    # deployed app's address and the client's registration.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    allowed_redirect_uris = list(string)

    # Managed client secrets (CONFIDENTIAL_CLIENT only). Each entry
    # creates one credential whose secret value GCP generates server-side
    # — read it from this resource's client_secret output. Rotation story:
    # add a second credential, cut consumers over, remove the first.
    credentials = optional(list(object({
      # The credential's resource ID (the last segment of its name), e.g.
      # "primary". Immutable.
      credential_id = string

      # Human-readable name recording what consumes this credential.
      display_name = optional(string, "")

      # When true, the credential cannot be used to authenticate. GCP
      # requires a credential to be DISABLED before it can be deleted —
      # removing an entry from this list while it is still enabled fails at
      # the API; disable it in one apply, remove it in the next.
      disabled = optional(bool, false)
    })), [])

    # Deletion policy — one switch governs the client AND its credentials:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the client and its credentials are deleted; the
    #                client ID remains reserved briefly by GCP
    #   "PREVENT" -- destroy FAILS; protects a client live apps depend on
    #   "ABANDON" -- everything is removed from management but keeps
    #                working in GCP
    deletion_policy = optional(string, "")
  })
}
