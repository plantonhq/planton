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
  description = "GcpIdentityPlatformConfig specification"
  type = object({
    # The GCP project whose Identity Platform is configured. Can be a
    # literal project ID or a reference to a GcpProject resource. If
    # omitted, the provider's default project is used. The project must
    # have BILLING enabled — initialization fails without it.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # First-party sign-in methods. Each arm you set is sent explicitly
    # (including enabled=false) so the project's state always matches the
    # manifest; arms you omit are left unmanaged.
    sign_in = optional(object({
      # Email/password sign-in.
      email = optional(object({
        # Whether email sign-in is enabled. Required inside the arm: declaring
        # the arm takes the method under management, and the switch says which
        # way (false actively disables it).
        enabled = bool

        # Whether a password is required for email accounts. When false,
        # users can sign in via email link alone.
        password_required = optional(bool, false)
      }))

      # Phone-number (SMS code) sign-in.
      phone_number = optional(object({
        # Whether phone sign-in is enabled. Required inside the arm: declaring
        # the arm takes the method under management, and the switch says which
        # way (false actively disables it).
        enabled = bool

        # Test phone numbers mapped to fixed verification codes (e.g.
        # "+15555550100" -> "123456") — lets CI and reviewers exercise the
        # phone flow without receiving real SMS. Never ship real numbers here.
        test_phone_numbers = optional(map(string), {})
      }))

      # Anonymous (guest) sign-in — accounts created without credentials,
      # typically upgraded to a real provider later.
      anonymous = optional(object({
        # Whether anonymous (guest) sign-in is enabled. Required inside the arm:
        # declaring the arm takes the method under management, and the switch
        # says which way (false actively disables it).
        enabled = bool
      }))

      # Whether multiple accounts may share one email address. Most apps
      # leave this false so an email maps to exactly one account.
      allow_duplicate_emails = optional(bool, false)
    }))

    # Domains authorized for OAuth redirects and hosted sign-in flows
    # (e.g. "myapp.example.com"). localhost and the project's Firebase
    # subdomains are authorized by default when this list is left empty.
    authorized_domains = optional(list(string), [])

    # Multi-factor authentication policy for the project.
    mfa = optional(object({
      # Project-wide MFA state (provider-validated values):
      #   "DISABLED"  -- MFA cannot be used
      #   "ENABLED"   -- users may enroll a second factor
      #   "MANDATORY" -- every user must present a second factor
      state = optional(string, "")

      # Usable second factors. The provider accepts only "PHONE_SMS" here;
      # TOTP is configured through provider_configs.
      enabled_providers = optional(list(string), [])

      # Per-provider MFA configuration (currently TOTP authenticator apps).
      provider_configs = optional(list(object({
        # This provider's state: DISABLED, ENABLED, or MANDATORY.
        state = optional(string, "")

        # TOTP (authenticator-app) settings.
        totp_provider_config = optional(object({
          # How many adjacent 30-second code windows are accepted (clock-skew
          # tolerance). GCP accepts up to 10.
          adjacent_intervals = optional(number, 0)
        }))
      })), [])
    }))

    # Blocking functions — Cloud Functions invoked synchronously during
    # sign-up/sign-in that can allow, block, or modify the operation
    # (custom claims, domain allowlists, fraud checks).
    blocking_functions = optional(object({
      # The trigger points and the Cloud Functions they invoke. GCP supports
      # the "beforeCreate" (before an account is created) and "beforeSignIn"
      # (before a sign-in completes) event types.
      triggers = list(object({
        # The trigger point: "beforeCreate" or "beforeSignIn".
        event_type = string

        # The HTTPS endpoint of the Cloud Function to invoke — a literal URL
        # or a reference to a GcpCloudFunction resource (its function_url
        # output is exactly this value).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        function_uri = string
      }))

      # Which of the user's tokens are forwarded to the blocking function —
      # grant only what the function's logic actually inspects.
      forward_inbound_credentials = optional(object({
        # Forward the user's OAuth access token.
        access_token = optional(bool, false)

        # Forward the user's OIDC ID token.
        id_token = optional(bool, false)

        # Forward the user's refresh token.
        refresh_token = optional(bool, false)
      }))
    }))

    # Temporary sign-up quota override — all three fields are set together
    # (GCP needs the ceiling, the window length, and when it starts).
    sign_up_quota = optional(object({
      # Sign-ups allowed during the window — between 1 and 1000 (the API's
      # documented range).
      quota = optional(number, 0)

      # How long the override stays active, as a seconds duration (e.g.
      # "86400s" for one day).
      quota_duration = optional(string, "")

      # When the override takes effect (RFC3339 UTC, e.g.
      # "2026-09-01T00:00:00Z").
      start_time = optional(string, "")
    }))

    # Which regions may receive SMS (verification codes, MFA) — the
    # toll-fraud control. Exactly one policy arm.
    sms_region_config = optional(object({
      # Allow SMS to every region EXCEPT the listed ones.
      allow_by_default = optional(object({
        # Two-letter CLDR region codes to disallow (e.g. "KP", "RU").
        disallowed_regions = optional(list(string), [])
      }))

      # Allow SMS ONLY to the listed regions — the tighter toll-fraud
      # posture.
      allowlist_only = optional(object({
        # Two-letter CLDR region codes to allow (e.g. "US", "DE").
        allowed_regions = optional(list(string), [])
      }))
    }))

    # Restrictions on what client applications can do through the
    # Identity Toolkit API directly.
    client_permissions = optional(object({
      # When true, end users CANNOT sign up themselves through the API —
      # accounts are created only by your backend (admin SDK).
      disabled_user_signup = optional(bool, false)

      # When true, end users CANNOT delete their own accounts through the
      # API — deletion happens only through your backend.
      disabled_user_deletion = optional(bool, false)
    }))

    # Whether sign-in/sign-up requests are written to Cloud Logging. Sent
    # explicitly when set (true or false); leave unset to keep GCP's
    # current value unmanaged.
    request_logging_enabled = optional(bool)

    # Multi-tenancy: allow_tenants must be true before any
    # GcpIdentityPlatformTenant can be created in this project.
    multi_tenant = optional(object({
      # Whether this project may contain tenants. Must be true before any
      # GcpIdentityPlatformTenant is created in the project.
      allow_tenants = optional(bool, false)

      # Default GCP location for tenant data (e.g. "global").
      default_tenant_location = optional(string, "")
    }))

    # When true, anonymous users are deleted automatically after ~30 days
    # of inactivity — keeps abandoned guest sessions from accumulating.
    autodelete_anonymous_users = optional(bool, false)

    # Default supported identity providers (Google, Facebook, Apple, ...)
    # enabled for the whole project, each with the OAuth client obtained
    # from that provider's own developer console.
    default_supported_idps = optional(list(object({
      # Which provider this is — the provider's canonical IdP ID:
      # apple.com, facebook.com, gc.apple.com, github.com, google.com,
      # linkedin.com, microsoft.com, playgames.google.com, twitter.com,
      # yahoo.com. Immutable: changing it replaces the IdP config.
      idp_id = string

      # The OAuth client ID issued by the identity provider's own developer
      # console (e.g. Google Cloud Console for google.com, Meta for
      # facebook.com). Consent-screen clients have no programmatic creation
      # path — obtaining these is a documented console step.
      client_id = string

      # The OAuth client secret paired with client_id. A secret value: the
      # platform stores it as a managed-secret reference and resolves it
      # just-in-time at deploy — it never sits in plaintext in the control
      # plane. No Planton resource produces it (it comes from the external
      # provider's console), so it is a plain secret string.
      client_secret = string

      # Whether users can sign in with this provider (default true). Both
      # IaC engines send the value explicitly so behavior is identical
      # regardless of engine.
      enabled = optional(bool)
    })), [])

    # Custom OIDC identity providers for the whole project.
    oauth_idp_configs = optional(list(object({
      # Resource name for this OIDC config — must start with "oidc."
      # (e.g. "oidc.corp-sso"). The API's naming rule, validated up front.
      # Immutable.
      name = string

      # Human-readable name shown in consoles and sign-in UIs.
      display_name = optional(string, "")

      # The OIDC issuer URL (e.g. "https://accounts.example.com") — where
      # Identity Platform fetches the provider's discovery document.
      issuer = string

      # The OAuth client ID registered with the OIDC provider.
      client_id = string

      # The OAuth client secret — required for the authorization-code flow
      # (response_type.code), unused for the implicit id_token flow. A
      # secret value handled as a managed-secret reference end to end.
      client_secret = optional(string, "")

      # Whether users can sign in with this provider (default true). Sent
      # explicitly by both engines.
      enabled = optional(bool)

      # Which OAuth response type to request from the provider.
      response_type = optional(object({
        # Authorization-code flow — the provider returns a code exchanged
        # server-side (requires client_secret). The more secure flow.
        code = optional(bool, false)

        # Implicit flow — the provider returns an ID token directly.
        id_token = optional(bool, false)
      }))
    })), [])

    # Inbound SAML identity providers (enterprise SSO) for the whole
    # project.
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

      # This project's side of the SAML exchange (the service provider).
      sp_config = optional(object({
        # Where the IdP posts SAML responses — must be an https:// URL (the
        # API's rule, validated up front).
        callback_uri = optional(string, "")

        # This project's SP entity ID, as registered with the IdP.
        sp_entity_id = optional(string, "")
      }))
    })), [])

    # Deletion policy for the COMPOSED identity-provider configs (the
    # default-supported/OIDC/SAML entries above) — the project config
    # itself cannot be deleted (destroy always abandons it in place):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the IdP configs are deleted; users can no longer sign
    #                in through them (accounts themselves are untouched)
    #   "PREVENT" -- destroy FAILS; protects a live sign-in surface
    #   "ABANDON" -- the IdP configs are removed from management but keep
    #                working in GCP
    deletion_policy = optional(string, "")

    # Adopt the project's EXISTING Identity Platform configuration instead
    # of initializing it. Initialization is one-way and once-only: GCP
    # rejects a second initializeAuth with 400 "Identity Platform has
    # already been enabled for this project" (live-verified), and there is
    # no de-initialize. Set true for any project where Identity Platform
    # was ever enabled — by a console click, a Firebase Auth setup, or a
    # previous deployment of this kind (destroy abandons the configuration
    # in place, so a re-deploy after destroy also needs it). The module
    # then imports the singleton (projects/{project}/config) and applies
    # this spec as an update. Leave false only for a project whose
    # Identity Platform has never been enabled.
    adopt_existing = optional(bool, false)
  })
}
