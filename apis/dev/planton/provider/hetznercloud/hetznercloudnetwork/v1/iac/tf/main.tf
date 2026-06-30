resource "hcloud_network" "this" {
  name                    = local.network_name
  ip_range                = var.spec.ip_range
  labels                  = local.standard_labels
  delete_protection       = var.spec.delete_protection != null ? var.spec.delete_protection : false
  expose_routes_to_vswitch = var.spec.expose_routes_to_vswitch != null ? var.spec.expose_routes_to_vswitch : false
}

resource "hcloud_network_subnet" "this" {
  for_each = { for s in var.spec.subnets : s.ip_range => s }

  network_id   = hcloud_network.this.id
  type         = each.value.type
  network_zone = each.value.network_zone
  ip_range     = each.value.ip_range
  vswitch_id   = each.value.vswitch_id
}

resource "hcloud_network_route" "this" {
  for_each = { for r in (var.spec.routes != null ? var.spec.routes : []) : r.destination => r }

  network_id  = hcloud_network.this.id
  destination = each.value.destination
  gateway     = each.value.gateway
}
