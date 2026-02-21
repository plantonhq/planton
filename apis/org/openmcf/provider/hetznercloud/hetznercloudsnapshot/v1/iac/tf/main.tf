resource "hcloud_snapshot" "this" {
  server_id   = tonumber(var.spec.server_id)
  description = var.spec.description
  labels      = local.standard_labels
}
