# Create the Google Cloud DNS Record Set
# This creates a single DNS record within an existing Managed Zone
resource "google_dns_record_set" "record" {
  project      = local.project_id
  managed_zone = local.managed_zone
  name         = local.name
  type         = local.record_type
  ttl          = local.ttl_seconds
  rrdatas      = local.values
}
