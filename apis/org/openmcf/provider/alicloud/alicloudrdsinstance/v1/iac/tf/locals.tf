locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  instance_name = (
    var.spec.instance_name != ""
    ? var.spec.instance_name
    : var.metadata.name
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "alicloud_rds_instance"
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
    for db in var.spec.databases : db.name => db
  }

  default_charset = {
    "MySQL"      = "utf8mb4"
    "MariaDB"    = "utf8mb4"
    "PostgreSQL" = "UTF8"
    "PPAS"       = "UTF8"
    "SQLServer"  = "Chinese_PRC_CI_AS"
  }

  accounts_map = {
    for acct in var.spec.accounts : acct.account_name => acct
  }

  privileges_flat = flatten([
    for acct in var.spec.accounts : [
      for i, priv in acct.privileges : {
        key            = "${acct.account_name}-priv-${i}"
        account_name   = acct.account_name
        database_names = priv.database_names
        privilege      = priv.privilege
      }
    ]
  ])

  privileges_map = {
    for p in local.privileges_flat : p.key => p
  }
}
