resource "hcloud_primary_ip" "this" {
  name              = local.primary_ip_name
  type              = var.spec.type
  location          = var.spec.location
  assignee_type     = "server"
  auto_delete       = false
  labels            = local.standard_labels
  delete_protection = var.spec.delete_protection != null ? var.spec.delete_protection : false
}

resource "hcloud_rdns" "this" {
  count = var.spec.dns_ptr != null && var.spec.dns_ptr != "" ? 1 : 0

  primary_ip_id = hcloud_primary_ip.this.id
  ip_address    = hcloud_primary_ip.this.ip_address
  dns_ptr       = var.spec.dns_ptr
}
