output "name" {
  description = "The policy's resource name (projects/{project}/policy)"
  value       = "${google_binary_authorization_policy.this.id}/policy"
}

output "project_id" {
  description = "The project the policy governs"
  value       = google_binary_authorization_policy.this.project
}
