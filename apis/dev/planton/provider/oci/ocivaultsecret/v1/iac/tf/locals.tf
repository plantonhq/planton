locals {
  resource_id  = coalesce(var.metadata.id, var.metadata.name)
  display_name = var.spec.secret_name

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciVaultSecret"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  generation_type_map = {
    "bytes"      = "BYTES"
    "passphrase" = "PASSPHRASE"
    "ssh_key"    = "SSH_KEY"
  }

  rule_type_map = {
    "secret_expiry_rule" = "SECRET_EXPIRY_RULE"
    "secret_reuse_rule"  = "SECRET_REUSE_RULE"
  }

  target_system_type_map = {
    "adb"      = "ADB"
    "function" = "FUNCTION"
  }
}
