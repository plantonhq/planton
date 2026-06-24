locals {
  # Resource naming
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-kv-namespace")

  # Labels
  labels = merge({
    "name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Namespace configuration
  namespace_name = var.spec.namespace_name
  account_id     = var.spec.account_id
}
