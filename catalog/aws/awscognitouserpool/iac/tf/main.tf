# Cognito User Pool + its folded pool-scoped satellites (hosted-UI domain,
# log delivery).
#
# App clients, identity providers, and resource servers are separate kinds
# that compose onto the pool by reference -- this module deliberately creates
# nothing an application would hold its own identity to.

resource "aws_cognito_user_pool" "this" {
  name = local.resource_name

  # ---------------------------------------------------------------------------
  # Identity model (all ForceNew -- changing any of these replaces the pool
  # and destroys every user in it)
  # ---------------------------------------------------------------------------

  # Empty lists must reach AWS as absence, not as [] -- the two identity
  # models (username vs alias) are mutually exclusive and AWS rejects an
  # explicit empty set.
  username_attributes = length(var.spec.username_attributes) > 0 ? var.spec.username_attributes : null
  alias_attributes    = length(var.spec.alias_attributes) > 0 ? var.spec.alias_attributes : null

  username_configuration {
    case_sensitive = var.spec.username_case_sensitive
  }

  # ---------------------------------------------------------------------------
  # Pool-level posture
  # ---------------------------------------------------------------------------

  # AWS expresses deletion protection as an ACTIVE/INACTIVE string; the spec
  # keeps it an honest boolean and the module translates.
  deletion_protection = var.spec.deletion_protection ? "ACTIVE" : "INACTIVE"

  # Omitted tier means AWS's default (ESSENTIALS); only an explicit choice is
  # forwarded so manifests that predate tiers keep deploying unchanged.
  user_pool_tier = var.spec.user_pool_tier != "" ? var.spec.user_pool_tier : null

  # ---------------------------------------------------------------------------
  # Password and sign-in policy
  # ---------------------------------------------------------------------------

  # The zero-gates below are faithful, not lossy: AWS itself treats a
  # submitted 0 as null for temporary_password_validity_days (applying its
  # 7-day default), and 0 is AWS's own default posture for
  # password_history_size (history off) -- so 0 and absent are the same
  # policy at the API for all three numerics.
  dynamic "password_policy" {
    for_each = var.spec.password_policy != null ? [var.spec.password_policy] : []
    content {
      minimum_length                   = password_policy.value.minimum_length > 0 ? password_policy.value.minimum_length : null
      require_lowercase                = password_policy.value.require_lowercase
      require_uppercase                = password_policy.value.require_uppercase
      require_numbers                  = password_policy.value.require_numbers
      require_symbols                  = password_policy.value.require_symbols
      password_history_size            = password_policy.value.password_history_size > 0 ? password_policy.value.password_history_size : null
      temporary_password_validity_days = password_policy.value.temporary_password_validity_days > 0 ? password_policy.value.temporary_password_validity_days : null
    }
  }

  # The passwordless dial: listing first factors switches the pool to
  # choice-based sign-in for clients that enable ALLOW_USER_AUTH.
  dynamic "sign_in_policy" {
    for_each = length(var.spec.allowed_first_auth_factors) > 0 ? [1] : []
    content {
      allowed_first_auth_factors = var.spec.allowed_first_auth_factors
    }
  }

  # ---------------------------------------------------------------------------
  # MFA
  # ---------------------------------------------------------------------------

  mfa_configuration = var.spec.mfa_configuration != "" ? var.spec.mfa_configuration : null

  dynamic "software_token_mfa_configuration" {
    for_each = var.spec.software_token_mfa_enabled ? [1] : []
    content {
      enabled = true
    }
  }

  dynamic "email_mfa_configuration" {
    for_each = var.spec.email_mfa != null && coalesce(var.spec.email_mfa.enabled, true) ? [var.spec.email_mfa] : []
    content {
      message = email_mfa_configuration.value.message != "" ? email_mfa_configuration.value.message : null
      subject = email_mfa_configuration.value.subject != "" ? email_mfa_configuration.value.subject : null
    }
  }

  dynamic "web_authn_configuration" {
    for_each = var.spec.web_authn != null ? [var.spec.web_authn] : []
    content {
      relying_party_id  = web_authn_configuration.value.relying_party_id != "" ? web_authn_configuration.value.relying_party_id : null
      user_verification = web_authn_configuration.value.user_verification != "" ? web_authn_configuration.value.user_verification : null
    }
  }

  # ---------------------------------------------------------------------------
  # SMS delivery (Cognito assumes the referenced role to publish through SNS)
  # ---------------------------------------------------------------------------

  dynamic "sms_configuration" {
    for_each = var.spec.sms_configuration != null ? [var.spec.sms_configuration] : []
    content {
      external_id    = sms_configuration.value.external_id
      sns_caller_arn = sms_configuration.value.sns_caller_arn
      sns_region     = sms_configuration.value.sns_region != "" ? sms_configuration.value.sns_region : null
    }
  }

  sms_authentication_message = var.spec.sms_authentication_message != "" ? var.spec.sms_authentication_message : null

  # ---------------------------------------------------------------------------
  # Verification and recovery
  # ---------------------------------------------------------------------------

  auto_verified_attributes = length(var.spec.auto_verified_attributes) > 0 ? var.spec.auto_verified_attributes : null

  # Keep the previous value active until the new one verifies -- without this,
  # an unverified typo in an email update can lock a user out.
  dynamic "user_attribute_update_settings" {
    for_each = length(var.spec.attributes_require_verification_before_update) > 0 ? [1] : []
    content {
      attributes_require_verification_before_update = var.spec.attributes_require_verification_before_update
    }
  }

  dynamic "account_recovery_setting" {
    for_each = length(var.spec.account_recovery_mechanisms) > 0 ? [1] : []
    content {
      dynamic "recovery_mechanism" {
        for_each = var.spec.account_recovery_mechanisms
        content {
          name     = recovery_mechanism.value.name
          priority = recovery_mechanism.value.priority
        }
      }
    }
  }

  # ---------------------------------------------------------------------------
  # Email configuration
  # ---------------------------------------------------------------------------

  dynamic "email_configuration" {
    for_each = var.spec.email_configuration != null ? [var.spec.email_configuration] : []
    content {
      email_sending_account  = email_configuration.value.email_sending_account != "" ? email_configuration.value.email_sending_account : null
      source_arn             = email_configuration.value.source_arn != "" ? email_configuration.value.source_arn : null
      from_email_address     = email_configuration.value.from_email_address != "" ? email_configuration.value.from_email_address : null
      reply_to_email_address = email_configuration.value.reply_to_email_address != "" ? email_configuration.value.reply_to_email_address : null
      configuration_set      = email_configuration.value.configuration_set != "" ? email_configuration.value.configuration_set : null
    }
  }

  # The modern verification_message_template block is the single spelling this
  # module forwards -- the provider's legacy top-level message/subject fields
  # conflict with it and are deliberately not modeled.
  dynamic "verification_message_template" {
    for_each = var.spec.verification_message_template != null ? [var.spec.verification_message_template] : []
    content {
      default_email_option  = verification_message_template.value.default_email_option != "" ? verification_message_template.value.default_email_option : null
      email_message         = verification_message_template.value.email_message != "" ? verification_message_template.value.email_message : null
      email_subject         = verification_message_template.value.email_subject != "" ? verification_message_template.value.email_subject : null
      email_message_by_link = verification_message_template.value.email_message_by_link != "" ? verification_message_template.value.email_message_by_link : null
      email_subject_by_link = verification_message_template.value.email_subject_by_link != "" ? verification_message_template.value.email_subject_by_link : null
      sms_message           = verification_message_template.value.sms_message != "" ? verification_message_template.value.sms_message : null
    }
  }

  # ---------------------------------------------------------------------------
  # Admin create user (self-registration gate + invitation templates)
  # ---------------------------------------------------------------------------

  dynamic "admin_create_user_config" {
    for_each = (var.spec.allow_admin_create_user_only || var.spec.invite_message_template != null) ? [1] : []
    content {
      allow_admin_create_user_only = var.spec.allow_admin_create_user_only

      dynamic "invite_message_template" {
        for_each = var.spec.invite_message_template != null ? [var.spec.invite_message_template] : []
        content {
          email_message = invite_message_template.value.email_message != "" ? invite_message_template.value.email_message : null
          email_subject = invite_message_template.value.email_subject != "" ? invite_message_template.value.email_subject : null
          sms_message   = invite_message_template.value.sms_message != "" ? invite_message_template.value.sms_message : null
        }
      }
    }
  }

  # ---------------------------------------------------------------------------
  # Remembered devices
  # ---------------------------------------------------------------------------

  dynamic "device_configuration" {
    for_each = var.spec.device_configuration != null ? [var.spec.device_configuration] : []
    content {
      challenge_required_on_new_device      = coalesce(device_configuration.value.challenge_required_on_new_device, false)
      device_only_remembered_on_user_prompt = coalesce(device_configuration.value.device_only_remembered_on_user_prompt, false)
    }
  }

  # ---------------------------------------------------------------------------
  # Custom attributes. The pool schema is APPEND-ONLY in AWS: removing or
  # modifying an entry here errors instead of updating -- only additions
  # apply in place.
  # ---------------------------------------------------------------------------

  dynamic "schema" {
    for_each = var.spec.custom_attributes
    content {
      name                     = schema.value.name
      attribute_data_type      = schema.value.attribute_data_type
      mutable                  = schema.value.mutable
      required                 = schema.value.required
      developer_only_attribute = schema.value.developer_only_attribute

      dynamic "string_attribute_constraints" {
        for_each = schema.value.attribute_data_type == "String" && (schema.value.string_min_length != "" || schema.value.string_max_length != "") ? [1] : []
        content {
          min_length = schema.value.string_min_length != "" ? schema.value.string_min_length : null
          max_length = schema.value.string_max_length != "" ? schema.value.string_max_length : null
        }
      }

      dynamic "number_attribute_constraints" {
        for_each = schema.value.attribute_data_type == "Number" && (schema.value.number_min_value != "" || schema.value.number_max_value != "") ? [1] : []
        content {
          min_value = schema.value.number_min_value != "" ? schema.value.number_min_value : null
          max_value = schema.value.number_max_value != "" ? schema.value.number_max_value : null
        }
      }
    }
  }

  # ---------------------------------------------------------------------------
  # Lambda triggers. Referenced function ARNs arrive pre-resolved as plain
  # strings. The functions must grant Cognito invoke permission
  # (principal "cognito-idp.amazonaws.com") -- roles/permissions are owned by
  # the referenced resources, never mutated from here.
  # ---------------------------------------------------------------------------

  dynamic "lambda_config" {
    for_each = var.spec.lambda_config != null ? [var.spec.lambda_config] : []
    content {
      pre_sign_up         = lambda_config.value.pre_sign_up != "" ? lambda_config.value.pre_sign_up : null
      pre_authentication  = lambda_config.value.pre_authentication != "" ? lambda_config.value.pre_authentication : null
      post_authentication = lambda_config.value.post_authentication != "" ? lambda_config.value.post_authentication : null
      post_confirmation   = lambda_config.value.post_confirmation != "" ? lambda_config.value.post_confirmation : null
      # The plain field pins the V1_0 event; the config block below selects
      # the version explicitly. The spec's CEL keeps them mutually exclusive.
      pre_token_generation           = lambda_config.value.pre_token_generation != "" ? lambda_config.value.pre_token_generation : null
      custom_message                 = lambda_config.value.custom_message != "" ? lambda_config.value.custom_message : null
      user_migration                 = lambda_config.value.user_migration != "" ? lambda_config.value.user_migration : null
      define_auth_challenge          = lambda_config.value.define_auth_challenge != "" ? lambda_config.value.define_auth_challenge : null
      create_auth_challenge          = lambda_config.value.create_auth_challenge != "" ? lambda_config.value.create_auth_challenge : null
      verify_auth_challenge_response = lambda_config.value.verify_auth_challenge_response != "" ? lambda_config.value.verify_auth_challenge_response : null
      # Cognito encrypts the code/message payload delivered to custom sender
      # functions with this key -- the spec couples it to the senders.
      kms_key_id = lambda_config.value.kms_key_id != "" ? lambda_config.value.kms_key_id : null

      dynamic "pre_token_generation_config" {
        for_each = lambda_config.value.pre_token_generation_config != null ? [lambda_config.value.pre_token_generation_config] : []
        content {
          lambda_arn     = pre_token_generation_config.value.lambda_arn
          lambda_version = pre_token_generation_config.value.lambda_version
        }
      }

      dynamic "custom_email_sender" {
        for_each = lambda_config.value.custom_email_sender != null ? [lambda_config.value.custom_email_sender] : []
        content {
          lambda_arn     = custom_email_sender.value.lambda_arn
          lambda_version = custom_email_sender.value.lambda_version
        }
      }

      dynamic "custom_sms_sender" {
        for_each = lambda_config.value.custom_sms_sender != null ? [lambda_config.value.custom_sms_sender] : []
        content {
          lambda_arn     = custom_sms_sender.value.lambda_arn
          lambda_version = custom_sms_sender.value.lambda_version
        }
      }
    }
  }

  # ---------------------------------------------------------------------------
  # Threat protection
  # ---------------------------------------------------------------------------

  dynamic "user_pool_add_ons" {
    for_each = var.spec.user_pool_add_ons != null ? [var.spec.user_pool_add_ons] : []
    content {
      advanced_security_mode = user_pool_add_ons.value.advanced_security_mode

      dynamic "advanced_security_additional_flows" {
        for_each = user_pool_add_ons.value.custom_auth_mode != "" ? [1] : []
        content {
          custom_auth_mode = user_pool_add_ons.value.custom_auth_mode
        }
      }
    }
  }

  tags = local.aws_tags
}

