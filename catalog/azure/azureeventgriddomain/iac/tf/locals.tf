locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureeventgriddomain"
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

  # The spec enum's value names mapped to the provider's identity tokens.
  identity_type_map = {
    "SYSTEM_ASSIGNED" = "SystemAssigned"
    "USER_ASSIGNED"   = "UserAssigned"
  }

  # An input mapping block is sent only when it carries at least one
  # non-empty field -- the provider treats an empty block and an absent
  # one identically, and the built-in schemas need no mapping at all.
  input_mapping_fields_set = (
    var.spec.input_mapping_fields != null && length(compact([
      var.spec.input_mapping_fields.id,
      var.spec.input_mapping_fields.topic,
      var.spec.input_mapping_fields.event_time,
      var.spec.input_mapping_fields.event_type,
      var.spec.input_mapping_fields.subject,
      var.spec.input_mapping_fields.data_version,
    ])) > 0
  )

  input_mapping_default_values_set = (
    var.spec.input_mapping_default_values != null && length(compact([
      var.spec.input_mapping_default_values.event_type,
      var.spec.input_mapping_default_values.subject,
      var.spec.input_mapping_default_values.data_version,
    ])) > 0
  )
}
