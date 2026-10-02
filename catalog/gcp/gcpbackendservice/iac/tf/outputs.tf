# Exactly one of the two resources exists (see the count guards in main.tf),
# so each output selects whichever branch was created with one(concat(...)).

# The self-link — the value URL maps (and passthrough forwarding rules)
# reference; the composition handle that puts this backend service behind a
# load-balancer frontend. A regional service's link carries
# regions/{region} where the global one says global.
output "self_link" {
  description = "Self-link URI of the backend service"
  value       = one(concat(google_compute_backend_service.this[*].self_link, google_compute_region_backend_service.this[*].self_link))
}

# The name as it exists in GCP.
output "backend_service_name" {
  description = "Name of the backend service in GCP"
  value       = one(concat(google_compute_backend_service.this[*].name, google_compute_region_backend_service.this[*].name))
}

# Server-assigned numeric ID of the backend service.
output "generated_id" {
  description = "Server-assigned numeric ID of the backend service"
  value       = one(concat(google_compute_backend_service.this[*].generated_id, google_compute_region_backend_service.this[*].generated_id))
}

# Server-computed fingerprint for optimistic concurrency control.
output "fingerprint" {
  description = "Fingerprint of the backend service"
  value       = one(concat(google_compute_backend_service.this[*].fingerprint, google_compute_region_backend_service.this[*].fingerprint))
}

# Region of a regional backend service; empty for a global one, so
# downstream composition can confirm scope compatibility.
output "region" {
  description = "Region of the backend service (empty for global)"
  value       = local.is_regional ? var.spec.region : ""
}
