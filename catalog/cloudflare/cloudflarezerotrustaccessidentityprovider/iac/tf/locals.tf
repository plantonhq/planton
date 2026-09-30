locals {
  # Scope: exactly one of account_id or zone_id is set (enforced by the spec).
  account_id = try(var.spec.account_id, "")
  zone_id    = try(var.spec.zone_id, "")

  # The provider REQUIRES the config attribute even for types that need no
  # parameters (onetimepin) -- an all-null object (HCL's {}) is the correct
  # payload then. Unset fields are sent as null, never as "" (the provider's
  # per-type validators trigger on any non-null value). Every field is built
  # through try() so a wholly-absent spec config collapses each attribute to
  # null while keeping ONE object type (a `{} : {...}` conditional would make
  # Terraform reject the mismatched arms).
  config = {
    claims                      = try(length(var.spec.config.claims) > 0 ? var.spec.config.claims : null, null)
    client_id                   = try(var.spec.config.client_id != "" ? var.spec.config.client_id : null, null)
    client_secret               = try(var.spec.config.client_secret != "" ? var.spec.config.client_secret : null, null)
    email_claim_name            = try(var.spec.config.email_claim_name != "" ? var.spec.config.email_claim_name : null, null)
    pkce_enabled                = try(var.spec.config.pkce_enabled, null)
    conditional_access_enabled  = try(var.spec.config.conditional_access_enabled, null)
    directory_id                = try(var.spec.config.directory_id != "" ? var.spec.config.directory_id : null, null)
    prompt                      = try(var.spec.config.prompt != "" ? var.spec.config.prompt : null, null)
    support_groups              = try(var.spec.config.support_groups, null)
    centrify_account            = try(var.spec.config.centrify_account != "" ? var.spec.config.centrify_account : null, null)
    centrify_app_id             = try(var.spec.config.centrify_app_id != "" ? var.spec.config.centrify_app_id : null, null)
    apps_domain                 = try(var.spec.config.apps_domain != "" ? var.spec.config.apps_domain : null, null)
    auth_url                    = try(var.spec.config.auth_url != "" ? var.spec.config.auth_url : null, null)
    certs_url                   = try(var.spec.config.certs_url != "" ? var.spec.config.certs_url : null, null)
    scopes                      = try(length(var.spec.config.scopes) > 0 ? var.spec.config.scopes : null, null)
    token_url                   = try(var.spec.config.token_url != "" ? var.spec.config.token_url : null, null)
    authorization_server_id     = try(var.spec.config.authorization_server_id != "" ? var.spec.config.authorization_server_id : null, null)
    okta_account                = try(var.spec.config.okta_account != "" ? var.spec.config.okta_account : null, null)
    onelogin_account            = try(var.spec.config.onelogin_account != "" ? var.spec.config.onelogin_account : null, null)
    ping_env_id                 = try(var.spec.config.ping_env_id != "" ? var.spec.config.ping_env_id : null, null)
    attributes                  = try(length(var.spec.config.attributes) > 0 ? var.spec.config.attributes : null, null)
    email_attribute_name        = try(var.spec.config.email_attribute_name != "" ? var.spec.config.email_attribute_name : null, null)
    enable_encryption           = try(var.spec.config.enable_encryption, null)
    header_attributes           = try(length(var.spec.config.header_attributes) > 0 ? [
      for header_attribute in var.spec.config.header_attributes : {
        attribute_name = header_attribute.attribute_name != "" ? header_attribute.attribute_name : null
        header_name    = header_attribute.header_name != "" ? header_attribute.header_name : null
      }
    ] : null, null)
    idp_public_certs            = try(length(var.spec.config.idp_public_certs) > 0 ? var.spec.config.idp_public_certs : null, null)
    issuer_url                  = try(var.spec.config.issuer_url != "" ? var.spec.config.issuer_url : null, null)
    sign_request                = try(var.spec.config.sign_request, null)
    sso_target_url              = try(var.spec.config.sso_target_url != "" ? var.spec.config.sso_target_url : null, null)
    restrict_to_account_members = try(var.spec.config.restrict_to_account_members, null)
  }

  # SCIM: only sent when the spec configures it. identity_update_behavior left
  # empty means Cloudflare's default (no_action).
  scim_config = var.spec.scim_config == null ? null : {
    enabled                  = var.spec.scim_config.enabled
    identity_update_behavior = var.spec.scim_config.identity_update_behavior != "" ? var.spec.scim_config.identity_update_behavior : null
    seat_deprovision         = var.spec.scim_config.seat_deprovision
    user_deprovision         = var.spec.scim_config.user_deprovision
  }
}
