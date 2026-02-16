locals {
  resource_name = coalesce(try(var.metadata.name, null), "awselasticip")

  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # BYOIP settings — null when not configured.
  public_ipv4_pool     = try(var.spec.public_ipv4_pool, null) != "" ? try(var.spec.public_ipv4_pool, null) : null
  address              = try(var.spec.address, null) != "" ? try(var.spec.address, null) : null
  network_border_group = try(var.spec.network_border_group, null) != "" ? try(var.spec.network_border_group, null) : null
}
