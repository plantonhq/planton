# Exactly one of the two policy resources exists (see the count guards in
# main.tf), so each output selects whichever branch was created with
# one(concat(...)).

output "policy_id" {
  description = "Fully qualified resource ID of the security policy (projects/{project}/global/securityPolicies/{name}, or projects/{project}/regions/{region}/securityPolicies/{name} for a regional policy)"
  value       = one(concat(google_compute_security_policy.this[*].id, google_compute_region_security_policy.this[*].id))
}

output "policy_name" {
  description = "Name of the security policy as it exists in GCP"
  value       = one(concat(google_compute_security_policy.this[*].name, google_compute_region_security_policy.this[*].name))
}

# The composition handle a backend service attaches; a regional policy's
# link carries regions/{region} where the global one says global, and a
# backend service accepts only a policy of its own scope.
output "policy_self_link" {
  description = "Self-link URI of the security policy (used when attaching to backend services)"
  value       = one(concat(google_compute_security_policy.this[*].self_link, google_compute_region_security_policy.this[*].self_link))
}

output "fingerprint" {
  description = "Server-computed fingerprint of the policy for concurrency control"
  value       = one(concat(google_compute_security_policy.this[*].fingerprint, google_compute_region_security_policy.this[*].fingerprint))
}

# Region of a regional policy; empty for a global one, so downstream
# consumers can tell the scope from the outputs alone.
output "region" {
  description = "Region of a regional security policy; empty for a global policy"
  value       = local.is_regional ? var.spec.region : ""
}

output "network_edge_security_service_self_link" {
  description = "Self-link of the network edge security service enrolling the region in advanced network DDoS protection; empty when the spec declares none"
  value       = local.create_edge_service ? one(google_compute_network_edge_security_service.this[*].self_link) : ""
}
