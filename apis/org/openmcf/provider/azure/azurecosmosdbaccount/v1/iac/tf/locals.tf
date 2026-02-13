locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base tags for Azure resources
  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_cosmosdb_account"
    "resource_name" = var.metadata.name
  }

  # Organization tag only if var.metadata.org is non-empty
  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment tag only if var.metadata.env is non-empty
  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, and environment tags
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Flatten SQL containers for for_each (db_name + container_name as unique key)
  sql_containers = flatten([
    for db in var.spec.sql_databases : [
      for c in coalesce(db.containers, []) : {
        db_name                 = db.name
        name                    = c.name
        partition_key_paths     = c.partition_key_paths
        partition_key_kind       = coalesce(c.partition_key_kind, "Hash")
        throughput              = c.throughput
        autoscale_max_throughput = c.autoscale_max_throughput
        default_ttl             = c.default_ttl
      }
    ]
  ])

  sql_containers_map = {
    for c in local.sql_containers : "${c.db_name}/${c.name}" => c
  }

  # Flatten MongoDB collections for for_each
  mongo_collections = flatten([
    for db in var.spec.mongo_databases : [
      for c in coalesce(db.collections, []) : {
        db_name                 = db.name
        name                    = c.name
        shard_key               = c.shard_key
        throughput              = c.throughput
        autoscale_max_throughput = c.autoscale_max_throughput
        default_ttl_seconds      = c.default_ttl_seconds
        indexes                 = coalesce(c.indexes, [])
      }
    ]
  ])

  mongo_collections_map = {
    for c in local.mongo_collections : "${c.db_name}/${c.name}" => c
  }

  # Effective capabilities: add EnableMongo when kind is MongoDB if not present
  effective_capabilities = (
    var.spec.kind == "MongoDB" && !contains(var.spec.capabilities, "EnableMongo")
    ? concat(var.spec.capabilities, ["EnableMongo"])
    : var.spec.capabilities
  )
}
