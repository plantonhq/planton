output "name" {
  description = "Full resource name of the repository link (projects/{project}/locations/{location}/connections/{connection}/repositories/{repository_id})"
  value       = google_cloudbuildv2_repository.this.id
}

output "repository_id" {
  description = "The repository's ID in Cloud Build"
  value       = google_cloudbuildv2_repository.this.name
}

output "remote_uri" {
  description = "The repository's HTTPS clone URI on the code host"
  value       = google_cloudbuildv2_repository.this.remote_uri
}
