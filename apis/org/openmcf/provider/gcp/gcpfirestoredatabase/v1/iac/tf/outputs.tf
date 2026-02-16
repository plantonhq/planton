output "database_id" {
  description = "Fully qualified database ID (projects/{project}/databases/{name})"
  value       = "projects/${local.project_id}/databases/${google_firestore_database.this.name}"
}

output "database_name" {
  description = "Database name"
  value       = google_firestore_database.this.name
}

output "uid" {
  description = "Server-generated UUID4 for this database"
  value       = google_firestore_database.this.uid
}

output "create_time" {
  description = "Timestamp at which the database was created"
  value       = google_firestore_database.this.create_time
}

output "earliest_version_time" {
  description = "Earliest timestamp for point-in-time recovery reads"
  value       = google_firestore_database.this.earliest_version_time
}
