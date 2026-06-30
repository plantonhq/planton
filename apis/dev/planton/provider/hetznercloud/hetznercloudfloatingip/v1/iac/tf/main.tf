resource "hcloud_floating_ip" "this" {
  name              = local.floating_ip_name
  type              = var.spec.type
  home_location     = var.spec.home_location
  description       = var.spec.description
  server_id         = var.spec.server_id != null ? tonumber(var.spec.server_id) : null
  labels            = local.standard_labels
  delete_protection = var.spec.delete_protection != null ? var.spec.delete_protection : false
}

resource "hcloud_rdns" "this" {
  count = var.spec.dns_ptr != null && var.spec.dns_ptr != "" ? 1 : 0

  floating_ip_id = hcloud_floating_ip.this.id
  ip_address     = hcloud_floating_ip.this.ip_address
  dns_ptr        = var.spec.dns_ptr
}
