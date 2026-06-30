locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciPostgresqlDbSystem"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  nsg_ids = [for nsg in var.spec.network_details.nsg_ids : nsg.value]

  password_type_map = {
    "plain_text"   = "PLAIN_TEXT"
    "vault_secret" = "VAULT_SECRET"
  }

  backup_kind_map = {
    "daily"   = "DAILY"
    "weekly"  = "WEEKLY"
    "monthly" = "MONTHLY"
    "none"    = "NONE"
  }
}
