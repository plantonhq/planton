# -----------------------------------------------------------------------------
# Cognito User Pool
# -----------------------------------------------------------------------------

resource "aws_cognito_user_pool" "this" {
  name = local.name
  tags = local.tags

  # Identity model
  username_attributes = try(local.spec.username_attributes, null)
  alias_attributes    = try(local.spec.alias_attributes, null)

  username_configuration {
    case_sensitive = try(local.spec.username_case_sensitive, false)
  }

  # Password policy
  dynamic "password_policy" {
    for_each = try(local.spec.password_policy, null) != null ? [local.spec.password_policy] : []
    content {
      minimum_length                   = try(password_policy.value.minimum_length, null)
      require_lowercase                = try(password_policy.value.require_lowercase, null)
      require_uppercase                = try(password_policy.value.require_uppercase, null)
      require_numbers                  = try(password_policy.value.require_numbers, null)
      require_symbols                  = try(password_policy.value.require_symbols, null)
      temporary_password_validity_days = try(password_policy.value.temporary_password_validity_days, null)
    }
  }

  # MFA
  mfa_configuration = try(local.spec.mfa_configuration, "OFF")

  dynamic "software_token_mfa_configuration" {
    for_each = try(local.spec.software_token_mfa_enabled, false) ? [true] : []
    content {
      enabled = true
    }
  }

  # Auto-verified attributes
  auto_verified_attributes = try(local.spec.auto_verified_attributes, null)

  # Account recovery
  dynamic "account_recovery_setting" {
    for_each = length(try(local.spec.account_recovery_mechanisms, [])) > 0 ? [true] : []
    content {
      dynamic "recovery_mechanism" {
        for_each = local.spec.account_recovery_mechanisms
        content {
          name     = recovery_mechanism.value.name
          priority = recovery_mechanism.value.priority
        }
      }
    }
  }

  # Email configuration
  dynamic "email_configuration" {
    for_each = try(local.spec.email_configuration, null) != null ? [local.spec.email_configuration] : []
    content {
      email_sending_account  = try(email_configuration.value.email_sending_account, null)
      source_arn             = try(email_configuration.value.source_arn, null)
      from_email_address     = try(email_configuration.value.from_email_address, null)
      reply_to_email_address = try(email_configuration.value.reply_to_email_address, null)
      configuration_set      = try(email_configuration.value.configuration_set, null)
    }
  }

  # Admin create user config
  dynamic "admin_create_user_config" {
    for_each = try(local.spec.allow_admin_create_user_only, false) ? [true] : []
    content {
      allow_admin_create_user_only = true
    }
  }

  # Deletion protection
  deletion_protection = try(local.spec.deletion_protection, false) ? "ACTIVE" : "INACTIVE"

  # Custom attributes (schema)
  dynamic "schema" {
    for_each = try(local.spec.custom_attributes, [])
    content {
      name                = schema.value.name
      attribute_data_type = schema.value.attribute_data_type
      mutable             = try(schema.value.mutable, null)

      dynamic "string_attribute_constraints" {
        for_each = schema.value.attribute_data_type == "String" ? [true] : []
        content {
          min_length = try(schema.value.string_min_length, null)
          max_length = try(schema.value.string_max_length, null)
        }
      }

      dynamic "number_attribute_constraints" {
        for_each = schema.value.attribute_data_type == "Number" ? [true] : []
        content {
          min_value = try(schema.value.number_min_value, null)
          max_value = try(schema.value.number_max_value, null)
        }
      }
    }
  }

  # Lambda triggers
  dynamic "lambda_config" {
    for_each = try(local.spec.lambda_config, null) != null ? [local.spec.lambda_config] : []
    content {
      pre_sign_up                    = try(lambda_config.value.pre_sign_up, null)
      pre_authentication             = try(lambda_config.value.pre_authentication, null)
      post_authentication            = try(lambda_config.value.post_authentication, null)
      post_confirmation              = try(lambda_config.value.post_confirmation, null)
      pre_token_generation           = try(lambda_config.value.pre_token_generation, null)
      custom_message                 = try(lambda_config.value.custom_message, null)
      user_migration                 = try(lambda_config.value.user_migration, null)
      define_auth_challenge          = try(lambda_config.value.define_auth_challenge, null)
      create_auth_challenge          = try(lambda_config.value.create_auth_challenge, null)
      verify_auth_challenge_response = try(lambda_config.value.verify_auth_challenge_response, null)
    }
  }
}

# -----------------------------------------------------------------------------
# App Clients
# -----------------------------------------------------------------------------

resource "aws_cognito_user_pool_client" "this" {
  for_each = local.client_map

  name         = each.value.name
  user_pool_id = aws_cognito_user_pool.this.id

  generate_secret = try(each.value.generate_secret, null)

  # OAuth
  allowed_oauth_flows_user_pool_client = try(each.value.allowed_oauth_flows_user_pool_client, null)
  allowed_oauth_flows                  = try(each.value.allowed_oauth_flows, null)
  allowed_oauth_scopes                 = try(each.value.allowed_oauth_scopes, null)
  callback_urls                        = try(each.value.callback_urls, null)
  logout_urls                          = try(each.value.logout_urls, null)
  default_redirect_uri                 = try(each.value.default_redirect_uri, null)
  supported_identity_providers         = try(each.value.supported_identity_providers, null)

  # Auth flows
  explicit_auth_flows = try(each.value.explicit_auth_flows, null)

  # Token validity
  access_token_validity  = try(each.value.access_token_validity_minutes, null)
  id_token_validity      = try(each.value.id_token_validity_minutes, null)
  refresh_token_validity = try(each.value.refresh_token_validity_days, null)

  dynamic "token_validity_units" {
    for_each = (
      try(each.value.access_token_validity_minutes, 0) > 0 ||
      try(each.value.id_token_validity_minutes, 0) > 0 ||
      try(each.value.refresh_token_validity_days, 0) > 0
    ) ? [true] : []
    content {
      access_token  = "minutes"
      id_token      = "minutes"
      refresh_token = "days"
    }
  }

  # Security
  enable_token_revocation      = try(each.value.enable_token_revocation, null)
  prevent_user_existence_errors = try(each.value.prevent_user_existence_errors, null)
}

# -----------------------------------------------------------------------------
# Domain (optional)
# -----------------------------------------------------------------------------

resource "aws_cognito_user_pool_domain" "this" {
  count = local.has_domain ? 1 : 0

  domain          = local.spec.domain.domain
  user_pool_id    = aws_cognito_user_pool.this.id
  certificate_arn = local.is_custom_domain ? try(local.spec.domain.certificate_arn, null) : null
}
