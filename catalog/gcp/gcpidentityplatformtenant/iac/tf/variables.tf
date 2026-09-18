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
  description = "GcpIdentityPlatformTenant specification"
  type = object({
    # The GCP project containing the tenant. Can be a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used. The project's GcpIdentityPlatformConfig
    # must set multi_tenant.allow_tenants = true.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Human-readable tenant name (e.g. the customer organization) — shown
    # in consoles and usable in sign-in UIs. The tenant's resource ID is
    # server-generated and separate from this.
    display_name = string

    # Whether users can sign up with email/password in this tenant.
    allow_password_signup = optional(bool, false)

    # Whether users can sign in via emailed magic link (passwordless).
    enable_email_link_signin = optional(bool, false)

    # When true, ALL authentication in this tenant is disabled — the
    # maintenance/kill switch. Existing sessions stop refreshing.
    disable_auth = optional(bool, false)

    # Restrictions on what client applications can do against this tenant
    # through the Identity Toolkit API directly.
    client_permissions = optional(object({
      # When true, end users CANNOT sign up themselves through the API —
      # accounts are created only by your backend (admin SDK).
      disabled_user_signup = optional(bool, false)

      # When true, end users CANNOT delete their own accounts through the
      # API — deletion happens only through your backend.
      disabled_user_deletion = optional(bool, false)
    }))

    # Default supported identity providers (Google, Facebook, Apple, ...)
    # enabled for THIS tenant, each with the OAuth client obtained from
    # that provider's own developer console.
    default_supported_idps = optional(list(object({
      # Which provider this is — the provider's canonical IdP ID:
      # apple.com, facebook.com, gc.apple.com, github.com, google.com,
      # linkedin.com, microsoft.com, playgames.google.com, twitter.com,
      # yahoo.com. Immutable: changing it replaces the IdP config.
      idp_id = string

      # The OAuth client ID issued by the identity provider's own developer
      # console. Consent-screen clients have no programmatic creation path —
      # obtaining these is a documented console step.
      client_id = string

      # The OAuth client secret paired with client_id. A secret value: the
      # platform stores it as a managed-secret reference and resolves it
      # just-in-time at deploy. No Planton resource produces it (it comes
      # from the external provider's console), so it is a plain secret
      # string.
      client_secret = string

      # Whether users can sign in with this provider (default true). Both
      # IaC engines send the value explicitly so behavior is identical
      # regardless of engine.
      enabled = optional(bool)
    })), [])

    # Custom OIDC identity providers for this tenant.
    oauth_idp_configs = optional(list(object({
      # Resource name for this OIDC config — must start with "oidc."
      # (e.g. "oidc.corp-sso"). The API's naming rule, validated up front.
      # Immutable.
      name = string

      # Human-readable name shown in consoles and sign-in UIs. Required at
      # tenant level (the API's own difference from the project level).
      display_name = string

      # The OIDC issuer URL (e.g. "https://accounts.example.com") — where
      # Identity Platform fetches the provider's discovery document.
      issuer = string

      # The OAuth client ID registered with the OIDC provider.
      client_id = string

      # The OAuth client secret — needed when the provider exchange uses a
      # code flow. A secret value handled as a managed-secret reference end
      # to end. (Tenant-level OIDC has no response_type selection — the
      # API's own difference from the project level.)
      client_secret = optional(string, "")

      # Whether users can sign in with this provider (default true). Sent
      # explicitly by both engines.
      enabled = optional(bool)
    })), [])

    # Inbound SAML identity providers (enterprise SSO) for this tenant.
    inbound_saml_configs = optional(list(object({
      # Resource name for this SAML config — must start with "saml.",
      # contain only alphanumerics/hyphens/underscores/periods, and the part
      # after "saml." must start with a lowercase letter, end alphanumeric,
      # and be at least 2 characters (the API's naming rule, validated up
      # front). E.g. "saml.okta-prod". Immutable.
      name = string

      # Human-readable name shown in consoles and sign-in UIs.
      display_name = string

      # Whether users can sign in with this provider (default true). Sent
      # explicitly by both engines.
      enabled = optional(bool)

      # The external identity provider's side of the SAML exchange.
      idp_config = object({
        # The IdP's entity ID from its metadata XML.
        idp_entity_id = string

        # The IdP's SSO URL — where users are sent to authenticate.
        sso_url = string

        # Whether to sign outbound authentication requests.
        sign_request = optional(bool, false)

        # The IdP's X.509 signing certificates (PEM) used to verify SAML
        # responses. Public certificates, not secrets.
        idp_certificates = optional(list(object({
          # The certificate in PEM format.
          x509_certificate = optional(string, "")
        })), [])
      })

      # This tenant's side of the SAML exchange. Required at tenant level
      # (the API's own difference from the project level, where it is
      # optional).
      sp_config = object({
        # Where the IdP posts SAML responses — must be an https:// URL (the
        # API's rule, validated up front).
        callback_uri = string

        # This tenant's SP entity ID, as registered with the IdP.
        sp_entity_id = string
      })
    })), [])

    # Deletion policy — one switch governs the tenant AND its composed
    # IdP configs:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the tenant is deleted with ALL its users and
    #                providers; user accounts unrecoverable
    #   "PREVENT" -- destroy FAILS; protects a live customer user pool
    #   "ABANDON" -- everything is removed from management but keeps
    #                working in GCP
    deletion_policy = optional(string, "")
  })
}
