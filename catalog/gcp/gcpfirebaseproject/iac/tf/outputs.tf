# The GCP project Firebase is enabled on.
output "project_id" {
  description = "The GCP project Firebase is enabled on"
  value       = google_firebase_project.this.project
}

# The project's number -- the Firebase Cloud Messaging sender id a client
# registers with.
output "project_number" {
  description = "The project number (the FCM sender id)"
  value       = google_firebase_project.this.project_number
}

# The project's display name as Firebase shows it.
output "display_name" {
  description = "The project's Firebase display name"
  value       = google_firebase_project.this.display_name
}

# Admin SDK configuration values -- each present only when the project has
# the corresponding resource, so each degrades to "".
output "database_url" {
  description = "The default Firebase Realtime Database URL, when the project has one"
  value       = try(data.google_firebase_admin_sdk_config.this.database_url, "")
}

output "storage_bucket" {
  description = "The default Cloud Storage for Firebase bucket name, when the project has one"
  value       = try(data.google_firebase_admin_sdk_config.this.storage_bucket, "")
}

output "location_id" {
  description = "The project's default GCP resource location, once finalized"
  value       = try(data.google_firebase_admin_sdk_config.this.location_id, "")
}