# ---------------------------------------------------------------------------
# Hosted-UI domain (one per pool -- honestly folded into the pool resource).
#
# Two shapes, distinguished by the presence of a dot in the domain:
# - Prefix domain ("myapp-auth"): served at
#   {prefix}.auth.{region}.amazoncognito.com; ready in about a minute.
# - Custom domain ("auth.example.com"): AWS fronts the UI with a managed
#   CloudFront distribution (creation can take tens of minutes); the deployer
#   points DNS at the exported distribution domain.
# ---------------------------------------------------------------------------

resource "aws_cognito_user_pool_domain" "this" {
  count = local.has_domain ? 1 : 0

  domain       = var.spec.domain.domain
  user_pool_id = aws_cognito_user_pool.this.id

  # A certificate is what makes the domain "custom" to AWS. The spec's CEL
  # already requires it for dotted domains.
  certificate_arn = local.is_custom_domain && var.spec.domain.certificate_arn != "" ? var.spec.domain.certificate_arn : null

  # Only an explicit choice is forwarded -- omitted means AWS's default
  # (managed login for new domains).
  managed_login_version = var.spec.domain.managed_login_version
}

# ---------------------------------------------------------------------------
# Log delivery. AWS models this as ONE pool-scoped configuration carrying
# every route, so all spec entries materialize into a single resource (two
# resources would fight over the same setting on every apply).
# ---------------------------------------------------------------------------

