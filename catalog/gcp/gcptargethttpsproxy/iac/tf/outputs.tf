# Exactly one of the two resources exists (see the count guards in main.tf),
# so each output selects whichever branch was created with one(concat(...)).

# The self-link — the value a forwarding rule references as its target; the
# composition handle that puts a VIP in front of this proxy. A regional
# proxy's link carries regions/{region} where the global one says global.
output "self_link" {
  description = "Self-link URI of the target HTTPS proxy"
  value       = one(concat(google_compute_target_https_proxy.this[*].self_link, google_compute_region_target_https_proxy.this[*].self_link))
}

# The name as it exists in GCP.
output "proxy_name" {
  description = "Name of the target HTTPS proxy in GCP"
  value       = one(concat(google_compute_target_https_proxy.this[*].name, google_compute_region_target_https_proxy.this[*].name))
}

# Server-assigned numeric ID of the proxy.
output "proxy_id" {
  description = "Server-assigned numeric ID of the target HTTPS proxy"
  value       = one(concat(google_compute_target_https_proxy.this[*].proxy_id, google_compute_region_target_https_proxy.this[*].proxy_id))
}

# Server-computed fingerprint for optimistic concurrency control. The
# regional API collection carries no fingerprint, so a regional proxy
# reports an empty string here.
output "fingerprint" {
  description = "Fingerprint of the target HTTPS proxy (empty for a regional proxy)"
  value       = local.is_regional ? "" : one(google_compute_target_https_proxy.this[*].fingerprint)
}

# Region of a regional proxy; empty for a global one, so downstream
# composition can confirm scope compatibility.
output "region" {
  description = "Region of the target HTTPS proxy (empty for global)"
  value       = local.is_regional ? var.spec.region : ""
}
