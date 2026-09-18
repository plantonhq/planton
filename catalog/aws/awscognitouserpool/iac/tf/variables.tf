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
  description = "AwsCognitoUserPool specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Attributes that users can use as their username when signing in. When set,
    # the chosen attribute(s) become the sign-in identifier. Common choice: ["email"]
    # to let users sign in with their email address.
    #
    # Valid values: "email", "phone_number".
    # Mutually exclusive with `alias_attributes`. ForceNew.
    username_attributes = optional(list(string), [])

    # Attributes that can be used as aliases for the username. Unlike `username_attributes`,
    # users always have a separate username and the alias attributes are alternative sign-in
    # identifiers. Common choice: ["email", "preferred_username"].
    #
    # Valid values: "email", "phone_number", "preferred_username".
    # Mutually exclusive with `username_attributes`. ForceNew.
    alias_attributes = optional(list(string), [])

    # Whether usernames are case-sensitive. When false (default in AWS), "User" and "user"
    # are treated as the same username. ForceNew -- cannot be changed after pool creation.
    username_case_sensitive = optional(bool, false)

    # Enable deletion protection. When true, the user pool cannot be deleted
    # without first disabling this setting. Recommended for production pools --
    # a deleted pool takes every user account with it, unrecoverably.
    deletion_protection = optional(bool, false)

    # The feature tier of the user pool. Governs which capabilities AWS makes
    # available (and how the pool is billed):
    # - "LITE": lowest-cost tier; core sign-in only.
    # - "ESSENTIALS": the AWS default for new pools; adds passwordless
    #   sign-in (email/SMS OTP, passkeys) and managed login branding.
    # - "PLUS": adds threat protection (advanced security features).
    # When omitted, AWS creates the pool on the ESSENTIALS tier. Downgrading a
    # tier is allowed but AWS rejects it while a feature exclusive to the
    # higher tier is still configured (e.g. PLUS -> ESSENTIALS with threat
    # protection enforced).
    user_pool_tier = optional(string, "")

    # Password policy configuration. Controls password strength requirements for
    # user self-registration and admin-created passwords.
    password_policy = optional(object({
      # Minimum password length. Range: 6-99. AWS default: 8.
      minimum_length = optional(number, 0)

      # Require at least one lowercase letter.
      require_lowercase = optional(bool, false)

      # Require at least one uppercase letter.
      require_uppercase = optional(bool, false)

      # Require at least one digit.
      require_numbers = optional(bool, false)

      # Require at least one special character.
      require_symbols = optional(bool, false)

      # Number of previous passwords a new password must differ from. Range: 0-24.
      # 0 is AWS's default posture (history checking off), so leaving this unset
      # and setting 0 are the same policy. Password history requires the pool to
      # be on the ESSENTIALS tier or higher.
      password_history_size = optional(number, 0)

      # Number of days temporary passwords (admin-created) are valid. Range:
      # 0-365. AWS default: 7 -- and AWS treats a submitted 0 as null, applying
      # that default, so 0 is NOT "no expiration"; there is no way to make
      # temporary passwords permanent. Leaving this unset keeps the 7-day default.
      temporary_password_validity_days = optional(number, 0)
    }))

    # The authentication factors users may present as their FIRST factor when
    # signing in. This is the passwordless dial: leaving it empty keeps the
    # classic password-first behavior, while setting it enables the choice-based
    # sign-in flow (USER_AUTH) with the listed factors.
    #
    # Valid values:
    # - "PASSWORD": classic password sign-in.
    # - "EMAIL_OTP": one-time code delivered by email (ESSENTIALS tier or higher).
    # - "SMS_OTP": one-time code delivered by SMS (requires `sms_configuration`).
    # - "WEB_AUTHN": passkeys/security keys (configure `web_authn_configuration`
    #   to pin the relying-party ID your apps expect).
    allowed_first_auth_factors = optional(list(string), [])

    # Multi-factor authentication enforcement level.
    # - "OFF": MFA is not used (default).
    # - "OPTIONAL": Users can opt in to MFA.
    # - "ON": MFA is required for all users.
    mfa_configuration = optional(string, "")

    # Enable TOTP-based (time-based one-time password) software token MFA. When
    # true, users can configure authenticator apps like Google Authenticator or
    # Authy. Requires `mfa_configuration` to be "OPTIONAL" or "ON".
    software_token_mfa_enabled = optional(bool, false)

    # Email-based MFA: Cognito emails a one-time code as the second factor.
    # Requires `mfa_configuration` to be "OPTIONAL" or "ON", and AWS requires
    # the pool to send email through SES (`email_configuration` in "DEVELOPER"
    # mode) because MFA codes exceed the built-in sender's delivery guarantees.
    email_mfa = optional(object({
      # The email body for MFA codes. Must contain the "{####}" placeholder where
      # Cognito injects the code. 6-20000 characters. When omitted, AWS uses its
      # default message.
      message = optional(string, "")

      # The email subject for MFA codes. 1-140 characters. When omitted, AWS uses
      # its default subject.
      subject = optional(string, "")

      # Whether email MFA is on. Unset means on: declaring the block has always
      # meant enabling it, and this switch lets a manifest say the opposite out
      # loud.
      enabled = optional(bool)
    }))

    # WebAuthn (passkey / security key) relying-party configuration. Configure
    # this when "WEB_AUTHN" is an allowed first factor so registered passkeys
    # are bound to the domain your applications serve.
    web_authn = optional(object({
      # The relying-party ID passkeys are registered against -- the domain your
      # applications serve (e.g. "auth.example.com"). A passkey registered for one
      # relying-party ID does not work for another, so set this BEFORE users start
      # registering passkeys; changing it later invalidates existing registrations.
      # When omitted, Cognito uses the pool's own domain.
      relying_party_id = optional(string, "")

      # Whether the authenticator must verify the user (PIN, biometric):
      # - "required": authenticators must perform user verification.
      # - "preferred": verification is requested but not required (AWS default).
      user_verification = optional(string, "")
    }))

    # SMS delivery configuration. Cognito publishes SMS messages (verification
    # codes, MFA codes, invitations) through Amazon SNS by assuming the IAM role
    # referenced here. Required whenever any SMS-dependent feature is enabled:
    # phone_number in `auto_verified_attributes`, "SMS_OTP" as a first factor,
    # or SMS MFA.
    sms_configuration = optional(object({
      # The IAM role Cognito assumes to send SMS via SNS. Accepts a direct role
      # ARN or a reference to an AwsIamRole resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      sns_caller_arn = string

      # The external ID Cognito presents when assuming the role -- the standard
      # confused-deputy guard. Must match the sts:ExternalId condition in the
      # role's trust policy.
      external_id = string

      # The AWS region of the SNS topic-less publish (where SMS messages
      # originate). When omitted, AWS uses the pool's region. Set this when the
      # pool's region does not support SMS sending.
      sns_region = optional(string, "")
    }))

    # The SMS message sent for sign-in authentication codes. Must contain the
    # "{####}" placeholder where Cognito injects the code. 6-140 characters.
    # When omitted, AWS uses its default message.
    sms_authentication_message = optional(string, "")

    # Attributes to auto-verify when users sign up. Cognito sends a verification
    # code to these attributes. Common values: ["email"].
    # Valid values: "email", "phone_number".
    auto_verified_attributes = optional(list(string), [])

    # Attributes that must be verified before an update to them takes effect.
    # While the new value is pending verification, Cognito keeps the previous
    # value active -- without this, an unverified typo in an email update can
    # lock a user out of account recovery. Valid values: "email", "phone_number".
    attributes_require_verification_before_update = optional(list(string), [])

    # Account recovery mechanisms in priority order. Each mechanism defines a
    # fallback method for users who forget their password.
    account_recovery_mechanisms = optional(list(object({
      # Recovery method name. Valid values:
      # - "verified_email": send a recovery code to the verified email
      # - "verified_phone_number": send a recovery code via SMS
      # - "admin_only": only administrators can reset passwords
      name = string

      # Priority of this recovery method. 1 = primary, 2 = fallback.
      priority = number
    })), [])

    # Email sending configuration. Controls whether Cognito sends emails using
    # its built-in service (limited to 50/day in sandbox) or your verified SES
    # identity (production sending).
    email_configuration = optional(object({
      # Email sending mode. Valid values:
      # - "COGNITO_DEFAULT": Cognito sends emails using its built-in service.
      #   Limited to 50 emails/day in sandbox mode. No SES setup required.
      # - "DEVELOPER": Cognito sends emails through your verified SES identity.
      #   Requires `source_arn`. Supports production-level sending volumes.
      email_sending_account = optional(string, "")

      # SES verified identity ARN for sending emails. Required when
      # `email_sending_account` is "DEVELOPER". Accepts a direct identity ARN or
      # a reference to an AwsSesEmailIdentity resource. Example:
      # "arn:aws:ses:us-east-1:123456789012:identity/noreply@example.com"
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      source_arn = optional(string, "")

      # "From" email address shown to recipients. Only applicable when using
      # DEVELOPER mode with SES. Example: "No Reply <noreply@example.com>"
      from_email_address = optional(string, "")

      # Reply-to email address. When set, user replies go to this address
      # instead of the "from" address.
      reply_to_email_address = optional(string, "")

      # SES configuration set for tracking email delivery metrics (opens, bounces,
      # complaints). Accepts a direct configuration-set name or a reference to an
      # AwsSesConfigurationSet resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      configuration_set = optional(string, "")
    }))

    # Templates for the verification message Cognito sends when users sign up
    # or change a verified attribute. Also selects between code-based and
    # link-based email verification.
    verification_message_template = optional(object({
      # How email verification is performed:
      # - "CONFIRM_WITH_CODE": the email carries a code the user types back
      #   (AWS default; uses `email_message`/`email_subject`).
      # - "CONFIRM_WITH_LINK": the email carries a click-through confirmation
      #   link (uses `email_message_by_link`/`email_subject_by_link`).
      default_email_option = optional(string, "")

      # Email body for code-based verification. Must contain the "{####}"
      # placeholder where Cognito injects the code. 6-20000 characters.
      email_message = optional(string, "")

      # Email subject for code-based verification. 1-140 characters.
      email_subject = optional(string, "")

      # Email body for link-based verification. Must contain the "{##...##}"
      # placeholder pair wrapping the link text, e.g.
      # "Click {##here##} to verify your address." 6-20000 characters.
      email_message_by_link = optional(string, "")

      # Email subject for link-based verification. 1-140 characters.
      email_subject_by_link = optional(string, "")

      # SMS body for phone verification. Must contain the "{####}" placeholder.
      # 6-140 characters.
      sms_message = optional(string, "")
    }))

    # When true, only administrators can create users -- self-registration is
    # disabled. Users must be created via the admin API or AWS console.
    allow_admin_create_user_only = optional(bool, false)

    # Templates for the invitation message sent to admin-created users with
    # their temporary credentials.
    invite_message_template = optional(object({
      # Email body for invitations. Must contain BOTH the "{username}" and
      # "{####}" placeholders (Cognito injects the username and the temporary
      # password). 6-20000 characters.
      email_message = optional(string, "")

      # Email subject for invitations. 1-140 characters.
      email_subject = optional(string, "")

      # SMS body for invitations. Must contain BOTH the "{username}" and "{####}"
      # placeholders. 6-140 characters.
      sms_message = optional(string, "")
    }))

    # Remembered-device configuration. When devices are remembered, users can
    # skip MFA on trusted devices, and sign-in events carry a device key.
    device_configuration = optional(object({
      # When true, a remembered device still requires a challenge (MFA) the first
      # time it is seen -- remembering only suppresses challenges afterwards.
      # Unset is false.
      challenge_required_on_new_device = optional(bool)

      # When true, devices are remembered only after the user opts in when
      # prompted. When false (or unset), every device is remembered automatically.
      device_only_remembered_on_user_prompt = optional(bool)
    }))

    # Custom user attributes beyond the standard set (email, phone, name, etc.).
    # Each attribute is added to the pool's schema. The schema is APPEND-ONLY:
    # new attributes can be added at any time, but an existing attribute can
    # never be modified or removed (AWS has no API for it -- removing one from
    # this list errors instead of recreating the pool). The `mutable` and
    # `required` flags are likewise fixed at the moment the attribute is added.
    # Maximum 50 custom attributes per pool.
    custom_attributes = optional(list(object({
      # Attribute name (1-20 characters). Cognito auto-prefixes with "custom:".
      name = string

      # Data type. Valid values: "String", "Number", "DateTime", "Boolean".
      attribute_data_type = string

      # Whether the attribute value can be changed after creation. Fixed at the
      # moment the attribute is added to the pool (the schema is append-only).
      mutable = optional(bool, false)

      # Whether the attribute is required during user registration. Fixed at the
      # moment the attribute is added to the pool.
      required = optional(bool, false)

      # When true, only administrators (not the user themselves) can read and
      # write this attribute -- for backend bookkeeping like an external system's
      # record ID. Fixed at the moment the attribute is added.
      developer_only_attribute = optional(bool, false)

      # Minimum string length (for String attributes). Leave at 0 for no minimum.
      string_min_length = optional(string, "")

      # Maximum string length (for String attributes). Leave empty for no maximum.
      string_max_length = optional(string, "")

      # Minimum numeric value (for Number attributes). Leave empty for no minimum.
      number_min_value = optional(string, "")

      # Maximum numeric value (for Number attributes). Leave empty for no maximum.
      number_max_value = optional(string, "")
    })), [])

    # Lambda trigger configuration for custom authentication and user lifecycle
    # hooks. Each trigger is an optional Lambda function ARN that Cognito invokes
    # at the corresponding lifecycle point.
    lambda_config = optional(object({
      # Invoked before user sign-up to perform custom validation or auto-confirm.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      pre_sign_up = optional(string, "")

      # Invoked before authentication to perform custom validation.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      pre_authentication = optional(string, "")

      # Invoked after successful authentication for logging or custom logic.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      post_authentication = optional(string, "")

      # Invoked after user confirmation to trigger welcome emails or provisioning.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      post_confirmation = optional(string, "")

      # Invoked before token generation to add, remove, or modify claims. This is
      # the V1_0 trigger event; use `pre_token_generation_config` instead when the
      # function needs the V2_0/V3_0 event shape (access-token customization).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      pre_token_generation = optional(string, "")

      # Versioned pre-token-generation trigger. Unlike the plain
      # `pre_token_generation` field (which pins the V1_0 event), this selects the
      # trigger event version -- V2_0/V3_0 deliver the richer event that can also
      # customize ACCESS tokens, not just identity tokens. Set one or the other,
      # never both.
      pre_token_generation_config = optional(object({
        # The Lambda function to invoke. Accepts a direct ARN or a reference to an
        # AwsLambda resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        lambda_arn = string

        # The trigger event version:
        # - "V1_0": identity-token customization only.
        # - "V2_0": adds access-token customization.
        # - "V3_0": V2_0 plus group/role overrides for machine-to-machine flows.
        # V2_0/V3_0 require the ESSENTIALS tier or higher.
        lambda_version = string
      }))

      # Invoked to customize verification and invitation messages.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      custom_message = optional(string, "")

      # Invoked during user migration from an external identity provider.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      user_migration = optional(string, "")

      # Invoked to define a custom authentication challenge (custom auth flow).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      define_auth_challenge = optional(string, "")

      # Invoked to create a custom authentication challenge (custom auth flow).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      create_auth_challenge = optional(string, "")

      # Invoked to verify the response to a custom authentication challenge.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      verify_auth_challenge_response = optional(string, "")

      # Custom email sender: Cognito calls this function to deliver email itself
      # (instead of SES). The event carries the message encrypted with
      # `kms_key_id`, so that key is required.
      custom_email_sender = optional(object({
        # The Lambda function Cognito invokes to deliver the message. Accepts a
        # direct ARN or a reference to an AwsLambda resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        lambda_arn = string

        # The sender event version. "V1_0" is the only version AWS currently
        # defines; modeled as a field so new versions are additive, not breaking.
        lambda_version = string
      }))

      # Custom SMS sender: Cognito calls this function to deliver SMS itself
      # (instead of SNS). The event carries the message encrypted with
      # `kms_key_id`, so that key is required.
      custom_sms_sender = optional(object({
        # The Lambda function Cognito invokes to deliver the message. Accepts a
        # direct ARN or a reference to an AwsLambda resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        lambda_arn = string

        # The sender event version. "V1_0" is the only version AWS currently
        # defines; modeled as a field so new versions are additive, not breaking.
        lambda_version = string
      }))

      # The KMS key Cognito uses to encrypt the code/message payload delivered to
      # custom sender functions. Required when either custom sender is set.
      # Accepts a direct key ARN or a reference to an AwsKmsKey resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_id = optional(string, "")
    }))

    # Threat protection posture (AWS "advanced security features"). Beyond
    # audit-only logging this requires the PLUS feature tier.
    user_pool_add_ons = optional(object({
      # Threat protection mode for standard authentication flows:
      # - "OFF": no threat protection.
      # - "AUDIT": Cognito gathers risk signals and metrics without acting.
      # - "ENFORCED": Cognito acts on risk (blocks, requires MFA) per the pool's
      #   risk configuration.
      # AUDIT and ENFORCED require the PLUS feature tier.
      advanced_security_mode = string

      # Extends threat protection to CUSTOM authentication flows (Lambda-driven
      # challenges), which standard mode does not cover. Valid values: "AUDIT",
      # "ENFORCED". When omitted, custom auth flows are not covered.
      custom_auth_mode = optional(string, "")
    }))

    # Pool-wide risk configuration for threat protection: automated responses
    # to account-takeover and compromised-credentials risk, notification
    # templates, and IP allow/block exceptions. Applies to every app client
    # that does not carry its own client-scoped risk configuration (a client
    # sets that on the AwsCognitoUserPoolClient spec). Requires threat
    # protection to be active (`user_pool_add_ons.advanced_security_mode` of
    # "AUDIT" or "ENFORCED"); automated responses act only in ENFORCED mode.
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

    # Where Cognito delivers detailed event logs. Each entry routes one event
    # source (user notifications, or auth events from threat protection) to
    # exactly one destination: a CloudWatch log group, a Firehose stream, or an
    # S3 bucket. "userAuthEvents" logging requires threat protection (PLUS tier).
    log_configurations = optional(list(object({
      # The event source to deliver:
      # - "userNotification": message-delivery errors (email/SMS send failures).
      #   Supports the ERROR log level.
      # - "userAuthEvents": detailed auth events from threat protection. Requires
      #   the PLUS tier with advanced security enabled; supports the INFO level.
      event_source = string

      # The log level to deliver: "ERROR" (userNotification) or "INFO"
      # (userAuthEvents).
      log_level = string

      # CloudWatch Logs destination. Accepts a direct log group ARN or a
      # reference to an AwsCloudwatchLogGroup resource. Exactly one destination
      # must be set per entry.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      cloudwatch_log_group_arn = optional(string, "")

      # Kinesis Data Firehose destination. Accepts a direct delivery-stream ARN
      # or a reference to an AwsKinesisFirehose resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      firehose_stream_arn = optional(string, "")

      # S3 destination. Accepts a direct bucket ARN or a reference to an
      # AwsS3Bucket resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      s3_bucket_arn = optional(string, "")
    })), [])

    # Domain configuration for the hosted sign-in UI and OAuth2 endpoints. When
    # set, Cognito provides a hosted login page at:
    #   https://{domain}.auth.{region}.amazoncognito.com
    # (for Cognito prefix domains) or at the custom domain you specify.
    #
    # Required for OAuth flows that use the Authorization Code grant with a
    # hosted UI redirect.
    domain = optional(object({
      # Domain prefix (for Cognito-hosted domains) or fully-qualified domain name
      # (for custom domains). For Cognito-hosted: provide a unique prefix like
      # "myapp-auth" which creates "myapp-auth.auth.{region}.amazoncognito.com" --
      # AWS rejects prefixes containing the reserved words "aws", "amazon", or
      # "cognito". For custom domains: provide the FQDN like "auth.example.com".
      # ForceNew -- the domain cannot be changed after creation.
      domain = string

      # ACM certificate ARN for custom domains. Required when `domain` contains
      # a dot (custom domain). The certificate must be in us-east-1 regardless
      # of the user pool's region (Cognito uses CloudFront for custom domains).
      # Not applicable for Cognito-hosted prefix domains. Adding or removing the
      # certificate replaces the domain (switching custom <-> prefix); rotating
      # one certificate ARN to another updates in place.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      certificate_arn = optional(string, "")

      # The sign-in page generation served on this domain:
      # - 1: classic hosted UI.
      # - 2: managed login (the newer, branding-designer experience; requires a
      #      branding style to be assigned before sign-in pages render).
      # When omitted, AWS defaults new domains to managed login (2).
      managed_login_version = optional(number)
    }))

    # User groups in this pool. Groups organize users for authorization: a
    # user's group memberships appear in the "cognito:groups" claim of their
    # tokens, and a group's IAM role flows into credentials vended through
    # identity-pool federation. Groups are pool-scoped configuration with no
    # independent AWS lifecycle, so they live here rather than as a separate
    # component. Users are ASSIGNED to groups at runtime (admin API) -- group
    # membership is data-plane content, not infrastructure.
    user_groups = optional(list(object({
      # Group name. Unique within the pool. 1-128 characters. Renaming a group
      # replaces it (AWS has no rename API) -- memberships do not survive.
      name = string

      # Human-readable purpose of the group. Up to 2048 characters.
      description = optional(string, "")

      # Priority when a user belongs to multiple groups: the LOWEST value wins
      # for the "cognito:preferred_role" claim. Note an AWS/provider asymmetry:
      # AWS accepts precedence 0 (the strongest priority) and defaults to null,
      # but the Terraform provider's zero-value gating cannot send an explicit 0
      # -- through either engine -- so 0 here means "no precedence" (AWS null).
      # Use 1 as the strongest expressible priority.
      precedence = optional(number, 0)

      # The IAM role whose ARN lands in members' "cognito:roles" and
      # "cognito:preferred_role" claims (and is assumable through identity-pool
      # federation). Accepts a direct role ARN or a reference to an AwsIamRole
      # resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = optional(string, "")
    })), [])
  })
}
