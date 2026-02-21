resource "hcloud_server" "this" {
  name                   = local.server_name
  server_type            = var.spec.server_type
  image                  = var.spec.image
  location               = var.spec.location
  labels                 = local.standard_labels
  ssh_keys               = var.spec.ssh_keys
  user_data              = var.spec.user_data
  placement_group_id     = var.spec.placement_group_id != null ? tonumber(var.spec.placement_group_id) : null
  firewall_ids           = var.spec.firewall_ids != null ? [for id in var.spec.firewall_ids : tonumber(id)] : null
  backups                = var.spec.backups != null ? var.spec.backups : false
  keep_disk              = var.spec.keep_disk != null ? var.spec.keep_disk : false
  delete_protection      = var.spec.delete_protection != null ? var.spec.delete_protection : false
  rebuild_protection     = var.spec.rebuild_protection != null ? var.spec.rebuild_protection : false
  shutdown_before_deletion = var.spec.shutdown_before_deletion != null ? var.spec.shutdown_before_deletion : false

  dynamic "public_net" {
    for_each = var.spec.public_net != null ? [var.spec.public_net] : []
    content {
      ipv4_enabled = public_net.value.ipv4_enabled != null ? public_net.value.ipv4_enabled : true
      ipv6_enabled = public_net.value.ipv6_enabled != null ? public_net.value.ipv6_enabled : true
      ipv4         = public_net.value.ipv4 != null ? tonumber(public_net.value.ipv4) : null
      ipv6         = public_net.value.ipv6 != null ? tonumber(public_net.value.ipv6) : null
    }
  }

  dynamic "network" {
    for_each = var.spec.networks != null ? { for n in var.spec.networks : n.network_id => n } : {}
    content {
      network_id = tonumber(network.value.network_id)
      ip         = network.value.ip
      alias_ips  = network.value.alias_ips != null ? network.value.alias_ips : []
    }
  }
}

resource "hcloud_rdns" "this" {
  count = var.spec.dns_ptr != null && var.spec.dns_ptr != "" ? 1 : 0

  server_id  = hcloud_server.this.id
  ip_address = hcloud_server.this.ipv4_address
  dns_ptr    = var.spec.dns_ptr
}
