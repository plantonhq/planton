locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciFileSystem"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  access_map = {
    "read_write" = "READ_WRITE"
    "read_only"  = "READ_ONLY"
  }

  identity_squash_map = {
    "no_squash"   = "NONE"
    "root_squash" = "ROOT"
    "all_squash"  = "ALL"
  }
}
