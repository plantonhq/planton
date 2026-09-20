# Exactly one of the two resources exists (see the count guards in main.tf),
# so each output selects whichever branch was created with one(concat(...)).

# The self-link — the value target HTTP(S) proxies reference as their url_map;
# the composition handle that puts this routing brain behind a load-balancer
# frontend. A regional map's link carries regions/{region} where the global
# one says global.
output "self_link" {
  description = "Self-link URI of the URL map"
  value       = one(concat(google_compute_url_map.this[*].self_link, google_compute_region_url_map.this[*].self_link))
}

# The name as it exists in GCP.
output "url_map_name" {
  description = "Name of the URL map in GCP"
  value       = one(concat(google_compute_url_map.this[*].name, google_compute_region_url_map.this[*].name))
}

# Server-assigned numeric ID of the URL map.
output "map_id" {
  description = "Server-assigned numeric ID of the URL map"
  value       = one(concat(google_compute_url_map.this[*].map_id, google_compute_region_url_map.this[*].map_id))
}

# Server-computed fingerprint for optimistic concurrency control.
output "fingerprint" {
  description = "Fingerprint of the URL map"
  value       = one(concat(google_compute_url_map.this[*].fingerprint, google_compute_region_url_map.this[*].fingerprint))
}

# Region of a regional URL map; empty for a global one, so downstream
# composition can confirm scope compatibility.
output "region" {
  description = "Region of the URL map (empty for global)"
  value       = local.is_regional ? var.spec.region : ""
}
