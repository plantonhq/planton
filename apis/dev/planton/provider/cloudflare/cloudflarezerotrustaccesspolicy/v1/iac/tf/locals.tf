locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-zero-trust-access-policy")

  # Access rules pass straight through to the provider: each element already
  # carries exactly one variant (the proto oneof) and the proto field names match
  # the provider's attribute names 1:1.
  include = var.spec.include
  exclude = try(var.spec.exclude, [])
  require = try(var.spec.require, [])

  approval_groups = [
    for g in try(var.spec.approval_groups, []) : {
      approvals_needed = g.approvals_needed
      email_addresses  = length(try(g.email_addresses, [])) > 0 ? g.email_addresses : null
      email_list_uuid  = try(g.email_list_uuid, "") != "" ? g.email_list_uuid : null
    }
  ]

  connection_rules = try(var.spec.connection_rules, null) == null ? null : {
    rdp = try(var.spec.connection_rules.rdp, null) == null ? null : {
      allowed_clipboard_local_to_remote_formats = length(try(var.spec.connection_rules.rdp.allowed_clipboard_local_to_remote_formats, [])) > 0 ? var.spec.connection_rules.rdp.allowed_clipboard_local_to_remote_formats : null
      allowed_clipboard_remote_to_local_formats = length(try(var.spec.connection_rules.rdp.allowed_clipboard_remote_to_local_formats, [])) > 0 ? var.spec.connection_rules.rdp.allowed_clipboard_remote_to_local_formats : null
    }
  }

  mfa_config = try(var.spec.mfa_config, null) == null ? null : {
    allowed_authenticators = length(try(var.spec.mfa_config.allowed_authenticators, [])) > 0 ? var.spec.mfa_config.allowed_authenticators : null
    mfa_disabled           = try(var.spec.mfa_config.mfa_disabled, false)
    session_duration       = try(var.spec.mfa_config.session_duration, "") != "" ? var.spec.mfa_config.session_duration : null
  }
}
