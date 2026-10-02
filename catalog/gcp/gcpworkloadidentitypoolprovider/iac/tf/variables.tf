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
  description = "GcpWorkloadIdentityPoolProvider specification"
  type = object({
    # The pool this provider belongs to — the bare pool ID (the final component
    # of the pool's resource name), not the full path.
    # Reference a GcpWorkloadIdentityPool resource — its
    # workload_identity_pool_id output is exactly this value.
    # Immutable: a provider cannot move between pools.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    workload_identity_pool_id = string

    # The ID for the provider, which becomes the final component of its
    # resource name. 4-32 characters of lowercase letters, digits, and hyphens;
    # the prefix "gcp-" is reserved by Google. Immutable: changing it destroys
    # and recreates the provider, invalidating tokens minted for the old
    # audience.
    workload_identity_pool_provider_id = string

    # The GCP project that owns the pool (and therefore this provider).
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the provider.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Human-readable name shown in the GCP console (max 32 characters). Mutable.
    display_name = optional(string, "")

    # Which issuer this provider trusts and any operational notes — write it
    # for the operator auditing trust boundaries later (max 256 characters).
    # Mutable.
    description = optional(string, "")

    # Emergency kill switch: a disabled provider rejects new token exchanges
    # (already-issued Google credentials remain valid until they expire).
    # Prefer disabling over deleting — deletion starts the 30-day soft-delete
    # clock and blocks ID reuse. Mutable.
    disabled = optional(bool, false)

    # Maps claims from the issuer's credential to Google attributes. Keys are
    # "google.subject" (required — the principal IAM authenticates),
    # "google.groups", or custom "attribute.<name>" entries (max 50); values
    # are CEL expressions over the credential's claims via the `assertion`
    # keyword, e.g. {"google.subject": "assertion.sub",
    # "attribute.repository": "assertion.repository"}.
    # Mapped attributes are what IAM bindings can target:
    # principalSet://iam.googleapis.com/<pool>/attribute.<name>/<value>.
    # OIDC providers must define it explicitly; AWS, SAML, and X.509 providers
    # fall back to sensible issuer-specific defaults when omitted.
    attribute_mapping = optional(map(string), {})

    # A CEL expression gating which otherwise valid credentials are accepted,
    # over the `assertion`, `google`, and `attribute` keywords (max 4096
    # characters), e.g. restricting a GitHub provider to one org:
    # assertion.repository_owner == "my-org".
    # Without a condition, ANY identity the issuer vouches for can federate —
    # for multi-tenant issuers like GitHub Actions, always set one.
    attribute_condition = optional(string, "")

    # Trust an Amazon Web Services account: workloads presenting AWS
    # credentials from this account can federate.
    aws = optional(object({
      # The 12-digit AWS account ID whose workloads may federate.
      account_id = string
    }))

    # Trust an OpenID Connect issuer — the workhorse for keyless CI/CD
    # (GitHub Actions, GitLab CI, Kubernetes clusters, custom issuers).
    oidc = optional(object({
      # The OIDC issuer URL, e.g. https://token.actions.githubusercontent.com
      # for GitHub Actions. Must match the `iss` claim of incoming tokens.
      issuer_uri = string

      # Acceptable `aud` (audience) values in incoming tokens; each at most 256
      # characters, at most 10 entries. When empty, the audience must equal the
      # provider's full canonical resource name (with or without the https:
      # prefix) — the safest default, since tokens minted for anything else are
      # rejected.
      allowed_audiences = optional(list(string), [])

      # OIDC JWKS (public signing keys) in JSON format, for issuers whose keys
      # cannot be fetched from the issuer_uri's .well-known discovery document
      # (e.g. issuers behind a private network). Leave unset to use discovery —
      # the normal path. These are public keys, not a secret.
      jwks_json = optional(string, "")
    }))

    # Trust a SAML 2.0 identity provider, typically an enterprise IdP.
    saml = optional(object({
      # The SAML identity provider's configuration metadata XML document, as
      # exported by the IdP. Public IdP metadata (entity ID, SSO endpoints,
      # signing certificates), not a secret.
      idp_metadata_xml = string
    }))

    # Trust an X.509 certificate authority: clients presenting certificates
    # that chain to the configured trust store can federate (certificate-based
    # workload identity without any token issuer).
    x509 = optional(object({
      # The trust store validating incoming end-entity certificates. Exactly one
      # trust store is supported per provider.
      trust_store = object({
        # Trust anchors: an incoming end-entity certificate must chain up to one of
        # these.
        trust_anchors = list(object({
          # PEM certificate of the PKI used for validation. Must contain exactly one
          # CA certificate (root or intermediate). Public key material, not a secret.
          pem_certificate = string
        }))

        # Intermediate CA certificates available for building the chain from an
        # end-entity certificate to a trust anchor.
        intermediate_cas = optional(list(object({
          # PEM certificate of the PKI used for validation. Must contain exactly one
          # CA certificate (either root or intermediate cert). Public key material,
          # not a secret.
          pem_certificate = string
        })), [])
      })
    }))

    # Deletion policy for the provider — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the provider is deleted (starts the ~30-day soft-delete
    #                clock and blocks the ID from reuse until it expires)
    #   "PREVENT" -- destroy FAILS; protects the keyless-auth path every
    #                pipeline federating through this issuer depends on
    #   "ABANDON" -- the provider is removed from management but keeps
    #                exchanging tokens in GCP
    deletion_policy = optional(string, "")
  })
}
