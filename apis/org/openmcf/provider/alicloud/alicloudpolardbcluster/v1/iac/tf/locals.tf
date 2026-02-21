locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  cluster_description = (
    var.spec.description != ""
    ? var.spec.description
    : var.metadata.name
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "alicloud_polardb_cluster"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.spec.tags)

  databases_map = {
    for db in var.spec.databases : db.db_name => db
  }

  default_charset = {
    "MySQL"      = "utf8"
    "PostgreSQL" = "UTF8"
    "Oracle"     = "UTF8"
  }

  accounts_map = {
    for acct in var.spec.accounts : acct.account_name => acct
  }

  privileges_flat = flatten([
    for acct in var.spec.accounts : [
      for i, priv in acct.privileges : {
        key               = "${acct.account_name}-priv-${i}"
        account_name      = acct.account_name
        db_names          = priv.db_names
        account_privilege = priv.account_privilege
      }
    ]
  ])

  privileges_map = {
    for p in local.privileges_flat : p.key => p
  }
}
