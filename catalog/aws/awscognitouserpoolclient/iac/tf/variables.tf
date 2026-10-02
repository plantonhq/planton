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
  description = "AwsCognitoUserPoolClient specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The Cognito User Pool this client authenticates against.
    # Format: "{region}_{poolId}" (e.g., "us-east-1_Ab1Cd2EfG").
    # ForceNew -- a client cannot be moved between pools.
    # Accepts a direct pool ID or a reference to an AwsCognitoUserPool resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    user_pool_id = string

    # Whether AWS generates a client secret. ForceNew -- cannot be changed after
    # creation. Set true for confidential clients (server-side applications that
    # can protect the secret, machine-to-machine clients); leave false for
    # public clients (SPAs, mobile apps), which authenticate with PKCE instead.
    generate_secret = optional(bool, false)

    # Enable OAuth 2.0 flows for this client. Must be true for
    # `allowed_oauth_flows` and `allowed_oauth_scopes` to take effect.
    allowed_oauth_flows_user_pool_client = optional(bool, false)

    # OAuth 2.0 grant types this client can use. Valid values:
    # - "code": Authorization Code grant (recommended for most apps; pair with
    #   PKCE for public clients).
    # - "implicit": Implicit grant (legacy; tokens leak into browser history --
    #   avoid for new applications).
    # - "client_credentials": Client Credentials grant (machine-to-machine;
    #   requires a client secret and custom scopes from a resource server, and
    #   cannot be combined with the user-facing grants on the same client).
    allowed_oauth_flows = optional(list(string), [])

    # OAuth 2.0 scopes this client can request. Built-in values: "openid",
    # "email", "profile", "phone", "aws.cognito.signin.user.admin". Custom
    # scopes minted by an AwsCognitoResourceServer use the form
    # "{resource-server-identifier}/{scope_name}" -- reference the resource
    # server's scope_identifiers output for the exact strings.
    allowed_oauth_scopes = optional(list(string), [])

    # Callback URLs for OAuth redirects after authentication. Required for
    # Authorization Code and Implicit grants. Maximum 100 URLs.
    callback_urls = optional(list(string), [])

    # URLs where Cognito redirects after sign-out. Maximum 100 URLs.
    logout_urls = optional(list(string), [])

    # Default redirect URI. Must be one of the `callback_urls`. Used when no
    # redirect_uri is specified in the authorization request.
    default_redirect_uri = optional(string, "")

    # Identity providers this client offers at sign-in. Use the literal
    # "COGNITO" for the pool's own user directory, and add federated providers
    # by name -- either as literals ("Google", "CorpOkta") or as references to
    # AwsCognitoIdentityProvider resources (which also gives the deployment
    # graph the right ordering: the provider exists before the client lists it).
    # When omitted, AWS enables all of the pool's providers for this client.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    supported_identity_providers = optional(list(string), [])

    # Explicit authentication flows enabled for this client. Controls which
    # authentication APIs the client can use. Valid values:
    # - "ALLOW_USER_SRP_AUTH": Secure Remote Password (recommended)
    # - "ALLOW_REFRESH_TOKEN_AUTH": Enable token refresh (always recommended)
    # - "ALLOW_USER_PASSWORD_AUTH": Direct username/password (less secure)
    # - "ALLOW_ADMIN_USER_PASSWORD_AUTH": Admin-initiated auth
    # - "ALLOW_CUSTOM_AUTH": Custom auth flow via Lambda triggers
    # - "ALLOW_USER_AUTH": Choice-based sign-in (passwordless first factors --
    #   pairs with the pool's allowed_first_auth_factors)
    # The legacy pre-ALLOW spellings ("ADMIN_NO_SRP_AUTH",
    # "CUSTOM_AUTH_FLOW_ONLY", "USER_PASSWORD_AUTH") are accepted for pools that
    # still use them, but cannot be mixed with ALLOW_* values on one client.
    explicit_auth_flows = optional(list(string), [])

    # How long the session token of one sign-in attempt (the challenge
    # handshake) stays valid, in minutes. Range: 3-15. AWS default: 3.
    auth_session_validity = optional(number)

    # Access token lifetime, in `token_validity_units.access_token` units
    # (default: hours). AWS bounds the result to 5 minutes - 24 hours.
    # AWS default: 1 hour.
    access_token_validity = optional(number)

    # ID token lifetime, in `token_validity_units.id_token` units (default:
    # hours). AWS bounds the result to 5 minutes - 24 hours. AWS default: 1 hour.
    id_token_validity = optional(number)

    # Refresh token lifetime, in `token_validity_units.refresh_token` units
    # (default: days). AWS bounds the result to 60 minutes - 10 years.
    # AWS default: 30 days.
    refresh_token_validity = optional(number)

    # The units the three token lifetimes are expressed in.
    token_validity_units = optional(object({
      # Unit for access_token_validity: "seconds", "minutes", "hours" (AWS
      # default), or "days".
      access_token = optional(string, "")

      # Unit for id_token_validity: "seconds", "minutes", "hours" (AWS default),
      # or "days".
      id_token = optional(string, "")

      # Unit for refresh_token_validity: "seconds", "minutes", "hours", or "days"
      # (AWS default).
      refresh_token = optional(string, "")
    }))

    # Refresh-token rotation: each refresh issues a NEW refresh token and
    # retires the old one, shrinking the blast radius of a stolen token.
    # When ENABLED, do not also list ALLOW_REFRESH_TOKEN_AUTH in
    # explicit_auth_flows -- rotation owns the refresh behavior and AWS
    # rejects the combination.
    refresh_token_rotation = optional(object({
      # "ENABLED" or "DISABLED".
      feature = string

      # How long (seconds) the RETIRED refresh token keeps working after a
      # rotation, absorbing clients that lose the response carrying the new token.
      # Range: 0-60, where an EXPLICIT 0 disables the grace period (immediate
      # retirement -- AWS's default posture). Tri-state: unset leaves the choice
      # to AWS, explicit 0 pins immediate retirement, 1-60 grants that many
      # seconds of grace.
      retry_grace_period_seconds = optional(number)
    }))

    # Whether tokens can be revoked (sign-out revokes the refresh token and the
    # access/ID tokens minted from it). AWS default: true. Only set false when
    # the marginal token-validation latency matters more than revocability.
    enable_token_revocation = optional(bool)

    # Propagate client IP and user-agent to Cognito threat protection for
    # server-side flows (where Cognito otherwise sees only the server's IP).
    # Requires a client secret and the pool's threat protection to be active.
    enable_propagate_additional_user_context_data = optional(bool, false)

    # How to handle requests for non-existent users. Valid values:
    # - "ENABLED": Return the same error for non-existent and incorrect-password
    #   users to prevent user enumeration attacks (recommended).
    # - "LEGACY": Return different errors (reveals whether a user exists).
    prevent_user_existence_errors = optional(string, "")

    # User attributes this client can READ (e.g. "email", "email_verified",
    # "custom:tenant_id"). When omitted, AWS grants read access to all
    # attributes.
    read_attributes = optional(list(string), [])

    # User attributes this client can WRITE. When omitted, AWS grants write
    # access to all mutable attributes.
    write_attributes = optional(list(string), [])

    # Amazon Pinpoint analytics wiring: Cognito publishes sign-in/sign-up events
    # to a Pinpoint project for user-journey analytics.
    analytics_configuration = optional(object({
      # The Pinpoint project (application) ARN. Mutually exclusive with
      # `application_id`; when set, AWS derives the publish role itself.
      application_arn = optional(string, "")

      # The Pinpoint application ID. Requires `external_id` and `role_arn`.
      application_id = optional(string, "")

      # The external ID for the role assumption (confused-deputy guard).
      external_id = optional(string, "")

      # The IAM role Cognito assumes to publish events to Pinpoint. Accepts a
      # direct role ARN or a reference to an AwsIamRole resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = optional(string, "")

      # Whether Cognito includes user data (endpoint attributes) in the events
      # it publishes.
      user_data_shared = optional(bool, false)
    }))

    # Client-scoped risk configuration for threat protection. When set, this
    # client stops following the pool-wide risk configuration (set on the
    # AwsCognitoUserPool spec) and uses these responses instead. Requires the
    # POOL to have threat protection active (`user_pool_add_ons.
    # advanced_security_mode` of "AUDIT" or "ENFORCED") -- a cross-resource
    # requirement this spec cannot validate; AWS rejects the configuration
    # when threat protection is off.
    risk_configuration = optional(object({
      # Automated responses to sign-in attempts Cognito assesses as possible
      # account takeover, by risk level, plus the notification templates for
      # "notify the user" actions.
      account_takeover = optional(object({
        # Response to a LOW-risk assessment.
        low_action = optional(object({
          # The response Cognito takes (acts only in ENFORCED mode; AUDIT gathers
          # metrics without acting):
          # - "BLOCK": reject the sign-in attempt.
          # - "MFA_IF_CONFIGURED": require MFA when the user has a factor set up,
          #   otherwise allow the attempt.
          # - "MFA_REQUIRED": require MFA; users without a factor are blocked.
          # - "NO_ACTION": allow the attempt (pair with notify to warn the user).
          event_action = string

          # Whether Cognito emails the user about the event using the templates in
          # notify_configuration.
          notify = optional(bool, false)
        }))

        # Response to a MEDIUM-risk assessment.
        medium_action = optional(object({
          # The response Cognito takes (acts only in ENFORCED mode; AUDIT gathers
          # metrics without acting):
          # - "BLOCK": reject the sign-in attempt.
          # - "MFA_IF_CONFIGURED": require MFA when the user has a factor set up,
          #   otherwise allow the attempt.
          # - "MFA_REQUIRED": require MFA; users without a factor are blocked.
          # - "NO_ACTION": allow the attempt (pair with notify to warn the user).
          event_action = string

          # Whether Cognito emails the user about the event using the templates in
          # notify_configuration.
          notify = optional(bool, false)
        }))

        # Response to a HIGH-risk assessment.
        high_action = optional(object({
          # The response Cognito takes (acts only in ENFORCED mode; AUDIT gathers
          # metrics without acting):
          # - "BLOCK": reject the sign-in attempt.
          # - "MFA_IF_CONFIGURED": require MFA when the user has a factor set up,
          #   otherwise allow the attempt.
          # - "MFA_REQUIRED": require MFA; users without a factor are blocked.
          # - "NO_ACTION": allow the attempt (pair with notify to warn the user).
          event_action = string

          # Whether Cognito emails the user about the event using the templates in
          # notify_configuration.
          notify = optional(bool, false)
        }))

        # SES sending configuration and message templates for the notification
        # emails sent when an action has notify enabled. Required by AWS when any
        # action notifies the user.
        notify_configuration = optional(object({
          # The SES identity Cognito sends notification emails from. Accepts a
          # direct identity ARN or a reference to an AwsSesEmailIdentity resource.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          source_arn = string

          # "From" address shown on notification emails. When omitted, AWS derives
          # it from the source identity.
          from = optional(string, "")

          # Reply-to address on notification emails.
          reply_to = optional(string, "")

          # Template for the "we blocked a sign-in attempt" notification.
          block_email = optional(object({
            # Email subject. 1-140 characters.
            subject = string

            # HTML email body. 6-20000 characters.
            html_body = string

            # Plain-text email body. 6-20000 characters.
            text_body = string
          }))

          # Template for the "we required MFA on a sign-in attempt" notification.
          mfa_email = optional(object({
            # Email subject. 1-140 characters.
            subject = string

            # HTML email body. 6-20000 characters.
            html_body = string

            # Plain-text email body. 6-20000 characters.
            text_body = string
          }))

          # Template for the "we observed a risky sign-in attempt" (no action taken)
          # notification.
          no_action_email = optional(object({
            # Email subject. 1-140 characters.
            subject = string

            # HTML email body. 6-20000 characters.
            html_body = string

            # Plain-text email body. 6-20000 characters.
            text_body = string
          }))
        }))
      }))

      # Automated response when Cognito detects the presented credentials in a
      # known-compromised set.
      compromised_credentials = optional(object({
        # The response Cognito takes (acts only in ENFORCED mode):
        # - "BLOCK": reject the attempt.
        # - "NO_ACTION": allow the attempt (still logged/audited).
        event_action = string

        # Which authentication events are checked against the compromised set.
        # When empty, AWS checks all supported events. Valid values: "SIGN_IN",
        # "PASSWORD_CHANGE", "SIGN_UP".
        event_filter = optional(list(string), [])
      }))

      # IP-range exceptions that bypass or force the risk decision.
      risk_exception = optional(object({
        # Always-BLOCK list: authentication requests from these CIDR ranges are
        # rejected regardless of risk assessment. Up to 200 entries.
        blocked_ip_ranges = optional(list(string), [])

        # Always-ALLOW list: risk detection is skipped for these CIDR ranges.
        # Up to 200 entries.
        skipped_ip_ranges = optional(list(string), [])
      }))
    }))
  })
}
