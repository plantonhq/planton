# The folder's numeric id -- what every child references (a nested folder's
# parent, a project's folder_id, a policy's or binding's folder scope).
output "folder_id" {
  description = "The folder's numeric ID"
  value       = google_folder.this.folder_id
}

# The folder's resource name, folders/{folder_id}.
output "name" {
  description = "The folder's resource name (folders/{folder_id})"
  value       = google_folder.this.name
}

# ACTIVE for a live folder, DELETE_REQUESTED during the soft-delete window.
output "lifecycle_state" {
  description = "The folder's lifecycle state"
  value       = google_folder.this.lifecycle_state
}

output "create_time" {
  description = "When the folder was created (RFC 3339)"
  value       = google_folder.this.create_time
}
