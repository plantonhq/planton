variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsCognitoUserPool specification"
  type = object({
    region = string
    username_attributes = optional(list(string), [])
    alias_attributes = optional(list(string), [])
    username_case_sensitive = optional(bool, false)
    deletion_protection = optional(bool, false)
    user_pool_tier = optional(string, "")
    password_policy = optional(object({
      minimum_length = optional(number, 0)
      require_lowercase = optional(bool, false)
      require_uppercase = optional(bool, false)
      require_numbers = optional(bool, false)
      require_symbols = optional(bool, false)
      password_history_size = optional(number, 0)
      temporary_password_validity_days = optional(number, 0)
    }))
    allowed_first_auth_factors = optional(list(string), [])
    mfa_configuration = optional(string, "")
    software_token_mfa_enabled = optional(bool, false)
    email_mfa = optional(object({
      message = optional(string, "")
      subject = optional(string, "")
      enabled = optional(bool)
    }))
    web_authn = optional(object({
      relying_party_id = optional(string, "")
      user_verification = optional(string, "")
    }))
    sms_configuration = optional(object({
      sns_caller_arn = string
      external_id = string
      sns_region = optional(string, "")
    }))
    sms_authentication_message = optional(string, "")
    auto_verified_attributes = optional(list(string), [])
    attributes_require_verification_before_update = optional(list(string), [])
    account_recovery_mechanisms = optional(list(object({
      name = string
      priority = number
    })), [])
    email_configuration = optional(object({
      email_sending_account = optional(string, "")
      source_arn = optional(string, "")
      from_email_address = optional(string, "")
      reply_to_email_address = optional(string, "")
      configuration_set = optional(string, "")
    }))
    verification_message_template = optional(object({
      default_email_option = optional(string, "")
      email_message = optional(string, "")
      email_subject = optional(string, "")
      email_message_by_link = optional(string, "")
      email_subject_by_link = optional(string, "")
      sms_message = optional(string, "")
    }))
    allow_admin_create_user_only = optional(bool, false)
    invite_message_template = optional(object({
      email_message = optional(string, "")
      email_subject = optional(string, "")
      sms_message = optional(string, "")
    }))
    device_configuration = optional(object({
      challenge_required_on_new_device = optional(bool)
      device_only_remembered_on_user_prompt = optional(bool)
    }))
    custom_attributes = optional(list(object({
      name = string
      attribute_data_type = string
      mutable = optional(bool, false)
      required = optional(bool, false)
      developer_only_attribute = optional(bool, false)
      string_min_length = optional(string, "")
      string_max_length = optional(string, "")
      number_min_value = optional(string, "")
      number_max_value = optional(string, "")
    })), [])
    lambda_config = optional(object({
      pre_sign_up = optional(string, "")
      pre_authentication = optional(string, "")
      post_authentication = optional(string, "")
      post_confirmation = optional(string, "")
      pre_token_generation = optional(string, "")
      pre_token_generation_config = optional(object({
        lambda_arn = string
        lambda_version = string
      }))
      custom_message = optional(string, "")
      user_migration = optional(string, "")
      define_auth_challenge = optional(string, "")
      create_auth_challenge = optional(string, "")
      verify_auth_challenge_response = optional(string, "")
      custom_email_sender = optional(object({
        lambda_arn = string
        lambda_version = string
      }))
      custom_sms_sender = optional(object({
        lambda_arn = string
        lambda_version = string
      }))
      kms_key_id = optional(string, "")
    }))
    user_pool_add_ons = optional(object({
      advanced_security_mode = string
      custom_auth_mode = optional(string, "")
    }))
    risk_configuration = optional(object({
      account_takeover = optional(object({
        low_action = optional(object({
          event_action = string
          notify = optional(bool, false)
        }))
        medium_action = optional(object({
          event_action = string
          notify = optional(bool, false)
        }))
        high_action = optional(object({
          event_action = string
          notify = optional(bool, false)
        }))
        notify_configuration = optional(object({
          source_arn = string
          from = optional(string, "")
          reply_to = optional(string, "")
          block_email = optional(object({
            subject = string
            html_body = string
            text_body = string
          }))
          mfa_email = optional(object({
            subject = string
            html_body = string
            text_body = string
          }))
          no_action_email = optional(object({
            subject = string
            html_body = string
            text_body = string
          }))
        }))
      }))
      compromised_credentials = optional(object({
        event_action = string
        event_filter = optional(list(string), [])
      }))
      risk_exception = optional(object({
        blocked_ip_ranges = optional(list(string), [])
        skipped_ip_ranges = optional(list(string), [])
      }))
    }))
    log_configurations = optional(list(object({
      event_source = string
      log_level = string
      cloudwatch_log_group_arn = optional(string, "")
      firehose_stream_arn = optional(string, "")
      s3_bucket_arn = optional(string, "")
    })), [])
    domain = optional(object({
      domain = string
      certificate_arn = optional(string, "")
      managed_login_version = optional(number)
    }))
    user_groups = optional(list(object({
      name = string
      description = optional(string, "")
      precedence = optional(number, 0)
      role_arn = optional(string, "")
    })), [])
  })
}
