resource "google_compute_global_address" "this" {
  project      = local.project_id
  name         = local.address_name
  address_type = local.address_type
  ip_version   = local.ip_version
  description  = local.description != "" ? local.description : null
  address      = local.address != "" ? local.address : null
  network      = local.network
  prefix_length = local.prefix_length
  purpose      = local.purpose
  labels       = var.labels
}
