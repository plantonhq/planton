variable "metadata" {
  description = "Catalog object metadata"
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
  description = "CloudflareZeroTrustOrganization specification"
  type = object({
    account_id                                  = optional(string, "")
    zone_id                                     = optional(string, "")
    auth_domain                                 = optional(string, "")
    name                                        = optional(string, "")
    session_duration                            = optional(string, "")
    warp_auth_session_duration                  = optional(string, "")
    user_seat_expiration_inactive_time          = optional(string, "")
    deny_unmatched_requests                     = optional(bool)
    deny_unmatched_requests_exempted_zone_names = optional(list(string), [])
    custom_pages = optional(object({
      forbidden       = optional(string, "")
      identity_denied = optional(string, "")
    }))
    login_design = optional(object({
      background_color = optional(string, "")
      text_color       = optional(string, "")
      logo_path        = optional(string, "")
      header_text      = optional(string, "")
      footer_text      = optional(string, "")
    }))
    mfa_config = optional(object({
      allowed_authenticators        = optional(list(string), [])
      session_duration              = optional(string, "")
      amr_matching_session_duration = optional(string, "")
      required_aaguids              = optional(string, "")
    }))
    mfa_ssh_piv_key_requirements = optional(object({
      pin_policy          = optional(string, "")
      touch_policy        = optional(string, "")
      ssh_key_type        = optional(list(string), [])
      ssh_key_size        = optional(list(number), [])
      require_fips_device = optional(bool)
    }))
    allow_authenticate_via_warp = optional(bool)
    auto_redirect_to_identity   = optional(bool)
    mfa_required_for_all_apps   = optional(bool)
    is_ui_read_only             = optional(bool)
    ui_read_only_toggle_reason  = optional(string, "")
    key_rotation_interval_days  = optional(number)
  })
}