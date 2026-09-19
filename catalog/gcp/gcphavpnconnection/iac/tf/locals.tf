locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project (null lets the google provider resolve its own).
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The gateway trio arrives as resolved references to the GcpHaVpnGateway
  # this connection rides: the gateway's self link, the router's name, and
  # their shared region.
  gateway = var.spec.gateway
  router  = var.spec.router
  region  = var.spec.region

  # The peer is exactly one of an external device or a Google Cloud gateway
  # (spec CEL). An external peer means this connection creates the external
  # VPN gateway resource and every tunnel names one of its interfaces.
  is_external_peer = var.spec.peer.external_gateway != null

  # The external gateway's name defaults to metadata.name when the spec
  # leaves it empty.
  external_gateway_name = (
    local.is_external_peer && var.spec.peer.external_gateway.name != ""
    ? var.spec.peer.external_gateway.name
    : var.metadata.name
  )

  # Empty defers to the provider default (DELETE); applied to every companion.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Tunnels keyed by name so a tunnel added or removed in the middle of the
  # list never renumbers (and so recreates) its neighbours. Each entry
  # carries the derived session and MD5 key names both engines share: the
  # session (router interface and BGP peer) is bgp_session.name or the
  # tunnel's name; the MD5 key is its declared name or `<tunnel>-md5`.
  tunnels = {
    for t in var.spec.tunnels : t.name => merge(t, {
      session_name = t.bgp_session.name != "" ? t.bgp_session.name : t.name
      md5_key_name = (
        t.bgp_session.md5_authentication_key != null
        ? (t.bgp_session.md5_authentication_key.name != "" ? t.bgp_session.md5_authentication_key.name : "${t.name}-md5")
        : ""
      )
      # Always sent: the provider defaults, made explicit so the spec is the
      # single source of truth (ike_version 2; enable true).
      ike_version = coalesce(t.ike_version, 2)
      enable      = coalesce(t.bgp_session.enable, true)
    })
  }

  # The same planton-ai_* label set the Pulumi module applies, merged into
  # every labeled companion (tunnels, external gateway) after the user's own
  # labels so the platform labels win on key conflicts.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcphavpnconnection"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  platform_labels = merge(local.base_labels, local.org_label, local.env_label, local.id_label)

  # Outputs in spec order (the lists are index-aligned with spec.tunnels).
  tunnel_order = [for t in var.spec.tunnels : t.name]
}
