output "policy_id" {
  description = "Fully qualified resource ID of the security policy (projects/{project}/global/securityPolicies/{name})"
  value       = google_compute_security_policy.this.id
}

output "policy_name" {
  description = "Name of the security policy as it exists in GCP"
  value       = google_compute_security_policy.this.name
}

output "policy_self_link" {
  description = "Self-link URI of the security policy (used when attaching to backend services)"
  value       = google_compute_security_policy.this.self_link
}

output "fingerprint" {
  description = "Server-computed fingerprint of the policy for concurrency control"
  value       = google_compute_security_policy.this.fingerprint
}
