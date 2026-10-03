locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurevirtualnetworkgateway"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over them:
  # user tags deliberately win so an org's governance conventions (cost
  # center, owner) can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # Enum wire maps. tfvars carries FULL proto enum value names; the maps
  # translate them to azurerm's exact (case-sensitive) vocabulary. Absent
  # type/vpn_type fall back to the site-to-site defaults, sent explicitly
  # so both engines produce identical payloads.
  type_wire = {
    "VPN"           = "Vpn"
    "EXPRESS_ROUTE" = "ExpressRoute"
  }
  gateway_type = lookup(local.type_wire, coalesce(var.spec.type, "VPN"), "Vpn")

  vpn_type_wire = {
    "ROUTE_BASED"  = "RouteBased"
    "POLICY_BASED" = "PolicyBased"
  }
  vpn_type = lookup(local.vpn_type_wire, coalesce(var.spec.vpn_type, "ROUTE_BASED"), "RouteBased")

  # The SKU is spec-required (never empty for a valid manifest). The
  # non-AZ VpnGw1-5 rows are gone with their retired spec values: ARM
  # rejects new non-AZ VPN gateway creates
  # (NonAzSkusNotAllowedForVPNGateway, live-confirmed).
  sku_wire = {
    "BASIC"             = "Basic"
    "STANDARD"          = "Standard"
    "HIGH_PERFORMANCE"  = "HighPerformance"
    "ULTRA_PERFORMANCE" = "UltraPerformance"
    "VPN_GW_1_AZ"       = "VpnGw1AZ"
    "VPN_GW_2_AZ"       = "VpnGw2AZ"
    "VPN_GW_3_AZ"       = "VpnGw3AZ"
    "VPN_GW_4_AZ"       = "VpnGw4AZ"
    "VPN_GW_5_AZ"       = "VpnGw5AZ"
    "ER_GW_1_AZ"        = "ErGw1AZ"
    "ER_GW_2_AZ"        = "ErGw2AZ"
    "ER_GW_3_AZ"        = "ErGw3AZ"
    "ER_GW_SCALE"       = "ErGwScale"
  }
  sku = lookup(local.sku_wire, var.spec.sku, null)

  # Sent only when specified -- the provider treats generation as
  # Computed, so omission lets Azure pick the SKU's default.
  generation_wire = {
    "GENERATION1" = "Generation1"
    "GENERATION2" = "Generation2"
    "NONE"        = "None"
  }
  generation = (
    var.spec.generation != null && var.spec.generation != ""
  ) ? lookup(local.generation_wire, var.spec.generation, null) : null

  allocation_wire = {
    "DYNAMIC" = "Dynamic"
    "STATIC"  = "Static"
  }

  nat_mode_wire = {
    "EGRESS_SNAT"  = "EgressSnat"
    "INGRESS_SNAT" = "IngressSnat"
  }

  nat_type_wire = {
    "STATIC_NAT"  = "Static"
    "DYNAMIC_NAT" = "Dynamic"
  }
}
