locals {
  resource_id = var.metadata.id

  base_tags = {
    resource      = "true"
    resource_name = var.metadata.name
    resource_kind = "azureprivateendpoint"
  }

  org_tag = var.metadata.org != "" ? { organization = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { environment = var.metadata.env } : {}
  id_tag  = local.resource_id != "" ? { resource_id = local.resource_id } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.metadata.tags)

  # Connection name derived from the resource metadata name
  connection_name = "${var.metadata.name}-connection"

  # DNS zone group name derived from the resource metadata name
  dns_zone_group_name = "${var.metadata.name}-dns-zone-group"
}
