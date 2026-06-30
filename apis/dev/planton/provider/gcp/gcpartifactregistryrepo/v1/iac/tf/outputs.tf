#############################
# Repository Outputs
#############################

output "repo_name" {
  description = "The repository ID for the repository"
  value       = google_artifact_registry_repository.repo.repository_id
}

# Example: us-west2-docker.pkg.dev
output "repo_hostname" {
  description = "The Repository hostname (region-docker.pkg.dev)"
  value       = "${google_artifact_registry_repository.repo.location}-docker.pkg.dev"
}

# Example: us-west2-docker.pkg.dev/<project>/<repo-id>
output "repo_url" {
  description = "Full Repository URL"
  value       = "${google_artifact_registry_repository.repo.location}-docker.pkg.dev/${google_artifact_registry_repository.repo.project}/${google_artifact_registry_repository.repo.repository_id}"
}

#############################
# Service Account Outputs
#############################

output "reader_service_account_email" {
  description = "Email address of the reader service account"
  value       = google_service_account.reader.email
}

output "reader_service_account_key_base64" {
  description = "Base64-encoded private key of the reader service account"
  value       = google_service_account_key.reader_key.private_key
  sensitive   = true
}

output "writer_service_account_email" {
  description = "Email address of the writer service account"
  value       = google_service_account.writer.email
}

output "writer_service_account_key_base64" {
  description = "Base64-encoded private key of the writer service account"
  value       = google_service_account_key.writer_key.private_key
  sensitive   = true
}