resource "aws_cognito_log_delivery_configuration" "this" {
  count = length(var.spec.log_configurations) > 0 ? 1 : 0

  user_pool_id = aws_cognito_user_pool.this.id

  dynamic "log_configurations" {
    for_each = var.spec.log_configurations
    content {
      event_source = log_configurations.value.event_source
      log_level    = log_configurations.value.log_level

      # The spec's CEL guarantees exactly one destination per entry.
      dynamic "cloud_watch_logs_configuration" {
        for_each = log_configurations.value.cloudwatch_log_group_arn != "" ? [1] : []
        content {
          log_group_arn = log_configurations.value.cloudwatch_log_group_arn
        }
      }

      dynamic "firehose_configuration" {
        for_each = log_configurations.value.firehose_stream_arn != "" ? [1] : []
        content {
          stream_arn = log_configurations.value.firehose_stream_arn
        }
      }

      dynamic "s3_configuration" {
        for_each = log_configurations.value.s3_bucket_arn != "" ? [1] : []
        content {
          bucket_arn = log_configurations.value.s3_bucket_arn
        }
      }
    }
  }
}

# ---------------------------------------------------------------------------
# User groups. Pool-scoped configuration with no independent AWS lifecycle;
# membership (which users are in a group) is data-plane content managed at
# runtime, never from here.
# ---------------------------------------------------------------------------

