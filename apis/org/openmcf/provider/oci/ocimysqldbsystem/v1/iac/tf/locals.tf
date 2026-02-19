locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciMysqlDbSystem"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  nsg_ids = [for nsg in var.spec.nsg_ids : nsg.value]

  key_generation_type_map = {
    "system" = "SYSTEM"
    "byok"   = "BYOK"
  }

  certificate_generation_type_map = {
    "system_cert" = "SYSTEM"
    "byoc"        = "BYOC"
  }

  maintenance_schedule_type_map = {
    "early"   = "EARLY"
    "regular" = "REGULAR"
  }

  version_preference_map = {
    "oldest"        = "OLDEST"
    "second_newest" = "SECOND_NEWEST"
    "newest"        = "NEWEST"
  }

  version_track_preference_map = {
    "long_term_support" = "LONG_TERM_SUPPORT"
    "innovation"        = "INNOVATION"
    "follow"            = "FOLLOW"
  }

  database_console_status_map = {
    "enabled"  = "ENABLED"
    "disabled" = "DISABLED"
  }
}
