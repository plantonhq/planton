locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null lets the google provider resolve its own
  # project from configuration or the GOOGLE_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The gateway's name defaults to metadata.name when the spec leaves it
  # empty -- the same naming basis every kind uses.
  gateway_name = var.spec.gateway_name != "" ? var.spec.gateway_name : var.metadata.name

  # The router's name defaults to the gateway's: a router and a gateway are
  # different resource types, so the shared name never collides.
  router_name = var.spec.router.name != "" ? var.spec.router.name : local.gateway_name

  # Empty defers to the provider default (DELETE); applied to the gateway and
  # the router alike.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, so a resource
  # is attributable to its Planton object regardless of the engine that
  # created it. Conditional labels appear under the same conditions on both
  # sides.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.gateway_name
    "planton-ai_kind"     = "gcphavpngateway"
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

  # User labels first: the platform labels win on key conflicts.
  final_labels = merge(var.spec.labels, local.base_labels, local.org_label, local.env_label, local.id_label)

  # The two interface addresses are API-assigned for an internet-facing
  # gateway; the list is read back with the interface ids Google assigns
  # (0 and 1), so each address is picked by id, not position -- and is empty
  # for an Interconnect-backed interface, which has no public address.
  interface_ips = { for i in google_compute_ha_vpn_gateway.this.vpn_interfaces : tostring(i.id) => i.ip_address }
}
