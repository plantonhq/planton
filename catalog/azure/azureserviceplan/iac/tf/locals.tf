locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureserviceplan"
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

  # Metadata-derived tags first, then the user's spec tags merged over
  # them: user tags deliberately win so an org's governance conventions
  # (cost center, owner) can override the derived values where they
  # collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # OS type: spec enum name -> Azure's wire value. Absent means Linux --
  # the catalog's app kinds are Linux-based.
  os_type_map = {
    "LINUX"             = "Linux"
    "WINDOWS"           = "Windows"
    "WINDOWS_CONTAINER" = "WindowsContainer"
  }
  os_type = local.os_type_map[coalesce(var.spec.os_type, "LINUX")]

  # The spec's sku enum arrives as the FULL proto value name
  # (PREMIUM_P1V3 style); Azure spells the same SKU in its own mixed case
  # (P1v3 style). Spelled out row by row so the plan renders the exact
  # wire value and a vocabulary drift fails loudly at plan time.
  sku_map = {
    "FREE_F1"              = "F1"
    "SHARED_D1"            = "D1"
    "SHARED"               = "SHARED"
    "BASIC_B1"             = "B1"
    "BASIC_B2"             = "B2"
    "BASIC_B3"             = "B3"
    "STANDARD_S1"          = "S1"
    "STANDARD_S2"          = "S2"
    "STANDARD_S3"          = "S3"
    "PREMIUM_P1V2"         = "P1v2"
    "PREMIUM_P2V2"         = "P2v2"
    "PREMIUM_P3V2"         = "P3v2"
    "PREMIUM_P0V3"         = "P0v3"
    "PREMIUM_P1V3"         = "P1v3"
    "PREMIUM_P2V3"         = "P2v3"
    "PREMIUM_P3V3"         = "P3v3"
    "PREMIUM_P1MV3"        = "P1mv3"
    "PREMIUM_P2MV3"        = "P2mv3"
    "PREMIUM_P3MV3"        = "P3mv3"
    "PREMIUM_P4MV3"        = "P4mv3"
    "PREMIUM_P5MV3"        = "P5mv3"
    "PREMIUM_P0V4"         = "P0v4"
    "PREMIUM_P1V4"         = "P1v4"
    "PREMIUM_P2V4"         = "P2v4"
    "PREMIUM_P3V4"         = "P3v4"
    "PREMIUM_P1MV4"        = "P1mv4"
    "PREMIUM_P2MV4"        = "P2mv4"
    "PREMIUM_P3MV4"        = "P3mv4"
    "PREMIUM_P4MV4"        = "P4mv4"
    "PREMIUM_P5MV4"        = "P5mv4"
    "CONSUMPTION_Y1"       = "Y1"
    "ELASTIC_PREMIUM_EP1"  = "EP1"
    "ELASTIC_PREMIUM_EP2"  = "EP2"
    "ELASTIC_PREMIUM_EP3"  = "EP3"
    "FLEX_CONSUMPTION_FC1" = "FC1"
    "ISOLATED_I1"          = "I1"
    "ISOLATED_I2"          = "I2"
    "ISOLATED_I3"          = "I3"
    "ISOLATED_I1V2"        = "I1v2"
    "ISOLATED_I2V2"        = "I2v2"
    "ISOLATED_I3V2"        = "I3v2"
    "ISOLATED_I4V2"        = "I4v2"
    "ISOLATED_I5V2"        = "I5v2"
    "ISOLATED_I6V2"        = "I6v2"
    "ISOLATED_I1MV2"       = "I1mv2"
    "ISOLATED_I2MV2"       = "I2mv2"
    "ISOLATED_I3MV2"       = "I3mv2"
    "ISOLATED_I4MV2"       = "I4mv2"
    "ISOLATED_I5MV2"       = "I5mv2"
    "WORKFLOW_WS1"         = "WS1"
    "WORKFLOW_WS2"         = "WS2"
    "WORKFLOW_WS3"         = "WS3"
  }
  sku_name = local.sku_map[var.spec.sku_name]
}
