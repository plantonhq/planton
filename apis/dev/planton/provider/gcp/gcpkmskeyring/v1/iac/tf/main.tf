resource "google_kms_key_ring" "this" {
  project  = local.project_id
  name     = local.key_ring_name
  location = local.location
}
