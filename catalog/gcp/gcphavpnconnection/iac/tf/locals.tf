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

  # Keys are never required. A tunnel uses its own shared_secret, else the
  # connection's, else ONE key minted for every tunnel that declares none
  # (Google's HA VPN shape: one key for all tunnels). The same rule holds for
  # the MD5 key of every session that declares an md5_authentication_key
  # block. A key is minted exactly when the connection declares none and at
  # least one tunnel (session) leaves its own empty. Identical predicates in
  # the Pulumi module's locals.go.
  generate_shared_secret = var.spec.shared_secret == "" && anytrue([
    for t in var.spec.tunnels : t.shared_secret == ""
  ])
  generate_md5_authentication_key = var.spec.md5_authentication_key == "" && anytrue([
    for t in var.spec.tunnels : try(t.bgp_session.md5_authentication_key.key == "", false)
  ])

  connection_shared_secret          = local.generate_shared_secret ? random_password.shared_secret[0].result : var.spec.shared_secret
  connection_md5_authentication_key = local.generate_md5_authentication_key ? random_password.md5_authentication_key[0].result : var.spec.md5_authentication_key

  # Tunnels keyed by name so a tunnel added or removed in the middle of the
  # list never renumbers (and so recreates) its neighbours. Each entry
  # carries the derivations both engines share: the session (router
  # interface and BGP peer) is bgp_session.name or the tunnel's name; the
  # MD5 key is its declared name or `<tunnel>-md5`; the effective keys
  # follow the own-else-connection-else-minted rule above.
  tunnels = {
    for t in var.spec.tunnels : t.name => merge(t, {
      session_name = t.bgp_session.name != "" ? t.bgp_session.name : t.name
      md5_key_name = (
        t.bgp_session.md5_authentication_key != null
        ? (t.bgp_session.md5_authentication_key.name != "" ? t.bgp_session.md5_authentication_key.name : "${t.name}-md5")
        : ""
      )
      shared_secret = t.shared_secret != "" ? t.shared_secret : local.connection_shared_secret
      md5_key = (
        t.bgp_session.md5_authentication_key != null
        ? (t.bgp_session.md5_authentication_key.key != "" ? t.bgp_session.md5_authentication_key.key : local.connection_md5_authentication_key)
        : null
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
