# The account's default WARP device profile: the settings every enrolled
# device receives unless a custom profile matches it first.
#
# The profile always exists on the account -- create and update are the same
# PATCH against the singleton, and DESTROY IS A NO-OP that leaves the
# last-applied values standing. Unset spec fields are never sent, leaving the
# live value (or Cloudflare's default) untouched.
resource "cloudflare_zero_trust_device_default_profile" "main" {
  account_id = var.spec.account_id

  allow_mode_switch              = try(var.spec.allow_mode_switch, null)
  allow_updates                  = try(var.spec.allow_updates, null)
  allowed_to_leave               = try(var.spec.allowed_to_leave, null)
  auto_connect                   = try(var.spec.auto_connect, null)
  captive_portal                 = try(var.spec.captive_portal, null)
  disable_auto_fallback          = try(var.spec.disable_auto_fallback, null)
  exclude_office_ips             = try(var.spec.exclude_office_ips, null)
  register_interface_ip_with_dns = try(var.spec.register_interface_ip_with_dns, null)
  sccm_vpn_boundary_support      = try(var.spec.sccm_vpn_boundary_support, null)
  switch_locked                  = try(var.spec.switch_locked, null)
  support_url                    = try(var.spec.support_url, "") != "" ? var.spec.support_url : null
  tunnel_protocol                = try(var.spec.tunnel_protocol, "") != "" ? var.spec.tunnel_protocol : null
  lan_allow_minutes              = try(var.spec.lan_allow_minutes, null)
  lan_allow_subnet_size          = try(var.spec.lan_allow_subnet_size, null)

  # Split tunnel: exclude and include are mutually exclusive (spec validation
  # enforces it). Each entry carries exactly one of address or host; the
  # unset one is never sent.
  exclude = length(try(var.spec.exclude, [])) > 0 ? [
    for entry in var.spec.exclude : {
      address     = entry.address != "" ? entry.address : null
      host        = entry.host != "" ? entry.host : null
      description = try(entry.description, "") != "" ? entry.description : null
    }
  ] : null

  include = length(try(var.spec.include, [])) > 0 ? [
    for entry in var.spec.include : {
      address     = entry.address != "" ? entry.address : null
      host        = entry.host != "" ? entry.host : null
      description = try(entry.description, "") != "" ? entry.description : null
    }
  ] : null

  service_mode_v2 = try(var.spec.service_mode_v2, null) != null ? {
    mode = var.spec.service_mode_v2.mode
    port = try(var.spec.service_mode_v2.port, null)
  } : null

  # The provider's attribute for the default virtual network is named
  # "default"; the spec names it default_virtual_network_id for clarity.
  virtual_networks = try(var.spec.virtual_networks, null) != null ? {
    allowed = var.spec.virtual_networks.allowed
    default = var.spec.virtual_networks.default_virtual_network_id
  } : null

  # Always send the list, empty included. The API echoes [] for an empty
  # suffix list, and the provider's Computed+Optional attribute carries no
  # state-preserving plan modifier -- a null config re-plans an in-place
  # update forever (measured live 2026-08-27 at v5.23.0). Sending [] matches
  # the spec's documented contract (an empty list clears the account's
  # list; proto3 cannot distinguish unset from empty), and converges.
  dns_search_suffixes = [
    for row in try(var.spec.dns_search_suffixes, []) : {
      suffix      = row.suffix
      description = try(row.description, "") != "" ? row.description : null
    }
  ]

  # policy_id is a pure-computed server echo (the singleton's stable id)
  # with no UseStateForUnknown modifier at v5.23.0: every refresh-inclusive
  # plan re-marks it "(known after apply)" and proposes a no-op update
  # forever (measured live 2026-08-27; the value never actually changes).
  # Ignoring it is safe -- the attribute is never sent, and the
  # output still reads the real value from state after apply.
  lifecycle {
    ignore_changes = [policy_id]
  }
}

# The profile's local-DNS fallback list. The profile resource reports
# fallback_domains READ-ONLY; this dedicated companion is the only write
# path, and it replaces the whole list on every apply (declarative: what the
# spec lists is exactly what exists). Its destroy is also a no-op -- the last
# applied list stands. Deployed only when the spec declares rows, so an
# unmanaged list is never touched.
resource "cloudflare_zero_trust_device_default_profile_local_domain_fallback" "main" {
  count = length(try(var.spec.fallback_domains, [])) > 0 ? 1 : 0

  account_id = var.spec.account_id

  domains = [
    for row in var.spec.fallback_domains : {
      suffix      = row.suffix
      description = try(row.description, "") != "" ? row.description : null
      dns_server  = length(try(row.dns_server, [])) > 0 ? row.dns_server : null
    }
  ]

  depends_on = [cloudflare_zero_trust_device_default_profile.main]
}

# Per-zone WARP client certificate provisioning -- the one ZONE-scoped
# surface on this account-scoped kind. Cloudflare offers no delete for it
# (removing this resource leaves the toggle as last applied) and no import.
# Deployed only when the spec declares the fold.
resource "cloudflare_zero_trust_device_default_profile_certificates" "main" {
  count = try(var.spec.zone_certificates, null) != null ? 1 : 0

  zone_id = var.spec.zone_certificates.zone_id
  enabled = var.spec.zone_certificates.enabled
}