resource "aws_cognito_user_group" "this" {
  for_each = { for g in var.spec.user_groups : g.name => g }

  user_pool_id = aws_cognito_user_pool.this.id
  name         = each.value.name
  description  = each.value.description != "" ? each.value.description : null

  # AWS accepts precedence 0 (the strongest priority) but the provider's own
  # zero-value gating cannot send it, so 0 carries "no precedence" here --
  # the spec documents 1 as the strongest expressible value.
  precedence = each.value.precedence > 0 ? each.value.precedence : null

  role_arn = each.value.role_arn != "" ? each.value.role_arn : null
}

# ---------------------------------------------------------------------------
# Pool-wide risk configuration (threat protection's automated responses).
# AWS applies this to every app client without a client-scoped configuration
# of its own; a client sets that on its AwsCognitoUserPoolClient spec, never
# here -- the two scopes are separate AWS configurations that do not fight.
# ---------------------------------------------------------------------------

resource "aws_cognito_risk_configuration" "this" {
  count = var.spec.risk_configuration != null ? 1 : 0

  user_pool_id = aws_cognito_user_pool.this.id

  dynamic "account_takeover_risk_configuration" {
    for_each = var.spec.risk_configuration.account_takeover != null ? [var.spec.risk_configuration.account_takeover] : []
    content {
      # The provider requires the actions block; the spec's CEL requires at
      # least one action inside it.
      actions {
        dynamic "low_action" {
          for_each = account_takeover_risk_configuration.value.low_action != null ? [account_takeover_risk_configuration.value.low_action] : []
          content {
            event_action = low_action.value.event_action
            notify       = low_action.value.notify
          }
        }

        dynamic "medium_action" {
          for_each = account_takeover_risk_configuration.value.medium_action != null ? [account_takeover_risk_configuration.value.medium_action] : []
          content {
            event_action = medium_action.value.event_action
            notify       = medium_action.value.notify
          }
        }

        dynamic "high_action" {
          for_each = account_takeover_risk_configuration.value.high_action != null ? [account_takeover_risk_configuration.value.high_action] : []
          content {
            event_action = high_action.value.event_action
            notify       = high_action.value.notify
          }
        }
      }

      dynamic "notify_configuration" {
        for_each = account_takeover_risk_configuration.value.notify_configuration != null ? [account_takeover_risk_configuration.value.notify_configuration] : []
        content {
          source_arn = notify_configuration.value.source_arn
          from       = notify_configuration.value.from != "" ? notify_configuration.value.from : null
          reply_to   = notify_configuration.value.reply_to != "" ? notify_configuration.value.reply_to : null

          dynamic "block_email" {
            for_each = notify_configuration.value.block_email != null ? [notify_configuration.value.block_email] : []
            content {
              subject   = block_email.value.subject
              html_body = block_email.value.html_body
              text_body = block_email.value.text_body
            }
          }

          dynamic "mfa_email" {
            for_each = notify_configuration.value.mfa_email != null ? [notify_configuration.value.mfa_email] : []
            content {
              subject   = mfa_email.value.subject
              html_body = mfa_email.value.html_body
              text_body = mfa_email.value.text_body
            }
          }

          dynamic "no_action_email" {
            for_each = notify_configuration.value.no_action_email != null ? [notify_configuration.value.no_action_email] : []
            content {
              subject   = no_action_email.value.subject
              html_body = no_action_email.value.html_body
              text_body = no_action_email.value.text_body
            }
          }
        }
      }
    }
  }

  dynamic "compromised_credentials_risk_configuration" {
    for_each = var.spec.risk_configuration.compromised_credentials != null ? [var.spec.risk_configuration.compromised_credentials] : []
    content {
      actions {
        event_action = compromised_credentials_risk_configuration.value.event_action
      }

      # Empty means AWS's default (all supported events) -- send absence.
      event_filter = length(compromised_credentials_risk_configuration.value.event_filter) > 0 ? compromised_credentials_risk_configuration.value.event_filter : null
    }
  }

  dynamic "risk_exception_configuration" {
    for_each = var.spec.risk_configuration.risk_exception != null ? [var.spec.risk_configuration.risk_exception] : []
    content {
      blocked_ip_range_list = length(risk_exception_configuration.value.blocked_ip_ranges) > 0 ? risk_exception_configuration.value.blocked_ip_ranges : null
      skipped_ip_range_list = length(risk_exception_configuration.value.skipped_ip_ranges) > 0 ? risk_exception_configuration.value.skipped_ip_ranges : null
    }
  }
}
