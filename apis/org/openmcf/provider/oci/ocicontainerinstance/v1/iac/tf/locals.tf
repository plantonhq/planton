locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciContainerInstance"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  restart_policy_map = {
    "always"     = "ALWAYS"
    "never"      = "NEVER"
    "on_failure" = "ON_FAILURE"
  }

  health_check_type_map = {
    "http" = "HTTP"
    "tcp"  = "TCP"
  }

  failure_action_map = {
    "kill" = "KILL"
    "none" = "NONE"
  }

  volume_type_map = {
    "emptydir"   = "EMPTYDIR"
    "configfile" = "CONFIGFILE"
  }

  secret_type_map = {
    "basic" = "BASIC"
    "vault" = "VAULT"
  }
}
