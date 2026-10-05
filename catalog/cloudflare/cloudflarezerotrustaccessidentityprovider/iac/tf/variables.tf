variable "metadata" {
  description = "Catalog object metadata"
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
  description = "CloudflareZeroTrustAccessIdentityProvider specification"
  type = object({
    account_id = optional(string, "")
    zone_id = optional(string, "")
    name = string
    type = string
    config = optional(object({
      claims = optional(list(string), [])
      client_id = optional(string, "")
      client_secret = optional(string, "")
      email_claim_name = optional(string, "")
      pkce_enabled = optional(bool)
      conditional_access_enabled = optional(bool)
      directory_id = optional(string, "")
      prompt = optional(string, "")
      support_groups = optional(bool)
      centrify_account = optional(string, "")
      centrify_app_id = optional(string, "")
      apps_domain = optional(string, "")
      auth_url = optional(string, "")
      certs_url = optional(string, "")
      scopes = optional(list(string), [])
      token_url = optional(string, "")
      authorization_server_id = optional(string, "")
      okta_account = optional(string, "")
      onelogin_account = optional(string, "")
      ping_env_id = optional(string, "")
      attributes = optional(list(string), [])
      email_attribute_name = optional(string, "")
      enable_encryption = optional(bool)
      header_attributes = optional(list(object({
        attribute_name = optional(string, "")
        header_name = optional(string, "")
      })), [])
      idp_public_certs = optional(list(string), [])
      issuer_url = optional(string, "")
      sign_request = optional(bool)
      sso_target_url = optional(string, "")
      restrict_to_account_members = optional(bool)
    }))
    saml_certificate_set_id = optional(string, "")
    scim_config = optional(object({
      enabled = optional(bool, false)
      identity_update_behavior = optional(string, "")
      seat_deprovision = optional(bool, false)
      user_deprovision = optional(bool, false)
    }))
    read_only = optional(bool, false)
  })
}