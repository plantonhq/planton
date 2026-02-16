locals {
  project_id    = var.spec.project_id.value
  address_name  = var.spec.address_name
  address       = var.spec.address
  address_type  = var.spec.address_type
  description   = var.spec.description
  ip_version    = var.spec.ip_version
  network       = var.spec.network != null ? var.spec.network.value : null
  prefix_length = var.spec.prefix_length
  purpose       = var.spec.purpose != "" ? var.spec.purpose : null
}
