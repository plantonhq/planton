locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurednsrecord"
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

  # The platform materializes the proto default (300) before the module
  # runs; the coalesce is a same-value safety net for direct iac-input
  # paths, never a different fallback.
  ttl = coalesce(var.spec.ttl_seconds, 300)

  # The CAA tag arrives as the proto enum value name (ISSUE, ISSUEWILD,
  # IODEF, CONTACTEMAIL); Azure's wire vocabulary is the lowercase form
  # and the provider validates it case-sensitively.
  caa_tag_map = {
    "ISSUE"        = "issue"
    "ISSUEWILD"    = "issuewild"
    "IODEF"        = "iodef"
    "CONTACTEMAIL" = "contactemail"
  }
}
