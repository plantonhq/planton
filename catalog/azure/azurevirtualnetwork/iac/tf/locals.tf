locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurevirtualnetwork"
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

  # Map the spec enum's name string to ARM's enforcement value. null keeps
  # the encryption block absent entirely (ARM's default: encryption off),
  # so an unspecified spec and Azure's default deploy identically on both
  # engines.
  encryption_enforcement = (
    var.spec.encryption == "ALLOW_UNENCRYPTED" ? "AllowUnencrypted" :
    var.spec.encryption == "DROP_UNENCRYPTED" ? "DropUnencrypted" : null
  )

  # Map the spec enum's name string to ARM's policy value. null lets
  # azurerm apply ARM's default ("Disabled"); only the opt-in "Basic" mode
  # is ever sent.
  private_endpoint_vnet_policies = (
    var.spec.private_endpoint_vnet_policies == "BASIC" ? "Basic" : null
  )
}
