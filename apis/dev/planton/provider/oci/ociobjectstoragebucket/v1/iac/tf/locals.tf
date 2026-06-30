locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciObjectStorageBucket"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  access_type_map = {
    "no_public_access"        = "NoPublicAccess"
    "object_read"             = "ObjectRead"
    "object_read_without_list" = "ObjectReadWithoutList"
  }

  storage_tier_map = {
    "standard" = "Standard"
    "archive"  = "Archive"
  }

  versioning_map = {
    "enabled"   = "Enabled"
    "disabled"  = "Disabled"
    "suspended" = "Suspended"
  }

  auto_tiering_map = {
    "auto_tiering_disabled" = "Disabled"
    "infrequent_access"     = "InfrequentAccess"
  }

  lifecycle_action_map = {
    "lifecycle_archive"            = "ARCHIVE"
    "lifecycle_infrequent_access"  = "INFREQUENT_ACCESS"
    "lifecycle_delete"             = "DELETE"
    "lifecycle_abort"              = "ABORT"
  }

  time_unit_map = {
    "days"  = "DAYS"
    "years" = "YEARS"
  }
}
