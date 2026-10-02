output "name" {
  description = "Full resource name of the target (projects/{project}/locations/{location}/targets/{target_id})"
  value       = google_clouddeploy_target.this.id
}

output "target_id" {
  description = "The target's ID, what delivery pipeline stages and multi-targets reference"
  value       = google_clouddeploy_target.this.target_id
}

output "uid" {
  description = "Google's unique identifier for the target"
  value       = google_clouddeploy_target.this.uid
}
