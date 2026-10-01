output "name" {
  description = "Full resource name of the scope (projects/{project}/locations/global/scopes/{scope_id})"
  value       = google_gke_hub_scope.this.name
}

output "scope_id" {
  description = "The scope's ID"
  value       = google_gke_hub_scope.this.scope_id
}

output "uid" {
  description = "Google's unique identifier for the scope"
  value       = google_gke_hub_scope.this.uid
}
