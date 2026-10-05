locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurevpngateway"
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

  # The spec's enum NAMES mapped onto ARM's vocabulary. Unset optionals
  # apply ARM's defaults explicitly so the rendered plan shows the real
  # value -- mirroring the Pulumi module's nil handling. Note the wire
  # value's space: "Microsoft Network".
  routing_preference_wire = {
    "MICROSOFT_NETWORK" = "Microsoft Network"
    "INTERNET"          = "Internet"
  }
  routing_preference = (
    var.spec.routing_preference == null || var.spec.routing_preference == ""
    ? "Microsoft Network"
    : lookup(local.routing_preference_wire, var.spec.routing_preference, var.spec.routing_preference)
  )

  nat_rule_mode_wire = {
    "EGRESS_SNAT"  = "EgressSnat"
    "INGRESS_SNAT" = "IngressSnat"
  }

  nat_rule_type_wire = {
    "STATIC_NAT"  = "Static"
    "DYNAMIC_NAT" = "Dynamic"
  }

  # Unspecified (the enum's zero value renders as "") deliberately maps
  # to nothing -- the lookup's null default keeps the rule on both
  # instances.
  nat_rule_ip_configuration_wire = {
    "INSTANCE_0" = "Instance0"
    "INSTANCE_1" = "Instance1"
  }
}
