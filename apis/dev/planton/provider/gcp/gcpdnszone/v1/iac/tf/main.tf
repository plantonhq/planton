# Create the Google Cloud DNS Managed Zone
# This is a public DNS zone that will be authoritative for the specified domain
resource "google_dns_managed_zone" "managed_zone" {
  name        = local.managed_zone_name
  project     = var.spec.project_id.value
  description = "managed-zone for ${var.metadata.name}"
  dns_name    = local.zone_dns_name
  visibility  = "public"
}

# Grant DNS management permissions to specified service accounts
# This is typically used to grant permissions to Kubernetes workload identities
# such as cert-manager (for DNS-01 ACME challenges) and external-dns (for Ingress automation)
# 
# Note: This currently uses project-level IAM binding (roles/dns.admin) which grants
# access to all zones in the project. This is a temporary workaround until per-zone
# IAM bindings become available in the Google provider.
# See: https://cloud.google.com/dns/docs/zones/iam-per-resource-zones
resource "google_project_iam_binding" "dns_admin" {
  count   = length(local.iam_binding_members) > 0 ? 1 : 0
  project = var.spec.project_id.value
  role    = "roles/dns.admin"
  members = local.iam_binding_members

  depends_on = [
    google_dns_managed_zone.managed_zone
  ]
}

# Create DNS records within the managed zone
# These are static, foundational records managed by Infrastructure-as-Code
# Dynamic application records (for Kubernetes Ingresses) should be managed by external-dns
# Temporary ACME challenge records should be managed by cert-manager
# This follows the "split state" pattern where different lifecycle owners manage different records
resource "google_dns_record_set" "records" {
  for_each = { for idx, rec in var.spec.records : idx => rec }

  managed_zone = google_dns_managed_zone.managed_zone.name
  project      = var.spec.project_id.value
  name         = each.value.name
  type         = each.value.record_type
  ttl          = each.value.ttl_seconds
  rrdatas      = each.value.values

  depends_on = [
    google_dns_managed_zone.managed_zone
  ]
}
