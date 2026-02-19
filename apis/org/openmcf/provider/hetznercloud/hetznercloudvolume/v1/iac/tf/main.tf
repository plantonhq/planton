resource "hcloud_volume" "this" {
  name              = local.volume_name
  size              = var.spec.size
  location          = var.spec.location
  format            = var.spec.format != null && var.spec.format != "format_unspecified" ? var.spec.format : null
  labels            = local.standard_labels
  delete_protection = var.spec.delete_protection != null ? var.spec.delete_protection : false
}

resource "hcloud_volume_attachment" "this" {
  count = var.spec.server_id != null ? 1 : 0

  volume_id = hcloud_volume.this.id
  server_id = tonumber(var.spec.server_id)
  automount = var.spec.automount != null ? var.spec.automount : false
}
