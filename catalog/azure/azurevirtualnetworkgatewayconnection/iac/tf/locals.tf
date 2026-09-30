locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurevirtualnetworkgatewayconnection"
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
  # translate them to azurerm's exact (case-sensitive) vocabulary. The
  # connection type is spec-required (never empty for a valid manifest).
  type_wire = {
    "IPSEC"         = "IPsec"
    "VNET_TO_VNET"  = "Vnet2Vnet"
    "EXPRESS_ROUTE" = "ExpressRoute"
  }
  connection_type = lookup(local.type_wire, var.spec.type, null)

  # Sent only when specified -- the provider treats the protocol as
  # Computed, so omission lets Azure apply its default (IKEv2).
  protocol_wire = {
    "IKE_V1" = "IKEv1"
    "IKE_V2" = "IKEv2"
  }
  connection_protocol = (
    var.spec.connection_protocol != null && var.spec.connection_protocol != ""
  ) ? lookup(local.protocol_wire, var.spec.connection_protocol, null) : null

  # Sent explicitly (Default when unspecified) -- deterministic payloads
  # on both engines.
  mode_wire = {
    "DEFAULT"        = "Default"
    "INITIATOR_ONLY" = "InitiatorOnly"
    "RESPONDER_ONLY" = "ResponderOnly"
  }
  connection_mode = lookup(local.mode_wire, coalesce(var.spec.connection_mode, "DEFAULT"), "Default")
}
