variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareZeroTrustAccessPolicySpec defines a reusable, account-scoped Access policy"
  # NOTE: StringValueOrRef fields are flattened to plain strings and enums to their
  # string names by the proto->tfvars converter; each access rule is an object with
  # exactly one variant key set (the proto oneof), the rest null.
  type = object({
    account_id = string
    name       = string
    decision   = string

    include = list(object({
      email                     = optional(object({ email = string }))
      email_domain              = optional(object({ domain = string }))
      email_list                = optional(object({ id = string }))
      everyone                  = optional(object({}))
      ip                        = optional(object({ ip = string }))
      ip_list                   = optional(object({ id = string }))
      certificate               = optional(object({}))
      group                     = optional(object({ id = string }))
      azure_ad                  = optional(object({ id = string, identity_provider_id = string }))
      github_organization       = optional(object({ identity_provider_id = string, name = string, team = optional(string) }))
      gsuite                    = optional(object({ email = string, identity_provider_id = string }))
      okta                      = optional(object({ name = string, identity_provider_id = string }))
      saml                      = optional(object({ attribute_name = string, attribute_value = string, identity_provider_id = string }))
      oidc                      = optional(object({ claim_name = string, claim_value = string, identity_provider_id = string }))
      auth_context              = optional(object({ id = string, ac_id = string, identity_provider_id = string }))
      auth_method               = optional(object({ auth_method = string }))
      common_name               = optional(object({ common_name = string }))
      geo                       = optional(object({ country_code = string }))
      device_posture            = optional(object({ integration_uid = string }))
      external_evaluation       = optional(object({ evaluate_url = string, keys_url = string }))
      login_method              = optional(object({ id = string }))
      service_token             = optional(object({ token_id = string }))
      any_valid_service_token   = optional(object({}))
      linked_app_token          = optional(object({ app_uid = string }))
      user_risk_score           = optional(object({ user_risk_score = list(string) }))
      cloudflare_account_member = optional(object({ account_id = optional(string) }))
    }))

    exclude = optional(list(object({
      email                     = optional(object({ email = string }))
      email_domain              = optional(object({ domain = string }))
      email_list                = optional(object({ id = string }))
      everyone                  = optional(object({}))
      ip                        = optional(object({ ip = string }))
      ip_list                   = optional(object({ id = string }))
      certificate               = optional(object({}))
      group                     = optional(object({ id = string }))
      azure_ad                  = optional(object({ id = string, identity_provider_id = string }))
      github_organization       = optional(object({ identity_provider_id = string, name = string, team = optional(string) }))
      gsuite                    = optional(object({ email = string, identity_provider_id = string }))
      okta                      = optional(object({ name = string, identity_provider_id = string }))
      saml                      = optional(object({ attribute_name = string, attribute_value = string, identity_provider_id = string }))
      oidc                      = optional(object({ claim_name = string, claim_value = string, identity_provider_id = string }))
      auth_context              = optional(object({ id = string, ac_id = string, identity_provider_id = string }))
      auth_method               = optional(object({ auth_method = string }))
      common_name               = optional(object({ common_name = string }))
      geo                       = optional(object({ country_code = string }))
      device_posture            = optional(object({ integration_uid = string }))
      external_evaluation       = optional(object({ evaluate_url = string, keys_url = string }))
      login_method              = optional(object({ id = string }))
      service_token             = optional(object({ token_id = string }))
      any_valid_service_token   = optional(object({}))
      linked_app_token          = optional(object({ app_uid = string }))
      user_risk_score           = optional(object({ user_risk_score = list(string) }))
      cloudflare_account_member = optional(object({ account_id = optional(string) }))
    })), [])

    require = optional(list(object({
      email                     = optional(object({ email = string }))
      email_domain              = optional(object({ domain = string }))
      email_list                = optional(object({ id = string }))
      everyone                  = optional(object({}))
      ip                        = optional(object({ ip = string }))
      ip_list                   = optional(object({ id = string }))
      certificate               = optional(object({}))
      group                     = optional(object({ id = string }))
      azure_ad                  = optional(object({ id = string, identity_provider_id = string }))
      github_organization       = optional(object({ identity_provider_id = string, name = string, team = optional(string) }))
      gsuite                    = optional(object({ email = string, identity_provider_id = string }))
      okta                      = optional(object({ name = string, identity_provider_id = string }))
      saml                      = optional(object({ attribute_name = string, attribute_value = string, identity_provider_id = string }))
      oidc                      = optional(object({ claim_name = string, claim_value = string, identity_provider_id = string }))
      auth_context              = optional(object({ id = string, ac_id = string, identity_provider_id = string }))
      auth_method               = optional(object({ auth_method = string }))
      common_name               = optional(object({ common_name = string }))
      geo                       = optional(object({ country_code = string }))
      device_posture            = optional(object({ integration_uid = string }))
      external_evaluation       = optional(object({ evaluate_url = string, keys_url = string }))
      login_method              = optional(object({ id = string }))
      service_token             = optional(object({ token_id = string }))
      any_valid_service_token   = optional(object({}))
      linked_app_token          = optional(object({ app_uid = string }))
      user_risk_score           = optional(object({ user_risk_score = list(string) }))
      cloudflare_account_member = optional(object({ account_id = optional(string) }))
    })), [])

    session_duration               = optional(string, "")
    approval_required              = optional(bool, false)
    isolation_required             = optional(bool, false)
    purpose_justification_required = optional(bool, false)
    purpose_justification_prompt   = optional(string, "")

    approval_groups = optional(list(object({
      approvals_needed = number
      email_addresses  = optional(list(string), [])
      email_list_uuid  = optional(string, "")
    })), [])

    connection_rules = optional(object({
      rdp = optional(object({
        allowed_clipboard_local_to_remote_formats = optional(list(string), [])
        allowed_clipboard_remote_to_local_formats = optional(list(string), [])
      }))
    }))

    mfa_config = optional(object({
      allowed_authenticators = optional(list(string), [])
      mfa_disabled           = optional(bool, false)
      session_duration       = optional(string, "")
    }))
  })
}
