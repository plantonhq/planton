locals {
  resource_id = var.metadata.id

  base_tags = {
    resource      = "true"
    resource_name = var.metadata.name
    resource_kind = "azureprivatednszone"
  }

  org_tag = var.metadata.org != "" ? { organization = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { environment = var.metadata.env } : {}
  id_tag  = local.resource_id != "" ? { resource_id = local.resource_id } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.metadata.tags)

  # VNet link name derived from the resource metadata name
  vnet_link_name = "${var.metadata.name}-vnet-link"
}
